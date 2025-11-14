package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/forbest/bybitgo/internal/backtest"
	"github.com/forbest/bybitgo/internal/market"
	"github.com/forbest/bybitgo/internal/portfolio"
	"github.com/forbest/bybitgo/internal/risk"
)

// Dashboard represents the web dashboard for the trading bot
type Dashboard struct {
	PortfolioManager *portfolio.PortfolioManager
	RiskManager      *risk.RiskManager
	MarketAnalyzer   *market.MarketAnalyzer
	Server           *http.Server
	// Add a channel for manual override commands
	OverrideChannel chan OverrideCommand
	// Add backtest result storage
	BacktestResults map[string]*backtest.BacktestResult
}

// OverrideCommand represents a manual override command
type OverrideCommand struct {
	Command   string            // "start", "stop", "rebalance", "emergency_stop"
	Symbol    string            // Optional, for symbol-specific commands
	Arguments map[string]string // Additional arguments
}

// NewDashboard creates a new Dashboard
func NewDashboard(portfolioManager *portfolio.PortfolioManager, riskManager *risk.RiskManager, marketAnalyzer *market.MarketAnalyzer) *Dashboard {
	return &Dashboard{
		PortfolioManager: portfolioManager,
		RiskManager:      riskManager,
		MarketAnalyzer:   marketAnalyzer,
		OverrideChannel:  make(chan OverrideCommand, 10), // Buffered channel
		BacktestResults:  make(map[string]*backtest.BacktestResult),
	}
}

// Start starts the web dashboard server
func (d *Dashboard) Start(port string) error {
	// Serve static files
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static/"))))

	// Register API handlers
	http.HandleFunc("/api/metrics", d.metricsHandler)
	http.HandleFunc("/api/trades", d.tradesHandler)
	http.HandleFunc("/api/performance", d.performanceHandler)
	http.HandleFunc("/api/risk", d.riskHandler)
	http.HandleFunc("/api/market", d.marketHandler)
	http.HandleFunc("/api/override", d.overrideHandler)
	http.HandleFunc("/api/backtest", d.backtestHandler)
	http.HandleFunc("/api/portfolio", d.portfolioHandler)

	// Serve the main dashboard page
	http.HandleFunc("/", d.dashboardHandler)

	// Create server
	d.Server = &http.Server{
		Addr:    ":" + port,
		Handler: nil,
	}

	fmt.Printf("Starting dashboard server on port %s\n", port)
	return d.Server.ListenAndServe()
}

// Stop stops the web dashboard server
func (d *Dashboard) Stop() error {
	if d.Server != nil {
		return d.Server.Close()
	}
	return nil
}

// dashboardHandler serves the main dashboard page
func (d *Dashboard) dashboardHandler(w http.ResponseWriter, r *http.Request) {
	// Only serve the dashboard for the root path
	if r.URL.Path != "/" && !strings.HasPrefix(r.URL.Path, "/static/") {
		http.NotFound(w, r)
		return
	}

	// If requesting static files, let the file server handle it
	if strings.HasPrefix(r.URL.Path, "/static/") {
		http.StripPrefix("/static/", http.FileServer(http.Dir("web/static/"))).ServeHTTP(w, r)
		return
	}

	// Serve the main index.html file
	http.ServeFile(w, r, "web/static/index.html")
}

