package market

import (
	"context"
	"fmt"
	"math"
	"sort"

	"github.com/forbest/bybitgo/internal/bybit"
)

// MarketAnalyzer analyzes market conditions for strategy selection
type MarketAnalyzer struct {
	VolatilityTracker map[string]*VolatilityData
	TrendIndicator    map[string]*TrendData
	VolumeAnalysis    map[string]*VolumeProfile
	CorrelationMatrix map[string]map[string]float64
	PriceHistory      map[string][]float64 // Store price history for correlation calculation
}

// VolatilityData tracks volatility for a symbol
type VolatilityData struct {
	Symbol             string
	RecentVolatility   float64
	LongTermVolatility float64
	VolatilityRegime   string // "high", "medium", "low"
}

// TrendData tracks trend information for a symbol
type TrendData struct {
	Symbol         string
	TrendStrength  float64 // 0-1 scale
	TrendDirection string  // "up", "down", "sideways"
	ADX            float64
}

// VolumeProfile tracks volume characteristics
type VolumeProfile struct {
	Symbol        string
	CurrentVolume float64
	AverageVolume float64
	VolumeRatio   float64 // Current vs average
	VolumeTrend   string  // "increasing", "decreasing", "stable"
}

// MarketRegime represents the current market condition
type MarketRegime struct {
	Volatility string // "high_volatility", "low_volatility"
	Trend      string // "trending_up", "trending_down", "ranging"
	Volume     string // "high_volume", "low_volume"
}

// MACDResult represents MACD indicator results
type MACDResult struct {
	MACDLine   float64
	SignalLine float64
	Histogram  float64
}

// StochasticRSIResult represents Stochastic RSI indicator results
type StochasticRSIResult struct {
	K float64 // %K line
	D float64 // %D line (SMA of %K)
}

// VWAPResult represents Volume Weighted Average Price results
type VWAPResult struct {
	Value     float64
	UpperBand float64
	LowerBand float64
	Bandwidth float64
}

// IndicatorCombination represents a combination of multiple indicators
type IndicatorCombination struct {
	Name        string
	Indicators  []string
	Weights     []float64
	Threshold   float64
	Description string
}

// CombinedSignal represents a signal generated from multiple indicators
type CombinedSignal struct {
	Symbol     string
	Score      float64
	Confidence float64
	Components map[string]float64
	Signal     string // "BUY", "SELL", "HOLD"
	Reason     string
}

// VolumeWeightedSignal represents a signal that incorporates volume analysis
type VolumeWeightedSignal struct {
	Symbol            string
	BaseSignal        string  // "BUY", "SELL", "HOLD"
	VolumeConfidence  float64 // 0-1 scale based on volume analysis
	PriceConfidence   float64 // 0-1 scale based on price analysis
	OverallConfidence float64 // Combined confidence
	Reason            string
}

// NewMarketAnalyzer creates a new MarketAnalyzer
func NewMarketAnalyzer() *MarketAnalyzer {
	return &MarketAnalyzer{
		VolatilityTracker: make(map[string]*VolatilityData),
		TrendIndicator:    make(map[string]*TrendData),
		VolumeAnalysis:    make(map[string]*VolumeProfile),
		CorrelationMatrix: make(map[string]map[string]float64),
		PriceHistory:      make(map[string][]float64),
	}
}

// AnalyzeMarketConditions analyzes market data and updates internal trackers
func (ma *MarketAnalyzer) AnalyzeMarketConditions(ctx context.Context, symbol string, data *bybit.MarketData) (*MarketRegime, error) {
	// Calculate volatility
	volatility := ma.calculateVolatility(data)

	// Calculate trend
	trend := ma.calculateTrend(data)

	// Calculate volume profile
	volume := ma.calculateVolumeProfile(data)

	// Update price history for correlation analysis
	ma.updatePriceHistory(symbol, data)

	// Update trackers
	ma.VolatilityTracker[symbol] = volatility
	ma.TrendIndicator[symbol] = trend
	ma.VolumeAnalysis[symbol] = volume

	// Determine market regime
	regime := &MarketRegime{
		Volatility: ma.determineVolatilityRegime(volatility),
		Trend:      ma.determineTrendRegime(trend),
		Volume:     ma.determineVolumeRegime(volume),
	}

	return regime, nil
}

