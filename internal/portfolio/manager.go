package portfolio

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/forbest/bybitgo/internal/bybit"
	"github.com/forbest/bybitgo/internal/config"
	"github.com/forbest/bybitgo/internal/market"
)

// TradeLogEntry represents a single trade log entry
type TradeLogEntry struct {
	Timestamp     time.Time
	Symbol        string
	Action        string // "BUY", "SELL", "HOLD"
	Quantity      float64
	Price         float64
	Strategy      string
	Confidence    float64
	Reason        string
	PnL           float64 // Profit and Loss for this trade
	CumulativePnL float64 // Cumulative PnL for this symbol
}

// PerformanceMetrics tracks performance metrics for the portfolio
type PerformanceMetrics struct {
	TotalTrades   int
	WinningTrades int
	LosingTrades  int
	WinRate       float64
	TotalPnL      float64
	AveragePnL    float64
	MaxDrawdown   float64
	SharpeRatio   float64
	SortinoRatio  float64
}

// PortfolioManager manages the portfolio of cryptocurrencies
type PortfolioManager struct {
	Symbols            []string
	Allocations        map[string]float64
	Performance        map[string]float64 // Track performance of each symbol
	TradeLog           []TradeLogEntry    // Detailed trade log
	PerformanceMetrics PerformanceMetrics // Overall performance metrics
	RebalanceInterval  time.Duration
	BybitClient        *bybit.Client
	Config             *config.Config
	MarketAnalyzer     *market.MarketAnalyzer
}

// NewPortfolioManager creates a new PortfolioManager
func NewPortfolioManager(client *bybit.Client, cfg *config.Config) *PortfolioManager {
	return &PortfolioManager{
		Symbols:           make([]string, 0),
		Allocations:       make(map[string]float64),
		Performance:       make(map[string]float64),
		RebalanceInterval: time.Duration(cfg.RebalanceMinutes) * time.Minute,
		BybitClient:       client,
		Config:            cfg,
		MarketAnalyzer:    market.NewMarketAnalyzer(),
	}
}

// UpdateTopCoins updates the list of top coins based on trading volume
func (pm *PortfolioManager) UpdateTopCoins(ctx context.Context) error {
	// Get top 6 coins from Bybit
	topCoins, err := pm.BybitClient.GetTopCoins(ctx, 6)
	if err != nil {
		return fmt.Errorf("failed to get top coins: %w", err)
	}

	pm.Symbols = topCoins

	// Reset allocations
	pm.Allocations = make(map[string]float64)

	// Equal allocation for now (can be improved with market cap weighting)
	allocation := 1.0 / float64(len(pm.Symbols))
	for _, symbol := range pm.Symbols {
		pm.Allocations[symbol] = allocation
	}

	return nil
}

// GetAllocation returns the capital allocation for a symbol
func (pm *PortfolioManager) GetAllocation(symbol string) float64 {
	if alloc, exists := pm.Allocations[symbol]; exists {
		return alloc
	}
	return 0
}

// GetPerformanceBasedAllocation returns the capital allocation for a symbol adjusted for performance
func (pm *PortfolioManager) GetPerformanceBasedAllocation(symbol string) float64 {
	// Get base allocation
	baseAllocation := pm.GetAllocation(symbol)

	// Get performance data
	performance, exists := pm.Performance[symbol]
	if !exists {
		// If no performance data, return base allocation
		return baseAllocation
	}

	// Adjust allocation based on performance
	// Higher performance = higher allocation, lower performance = lower allocation
	// This is a simplified relationship
	performanceFactor := 1.0

	// If performance is positive, increase allocation
	if performance > 0 {
		performanceFactor = 1.0 + (performance / 100.0) // Scale by percentage
	} else if performance < 0 {
		// If performance is negative, decrease allocation
		performanceFactor = 1.0 + (performance / 100.0) // This will reduce allocation
		// Ensure the factor doesn't go below 0.1 (10% of original allocation)
		if performanceFactor < 0.1 {
			performanceFactor = 0.1
		}
	}

	// Ensure the factor is reasonable (between 0.1 and 2.0)
	if performanceFactor < 0.1 {
		performanceFactor = 0.1
	} else if performanceFactor > 2.0 {
		performanceFactor = 2.0
	}

	return baseAllocation * performanceFactor
}

