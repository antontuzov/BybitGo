package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/forbest/bybitgo/internal/bybit"
	"github.com/forbest/bybitgo/internal/config"
	"github.com/forbest/bybitgo/internal/market"
	"github.com/forbest/bybitgo/internal/notifications"
	"github.com/forbest/bybitgo/internal/portfolio"
	"github.com/forbest/bybitgo/internal/risk"
	"github.com/forbest/bybitgo/internal/strategy"
	"github.com/forbest/bybitgo/internal/web"
	"github.com/joho/godotenv"
)

// TradingBot represents the main trading bot
type TradingBot struct {
	Config           *config.Config
	BybitClient      *bybit.Client
	PortfolioManager *portfolio.PortfolioManager
	MarketAnalyzer   *market.MarketAnalyzer
	StrategyAI       *strategy.StrategyAI
	RiskManager      *risk.RiskManager
	Strategies       map[strategy.StrategyType]strategy.Strategy
	CircuitBreaker   *risk.CircuitBreaker
	Dashboard        *web.Dashboard
	Server           *http.Server
	Notifier         *notifications.Notifier
	// Add fields for manual override control
	IsRunning bool
	StopChan  chan struct{}
}

// NewTradingBot creates a new TradingBot
func NewTradingBot() (*TradingBot, error) {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Create Bybit client
	bybitClient := bybit.NewClient(cfg.BybitAPIKey, cfg.BybitAPISecret, cfg.Testnet)

	// Create market analyzer
	marketAnalyzer := market.NewMarketAnalyzer()

	// Create portfolio manager
	portfolioManager := portfolio.NewPortfolioManager(bybitClient, cfg)
	// Set the market analyzer reference
	portfolioManager.MarketAnalyzer = marketAnalyzer

	// Create strategy AI
	strategyAI := strategy.NewStrategyAI(marketAnalyzer)

	// Create risk manager
	riskManager := risk.NewRiskManager(cfg)

	// Create circuit breaker (10 seconds timeout, 5 failure threshold)
	circuitBreaker := risk.NewCircuitBreaker(10*time.Second, 5)

	// Create strategy implementations
	strategies := map[strategy.StrategyType]strategy.Strategy{
		strategy.MarketMaking:       strategy.NewMarketMakingStrategy(),
		strategy.Momentum:           strategy.NewMomentumStrategy(),
		strategy.MeanReversion:      strategy.NewMeanReversionStrategy(),
		strategy.VolatilityBreakout: strategy.NewVolatilityBreakoutStrategy(),
	}

	// Create dashboard
	dashboard := web.NewDashboard(portfolioManager, riskManager, marketAnalyzer)

	// Create notifier
	notifier := notifications.NewNotifier()

	return &TradingBot{
		Config:           cfg,
		BybitClient:      bybitClient,
		PortfolioManager: portfolioManager,
		MarketAnalyzer:   marketAnalyzer,
		StrategyAI:       strategyAI,
		RiskManager:      riskManager,
		CircuitBreaker:   circuitBreaker,
		Strategies:       strategies,
		Dashboard:        dashboard,
		Notifier:         notifier,
		IsRunning:        true, // Start running by default
		StopChan:         make(chan struct{}),
	}, nil
}

// Run starts the trading bot
func (bot *TradingBot) Run(ctx context.Context) error {
	log.Println("Starting trading bot...")

	// Start the web dashboard in a separate goroutine
	go func() {
		log.Println("Starting web dashboard on port 8080...")
		if err := bot.Dashboard.Start("8080"); err != nil && err != http.ErrServerClosed {
			log.Printf("Dashboard server error: %v", err)
		}
	}()

	// Start the override command handler in a separate goroutine
	go bot.handleOverrideCommands()

	// Initialize portfolio with top coins
	if err := bot.PortfolioManager.UpdateTopCoins(ctx); err != nil {
		return fmt.Errorf("failed to initialize portfolio: %w", err)
	}

	log.Printf("Initialized portfolio with symbols: %v", bot.PortfolioManager.Symbols)

	// Start the main trading loop
	return bot.tradingLoop(ctx)
}

