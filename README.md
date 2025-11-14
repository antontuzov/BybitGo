# Bybit Trading Bot

An advanced Golang trading bot for Bybit testnet that combines market-making strategies with portfolio management across top cryptocurrencies.

## Features

### Core Functionality
- **Multi-Coin Portfolio Management**: Automatically trades the top 6 cryptocurrencies
- **Dynamic Strategy Selection**: Updates strategy every 5 minutes based on market conditions
- **Real Bybit API Integration**: Connects to Bybit testnet for live trading

### Trading Strategies
- **Market Making**: Avellaneda-Stoikov market making model
- **Momentum Trading**: Trend-following strategy
- **Mean Reversion**: Contrarian strategy for overbought/oversold conditions
- **Volatility Breakout**: Strategy for high volatility market conditions

### Risk Management
- **Stop-Loss and Take-Profit**: Configurable levels per trade
- **Position Sizing**: Based on volatility analysis
- **Maximum Drawdown Limits**: Per symbol and portfolio-wide
- **Trailing Stop**: Dynamic stop-loss adjustment
- **Correlation Analysis**: Diversification risk management

### Technical Analysis
- **MACD**: Moving Average Convergence Divergence
- **Stochastic RSI**: Momentum oscillator
- **VWAP**: Volume Weighted Average Price
- **Custom Indicator Combinations**: Configurable indicator weights
- **Volume-Weighted Strategies**: Incorporates volume analysis

### Monitoring & Analytics
- **Performance Dashboard**: Real-time monitoring of bot performance
- **Trade Logging**: Detailed record of all trading activity
- **Performance Metrics**: Sharpe ratio, Sortino ratio, win rate, etc.
- **Risk Metrics**: Exposure, drawdown, volatility tracking

### Notification System
- **Email/SMS Alerts**: Trade notifications
- **Telegram Integration**: Real-time updates
- **Emergency Stop Alerts**: Critical risk notifications

### Web Interface
- **Live Trading Dashboard**: Monitor performance in real-time
- **Manual Override Controls**: Start/stop trading, rebalance portfolio
- **Backtesting Visualization**: Strategy performance analysis

## Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/bybit-trading-bot.git
   cd bybit-trading-bot
   ```

2. Install dependencies:
   ```bash
   go mod tidy
   ```

3. Configure environment variables:
   Copy `.env.example` to `.env` and update with your Bybit API credentials:
   ```bash
   cp .env.example .env
   ```

4. Build the application:
   ```bash
   go build -o bot cmd/bot/main.go
   ```

## Configuration

The bot is configured through environment variables in the `.env` file:

- `BYBIT_API_KEY`: Your Bybit API key
- `BYBIT_API_SECRET`: Your Bybit API secret
- `TESTNET`: Set to "true" for testnet, "false" for mainnet
- `TOTAL_CAPITAL`: Total capital for portfolio management
- `MAX_POSITION_PER_COIN`: Maximum position size per coin
- `REBALANCE_MINUTES`: Portfolio rebalance interval in minutes
- `STOP_LOSS_PERCENT`: Stop-loss percentage
- `TAKE_PROFIT_PERCENT`: Take-profit percentage

## Usage

Start the trading bot:
```bash
./bot
```

Access the web dashboard at http://localhost:8080

## Automated Trading

The bot is configured to automatically trade every 5 minutes as specified by the `REBALANCE_MINUTES=5` setting in the `.env` file. The bot will:

1. Analyze market conditions for the top 6 cryptocurrencies
2. Select optimal strategies for each coin
3. Execute trades based on strategy signals
4. Rebalance the portfolio based on performance
5. Monitor risk metrics and adjust positions accordingly

To run the bot in the background:
```bash
./run-bot.sh
```

## Web Dashboard Deployment

### Vercel Deployment

The web dashboard can be deployed to Vercel for online access:

1. The static files are located in the `web/static/` directory
2. Vercel configuration is in `vercel.json`
3. Push this repository to GitHub
4. Connect your GitHub repository to Vercel
5. Vercel will automatically deploy the dashboard

### Local Development

For local development of the web interface:

1. Start the trading bot backend (runs on port 8080 by default)
2. Open `web/static/index.html` in a browser
3. The frontend will automatically connect to the backend API

Note: For local development, you may need to configure CORS settings in the backend if serving the frontend from a different port.

## API Endpoints

- `/api/metrics`: Performance metrics
- `/api/trades`: Recent trades
- `/api/performance`: Portfolio performance
- `/api/risk`: Risk metrics
- `/api/market`: Market conditions
- `/api/portfolio`: Portfolio details
- `/api/override`: Manual controls
- `/api/backtest`: Backtesting

## License

This project is licensed under the MIT License - see the LICENSE file for details.