// GetVolatilityAdjustedAllocation returns the capital allocation for a symbol adjusted for volatility
func (pm *PortfolioManager) GetVolatilityAdjustedAllocation(symbol string) float64 {
	// Get base allocation
	baseAllocation := pm.GetAllocation(symbol)

	// Get volatility data from market analyzer
	volData, exists := pm.MarketAnalyzer.VolatilityTracker[symbol]
	if !exists {
		// If no volatility data, return base allocation
		return baseAllocation
	}

	// Adjust allocation based on volatility
	// Lower volatility = higher allocation, higher volatility = lower allocation
	// This is a simplified inverse relationship
	if volData.RecentVolatility > 0 {
		// Scale allocation inversely with volatility
		// Higher volatility reduces position size
		volatilityFactor := 1.0 / (1.0 + volData.RecentVolatility*100)

		// Ensure the factor is between 0.1 and 2.0
		if volatilityFactor < 0.1 {
			volatilityFactor = 0.1
		} else if volatilityFactor > 2.0 {
			volatilityFactor = 2.0
		}

		return baseAllocation * volatilityFactor
	}

	return baseAllocation
}

// GetOptimalAllocation returns the capital allocation for a symbol considering both performance and volatility
func (pm *PortfolioManager) GetOptimalAllocation(symbol string) float64 {
	// Get performance-based allocation
	perfAllocation := pm.GetPerformanceBasedAllocation(symbol)

	// Get volatility-adjusted allocation
	volAllocation := pm.GetVolatilityAdjustedAllocation(symbol)

	// Combine both factors (simple average)
	return (perfAllocation + volAllocation) / 2.0
}

// UpdatePerformance updates the performance metrics for a symbol
func (pm *PortfolioManager) UpdatePerformance(symbol string, performance float64) {
	// Update performance with exponential moving average to smooth out fluctuations
	currentPerf, exists := pm.Performance[symbol]
	if !exists {
		pm.Performance[symbol] = performance
	} else {
		// Use 80% weight for current performance, 20% for historical
		pm.Performance[symbol] = (performance * 0.8) + (currentPerf * 0.2)
	}
}

// RebalancePortfolio rebalances the portfolio based on current allocations
func (pm *PortfolioManager) RebalancePortfolio(ctx context.Context) error {
	// This is a simplified implementation
	// In practice, you would:
	// 1. Check current positions
	// 2. Calculate target positions based on allocations
	// 3. Place orders to adjust positions

	fmt.Println("Rebalancing portfolio...")

	// Update top coins first
	if err := pm.UpdateTopCoins(ctx); err != nil {
		return fmt.Errorf("failed to update top coins: %w", err)
	}

	// For each symbol, calculate target position size
	for _, symbol := range pm.Symbols {
		// Use optimal allocation (considering both performance and volatility)
		allocation := pm.GetOptimalAllocation(symbol)
		targetValue := pm.Config.TotalCapital * allocation

		fmt.Printf("Symbol: %s, Target Allocation: %.2f%%, Target Value: $%.2f\n",
			symbol, allocation*100, targetValue)

		// Here you would place actual orders to achieve the target allocation
		// This requires checking current positions and placing appropriate orders
	}

	return nil
}

// GetCurrentPositions returns current positions for all symbols
func (pm *PortfolioManager) GetCurrentPositions(ctx context.Context) (map[string][]bybit.Position, error) {
	positions := make(map[string][]bybit.Position)

	for _, symbol := range pm.Symbols {
		pos, err := pm.BybitClient.GetPositions(ctx, symbol)
		if err != nil {
			return nil, fmt.Errorf("failed to get positions for %s: %w", symbol, err)
		}
		positions[symbol] = pos
	}

	return positions, nil
}