// updatePriceHistory updates the price history for a symbol
func (ma *MarketAnalyzer) updatePriceHistory(symbol string, data *bybit.MarketData) {
	var prices []float64
	for _, kline := range data.Kline {
		close, _ := kline.Close.Float64()
		prices = append(prices, close)
	}

	// Keep only the last 100 prices
	if len(prices) > 100 {
		prices = prices[len(prices)-100:]
	}

	ma.PriceHistory[symbol] = prices
}

// calculateVolatility calculates volatility metrics for a symbol
func (ma *MarketAnalyzer) calculateVolatility(data *bybit.MarketData) *VolatilityData {
	// Simplified volatility calculation based on price range
	// In practice, you would use more sophisticated methods like GARCH models

	var prices []float64
	for _, kline := range data.Kline {
		high, _ := kline.High.Float64()
		low, _ := kline.Low.Float64()
		prices = append(prices, (high+low)/2)
	}

	if len(prices) < 2 {
		return &VolatilityData{
			Symbol:             data.Symbol,
			RecentVolatility:   0,
			LongTermVolatility: 0,
			VolatilityRegime:   "low",
		}
	}

	// Calculate recent volatility (last 10 periods)
	recentVol := ma.simpleVolatility(prices[len(prices)-10:])

	// Calculate long-term volatility (entire series)
	longVol := ma.simpleVolatility(prices)

	// Determine regime based on comparison
	regime := "medium"
	if recentVol > longVol*1.2 {
		regime = "high"
	} else if recentVol < longVol*0.8 {
		regime = "low"
	}

	return &VolatilityData{
		Symbol:             data.Symbol,
		RecentVolatility:   recentVol,
		LongTermVolatility: longVol,
		VolatilityRegime:   regime,
	}
}

// simpleVolatility calculates a simple volatility measure
func (ma *MarketAnalyzer) simpleVolatility(prices []float64) float64 {
	if len(prices) < 2 {
		return 0
	}

	// Calculate percentage changes
	sum := 0.0
	count := 0
	for i := 1; i < len(prices); i++ {
		if prices[i-1] != 0 {
			change := math.Abs((prices[i] - prices[i-1]) / prices[i-1])
			sum += change
			count++
		}
	}

	if count == 0 {
		return 0
	}

	return sum / float64(count)
}

// calculateTrend calculates trend metrics for a symbol
func (ma *MarketAnalyzer) calculateTrend(data *bybit.MarketData) *TrendData {
	// Simplified trend calculation
	// In practice, you would use indicators like ADX, MACD, etc.

	var prices []float64
	for _, kline := range data.Kline {
		close, _ := kline.Close.Float64()
		prices = append(prices, close)
	}

	if len(prices) < 2 {
		return &TrendData{
			Symbol:         data.Symbol,
			TrendStrength:  0,
			TrendDirection: "sideways",
			ADX:            0,
		}
	}

	// Simple linear regression slope as trend indicator
	slope := ma.linearRegressionSlope(prices)

	direction := "sideways"
	strength := math.Abs(slope)

	if slope > 0.001 {
		direction = "up"
	} else if slope < -0.001 {
		direction = "down"
	}

	// Normalize strength to 0-1 scale (simplified)
	if strength > 0.05 {
		strength = 0.05
	}
	strength = strength / 0.05

	return &TrendData{
		Symbol:         data.Symbol,
		TrendStrength:  strength,
		TrendDirection: direction,
		ADX:            0, // Would calculate actual ADX in production
	}
}

// linearRegressionSlope calculates the slope of a linear regression
func (ma *MarketAnalyzer) linearRegressionSlope(values []float64) float64 {
	n := float64(len(values))

	if n < 2 {
		return 0
	}

	var sumX, sumY, sumXY, sumXX float64

	for i, value := range values {
		x := float64(i)
		sumX += x
		sumY += value
		sumXY += x * value
		sumXX += x * x
	}

	numerator := n*sumXY - sumX*sumY
	denominator := n*sumXX - sumX*sumX

	if denominator == 0 {
		return 0
	}

	return numerator / denominator
}

