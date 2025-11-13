package strategy

import (
	"github.com/forbest/bybitgo/internal/market"
)

// StrategyType represents different trading strategies
type StrategyType string

const (
	MarketMaking       StrategyType = "market_making"
	Momentum           StrategyType = "momentum"
	MeanReversion      StrategyType = "mean_reversion"
	VolatilityBreakout StrategyType = "volatility_breakout"
)

// StrategyAI selects the best strategy for each symbol based on market conditions
type StrategyAI struct {
	MarketAnalyzer  *market.MarketAnalyzer
	StrategyWeights map[string]map[string]float64 // symbol -> strategy -> weight
}

// NewStrategyAI creates a new StrategyAI
func NewStrategyAI(analyzer *market.MarketAnalyzer) *StrategyAI {
	return &StrategyAI{
		MarketAnalyzer:  analyzer,
		StrategyWeights: make(map[string]map[string]float64),
	}
}

// SelectStrategy selects the best strategy for a symbol based on market conditions
func (ai *StrategyAI) SelectStrategy(symbol string) StrategyType {
	// Get market regime for the symbol
	regime := ai.MarketAnalyzer.GetMarketRegime(symbol)

	// Calculate strategy weights based on market conditions
	weights := ai.calculateStrategyWeights(regime)

	// Store weights for reference
	if _, exists := ai.StrategyWeights[symbol]; !exists {
		ai.StrategyWeights[symbol] = make(map[string]float64)
	}

	for strategy, weight := range weights {
		ai.StrategyWeights[symbol][strategy] = weight
	}

	// Select strategy with highest weight
	bestStrategy := MarketMaking
	highestWeight := 0.0

	for strategy, weight := range weights {
		if weight > highestWeight {
			highestWeight = weight
			bestStrategy = StrategyType(strategy)
		}
	}

	return bestStrategy
}

// calculateStrategyWeights calculates weights for each strategy based on market regime
func (ai *StrategyAI) calculateStrategyWeights(regime *market.MarketRegime) map[string]float64 {
	weights := make(map[string]float64)

	// Base weights
	weights[string(MarketMaking)] = 0.25
	weights[string(Momentum)] = 0.25
	weights[string(MeanReversion)] = 0.25
	weights[string(VolatilityBreakout)] = 0.25

	// Adjust weights based on market regime
	switch regime.Volatility {
	case "high_volatility":
		weights[string(VolatilityBreakout)] += 0.3
		weights[string(MarketMaking)] -= 0.1
		weights[string(Momentum)] += 0.1
		weights[string(MeanReversion)] -= 0.3
	case "low_volatility":
		weights[string(MeanReversion)] += 0.3
		weights[string(MarketMaking)] += 0.1
		weights[string(Momentum)] -= 0.1
		weights[string(VolatilityBreakout)] -= 0.3
	}

	switch regime.Trend {
	case "trending_up", "trending_down":
		weights[string(Momentum)] += 0.4
		weights[string(MarketMaking)] -= 0.2
		weights[string(MeanReversion)] -= 0.2
	case "ranging":
		weights[string(MeanReversion)] += 0.4
		weights[string(MarketMaking)] += 0.1
		weights[string(Momentum)] -= 0.3
		weights[string(VolatilityBreakout)] -= 0.2
	}

	switch regime.Volume {
	case "high_volume":
		weights[string(Momentum)] += 0.2
		weights[string(VolatilityBreakout)] += 0.2
		weights[string(MarketMaking)] -= 0.2
		weights[string(MeanReversion)] -= 0.2
	case "low_volume":
		weights[string(MarketMaking)] += 0.3
		weights[string(MeanReversion)] += 0.1
		weights[string(Momentum)] -= 0.2
		weights[string(VolatilityBreakout)] -= 0.2
	}

	// Normalize weights to sum to 1.0
	total := 0.0
	for _, weight := range weights {
		total += weight
	}

	if total > 0 {
		for strategy := range weights {
			weights[strategy] = weights[strategy] / total
		}
	}

	return weights
}

// GetStrategyWeights returns the current strategy weights for a symbol
func (ai *StrategyAI) GetStrategyWeights(symbol string) map[string]float64 {
	if weights, exists := ai.StrategyWeights[symbol]; exists {
		return weights
	}
	return make(map[string]float64)
}

// CalculateVolatilityScore calculates a volatility score for strategy selection
func (ai *StrategyAI) CalculateVolatilityScore(regime *market.MarketRegime) float64 {
	switch regime.Volatility {
	case "high_volatility":
		return 0.9
	case "low_volatility":
		return 0.3
	default:
		return 0.6
	}
}

// CalculateTrendScore calculates a trend score for strategy selection
func (ai *StrategyAI) CalculateTrendScore(regime *market.MarketRegime) float64 {
	switch regime.Trend {
	case "trending_up", "trending_down":
		return 0.9
	case "ranging":
		return 0.3
	default:
		return 0.6
	}
}

// CalculateVolumeScore calculates a volume score for strategy selection
func (ai *StrategyAI) CalculateVolumeScore(regime *market.MarketRegime) float64 {
	switch regime.Volume {
	case "high_volume":
		return 0.9
	case "low_volume":
		return 0.3
	default:
		return 0.6
	}
}
