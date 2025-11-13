package web

import (
	"encoding/json"
	"fmt"
	"net/http"
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
	// Register handlers
	http.HandleFunc("/api/metrics", d.metricsHandler)
	http.HandleFunc("/api/trades", d.tradesHandler)
	http.HandleFunc("/api/performance", d.performanceHandler)
	http.HandleFunc("/api/risk", d.riskHandler)
	http.HandleFunc("/api/market", d.marketHandler)
	http.HandleFunc("/api/override", d.overrideHandler)
	http.HandleFunc("/api/backtest", d.backtestHandler)
	http.HandleFunc("/api/portfolio", d.portfolioHandler)
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
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	html := `<!DOCTYPE html>
<html>
<head>
    <title>Bybit Trading Bot Dashboard</title>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            margin: 0;
            padding: 20px;
            background-color: #f5f5f5;
            color: #333;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
        }
        header {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 20px;
            border-radius: 10px;
            margin-bottom: 20px;
            box-shadow: 0 4px 6px rgba(0,0,0,0.1);
        }
        h1 {
            margin: 0;
            font-size: 2em;
        }
        .grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
            gap: 20px;
            margin-bottom: 20px;
        }
        .card {
            background: white;
            border-radius: 10px;
            padding: 20px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        .card h2 {
            margin-top: 0;
            color: #667eea;
            border-bottom: 2px solid #f0f0f0;
            padding-bottom: 10px;
        }
        .metric {
            display: flex;
            justify-content: space-between;
            padding: 8px 0;
            border-bottom: 1px solid #f0f0f0;
        }
        .metric:last-child {
            border-bottom: none;
        }
        .metric-label {
            font-weight: 500;
        }
        .metric-value {
            font-weight: 600;
        }
        .positive {
            color: #4caf50;
        }
        .negative {
            color: #f44336;
        }
        .chart-container {
            height: 300px;
            margin: 20px 0;
            position: relative;
        }
        table {
            width: 100%;
            border-collapse: collapse;
            margin: 20px 0;
        }
        th, td {
            padding: 12px;
            text-align: left;
            border-bottom: 1px solid #ddd;
        }
        th {
            background-color: #f8f9fa;
            font-weight: 600;
        }
        tr:hover {
            background-color: #f5f5f5;
        }
        .action-buy {
            color: #4caf50;
            font-weight: bold;
        }
        .action-sell {
            color: #f44336;
            font-weight: bold;
        }
        .action-hold {
            color: #ff9800;
            font-weight: bold;
        }
        footer {
            text-align: center;
            margin-top: 40px;
            color: #666;
            font-size: 0.9em;
        }
        .refresh-btn {
            background: #667eea;
            color: white;
            border: none;
            padding: 10px 20px;
            border-radius: 5px;
            cursor: pointer;
            font-size: 1em;
            margin: 10px 0;
        }
        .refresh-btn:hover {
            background: #5a6fd8;
        }
        .controls {
            background: #e3f2fd;
            border-radius: 10px;
            padding: 20px;
            margin-bottom: 20px;
        }
        .control-btn {
            background: #2196f3;
            color: white;
            border: none;
            padding: 10px 15px;
            margin: 5px;
            border-radius: 5px;
            cursor: pointer;
            font-size: 0.9em;
        }
        .control-btn:hover {
            background: #1976d2;
        }
        .control-btn.stop {
            background: #f44336;
        }
        .control-btn.stop:hover {
            background: #d32f2f;
        }
        .control-btn.emergency {
            background: #ff9800;
        }
        .control-btn.emergency:hover {
            background: #f57c00;
        }
        .control-btn:disabled {
            background: #ccc;
            cursor: not-allowed;
        }
        .override-log {
            background: #f5f5f5;
            border-radius: 5px;
            padding: 10px;
            margin-top: 10px;
            max-height: 200px;
            overflow-y: auto;
            font-family: monospace;
            font-size: 0.9em;
        }
        .tabs {
            display: flex;
            margin-bottom: 20px;
            border-bottom: 1px solid #ddd;
        }
        .tab {
            padding: 10px 20px;
            cursor: pointer;
            background: #f5f5f5;
            border: 1px solid #ddd;
            border-bottom: none;
            border-radius: 5px 5px 0 0;
            margin-right: 5px;
        }
        .tab.active {
            background: white;
            font-weight: bold;
        }
        .tab-content {
            display: none;
        }
        .tab-content.active {
            display: block;
        }
        canvas {
            width: 100%;
            height: 100%;
        }
    </style>
</head>
<body>
    <div class="container">
        <header>
            <h1>Bybit Trading Bot Dashboard</h1>
            <p>Real-time monitoring of your automated trading strategies</p>
        </header>

        <!-- Manual Controls Section -->
        <div class="card controls">
            <h2>Manual Controls</h2>
            <button class="control-btn" onclick="sendCommand('start')">Start Trading</button>
            <button class="control-btn stop" onclick="sendCommand('stop')">Stop Trading</button>
            <button class="control-btn" onclick="sendCommand('rebalance')">Rebalance Portfolio</button>
            <button class="control-btn emergency" onclick="sendCommand('emergency_stop')">Emergency Stop</button>
            
            <div class="override-log" id="override-log">
                <p>Manual override commands will appear here...</p>
            </div>
        </div>

        <!-- Tab Navigation -->
        <div class="tabs">
            <div class="tab active" onclick="switchTab('live')">Live Trading</div>
            <div class="tab" onclick="switchTab('backtest')">Backtesting</div>
        </div>

        <!-- Live Trading Tab -->
        <div id="live-tab" class="tab-content active">
            <div class="grid">
                <div class="card">
                    <h2>Performance Metrics</h2>
                    <div id="performance-metrics">
                        <div class="metric">
                            <span class="metric-label">Total Trades:</span>
                            <span class="metric-value" id="total-trades">0</span>
                        </div>
                        <div class="metric">
                            <span class="metric-label">Win Rate:</span>
                            <span class="metric-value" id="win-rate">0%</span>
                        </div>
                        <div class="metric">
                            <span class="metric-label">Total PnL:</span>
                            <span class="metric-value positive" id="total-pnl">$0.00</span>
                        </div>
                        <div class="metric">
                            <span class="metric-label">Average PnL:</span>
                            <span class="metric-value" id="avg-pnl">$0.00</span>
                        </div>
                        <div class="metric">
                            <span class="metric-label">Sharpe Ratio:</span>
                            <span class="metric-value" id="sharpe-ratio">0.00</span>
                        </div>
                    </div>
                </div>

                <div class="card">
                    <h2>Risk Metrics</h2>
                    <div id="risk-metrics">
                        <div class="metric">
                            <span class="metric-label">Total Exposure:</span>
                            <span class="metric-value" id="total-exposure">$0.00</span>
                        </div>
                        <div class="metric">
                            <span class="metric-label">Portfolio Drawdown:</span>
                            <span class="metric-value negative" id="portfolio-drawdown">0%</span>
                        </div>
                        <div class="metric">
                            <span class="metric-label">Volatility:</span>
                            <span class="metric-value" id="volatility">0%</span>
                        </div>
                        <div class="metric">
                            <span class="metric-label">Correlation Risk:</span>
                            <span class="metric-value" id="correlation-risk">0</span>
                        </div>
                    </div>
                </div>

                <div class="card">
                    <h2>Portfolio Allocation</h2>
                    <div id="portfolio-allocation">
                        <!-- Portfolio allocation will be populated here -->
                    </div>
                </div>
                
                <div class="card">
                    <h2>Portfolio Details</h2>
                    <div id="portfolio-details">
                        <div class="metric">
                            <span class="metric-label">Total Capital:</span>
                            <span class="metric-value" id="total-capital">$0.00</span>
                        </div>
                        <div class="metric">
                            <span class="metric-label">Active Symbols:</span>
                            <span class="metric-value" id="active-symbols">0</span>
                        </div>
                    </div>
                </div>
            </div>

            <div class="card">
                <h2>Recent Trades</h2>
                <table id="trades-table">
                    <thead>
                        <tr>
                            <th>Time</th>
                            <th>Symbol</th>
                            <th>Action</th>
                            <th>Quantity</th>
                            <th>Price</th>
                            <th>Strategy</th>
                            <th>Confidence</th>
                            <th>PnL</th>
                        </tr>
                    </thead>
                    <tbody id="trades-body">
                        <!-- Trades will be populated here -->
                    </tbody>
                </table>
            </div>

            <div class="card">
                <h2>Market Conditions</h2>
                <div id="market-conditions">
                    <!-- Market conditions will be populated here -->
                </div>
            </div>
        </div>

        <!-- Backtesting Tab -->
        <div id="backtest-tab" class="tab-content">
            <div class="card">
                <h2>Backtest Configuration</h2>
                <div>
                    <label>Strategy: 
                        <select id="backtest-strategy">
                            <option value="market_making">Market Making</option>
                            <option value="momentum">Momentum</option>
                            <option value="mean_reversion">Mean Reversion</option>
                            <option value="volatility_breakout">Volatility Breakout</option>
                        </select>
                    </label>
                    <label>Initial Capital: $<input type="number" id="initial-capital" value="10000" min="1000"></label>
                    <label>Start Date: <input type="date" id="start-date" value="2023-01-01"></label>
                    <label>End Date: <input type="date" id="end-date" value="2023-12-31"></label>
                    <button class="control-btn" onclick="runBacktest()">Run Backtest</button>
                </div>
            </div>

            <div class="card">
                <h2>Backtest Results</h2>
                <div id="backtest-results">
                    <p>Run a backtest to see results...</p>
                </div>
            </div>

            <div class="card">
                <h2>Equity Curve</h2>
                <div class="chart-container">
                    <canvas id="equity-chart"></canvas>
                </div>
            </div>

            <div class="card">
                <h2>Trade History</h2>
                <table id="backtest-trades-table">
                    <thead>
                        <tr>
                            <th>Date</th>
                            <th>Symbol</th>
                            <th>Action</th>
                            <th>Quantity</th>
                            <th>Entry Price</th>
                            <th>Exit Price</th>
                            <th>PnL</th>
                        </tr>
                    </thead>
                    <tbody id="backtest-trades-body">
                        <!-- Backtest trades will be populated here -->
                    </tbody>
                </table>
            </div>
        </div>

        <footer>
            <p>Bybit Trading Bot Dashboard | Last Updated: <span id="last-updated"></span></p>
        </footer>
    </div>

    <script>
        // Tab switching functionality
        function switchTab(tabName) {
            // Hide all tab contents
            document.querySelectorAll('.tab-content').forEach(tab => {
                tab.classList.remove('active');
            });
            
            // Remove active class from all tabs
            document.querySelectorAll('.tab').forEach(tab => {
                tab.classList.remove('active');
            });
            
            // Show selected tab content
            document.getElementById(tabName + '-tab').classList.add('active');
            
            // Set active class on clicked tab
            event.target.classList.add('active');
        }

        // Function to refresh all data
        function refreshData() {
            fetchMetrics();
            fetchTrades();
            fetchPerformance();
            fetchRisk();
            fetchMarket();
            fetchPortfolio();
            document.getElementById('last-updated').textContent = new Date().toLocaleString();
        }

        // Fetch performance metrics
        function fetchMetrics() {
            fetch('/api/metrics')
                .then(response => response.json())
                .then(data => {
                    document.getElementById('total-trades').textContent = data.total_trades;
                    document.getElementById('win-rate').textContent = (data.win_rate * 100).toFixed(2) + '%';
                    document.getElementById('total-pnl').textContent = '$' + data.total_pnl.toFixed(2);
                    document.getElementById('avg-pnl').textContent = '$' + data.avg_pnl.toFixed(2);
                    document.getElementById('sharpe-ratio').textContent = data.sharpe_ratio.toFixed(2);
                    
                    // Update PnL color based on value
                    const pnlElement = document.getElementById('total-pnl');
                    if (data.total_pnl > 0) {
                        pnlElement.className = 'metric-value positive';
                    } else if (data.total_pnl < 0) {
                        pnlElement.className = 'metric-value negative';
                    } else {
                        pnlElement.className = 'metric-value';
                    }
                })
                .catch(error => console.error('Error fetching metrics:', error));
        }

        // Fetch recent trades
        function fetchTrades() {
            fetch('/api/trades')
                .then(response => response.json())
                .then(data => {
                    const tbody = document.getElementById('trades-body');
                    tbody.innerHTML = '';
                    
                    data.trades.slice(0, 10).forEach(trade => {
                        const row = document.createElement('tr');
                        row.innerHTML = '<td>' + new Date(trade.timestamp).toLocaleTimeString() + '</td>' +
                            '<td>' + trade.symbol + '</td>' +
                            '<td class="action-' + trade.action.toLowerCase() + '">' + trade.action + '</td>' +
                            '<td>' + trade.quantity.toFixed(4) + '</td>' +
                            '<td>$' + trade.price.toFixed(4) + '</td>' +
                            '<td>' + trade.strategy + '</td>' +
                            '<td>' + (trade.confidence * 100).toFixed(1) + '%</td>' +
                            '<td class="' + (trade.pnl > 0 ? 'positive' : trade.pnl < 0 ? 'negative' : '') + '">$' +
                                trade.pnl.toFixed(2) + '</td>';
                        tbody.appendChild(row);
                    });
                })
                .catch(error => console.error('Error fetching trades:', error));
        }

        // Fetch performance data
        function fetchPerformance() {
            fetch('/api/performance')
                .then(response => response.json())
                .then(data => {
                    const container = document.getElementById('portfolio-allocation');
                    container.innerHTML = '';
                    
                    Object.keys(data.allocations).forEach(symbol => {
                        const allocation = data.allocations[symbol];
                        const div = document.createElement('div');
                        div.className = 'metric';
                        div.innerHTML = '<span class="metric-label">' + symbol + ':</span>' +
                            '<span class="metric-value">' + (allocation * 100).toFixed(2) + '%</span>';
                        container.appendChild(div);
                    });
                })
                .catch(error => console.error('Error fetching performance:', error));
        }

        // Fetch risk data
        function fetchRisk() {
            fetch('/api/risk')
                .then(response => response.json())
                .then(data => {
                    document.getElementById('total-exposure').textContent = '$' + data.total_exposure.toFixed(2);
                    document.getElementById('portfolio-drawdown').textContent = (data.portfolio_drawdown * 100).toFixed(2) + '%';
                    document.getElementById('volatility').textContent = (data.volatility * 100).toFixed(2) + '%';
                    document.getElementById('correlation-risk').textContent = data.correlation_risk.toFixed(2);
                })
                .catch(error => console.error('Error fetching risk:', error));
        }

        // Fetch portfolio data
        function fetchPortfolio() {
            fetch('/api/portfolio')
                .then(response => response.json())
                .then(data => {
                    document.getElementById('total-capital').textContent = '$' + data.total_capital.toFixed(2);
                    document.getElementById('active-symbols').textContent = data.symbols.length;
                })
                .catch(error => console.error('Error fetching portfolio:', error));
        }

        // Fetch market data
        function fetchMarket() {
            fetch('/api/market')
                .then(response => response.json())
                .then(data => {
                    const container = document.getElementById('market-conditions');
                    container.innerHTML = '';
                    
                    Object.keys(data.conditions).forEach(symbol => {
                        const condition = data.conditions[symbol];
                        const div = document.createElement('div');
                        div.className = 'metric';
                        div.innerHTML = '<span class="metric-label">' + symbol + ':</span>' +
                            '<span class="metric-value">' + condition.volatility + ', ' + condition.trend + ', ' + condition.volume + '</span>';
                        container.appendChild(div);
                    });
                })
                .catch(error => console.error('Error fetching market:', error));
        }

        // Send manual override command
        function sendCommand(command) {
            fetch('/api/override', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({command: command}),
            })
            .then(response => response.json())
            .then(data => {
                // Log the command
                const log = document.getElementById('override-log');
                const entry = document.createElement('div');
                entry.textContent = '[' + new Date().toLocaleTimeString() + '] ' + data.message;
                log.appendChild(entry);
                log.scrollTop = log.scrollHeight;
                    
                // Show confirmation
                alert(data.message);
            })
            .catch(error => {
                console.error('Error sending command:', error);
                alert('Error sending command: ' + error.message);
            });
        }

        // Run backtest
        function runBacktest() {
            const strategy = document.getElementById('backtest-strategy').value;
            const initialCapital = document.getElementById('initial-capital').value;
            const startDate = document.getElementById('start-date').value;
            const endDate = document.getElementById('end-date').value;

            const data = {
                strategy: strategy,
                initial_capital: parseFloat(initialCapital),
                start_date: startDate,
                end_date: endDate
            };

            fetch('/api/backtest', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(data),
            })
            .then(response => response.json())
            .then(data => {
                displayBacktestResults(data);
            })
            .catch(error => {
                console.error('Error running backtest:', error);
                alert('Error running backtest: ' + error.message);
            });
        }

        // Display backtest results
        function displayBacktestResults(data) {
            const resultsDiv = document.getElementById('backtest-results');
            resultsDiv.innerHTML = '<div class="metric">' +
                '<span class="metric-label">Strategy:</span>' +
                '<span class="metric-value">' + data.strategy_name + '</span>' +
                '</div>' +
                '<div class="metric">' +
                '<span class="metric-label">Initial Capital:</span>' +
                '<span class="metric-value">$' + data.initial_capital.toFixed(2) + '</span>' +
                '</div>' +
                '<div class="metric">' +
                '<span class="metric-label">Final Capital:</span>' +
                '<span class="metric-value">$' + data.final_capital.toFixed(2) + '</span>' +
                '</div>' +
                '<div class="metric">' +
                '<span class="metric-label">Total Return:</span>' +
                '<span class="metric-value ' + (data.total_return >= 0 ? 'positive' : 'negative') + '">' + data.total_return.toFixed(2) + '%</span>' +
                '</div>' +
                '<div class="metric">' +
                '<span class="metric-label">Total Trades:</span>' +
                '<span class="metric-value">' + data.total_trades + '</span>' +
                '</div>' +
                '<div class="metric">' +
                '<span class="metric-label">Win Rate:</span>' +
                '<span class="metric-value">' + data.win_rate.toFixed(2) + '%</span>' +
                '</div>' +
                '<div class="metric">' +
                '<span class="metric-label">Max Drawdown:</span>' +
                '<span class="metric-value negative">' + data.max_drawdown.toFixed(2) + '%</span>' +
                '</div>' +
                '<div class="metric">' +
                '<span class="metric-label">Sharpe Ratio:</span>' +
                '<span class="metric-value">' + data.sharpe_ratio.toFixed(2) + '</span>' +
                '</div>' +
                '<div class="metric">' +
                '<span class="metric-label">Sortino Ratio:</span>' +
                '<span class="metric-value">' + data.sortino_ratio.toFixed(2) + '</span>' +
                '</div>';

            // Display trade history
            const tradesBody = document.getElementById('backtest-trades-body');
            tradesBody.innerHTML = '';
            data.trade_history.forEach(trade => {
                const row = document.createElement('tr');
                row.innerHTML = '<td>' + new Date(trade.timestamp).toLocaleDateString() + '</td>' +
                    '<td>' + trade.symbol + '</td>' +
                    '<td class="action-' + trade.action.toLowerCase() + '">' + trade.action + '</td>' +
                    '<td>' + trade.quantity.toFixed(4) + '</td>' +
                    '<td>$' + trade.entry_price.toFixed(2) + '</td>' +
                    '<td>$' + trade.exit_price.toFixed(2) + '</td>' +
                    '<td class="' + (trade.pnl >= 0 ? 'positive' : 'negative') + '">$' + trade.pnl.toFixed(2) + '</td>';
                tradesBody.appendChild(row);
            });

            // Update equity chart (simplified)
            updateEquityChart(data.equity_curve);
        }

        // Update equity chart (simplified implementation)
        function updateEquityChart(equityCurve) {
            const canvas = document.getElementById('equity-chart');
            const ctx = canvas.getContext('2d');
            
            // Clear canvas
            ctx.clearRect(0, 0, canvas.width, canvas.height);
            
            if (equityCurve.length === 0) return;
            
            // Simple line chart implementation
            ctx.beginPath();
            ctx.strokeStyle = '#2196f3';
            ctx.lineWidth = 2;
            
            const width = canvas.width;
            const height = canvas.height;
            const padding = 20;
            
            // Find min and max equity values
            let minEquity = equityCurve[0].equity;
            let maxEquity = equityCurve[0].equity;
            equityCurve.forEach(point => {
                if (point.equity < minEquity) minEquity = point.equity;
                if (point.equity > maxEquity) maxEquity = point.equity;
            });
            
            // Draw the line
            equityCurve.forEach((point, index) => {
                const x = padding + (index / (equityCurve.length - 1)) * (width - 2 * padding);
                const y = height - padding - ((point.equity - minEquity) / (maxEquity - minEquity)) * (height - 2 * padding);
                
                if (index === 0) {
                    ctx.moveTo(x, y);
                } else {
                    ctx.lineTo(x, y);
                }
            });
            
            ctx.stroke();
        }
        
        // Initial data load
        document.addEventListener('DOMContentLoaded', function() {
            refreshData();
            // Refresh data every 30 seconds
            setInterval(refreshData, 30000);
        });
    </script>
</body>
</html>`
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
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
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse the command from the request body
	var command OverrideCommand
	if err := json.NewDecoder(r.Body).Decode(&command); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Send the command to the override channel
	select {
	case d.OverrideChannel <- command:
		response := map[string]interface{}{
			"status":  "success",
			"message": fmt.Sprintf("Command '%s' sent successfully", command.Command),
		}
		w.Header().Set("Content-Type", "application/json")
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
