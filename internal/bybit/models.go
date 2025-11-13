package bybit

import (
	"time"

	"github.com/shopspring/decimal"
)

// KlineData represents kline/candlestick data
type KlineData struct {
	Open      decimal.Decimal
	High      decimal.Decimal
	Low       decimal.Decimal
	Close     decimal.Decimal
	Volume    decimal.Decimal
	Timestamp time.Time
}

// MarketData represents market data for a symbol
type MarketData struct {
	Symbol    string
	Timestamp time.Time
	Kline     []KlineData // List of kline data
	// Add other fields as needed for mock implementation
}

// Order represents a trading order
type Order struct {
	Symbol   string
	Side     string // BUY, SELL
	Type     string // MARKET, LIMIT
	Quantity decimal.Decimal
	Price    decimal.Decimal
}

// Position represents a trading position
type Position struct {
	Symbol        string
	Side          string
	Size          decimal.Decimal
	AvgPrice      decimal.Decimal
	UnrealisedPnl decimal.Decimal
}

// TradeSignal represents a trading signal
type TradeSignal struct {
	Symbol   string
	Action   string // BUY, SELL, HOLD
	Strength float64
	Reason   string
}