// metricsHandler serves performance metrics as JSON
func (d *Dashboard) metricsHandler(w http.ResponseWriter, r *http.Request) {
	metrics := d.PortfolioManager.CalculatePerformanceMetrics()

	response := map[string]interface{}{
		"total_trades":  metrics.TotalTrades,
		"win_rate":      metrics.WinRate,
		"total_pnl":     metrics.TotalPnL,
		"avg_pnl":       metrics.AveragePnL,
		"sharpe_ratio":  metrics.SharpeRatio,
		"sortino_ratio": metrics.SortinoRatio,
		"max_drawdown":  metrics.MaxDrawdown,
		"timestamp":     time.Now().Unix(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// tradesHandler serves recent trades as JSON
func (d *Dashboard) tradesHandler(w http.ResponseWriter, r *http.Request) {
	trades := d.PortfolioManager.GetRecentTrades(50)

	response := map[string]interface{}{
		"trades":    trades,
		"count":     len(trades),
		"timestamp": time.Now().Unix(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// performanceHandler serves performance data as JSON
func (d *Dashboard) performanceHandler(w http.ResponseWriter, r *http.Request) {
	allocations := make(map[string]float64)
	for symbol, allocation := range d.PortfolioManager.Allocations {
		allocations[symbol] = allocation
	}

	response := map[string]interface{}{
		"allocations": allocations,
		"performance": d.PortfolioManager.Performance,
		"timestamp":   time.Now().Unix(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// riskHandler serves risk metrics as JSON
func (d *Dashboard) riskHandler(w http.ResponseWriter, r *http.Request) {
	metrics := d.RiskManager.CalculateRiskMetrics()

	response := map[string]interface{}{
		"total_exposure":     metrics.TotalExposure,
		"portfolio_drawdown": metrics.PortfolioDrawdown,
		"volatility":         metrics.Volatility,
		"correlation_risk":   metrics.CorrelationRisk,
		"timestamp":          time.Now().Unix(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// marketHandler serves market conditions as JSON
func (d *Dashboard) marketHandler(w http.ResponseWriter, r *http.Request) {
	conditions := make(map[string]interface{})

	for _, symbol := range d.PortfolioManager.Symbols {
		regime := d.MarketAnalyzer.GetMarketRegime(symbol)
		conditions[symbol] = map[string]string{
			"volatility": regime.Volatility,
			"trend":      regime.Trend,
			"volume":     regime.Volume,
		}
	}

	response := map[string]interface{}{
		"conditions": conditions,
		"timestamp":  time.Now().Unix(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// portfolioHandler serves portfolio data as JSON
func (d *Dashboard) portfolioHandler(w http.ResponseWriter, r *http.Request) {
	// Get current positions
	allocations := make(map[string]float64)
	for symbol, allocation := range d.PortfolioManager.Allocations {
		allocations[symbol] = allocation
	}

	// Get symbols
	symbols := d.PortfolioManager.Symbols

	// Get performance data
	performance := d.PortfolioManager.Performance

	// Get trade log
	tradeLog := d.PortfolioManager.GetTradeLog()

	response := map[string]interface{}{
		"symbols":       symbols,
		"allocations":   allocations,
		"performance":   performance,
		"trade_log":     tradeLog,
		"total_capital": d.PortfolioManager.Config.TotalCapital,
		"timestamp":     time.Now().Unix(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Add overrideHandler to handle manual override commands
func (d *Dashboard) overrideHandler(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")

	// Handle preflight requests
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse the command from the request body
	var command OverrideCommand
	if err := json.NewDecoder(r.Body).Decode(&command); err != nil {
		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Send the command to the override channel
	select {
	case d.OverrideChannel <- command:
		response := map[string]interface{}{
			"status":  "success",
			"message": fmt.Sprintf("Command '%s' sent successfully", command.Command),
		}
		json.NewEncoder(w).Encode(response)
	default:
		http.Error(w, "Command queue full", http.StatusServiceUnavailable)
	}
}

// GetOverrideChannel returns the override channel for receiving commands
func (d *Dashboard) GetOverrideChannel() <-chan OverrideCommand {
	return d.OverrideChannel
}

// Add backtestHandler to handle backtest requests
func (d *Dashboard) backtestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse the backtest parameters from the request body
	var params struct {
		Strategy       string  `json:"strategy"`
		InitialCapital float64 `json:"initial_capital"`
		StartDate      string  `json:"start_date"`
		EndDate        string  `json:"end_date"`
	}

	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Parse dates
	startDate, err := time.Parse("2006-01-02", params.StartDate)
	if err != nil {
		http.Error(w, "Invalid start date", http.StatusBadRequest)
		return
	}

	endDate, err := time.Parse("2006-01-02", params.EndDate)
	if err != nil {
		http.Error(w, "Invalid end date", http.StatusBadRequest)
		return
	}

	// Create a backtest result (simplified)
	result := &backtest.BacktestResult{
		StrategyName:   params.Strategy,
		StartDate:      startDate,
		EndDate:        endDate,
		InitialCapital: params.InitialCapital,
		FinalCapital:   params.InitialCapital * 1.15, // Sample 15% return
		TotalReturn:    15.0,
		TotalTrades:    50,
		WinningTrades:  30,
		LosingTrades:   20,
		WinRate:        60.0,
		MaxDrawdown:    5.0,
		SharpeRatio:    1.5,
		SortinoRatio:   2.1,
		TradeHistory: []backtest.TradeRecord{
			{
				Timestamp:  startDate.Add(24 * time.Hour),
				Symbol:     "BTCUSDT",
				Action:     "BUY",
				Quantity:   0.1,
				EntryPrice: 50000.0,
				ExitPrice:  51000.0,
				PnL:        1000.0,
				Commission: 10.0,
			},
		},
		EquityCurve: []backtest.EquityPoint{
			{Timestamp: startDate, Equity: params.InitialCapital},
			{Timestamp: startDate.Add(24 * time.Hour), Equity: params.InitialCapital * 1.05},
			{Timestamp: startDate.Add(48 * time.Hour), Equity: params.InitialCapital * 1.10},
			{Timestamp: endDate, Equity: params.InitialCapital * 1.15},
		},
	}

	// Store the result
	d.BacktestResults[params.Strategy] = result

	// Convert to JSON response
	response := map[string]interface{}{
		"strategy_name":   result.StrategyName,
		"start_date":      result.StartDate.Format("2006-01-02"),
		"end_date":        result.EndDate.Format("2006-01-02"),
		"initial_capital": result.InitialCapital,
		"final_capital":   result.FinalCapital,
		"total_return":    result.TotalReturn,
		"total_trades":    result.TotalTrades,
		"winning_trades":  result.WinningTrades,
		"losing_trades":   result.LosingTrades,
		"win_rate":        result.WinRate,
		"max_drawdown":    result.MaxDrawdown,
		"sharpe_ratio":    result.SharpeRatio,
		"sortino_ratio":   result.SortinoRatio,
		"trade_history":   result.TradeHistory,
		"equity_curve":    result.EquityCurve,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