// calculateVolumeProfile calculates volume metrics for a symbol
func (ma *MarketAnalyzer) calculateVolumeProfile(data *bybit.MarketData) *VolumeProfile {
	// Simplified volume analysis
	var volumes []float64
	totalVolume := 0.0

	for _, kline := range data.Kline {
		volume, _ := kline.Volume.Float64()
		volumes = append(volumes, volume)
		totalVolume += volume
	}

	if len(volumes) == 0 {
		return &VolumeProfile{
			Symbol:        data.Symbol,
			CurrentVolume: 0,
			AverageVolume: 0,
			VolumeRatio:   0,
			VolumeTrend:   "stable",
		}
	}

	currentVolume := volumes[len(volumes)-1]
	averageVolume := totalVolume / float64(len(volumes))
	ratio := 1.0

	if averageVolume > 0 {
		ratio = currentVolume / averageVolume
	}

	trend := "stable"
	if ratio > 1.2 {
		trend = "increasing"
	} else if ratio < 0.8 {
		trend = "decreasing"
	}

	return &VolumeProfile{
		Symbol:        data.Symbol,
		CurrentVolume: currentVolume,
		AverageVolume: averageVolume,
		VolumeRatio:   ratio,
		VolumeTrend:   trend,
	}
}

// determineVolatilityRegime determines the volatility regime
func (ma *MarketAnalyzer) determineVolatilityRegime(volData *VolatilityData) string {
	return volData.VolatilityRegime + "_volatility"
}

// determineTrendRegime determines the trend regime
func (ma *MarketAnalyzer) determineTrendRegime(trendData *TrendData) string {
	switch trendData.TrendDirection {
	case "up":
		return "trending_up"
	case "down":
		return "trending_down"
	default:
		return "ranging"
	}
}

// determineVolumeRegime determines the volume regime
func (ma *MarketAnalyzer) determineVolumeRegime(volProfile *VolumeProfile) string {
	switch volProfile.VolumeTrend {
	case "increasing":
		return "high_volume"
	case "decreasing":
		return "low_volume"
	default:
		return "normal_volume"
	}
}

// GetMarketRegime returns the current market regime for a symbol
func (ma *MarketAnalyzer) GetMarketRegime(symbol string) *MarketRegime {
	volData, volExists := ma.VolatilityTracker[symbol]
	trendData, trendExists := ma.TrendIndicator[symbol]
	volProfile, volProfileExists := ma.VolumeAnalysis[symbol]

	if !volExists || !trendExists || !volProfileExists {
		return &MarketRegime{
			Volatility: "unknown",
			Trend:      "unknown",
			Volume:     "unknown",
		}
	}

	return &MarketRegime{
		Volatility: ma.determineVolatilityRegime(volData),
		Trend:      ma.determineTrendRegime(trendData),
		Volume:     ma.determineVolumeRegime(volProfile),
	}
}

// CalculateCorrelations calculates correlation matrix for all symbols
func (ma *MarketAnalyzer) CalculateCorrelations() map[string]map[string]float64 {
	// Initialize correlation matrix
	ma.CorrelationMatrix = make(map[string]map[string]float64)

	// Get all symbols
	symbols := make([]string, 0, len(ma.PriceHistory))
	for symbol := range ma.PriceHistory {
		symbols = append(symbols, symbol)
	}

	// Calculate correlations between all pairs
	for i, symbol1 := range symbols {
		if ma.CorrelationMatrix[symbol1] == nil {
			ma.CorrelationMatrix[symbol1] = make(map[string]float64)
		}

		for j, symbol2 := range symbols {
			if i == j {
				ma.CorrelationMatrix[symbol1][symbol2] = 1.0 // Perfect correlation with itself
			} else {
				corr := ma.calculateCorrelation(symbol1, symbol2)
				ma.CorrelationMatrix[symbol1][symbol2] = corr

				// Ensure symmetry
				if ma.CorrelationMatrix[symbol2] == nil {
					ma.CorrelationMatrix[symbol2] = make(map[string]float64)
				}
				ma.CorrelationMatrix[symbol2][symbol1] = corr
			}
		}
	}

	return ma.CorrelationMatrix
}