// handleOverrideCommands handles manual override commands from the web dashboard
func (bot *TradingBot) handleOverrideCommands() {
	for command := range bot.Dashboard.GetOverrideChannel() {
		log.Printf("Received manual override command: %s", command.Command)

		switch command.Command {
		case "start":
			bot.IsRunning = true
			log.Println("Trading bot started manually")
		case "stop":
			bot.IsRunning = false
			log.Println("Trading bot stopped manually")
		case "rebalance":
			// Trigger immediate rebalancing
			log.Println("Manual rebalancing triggered")
			// In a real implementation, you would trigger rebalancing here
		case "emergency_stop":
			bot.IsRunning = false
			log.Println("Emergency stop triggered manually")
			// Send emergency stop notification
			bot.Notifier.SendEmergencyStopAlert("Manual emergency stop triggered")
		default:
			log.Printf("Unknown command: %s", command.Command)
		}
	}
}

// tradingLoop runs the main trading loop
func (bot *TradingBot) tradingLoop(ctx context.Context) error {
	ticker := time.NewTicker(bot.PortfolioManager.RebalanceInterval)
	defer ticker.Stop()

	// Run initial cycle
	if err := bot.runTradingCycle(ctx); err != nil {
		log.Printf("Error in initial trading cycle: %v", err)
	}

	// Listen for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-ctx.Done():
			log.Println("Context cancelled, shutting down...")
			return nil
		case <-sigChan:
			log.Println("Received interrupt signal, shutting down...")
			return nil
		case <-ticker.C:
			// Check if bot is running (manual override)
			if bot.IsRunning {
				log.Println("Running trading cycle...")
				if err := bot.runTradingCycle(ctx); err != nil {
					log.Printf("Error in trading cycle: %v", err)
				}
			} else {
				log.Println("Trading bot is stopped (manual override), skipping trading cycle...")
			}
		case <-bot.StopChan:
			log.Println("Received stop signal, shutting down...")
			return nil
		}
	}
}

