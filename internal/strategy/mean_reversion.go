package strategy

import (
	"fmt"

	"github.com/forbest/bybitgo/internal/bybit"
)

// MeanReversionStrategy implements a mean reversion trading strategy
type MeanReversionStrategy struct {
	Parameters map[string]float64
}

// NewMeanReversionStrategy creates a new MeanReversionStrategy
func NewMeanReversionStrategy() *MeanReversionStrategy {
	return &MeanReversionStrategy{
		Parameters: map[string]float64{
			"bollinger_period": 20,
			"bollinger_std":    2.0,
			"rsi_period":       14,
			"rsi_overbought":   70,
			"rsi_oversold":     30,
		},
	}
}

// GetName returns the strategy name
func (mrs *MeanReversionStrategy) GetName() string {
	return string(MeanReversion)
}

// Analyze implements the mean reversion strategy analysis logic
func (mrs *MeanReversionStrategy) Analyze(marketData *bybit.MarketData) bybit.TradeSignal {
	if marketData == nil || len(marketData.Kline) == 0 {
		return bybit.TradeSignal{
			Symbol: marketData.Symbol,
			Action: "HOLD",
			Reason: "Insufficient market data",
		}
	}

	// Calculate Bollinger Bands
	middleBand, upperBand, lowerBand := mrs.calculateBollingerBands(marketData)

	// Calculate RSI
	rsi := mrs.calculateRSI(marketData)

	// Get current price
	currentPrice, _ := marketData.Kline[len(marketData.Kline)-1].Close.Float64()

	action := "HOLD"
	strength := 0.5
	reason := ""

	// Buy signal: Price below lower band and RSI oversold
	if currentPrice < lowerBand && rsi < mrs.Parameters["rsi_oversold"] {
		action = "BUY"
		// Strength based on how far below the band
		strength = (lowerBand - currentPrice) / lowerBand
		reason = fmt.Sprintf("Mean reversion buy signal: Price %.4f below lower band %.4f, RSI %.2f < %.2f",
			currentPrice, lowerBand, rsi, mrs.Parameters["rsi_oversold"])
	}

	// Sell signal: Price above upper band and RSI overbought
	if currentPrice > upperBand && rsi > mrs.Parameters["rsi_overbought"] {
		action = "SELL"
		// Strength based on how far above the band
		strength = (currentPrice - upperBand) / upperBand
		reason = fmt.Sprintf("Mean reversion sell signal: Price %.4f above upper band %.4f, RSI %.2f > %.2f",
			currentPrice, upperBand, rsi, mrs.Parameters["rsi_overbought"])
	}

	// No clear signal
	if action == "HOLD" {
		reason = fmt.Sprintf("Neutral conditions: Price %.4f, Middle Band %.4f, RSI %.2f",
			currentPrice, middleBand, rsi)
	}

	return bybit.TradeSignal{
		Symbol:   marketData.Symbol,
		Action:   action,
		Strength: strength,
		Reason:   reason,
	}
}

// Execute places mean reversion trades
func (mrs *MeanReversionStrategy) Execute(signal bybit.TradeSignal) error {
	if signal.Action == "HOLD" {
		return nil // Nothing to execute
	}

	// In a real implementation, this would place actual buy/sell orders
	fmt.Printf("Executing mean reversion strategy for %s: %s (%s)\n", signal.Symbol, signal.Action, signal.Reason)

	return nil
}

// GetParameters returns the strategy parameters
func (mrs *MeanReversionStrategy) GetParameters() map[string]float64 {
	return mrs.Parameters
}

// calculateBollingerBands calculates Bollinger Bands
func (mrs *MeanReversionStrategy) calculateBollingerBands(marketData *bybit.MarketData) (float64, float64, float64) {
	if len(marketData.Kline) < int(mrs.Parameters["bollinger_period"]) {
		return 0, 0, 0 // Not enough data
	}

	period := int(mrs.Parameters["bollinger_period"])
	stdDevMultiplier := mrs.Parameters["bollinger_std"]

	// Calculate simple moving average
	sum := 0.0
	prices := make([]float64, 0, period)

	for i := len(marketData.Kline) - period; i < len(marketData.Kline); i++ {
		price, _ := marketData.Kline[i].Close.Float64()
		sum += price
		prices = append(prices, price)
	}

	middleBand := sum / float64(period)

	// Calculate standard deviation
	varianceSum := 0.0
	for _, price := range prices {
		diff := price - middleBand
		varianceSum += diff * diff
	}

	stdDev := varianceSum / float64(period)
	if stdDev > 0 {
		stdDev = varianceSum / float64(period)
	}

	upperBand := middleBand + (stdDevMultiplier * stdDev)
	lowerBand := middleBand - (stdDevMultiplier * stdDev)

	return middleBand, upperBand, lowerBand
}

// calculateRSI calculates the Relative Strength Index (same as momentum strategy)
func (mrs *MeanReversionStrategy) calculateRSI(marketData *bybit.MarketData) float64 {
	if len(marketData.Kline) < int(mrs.Parameters["rsi_period"]) {
		return 50 // Neutral value when insufficient data
	}

	period := int(mrs.Parameters["rsi_period"])
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