// calculateCorrelation calculates the correlation between two symbols
func (ma *MarketAnalyzer) calculateCorrelation(symbol1, symbol2 string) float64 {
	prices1, ok1 := ma.PriceHistory[symbol1]
	prices2, ok2 := ma.PriceHistory[symbol2]

	// If either symbol doesn't have price history, return 0
	if !ok1 || !ok2 {
		return 0.0
	}

	// Use the minimum length to ensure we're comparing the same time periods
	minLen := len(prices1)
	if len(prices2) < minLen {
		minLen = len(prices2)
	}

	if minLen < 2 {
		return 0.0
	}

	// Trim to the same length
	prices1 = prices1[len(prices1)-minLen:]
	prices2 = prices2[len(prices2)-minLen:]

	// Calculate correlation using Pearson correlation coefficient
	return ma.pearsonCorrelation(prices1, prices2)
}

// pearsonCorrelation calculates the Pearson correlation coefficient
func (ma *MarketAnalyzer) pearsonCorrelation(x, y []float64) float64 {
	n := len(x)
	if n != len(y) || n < 2 {
		return 0.0
	}

	// Calculate means
	sumX, sumY := 0.0, 0.0
	for i := 0; i < n; i++ {
		sumX += x[i]
		sumY += y[i]
	}
	meanX := sumX / float64(n)
	meanY := sumY / float64(n)

	// Calculate numerator and denominators
	numerator := 0.0
	denomX, denomY := 0.0, 0.0

	for i := 0; i < n; i++ {
		diffX := x[i] - meanX
		diffY := y[i] - meanY
		numerator += diffX * diffY
		denomX += diffX * diffX
		denomY += diffY * diffY
	}

	if denomX == 0 || denomY == 0 {
		return 0.0
	}

	return numerator / math.Sqrt(denomX*denomY)
}

// GetHighlyCorrelatedAssets returns assets that are highly correlated with a given symbol
func (ma *MarketAnalyzer) GetHighlyCorrelatedAssets(symbol string, threshold float64) []string {
	correlations, exists := ma.CorrelationMatrix[symbol]
	if !exists {
		return []string{}
	}

	var highlyCorrelated []string
	for otherSymbol, correlation := range correlations {
		if otherSymbol != symbol && math.Abs(correlation) >= threshold {
			highlyCorrelated = append(highlyCorrelated, otherSymbol)
		}
	}

	// Sort by correlation strength (highest first)
	sort.Slice(highlyCorrelated, func(i, j int) bool {
		corrI := math.Abs(correlations[highlyCorrelated[i]])
		corrJ := math.Abs(correlations[highlyCorrelated[j]])
		return corrI > corrJ
	})

	return highlyCorrelated
}

// GetDiversificationScore calculates a diversification score for a portfolio
func (ma *MarketAnalyzer) GetDiversificationScore(symbols []string) float64 {
	if len(symbols) <= 1 {
		return 1.0 // Perfectly diversified (or not applicable)
	}

	// Calculate average correlation between all pairs
	totalCorrelation := 0.0
	count := 0

	for i := 0; i < len(symbols); i++ {
		for j := i + 1; j < len(symbols); j++ {
			symbol1 := symbols[i]
			symbol2 := symbols[j]

			correlation := 0.0
			if ma.CorrelationMatrix[symbol1] != nil {
				if corr, exists := ma.CorrelationMatrix[symbol1][symbol2]; exists {
					correlation = corr
				}
			}

			totalCorrelation += math.Abs(correlation)
			count++
		}
	}

	if count == 0 {
		return 1.0
	}

	averageCorrelation := totalCorrelation / float64(count)

	// Convert to diversification score (lower average correlation = higher diversification)
	// Score ranges from 0 (no diversification) to 1 (perfect diversification)
	return 1.0 - averageCorrelation
}