// LogTrade adds a trade entry to the trade log
func (pm *PortfolioManager) LogTrade(symbol, action string, quantity, price float64, strategy string, confidence float64, reason string) {
	entry := TradeLogEntry{
		Timestamp:     time.Now(),
		Symbol:        symbol,
		Action:        action,
		Quantity:      quantity,
		Price:         price,
		Strategy:      strategy,
		Confidence:    confidence,
		Reason:        reason,
		PnL:           0, // Will be calculated when position is closed
		CumulativePnL: 0, // Will be updated
	}

	pm.TradeLog = append(pm.TradeLog, entry)
}

// UpdateTradePnL updates the PnL for a trade when a position is closed
func (pm *PortfolioManager) UpdateTradePnL(symbol string, entryPrice, exitPrice float64, quantity float64, isLong bool) {
	pnl := 0.0
	if isLong {
		pnl = (exitPrice - entryPrice) * quantity
	} else {
		pnl = (entryPrice - exitPrice) * quantity
	}

	// Update the latest trade entry for this symbol
	for i := len(pm.TradeLog) - 1; i >= 0; i-- {
		if pm.TradeLog[i].Symbol == symbol {
			pm.TradeLog[i].PnL = pnl
			// Update cumulative PnL
			pm.TradeLog[i].CumulativePnL = pm.PerformanceMetrics.TotalPnL + pnl
			break
		}
	}

	// Update performance metrics
	pm.PerformanceMetrics.TotalPnL += pnl
	pm.PerformanceMetrics.TotalTrades++

	if pnl > 0 {
		pm.PerformanceMetrics.WinningTrades++
	} else {
		pm.PerformanceMetrics.LosingTrades++
	}

	if pm.PerformanceMetrics.TotalTrades > 0 {
		pm.PerformanceMetrics.WinRate = float64(pm.PerformanceMetrics.WinningTrades) / float64(pm.PerformanceMetrics.TotalTrades)
		pm.PerformanceMetrics.AveragePnL = pm.PerformanceMetrics.TotalPnL / float64(pm.PerformanceMetrics.TotalTrades)
	}
}

// GetTradeLog returns the trade log
func (pm *PortfolioManager) GetTradeLog() []TradeLogEntry {
	return pm.TradeLog
}

// GetPerformanceMetrics returns the current performance metrics
func (pm *PortfolioManager) GetPerformanceMetrics() PerformanceMetrics {
	return pm.PerformanceMetrics
}

// GetTradeLogForSymbol returns the trade log for a specific symbol
func (pm *PortfolioManager) GetTradeLogForSymbol(symbol string) []TradeLogEntry {
	var symbolTrades []TradeLogEntry
	for _, trade := range pm.TradeLog {
		if trade.Symbol == symbol {
			symbolTrades = append(symbolTrades, trade)
		}
	}
	return symbolTrades
}

// GetRecentTrades returns the most recent trades
func (pm *PortfolioManager) GetRecentTrades(count int) []TradeLogEntry {
	if len(pm.TradeLog) <= count {
		return pm.TradeLog
	}

	return pm.TradeLog[len(pm.TradeLog)-count:]
}

