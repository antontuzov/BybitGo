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
        body: JSON.stringify({Command: command}),
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