// runTradingCycle executes one complete trading cycle
func (bot *TradingBot) runTradingCycle(ctx context.Context) error {
	log.Println("=== Starting Trading Cycle ===")

	// Check circuit breaker state
	if bot.CircuitBreaker.State() == "open" {
		log.Println("WARNING: Circuit breaker is open, skipping trading cycle")
		return nil
	}

	// 1. Update top coins
	log.Println("1. Updating top coins...")
	err := bot.CircuitBreaker.Call(func() error {
		return bot.PortfolioManager.UpdateTopCoins(ctx)
	})
	if err != nil {
		return fmt.Errorf("failed to update top coins: %w", err)
	}

	// 2. Analyze market conditions for each coin
	log.Println("2. Analyzing market conditions...")
	marketData := make(map[string]*bybit.MarketData)
	currentPrices := make(map[string]float64)
	enhancedMarketData := make(map[string]*market.EnhancedMarketData)
	combinedSignals := make(map[string]*market.CombinedSignal)
	volumeWeightedSignals := make(map[string]*market.VolumeWeightedSignal)

	for _, symbol := range bot.PortfolioManager.Symbols {
		var data *bybit.MarketData
		err := bot.CircuitBreaker.Call(func() error {
			var err error
			data, err = bot.BybitClient.GetMarketData(ctx, symbol)
			return err
		})

		if err != nil {
			log.Printf("Warning: Failed to get market data for %s: %v", symbol, err)
			continue
		}

		marketData[symbol] = data

		// Extract current price from market data (use the latest close price)
		if len(data.Kline) > 0 {
			currentPrices[symbol], _ = data.Kline[len(data.Kline)-1].Close.Float64()
		}

		// Analyze enhanced market conditions with additional indicators
		enhancedData, err := bot.MarketAnalyzer.AnalyzeEnhancedMarketConditions(ctx, symbol, data)
		if err != nil {
			log.Printf("Warning: Failed to analyze enhanced market conditions for %s: %v", symbol, err)
		} else {
			enhancedMarketData[symbol] = enhancedData
			// Log some of the enhanced indicators
			if enhancedData.MACD != nil {
				log.Printf("  %s MACD: %.4f, Signal: %.4f, Histogram: %.4f",
					symbol, enhancedData.MACD.MACDLine, enhancedData.MACD.SignalLine, enhancedData.MACD.Histogram)
			}
			if enhancedData.StochasticRSI != nil {
				log.Printf("  %s Stochastic RSI: K=%.2f, D=%.2f",
					symbol, enhancedData.StochasticRSI.K, enhancedData.StochasticRSI.D)
			}
			if enhancedData.VWAP != nil {
				log.Printf("  %s VWAP: %.4f, Upper Band: %.4f, Lower Band: %.4f",
					symbol, enhancedData.VWAP.Value, enhancedData.VWAP.UpperBand, enhancedData.VWAP.LowerBand)
			}

			// Calculate combined signal
			combinedSignal := bot.MarketAnalyzer.CalculateCombinedSignal(symbol, enhancedData)
			combinedSignals[symbol] = combinedSignal
			log.Printf("  %s Combined Signal: %s (Score: %.2f, Confidence: %.2f) - %s",
				symbol, combinedSignal.Signal, combinedSignal.Score, combinedSignal.Confidence, combinedSignal.Reason)
		}

		// Analyze volume-weighted signals
		volumeSignal := bot.MarketAnalyzer.AnalyzeVolumeWeightedSignal(symbol, data)
		volumeWeightedSignals[symbol] = volumeSignal
		log.Printf("  %s Volume-Weighted Signal: %s (Confidence: %.2f) - %s",
			symbol, volumeSignal.BaseSignal, volumeSignal.OverallConfidence, volumeSignal.Reason)
	}

	// Update portfolio manager's market analyzer reference
	bot.PortfolioManager.MarketAnalyzer = bot.MarketAnalyzer

	// 3. Calculate correlations between assets
	log.Println("3. Calculating asset correlations...")
	bot.MarketAnalyzer.CalculateCorrelations()

	// Log highly correlated assets for each symbol
	for _, symbol := range bot.PortfolioManager.Symbols {
		highlyCorrelated := bot.MarketAnalyzer.GetHighlyCorrelatedAssets(symbol, 0.7) // 0.7 threshold for high correlation
		if len(highlyCorrelated) > 0 {
			log.Printf("  %s is highly correlated with: %v", symbol, highlyCorrelated)
		}
	}

	// Calculate and log portfolio diversification score
	diversificationScore := bot.MarketAnalyzer.GetDiversificationScore(bot.PortfolioManager.Symbols)
	log.Printf("  Portfolio diversification score: %.2f", diversificationScore)

	// 4. Check stop-loss and take-profit levels
	log.Println("4. Checking stop-loss and take-profit levels...")
	sltpActions := bot.RiskManager.CheckStopLossTakeProfit(currentPrices)
	for _, action := range sltpActions {
		log.Printf("  %s", action)
		// In a real implementation, you would execute the close order here
	}

	// 5. Check symbol drawdown limits
	log.Println("5. Checking symbol drawdown limits...")
	drawdownActions := bot.RiskManager.CheckSymbolDrawdown()
	for _, action := range drawdownActions {
		log.Printf("  %s", action)
		// In a real implementation, you would close positions that exceed drawdown limits
	}

	// 6. Select optimal strategy for each coin
	log.Println("6. Selecting strategies...")
	strategySelections := make(map[string]strategy.StrategyType)

	for _, symbol := range bot.PortfolioManager.Symbols {
		selectedStrategy := bot.StrategyAI.SelectStrategy(symbol)
		strategySelections[symbol] = selectedStrategy
		log.Printf("  %s: %s", symbol, selectedStrategy)
	}

	// 7. Execute strategy-specific logic for each coin and track performance
	log.Println("7. Executing strategies and tracking performance...")
	performanceData := make(map[string]float64)

	for _, symbol := range bot.PortfolioManager.Symbols {
		// Get selected strategy
		strategyType := strategySelections[symbol]
		strategyImpl, exists := bot.Strategies[strategyType]
		if !exists {
			log.Printf("Warning: No implementation for strategy %s", strategyType)
			continue
		}

		// Get market data
		data, exists := marketData[symbol]
		if !exists {
			log.Printf("Warning: No market data for %s", symbol)
			continue
		}

		// Analyze with strategy
		signal := strategyImpl.Analyze(data)
		log.Printf("  %s signal: %s (%.2f) - %s", symbol, signal.Action, signal.Strength, signal.Reason)

		// Execute strategy
		if err := strategyImpl.Execute(signal); err != nil {
			log.Printf("Warning: Failed to execute strategy for %s: %v", symbol, err)
		}

		// Log the trade
		var quantity float64
		var price float64
		if len(data.Kline) > 0 {
			price, _ = data.Kline[len(data.Kline)-1].Close.Float64()
			// Calculate quantity based on allocation and current price
			allocation := bot.PortfolioManager.GetOptimalAllocation(symbol)
			targetValue := bot.Config.TotalCapital * allocation
			quantity = targetValue / price
		}

		bot.PortfolioManager.LogTrade(
			symbol,
			signal.Action,
			quantity,
			price,
			string(strategyType),
			signal.Strength,
			signal.Reason,
		)

		// Send trade alert notification
		if signal.Action != "HOLD" {
			alert := notifications.TradeAlert{
				Symbol:     symbol,
				Action:     signal.Action,
				Quantity:   quantity,
				Price:      price,
				Strategy:   string(strategyType),
				Confidence: signal.Strength,
				Reason:     signal.Reason,
				Timestamp:  time.Now().Format("2006-01-02 15:04:05"),
			}
			bot.Notifier.SendTradeAlert(alert)
		}

		// Track performance based on signal strength and market conditions
		performanceData[symbol] = signal.Strength * 100 // Scale to percentage
	}

	// 8. Update portfolio performance metrics
	log.Println("8. Updating portfolio performance metrics...")
	for symbol, performance := range performanceData {
		bot.PortfolioManager.UpdatePerformance(symbol, performance)
	}

	// 9. Rebalance portfolio based on performance
	log.Println("9. Rebalancing portfolio...")
	err = bot.CircuitBreaker.Call(func() error {
		return bot.PortfolioManager.RebalancePortfolio(ctx)
	})
	if err != nil {
		return fmt.Errorf("failed to rebalance portfolio: %w", err)
	}

	// 10. Check risk metrics and log performance
	log.Println("10. Checking risk metrics and performance...")
	bot.RiskManager.CalculateRiskMetrics()
	log.Printf("Risk Report:\n%s", bot.RiskManager.GetRiskReport())

	// Log performance metrics
	performanceMetrics := bot.PortfolioManager.CalculatePerformanceMetrics()
	log.Printf("Performance Metrics:\n")
	log.Printf("  Total Trades: %d\n", performanceMetrics.TotalTrades)
	log.Printf("  Winning Trades: %d\n", performanceMetrics.WinningTrades)
	log.Printf("  Losing Trades: %d\n", performanceMetrics.LosingTrades)
	log.Printf("  Win Rate: %.2f%%\n", performanceMetrics.WinRate*100)
	log.Printf("  Total PnL: $%.2f\n", performanceMetrics.TotalPnL)
	log.Printf("  Average PnL: $%.2f\n", performanceMetrics.AveragePnL)
	log.Printf("  Max Drawdown: $%.2f\n", performanceMetrics.MaxDrawdown)
	log.Printf("  Sharpe Ratio: %.2f\n", performanceMetrics.SharpeRatio)
	log.Printf("  Sortino Ratio: %.2f\n", performanceMetrics.SortinoRatio)

	if bot.RiskManager.ShouldStopTrading() {
		log.Println("WARNING: Risk limits exceeded, consider stopping trading!")
		// Send emergency stop alert
		bot.Notifier.SendEmergencyStopAlert("Risk limits exceeded")
	}

	log.Println("=== Trading Cycle Complete ===")
	return nil
}

func main() {
	// Create context with cancel
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create trading bot
	bot, err := NewTradingBot()
	if err != nil {
		log.Fatalf("Failed to create trading bot: %v", err)
	}

	// Run the bot
	if err := bot.Run(ctx); err != nil {
		log.Fatalf("Trading bot error: %v", err)
	}
}