// calculateMACD calculates MACD indicator for a symbol
func (ma *MarketAnalyzer) calculateMACD(data *bybit.MarketData) *MACDResult {
	// Get closing prices
	var closes []float64
	for _, kline := range data.Kline {
		close, _ := kline.Close.Float64()
		closes = append(closes, close)
	}

	if len(closes) < 26 { // Need at least 26 periods for MACD
		return &MACDResult{0, 0, 0}
	}

	// Calculate 12-period EMA
	ema12 := ma.calculateEMA(closes, 12)

	// Calculate 26-period EMA
	ema26 := ma.calculateEMA(closes, 26)

	// MACD line is the difference between the two EMAs
	macdLine := ema12 - ema26

	// Calculate 9-period EMA of MACD line (signal line)
	// For simplicity, we'll use the last 9 MACD values
	macdValues := make([]float64, 9)
	for i := 0; i < 9; i++ {
		macdValues[i] = macdLine // Simplified - in practice would calculate historical MACD values
	}
	signalLine := ma.calculateEMA(macdValues, 9)

	// Histogram is the difference between MACD line and signal line
	histogram := macdLine - signalLine

	return &MACDResult{
		MACDLine:   macdLine,
		SignalLine: signalLine,
		Histogram:  histogram,
	}
}

// calculateEMA calculates Exponential Moving Average
func (ma *MarketAnalyzer) calculateEMA(prices []float64, period int) float64 {
	if len(prices) < period {
		return 0
	}

	// Calculate simple moving average for the first value
	sma := 0.0
	for i := 0; i < period; i++ {
		sma += prices[len(prices)-period+i]
	}
	sma /= float64(period)

	// Calculate multiplier
	multiplier := 2.0 / float64(period+1)

	// Calculate EMA
	ema := sma
	for i := len(prices) - period + 1; i < len(prices); i++ {
		ema = (prices[i]-ema)*multiplier + ema
	}

	return ema
}

// calculateStochasticRSI calculates Stochastic RSI indicator
func (ma *MarketAnalyzer) calculateStochasticRSI(data *bybit.MarketData) *StochasticRSIResult {
	// Get closing prices
	var closes []float64
	for _, kline := range data.Kline {
		close, _ := kline.Close.Float64()
		closes = append(closes, close)
	}

	if len(closes) < 14 { // Need at least 14 periods
		return &StochasticRSIResult{0, 0}
	}

	// Calculate RSI first
	rsi := ma.calculateRSI(closes, 14)

	// For Stochastic RSI, we need the highest and lowest RSI values over a period
	// This is a simplified implementation
	k := 0.0
	if rsi > 0 {
		k = (rsi - 0) / (100 - 0) * 100 // Normalize to 0-100
	}

	// Calculate %D as 3-period SMA of %K
	d := k // Simplified

	return &StochasticRSIResult{
		K: k,
		D: d,
	}
}

// calculateRSI calculates Relative Strength Index
func (ma *MarketAnalyzer) calculateRSI(prices []float64, period int) float64 {
	if len(prices) < period+1 {
		return 0
	}

	// Calculate price changes
	gains := 0.0
	losses := 0.0

	for i := len(prices) - period; i < len(prices); i++ {
		if i > 0 {
			change := prices[i] - prices[i-1]
			if change > 0 {
				gains += change
			} else {
				losses -= change
			}
		}
	}

	if losses == 0 {
		return 100
	}

	rs := gains / losses
	rsi := 100 - (100 / (1 + rs))

	return rsi
}

