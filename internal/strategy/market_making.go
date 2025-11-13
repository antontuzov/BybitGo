package strategy

import (
	"fmt"

	"github.com/forbest/bybitgo/internal/bybit"
	"github.com/shopspring/decimal"
)

// MarketMakingStrategy implements the Avellaneda-Stoikov market making model
type MarketMakingStrategy struct {
	Parameters map[string]float64
}

// NewMarketMakingStrategy creates a new MarketMakingStrategy
func NewMarketMakingStrategy() *MarketMakingStrategy {
	return &MarketMakingStrategy{
		Parameters: map[string]float64{
			"gamma":     0.1,  // Risk factor
			"k":         1.5,  // Order book liquidity factor
			"sigma":     0.02, // Volatility estimate
			"tick_size": 0.1,  // Minimum price increment
		},
	}
}

// GetName returns the strategy name
func (mms *MarketMakingStrategy) GetName() string {
	return string(MarketMaking)
}

// Analyze implements the strategy analysis logic
func (mms *MarketMakingStrategy) Analyze(marketData *bybit.MarketData) bybit.TradeSignal {
	if marketData == nil || len(marketData.Kline) == 0 {
		return bybit.TradeSignal{
			Symbol: marketData.Symbol,
			Action: "HOLD",
			Reason: "Insufficient market data",
		}
	}

	// Use the last kline data for price
	lastKline := marketData.Kline[len(marketData.Kline)-1]
	midPrice := lastKline.Close // Simplified - using close price as mid price

	// Calculate bid-ask spread using Avellaneda-Stoikov formula
	gamma := mms.Parameters["gamma"]

	// Simplified optimal spread calculation
	optimalSpread := gamma * mms.Parameters["sigma"] * mms.Parameters["sigma"]

	// Bid and ask prices
	bidPrice := midPrice.Sub(decimal.NewFromFloat(optimalSpread / 2))
	askPrice := midPrice.Add(decimal.NewFromFloat(optimalSpread / 2))

	signal := "HOLD"
	reason := fmt.Sprintf("Optimal spread: %.4f, Bid: %s, Ask: %s", optimalSpread, bidPrice.String(), askPrice.String())

	// Determine action based on spread and market conditions
	if optimalSpread > 0.01 { // Minimum threshold for profitable spread
		signal = "PLACE_ORDERS"
		reason = fmt.Sprintf("Market making opportunity detected. Spread: %.4f", optimalSpread)
	}

	return bybit.TradeSignal{
		Symbol:   marketData.Symbol,
		Action:   signal,
		Strength: 1.0 - optimalSpread, // Lower spread = higher strength
		Reason:   reason,
	}
}

// Execute places market making orders
func (mms *MarketMakingStrategy) Execute(signal bybit.TradeSignal) error {
	if signal.Action != "PLACE_ORDERS" {
		return nil // Nothing to execute
	}

	// In a real implementation, this would place bid and ask orders
	// based on the calculated optimal spread
	fmt.Printf("Executing market making strategy for %s: %s\n", signal.Symbol, signal.Reason)

	return nil
}

// GetParameters returns the strategy parameters
func (mms *MarketMakingStrategy) GetParameters() map[string]float64 {
	return mms.Parameters
}