// CalculatePerformanceMetrics calculates detailed performance metrics
func (pm *PortfolioManager) CalculatePerformanceMetrics() PerformanceMetrics {
	if len(pm.TradeLog) == 0 {
		return pm.PerformanceMetrics
	}

	// Reset metrics
	metrics := PerformanceMetrics{
		TotalTrades:   len(pm.TradeLog),
		WinningTrades: 0,
		LosingTrades:  0,
		TotalPnL:      0,
		MaxDrawdown:   0,
	}

	// Calculate basic metrics
	var profits []float64
	var losses []float64
	var cumulativePnL float64
	var peakPnL float64

	for _, trade := range pm.TradeLog {
		metrics.TotalPnL += trade.PnL
		cumulativePnL += trade.PnL

		if cumulativePnL > peakPnL {
			peakPnL = cumulativePnL
		}

		// Calculate drawdown
		drawdown := peakPnL - cumulativePnL
		if drawdown > metrics.MaxDrawdown {
			metrics.MaxDrawdown = drawdown
		}

		if trade.PnL > 0 {
			metrics.WinningTrades++
			profits = append(profits, trade.PnL)
		} else if trade.PnL < 0 {
			metrics.LosingTrades++
			losses = append(losses, math.Abs(trade.PnL))
		}
	}

	// Calculate win rate
	if metrics.TotalTrades > 0 {
		metrics.WinRate = float64(metrics.WinningTrades) / float64(metrics.TotalTrades)
	}

	// Calculate average PnL
	if metrics.TotalTrades > 0 {
		metrics.AveragePnL = metrics.TotalPnL / float64(metrics.TotalTrades)
	}

	// Calculate Sharpe ratio (simplified)
	if len(profits) > 0 || len(losses) > 0 {
		var returns []float64
		for _, trade := range pm.TradeLog {
			if trade.Quantity > 0 && trade.Price > 0 {
				returns = append(returns, trade.PnL/(trade.Quantity*trade.Price))
			} else {
				returns = append(returns, 0)
			}
		}

		// Calculate standard deviation of returns
		if len(returns) > 1 {
			sum := 0.0
			for _, r := range returns {
				sum += r
			}
			mean := sum / float64(len(returns))

			variance := 0.0
			for _, r := range returns {
				variance += math.Pow(r-mean, 2)
			}
			stdDev := math.Sqrt(variance / float64(len(returns)-1))

			// Sharpe ratio (assuming risk-free rate of 0)
			if stdDev > 0 {
				metrics.SharpeRatio = mean / stdDev
			}

			// Sortino ratio (considering only negative returns)
			if len(losses) > 0 {
				downsideSum := 0.0
				for _, loss := range losses {
					downsideSum += math.Pow(loss, 2)
				}
				downsideDev := math.Sqrt(downsideSum / float64(len(losses)))

				if downsideDev > 0 {
					metrics.SortinoRatio = mean / downsideDev
				}
			}
		}
	}

	// Update the stored metrics
	pm.PerformanceMetrics = metrics

	return metrics
}

// GetSymbolPerformanceMetrics returns performance metrics for a specific symbol
func (pm *PortfolioManager) GetSymbolPerformanceMetrics(symbol string) PerformanceMetrics {
	var symbolTrades []TradeLogEntry
	for _, trade := range pm.TradeLog {
		if trade.Symbol == symbol {
			symbolTrades = append(symbolTrades, trade)
		}
	}

	if len(symbolTrades) == 0 {
		return PerformanceMetrics{}
	}

	// Create a temporary PortfolioManager for this symbol
	tempPM := &PortfolioManager{
		TradeLog: symbolTrades,
	}

	return tempPM.CalculatePerformanceMetrics()
}

// GetPerformanceSummary returns a summary of performance metrics
func (pm *PortfolioManager) GetPerformanceSummary() string {
	metrics := pm.CalculatePerformanceMetrics()

	summary := fmt.Sprintf("Performance Summary:\n")
	summary += fmt.Sprintf("  Total Trades: %d\n", metrics.TotalTrades)
	summary += fmt.Sprintf("  Winning Trades: %d\n", metrics.WinningTrades)
	summary += fmt.Sprintf("  Losing Trades: %d\n", metrics.LosingTrades)
	summary += fmt.Sprintf("  Win Rate: %.2f%%\n", metrics.WinRate*100)
	summary += fmt.Sprintf("  Total PnL: $%.2f\n", metrics.TotalPnL)
	summary += fmt.Sprintf("  Average PnL: $%.2f\n", metrics.AveragePnL)
	summary += fmt.Sprintf("  Max Drawdown: $%.2f\n", metrics.MaxDrawdown)
	summary += fmt.Sprintf("  Sharpe Ratio: %.2f\n", metrics.SharpeRatio)
	summary += fmt.Sprintf("  Sortino Ratio: %.2f\n", metrics.SortinoRatio)

	return summary
}