// calculateVWAP calculates Volume Weighted Average Price
func (ma *MarketAnalyzer) calculateVWAP(data *bybit.MarketData) *VWAPResult {
	var totalPriceVolume float64
	var totalVolume float64

	for _, kline := range data.Kline {
		high, _ := kline.High.Float64()
		low, _ := kline.Low.Float64()
		close, _ := kline.Close.Float64()
		volume, _ := kline.Volume.Float64()

		// Typical price = (High + Low + Close) / 3
		typicalPrice := (high + low + close) / 3
		totalPriceVolume += typicalPrice * volume
		totalVolume += volume
	}

	if totalVolume == 0 {
		return &VWAPResult{0, 0, 0, 0}
	}

	vwap := totalPriceVolume / totalVolume

	// Calculate standard deviation for bands
	variance := 0.0
	count := 0
	for _, kline := range data.Kline {
		high, _ := kline.High.Float64()
		low, _ := kline.Low.Float64()
		close, _ := kline.Close.Float64()
		volume, _ := kline.Volume.Float64()

		typicalPrice := (high + low + close) / 3
		variance += math.Pow(typicalPrice-vwap, 2) * volume
		count += int(volume)
	}

	if count == 0 {
		return &VWAPResult{vwap, vwap, vwap, 0}
	}

	stdDev := math.Sqrt(variance / float64(count))
	upperBand := vwap + (2 * stdDev)
	lowerBand := vwap - (2 * stdDev)
	bandwidth := (upperBand - lowerBand) / vwap

	return &VWAPResult{
		Value:     vwap,
		UpperBand: upperBand,
		LowerBand: lowerBand,
		Bandwidth: bandwidth,
	}
}

// EnhancedMarketData represents enhanced market data with additional indicators
type EnhancedMarketData struct {
	Symbol        string
	BaseData      *bybit.MarketData
	MACD          *MACDResult
	StochasticRSI *StochasticRSIResult
	VWAP          *VWAPResult
}

// AnalyzeEnhancedMarketConditions analyzes market data with additional indicators
func (ma *MarketAnalyzer) AnalyzeEnhancedMarketConditions(ctx context.Context, symbol string, data *bybit.MarketData) (*EnhancedMarketData, error) {
	// Calculate additional indicators
	macd := ma.calculateMACD(data)
	stochasticRSI := ma.calculateStochasticRSI(data)
	vwap := ma.calculateVWAP(data)

	// Analyze base market conditions
	_, err := ma.AnalyzeMarketConditions(ctx, symbol, data)
	if err != nil {
		return nil, err
	}

	enhancedData := &EnhancedMarketData{
		Symbol:        symbol,
		BaseData:      data,
		MACD:          macd,
		StochasticRSI: stochasticRSI,
		VWAP:          vwap,
	}

	return enhancedData, nil
}

