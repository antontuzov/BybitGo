package strategy

import (
	"fmt"

	"github.com/forbest/bybitgo/internal/bybit"
)

// VolatilityBreakoutStrategy implements a volatility breakout trading strategy
type VolatilityBreakoutStrategy struct {
	Parameters map[string]float64
}

// NewVolatilityBreakoutStrategy creates a new VolatilityBreakoutStrategy
func NewVolatilityBreakoutStrategy() *VolatilityBreakoutStrategy {
	return &VolatilityBreakoutStrategy{
		Parameters: map[string]float64{
			"period":           20,
			"multiplier":       2.0,
			"min_volume_ratio": 1.5, // Minimum volume increase for breakout confirmation
		},
	}
}

// GetName returns the strategy name
func (vbs *VolatilityBreakoutStrategy) GetName() string {
	return string(VolatilityBreakout)
}

// Analyze implements the volatility breakout strategy analysis logic
func (vbs *VolatilityBreakoutStrategy) Analyze(marketData *bybit.MarketData) bybit.TradeSignal {
	if marketData == nil || len(marketData.Kline) == 0 {
		return bybit.TradeSignal{
			Symbol: marketData.Symbol,
			Action: "HOLD",
			Reason: "Insufficient market data",
		}
	}

	// Calculate volatility channel
	upperChannel, lowerChannel := vbs.calculateVolatilityChannel(marketData)

	// Get current and previous prices
	currentKline := marketData.Kline[len(marketData.Kline)-1]
	previousKline := marketData.Kline[len(marketData.Kline)-2]

	currentClose, _ := currentKline.Close.Float64()
	previousClose, _ := previousKline.Close.Float64()
	currentVolume, _ := currentKline.Volume.Float64()
	averageVolume := vbs.calculateAverageVolume(marketData)

	action := "HOLD"
	strength := 0.5
	reason := ""

	// Buy breakout: Price breaks above upper channel with increased volume
	if currentClose > upperChannel && previousClose <= upperChannel && currentVolume > averageVolume*vbs.Parameters["min_volume_ratio"] {
		action = "BUY"
		strength = (currentClose - upperChannel) / upperChannel
		reason = fmt.Sprintf("Buy breakout: Price %.4f broke above channel %.4f with volume %.2f > avg %.2f",
			currentClose, upperChannel, currentVolume, averageVolume)
	}

	// Sell breakout: Price breaks below lower channel with increased volume
	if currentClose < lowerChannel && previousClose >= lowerChannel && currentVolume > averageVolume*vbs.Parameters["min_volume_ratio"] {
		action = "SELL"
		strength = (lowerChannel - currentClose) / lowerChannel
		reason = fmt.Sprintf("Sell breakout: Price %.4f broke below channel %.4f with volume %.2f > avg %.2f",
			currentClose, lowerChannel, currentVolume, averageVolume)
	}

	// No clear signal
	if action == "HOLD" {
		reason = fmt.Sprintf("No breakout: Price %.4f, Channel range [%.4f - %.4f], Volume %.2f vs avg %.2f",
			currentClose, lowerChannel, upperChannel, currentVolume, averageVolume)
	}

	return bybit.TradeSignal{
		Symbol:   marketData.Symbol,
		Action:   action,
		Strength: strength,
		Reason:   reason,
	}
}

// Execute places volatility breakout trades
func (vbs *VolatilityBreakoutStrategy) Execute(signal bybit.TradeSignal) error {
	if signal.Action == "HOLD" {
		return nil // Nothing to execute
	}

	// In a real implementation, this would place actual buy/sell orders
	fmt.Printf("Executing volatility breakout strategy for %s: %s (%s)\n", signal.Symbol, signal.Action, signal.Reason)

	return nil
}

// GetParameters returns the strategy parameters
func (vbs *VolatilityBreakoutStrategy) GetParameters() map[string]float64 {
	return vbs.Parameters
}

// calculateVolatilityChannel calculates the volatility channel (Donchian channels)
func (vbs *VolatilityBreakoutStrategy) calculateVolatilityChannel(marketData *bybit.MarketData) (float64, float64) {
	if len(marketData.Kline) < int(vbs.Parameters["period"]) {
		return 0, 0 // Not enough data
	}

	period := int(vbs.Parameters["period"])

	highestHigh, _ := marketData.Kline[len(marketData.Kline)-period].High.Float64()
	lowestLow, _ := marketData.Kline[len(marketData.Kline)-period].Low.Float64()

	// Find highest high and lowest low over the period
	for i := len(marketData.Kline) - period; i < len(marketData.Kline); i++ {
		high, _ := marketData.Kline[i].High.Float64()
		low, _ := marketData.Kline[i].Low.Float64()

		if high > highestHigh {
			highestHigh = high
		}

		if low < lowestLow {
			lowestLow = low
		}
	}

	// Apply multiplier for channel expansion
	rangeSize := highestHigh - lowestLow
	expandedRange := rangeSize * vbs.Parameters["multiplier"]

	upperChannel := highestHigh + expandedRange/2
	lowerChannel := lowestLow - expandedRange/2

	return upperChannel, lowerChannel
}

// calculateAverageVolume calculates average volume over the period
func (vbs *VolatilityBreakoutStrategy) calculateAverageVolume(marketData *bybit.MarketData) float64 {
	if len(marketData.Kline) < int(vbs.Parameters["period"]) {
		return 0 // Not enough data
	}

	period := int(vbs.Parameters["period"])
	sum := 0.0

	for i := len(marketData.Kline) - period; i < len(marketData.Kline); i++ {
		volume, _ := marketData.Kline[i].Volume.Float64()
		sum += volume
	}

	return sum / float64(period)
}
