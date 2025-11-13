package strategy

import (
	"fmt"

	"github.com/forbest/bybitgo/internal/bybit"
)

// MomentumStrategy implements a momentum-based trading strategy
type MomentumStrategy struct {
	Parameters map[string]float64
}

// NewMomentumStrategy creates a new MomentumStrategy
func NewMomentumStrategy() *MomentumStrategy {
	return &MomentumStrategy{
		Parameters: map[string]float64{
			"rsi_period":     14,
			"rsi_overbought": 70,
			"rsi_oversold":   30,
			"macd_fast":      12,
			"macd_slow":      26,
			"macd_signal":    9,
		},
	}
}

// GetName returns the strategy name
func (ms *MomentumStrategy) GetName() string {
	return string(Momentum)
}

// Analyze implements the momentum strategy analysis logic
func (ms *MomentumStrategy) Analyze(marketData *bybit.MarketData) bybit.TradeSignal {
	if marketData == nil || len(marketData.Kline) == 0 {
		return bybit.TradeSignal{
			Symbol: marketData.Symbol,
			Action: "HOLD",
			Reason: "Insufficient market data",
		}
	}

	// Calculate RSI (simplified)
	rsi := ms.calculateRSI(marketData)

	// Calculate MACD (simplified)
	macd, signal := ms.calculateMACD(marketData)

	action := "HOLD"
	strength := 0.5
	reason := ""

	// Buy signals
	if rsi < ms.Parameters["rsi_oversold"] && macd > signal {
		action = "BUY"
		strength = (ms.Parameters["rsi_oversold"] - rsi) / ms.Parameters["rsi_oversold"]
		reason = fmt.Sprintf("Oversold conditions: RSI %.2f < %.2f and MACD %.4f > Signal %.4f",
			rsi, ms.Parameters["rsi_oversold"], macd, signal)
	}

	// Sell signals
	if rsi > ms.Parameters["rsi_overbought"] && macd < signal {
		action = "SELL"
		strength = (rsi - ms.Parameters["rsi_overbought"]) / (100 - ms.Parameters["rsi_overbought"])
		reason = fmt.Sprintf("Overbought conditions: RSI %.2f > %.2f and MACD %.4f < Signal %.4f",
			rsi, ms.Parameters["rsi_overbought"], macd, signal)
	}

	// No clear signal
	if action == "HOLD" {
		reason = fmt.Sprintf("Neutral conditions: RSI %.2f, MACD %.4f, Signal %.4f", rsi, macd, signal)
	}

	return bybit.TradeSignal{
		Symbol:   marketData.Symbol,
		Action:   action,
		Strength: strength,
		Reason:   reason,
	}
}

// Execute places momentum-based trades
func (ms *MomentumStrategy) Execute(signal bybit.TradeSignal) error {
	if signal.Action == "HOLD" {
		return nil // Nothing to execute
	}

	// In a real implementation, this would place actual buy/sell orders
	fmt.Printf("Executing momentum strategy for %s: %s (%s)\n", signal.Symbol, signal.Action, signal.Reason)

	return nil
}

// GetParameters returns the strategy parameters
func (ms *MomentumStrategy) GetParameters() map[string]float64 {
	return ms.Parameters
}

// calculateRSI calculates the Relative Strength Index (simplified)
func (ms *MomentumStrategy) calculateRSI(marketData *bybit.MarketData) float64 {
	if len(marketData.Kline) < int(ms.Parameters["rsi_period"]) {
		return 50 // Neutral value when insufficient data
	}

	period := int(ms.Parameters["rsi_period"])
	gains := 0.0
	losses := 0.0

	// Calculate average gains and losses
	for i := len(marketData.Kline) - period; i < len(marketData.Kline)-1; i++ {
		currentClose, _ := marketData.Kline[i].Close.Float64()
		previousClose, _ := marketData.Kline[i-1].Close.Float64()

		change := currentClose - previousClose
		if change > 0 {
			gains += change
		} else {
			losses -= change
		}
	}

	if gains+losses == 0 {
		return 50 // Neutral value
	}

	rs := gains / losses
	rsi := 100 - (100 / (1 + rs))

	return rsi
}

// calculateMACD calculates the MACD indicator (simplified)
func (ms *MomentumStrategy) calculateMACD(marketData *bybit.MarketData) (float64, float64) {
	if len(marketData.Kline) < int(ms.Parameters["macd_slow"]) {
		return 0, 0 // Not enough data
	}

	// Simplified EMA calculation
	fastPeriod := int(ms.Parameters["macd_fast"])
	slowPeriod := int(ms.Parameters["macd_slow"])

	// Calculate fast EMA
	fastEMA := ms.calculateEMA(marketData, fastPeriod)

	// Calculate slow EMA
	slowEMA := ms.calculateEMA(marketData, slowPeriod)

	// MACD line
	macd := fastEMA - slowEMA

	// Signal line (EMA of MACD)
	// Simplified - in practice would need historical MACD values
	signal := macd * 0.9 // Approximation

	return macd, signal
}

// calculateEMA calculates Exponential Moving Average (simplified)
func (ms *MomentumStrategy) calculateEMA(marketData *bybit.MarketData, period int) float64 {
	if len(marketData.Kline) < period {
		return 0
	}

	// Simple moving average for first value
	sum := 0.0
	for i := len(marketData.Kline) - period; i < len(marketData.Kline); i++ {
		close, _ := marketData.Kline[i].Close.Float64()
		sum += close
	}

	sma := sum / float64(period)

	// Simplified EMA calculation
	// In practice, would use proper EMA formula with smoothing factor
	return sma
}