// CalculateCombinedSignal calculates a combined signal from multiple indicators
func (ma *MarketAnalyzer) CalculateCombinedSignal(symbol string, enhancedData *EnhancedMarketData) *CombinedSignal {
	// Initialize components map
	components := make(map[string]float64)

	// Calculate individual indicator scores (normalized to 0-1 scale)
	var macdScore, rsiScore, vwapScore float64

	// MACD score (positive when MACD line > Signal line)
	if enhancedData.MACD != nil {
		if enhancedData.MACD.SignalLine != 0 {
			macdScore = (enhancedData.MACD.MACDLine - enhancedData.MACD.SignalLine) / math.Abs(enhancedData.MACD.SignalLine)
			// Normalize to 0-1 range
			macdScore = (macdScore + 1) / 2
			if macdScore < 0 {
				macdScore = 0
			} else if macdScore > 1 {
				macdScore = 1
			}
		}
		components["MACD"] = macdScore
	}

	// Stochastic RSI score (based on K and D lines)
	if enhancedData.StochasticRSI != nil {
		// Normalize Stochastic RSI to 0-1 range (already 0-100, so divide by 100)
		rsiScore = enhancedData.StochasticRSI.K / 100.0
		components["StochasticRSI"] = rsiScore
	}

	// VWAP score (price position relative to VWAP bands)
	if enhancedData.VWAP != nil && enhancedData.VWAP.UpperBand != enhancedData.VWAP.LowerBand {
		// Get current price from base data
		var currentPrice float64
		if len(enhancedData.BaseData.Kline) > 0 {
			currentPrice, _ = enhancedData.BaseData.Kline[len(enhancedData.BaseData.Kline)-1].Close.Float64()
		}

		// Position between bands (0 = lower band, 1 = upper band)
		if enhancedData.VWAP.UpperBand != enhancedData.VWAP.LowerBand {
			vwapScore = (currentPrice - enhancedData.VWAP.LowerBand) / (enhancedData.VWAP.UpperBand - enhancedData.VWAP.LowerBand)
			// Clamp to 0-1 range
			if vwapScore < 0 {
				vwapScore = 0
			} else if vwapScore > 1 {
				vwapScore = 1
			}
		}
		components["VWAP"] = vwapScore
	}

	// Calculate weighted average score
	// Equal weights for now (0.33 each)
	totalWeight := 0.33 + 0.33 + 0.33
	weightedScore := (macdScore*0.33 + rsiScore*0.33 + vwapScore*0.33) / totalWeight

	// Calculate confidence based on agreement between indicators
	agreement := 0.0
	if macdScore > 0.5 && rsiScore > 0.5 && vwapScore > 0.5 {
		agreement = 1.0 // Strong buy agreement
	} else if macdScore < 0.5 && rsiScore < 0.5 && vwapScore < 0.5 {
		agreement = -1.0 // Strong sell agreement
	} else {
		// Mixed signals, lower confidence
		agreement = (macdScore + rsiScore + vwapScore - 1.5) / 1.5
	}

	// Determine signal based on score and agreement
	signal := "HOLD"
	reason := "Neutral conditions"
	if weightedScore > 0.6 && agreement > 0.5 {
		signal = "BUY"
		reason = fmt.Sprintf("Strong buy signal: Score %.2f, Agreement %.2f", weightedScore, agreement)
	} else if weightedScore < 0.4 && agreement < -0.5 {
		signal = "SELL"
		reason = fmt.Sprintf("Strong sell signal: Score %.2f, Agreement %.2f", weightedScore, agreement)
	} else if weightedScore > 0.55 {
		signal = "BUY"
		reason = fmt.Sprintf("Moderate buy signal: Score %.2f", weightedScore)
	} else if weightedScore < 0.45 {
		signal = "SELL"
		reason = fmt.Sprintf("Moderate sell signal: Score %.2f", weightedScore)
	}

	// Confidence is based on how close the score is to 0 or 1, and agreement level
	confidence := math.Abs(weightedScore-0.5) * 2 // 0-1 range
	confidence = (confidence + math.Abs(agreement)) / 2

	return &CombinedSignal{
		Symbol:     symbol,
		Score:      weightedScore,
		Confidence: confidence,
		Components: components,
		Signal:     signal,
		Reason:     reason,
	}
}

