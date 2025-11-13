package strategy

import (
	"github.com/forbest/bybitgo/internal/bybit"
)

// Strategy defines the interface for trading strategies
type Strategy interface {
	Analyze(marketData *bybit.MarketData) bybit.TradeSignal
	Execute(signal bybit.TradeSignal) error
	GetName() string
	GetParameters() map[string]float64
}