// AnalyzeVolumeWeightedSignal analyzes market conditions with volume weighting
func (ma *MarketAnalyzer) AnalyzeVolumeWeightedSignal(symbol string, data *bybit.MarketData) *VolumeWeightedSignal {
	// Get the latest price and volume data
	if len(data.Kline) < 2 {
		return &VolumeWeightedSignal{
			Symbol:            symbol,
			BaseSignal:        "HOLD",
			VolumeConfidence:  0,
			PriceConfidence:   0,
			OverallConfidence: 0,
			Reason:            "Insufficient data",
		}
	}

	// Get latest two klines for comparison
	latest := data.Kline[len(data.Kline)-1]
	previous := data.Kline[len(data.Kline)-2]

	// Extract price and volume data
	latestClose, _ := latest.Close.Float64()
	previousClose, _ := previous.Close.Float64()
	latestVolume, _ := latest.Volume.Float64()
	previousVolume, _ := previous.Volume.Float64()

	// Calculate price change
	priceChange := latestClose - previousClose
	priceChangePercent := 0.0
	if previousClose != 0 {
		priceChangePercent = (priceChange / previousClose) * 100
	}

	// Calculate volume change
	volumeChange := latestVolume - previousVolume
	volumeChangePercent := 0.0
	if previousVolume != 0 {
		volumeChangePercent = (volumeChange / previousVolume) * 100
	}

	// Determine base signal based on price action
	baseSignal := "HOLD"
	priceConfidence := 0.0

	if priceChangePercent > 1.0 {
		baseSignal = "BUY"
		priceConfidence = math.Min(priceChangePercent/5.0, 1.0) // Cap at 1.0
	} else if priceChangePercent < -1.0 {
		baseSignal = "SELL"
		priceConfidence = math.Min(math.Abs(priceChangePercent)/5.0, 1.0) // Cap at 1.0
	}

	// Calculate volume confirmation
	volumeConfidence := 0.0
	reason := "Neutral conditions"

	// Volume-weighted analysis
	if baseSignal == "BUY" {
		if volumeChangePercent > 50.0 {
			// Strong volume confirmation for buy signal
			volumeConfidence = 1.0
			reason = fmt.Sprintf("Strong buy: Price up %.2f%% with volume surge %.2f%%", priceChangePercent, volumeChangePercent)
		} else if volumeChangePercent > 0 {
			// Weak volume confirmation for buy signal
			volumeConfidence = 0.5
			reason = fmt.Sprintf("Moderate buy: Price up %.2f%% with volume increase %.2f%%", priceChangePercent, volumeChangePercent)
		} else {
			// No volume confirmation (or negative volume)
			volumeConfidence = 0.2
			reason = fmt.Sprintf("Weak buy: Price up %.2f%% but volume down %.2f%%", priceChangePercent, math.Abs(volumeChangePercent))
		}
	} else if baseSignal == "SELL" {
		if volumeChangePercent > 50.0 {
			// Strong volume confirmation for sell signal
			volumeConfidence = 1.0
			reason = fmt.Sprintf("Strong sell: Price down %.2f%% with volume surge %.2f%%", priceChangePercent, volumeChangePercent)
		} else if volumeChangePercent > 0 {
			// Weak volume confirmation for sell signal
			volumeConfidence = 0.5
			reason = fmt.Sprintf("Moderate sell: Price down %.2f%% with volume increase %.2f%%", priceChangePercent, volumeChangePercent)
		} else {
			// No volume confirmation (or negative volume)
			volumeConfidence = 0.2
			reason = fmt.Sprintf("Weak sell: Price down %.2f%% but volume down %.2f%%", priceChangePercent, math.Abs(volumeChangePercent))
		}
	} else {
		// HOLD signal
		if volumeChangePercent > 100.0 {
			// High volume with no price movement - potential breakout
			volumeConfidence = 0.7
			reason = fmt.Sprintf("Accumulation: High volume (%.2f%%) with no significant price change", volumeChangePercent)
		} else {
			volumeConfidence = 0.3
			reason = fmt.Sprintf("Low activity: Volume change %.2f%%, price change %.2f%%", volumeChangePercent, priceChangePercent)
		}
	}

	// Calculate overall confidence as weighted average
	overallConfidence := (priceConfidence*0.6 + volumeConfidence*0.4)

	// Adjust signal based on confidence
	if overallConfidence < 0.3 {
		baseSignal = "HOLD"
	}

	return &VolumeWeightedSignal{
		Symbol:            symbol,
		BaseSignal:        baseSignal,
		VolumeConfidence:  volumeConfidence,
		PriceConfidence:   priceConfidence,
		OverallConfidence: overallConfidence,
		Reason:            reason,
	}
}

// GetVolumeProfileAnalysis returns detailed volume profile analysis
func (ma *MarketAnalyzer) GetVolumeProfileAnalysis(symbol string, data *bybit.MarketData) *VolumeProfile {
	// This is already implemented in calculateVolumeProfile method
	return ma.calculateVolumeProfile(data)
}

// GetDefaultIndicatorCombinations returns default indicator combinations
func (ma *MarketAnalyzer) GetDefaultIndicatorCombinations() []IndicatorCombination {
	return []IndicatorCombination{
		{
			Name:        "TrendFollowing",
			Indicators:  []string{"MACD", "VWAP"},
			Weights:     []float64{0.6, 0.4},
			Threshold:   0.55,
			Description: "Combination of MACD and VWAP for trend following",
		},
		{
			Name:        "Momentum",
			Indicators:  []string{"StochasticRSI", "MACD"},
			Weights:     []float64{0.5, 0.5},
			Threshold:   0.6,
			Description: "Combination of Stochastic RSI and MACD for momentum trading",
		},
		{
			Name:        "MeanReversion",
			Indicators:  []string{"StochasticRSI", "VWAP"},
			Weights:     []float64{0.4, 0.6},
			Threshold:   0.4,
			Description: "Combination of Stochastic RSI and VWAP for mean reversion",
		},
	}
}
