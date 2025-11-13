package risk

import (
	"fmt"

	"github.com/forbest/bybitgo/internal/bybit"
	"github.com/forbest/bybitgo/internal/config"
)

// RiskManager handles risk management for the trading bot
type RiskManager struct {
	Config    *config.Config
	Positions map[string]PositionRisk
}

// PositionRisk tracks risk metrics for a position
type PositionRisk struct {
	Symbol            string
	CurrentSize       float64
	EntryPrice        float64
	CurrentPrice      float64
	UnrealizedPnL     float64
	MaxDrawdown       float64
	CorrelationRisk   float64
	StopLossLevel     float64
	TakeProfitLevel   float64
	PeakValue         float64 // Track peak value for drawdown calculation
	TrailingStopLevel float64 // Trailing stop level
	IsTrailingStopSet bool    // Whether trailing stop is active
}

// RiskMetrics tracks overall portfolio risk
type RiskMetrics struct {
	TotalExposure     float64
	PortfolioDrawdown float64
	Volatility        float64
	CorrelationRisk   float64
}

// NewRiskManager creates a new RiskManager
func NewRiskManager(cfg *config.Config) *RiskManager {
	return &RiskManager{
		Config:    cfg,
		Positions: make(map[string]PositionRisk),
	}
}

// CheckPositionRisk checks if a position exceeds risk limits
func (rm *RiskManager) CheckPositionRisk(symbol string, orderSize float64, price float64) error {
	// Check position size limit
	if orderSize > rm.Config.MaxPositionPerCoin {
		return fmt.Errorf("order size %.2f exceeds maximum position limit %.2f for %s",
			orderSize, rm.Config.MaxPositionPerCoin, symbol)
	}

	// Check if adding this position would exceed total capital
	currentExposure := rm.GetTotalExposure()
	newExposure := currentExposure + (orderSize * price)

	if newExposure > rm.Config.TotalCapital {
		return fmt.Errorf("new position would exceed total capital: current %.2f + new %.2f > total %.2f",
			currentExposure, orderSize*price, rm.Config.TotalCapital)
	}

	return nil
}

// CheckPortfolioRisk checks overall portfolio risk
func (rm *RiskManager) CheckPortfolioRisk() error {
	metrics := rm.CalculateRiskMetrics()

	// Check maximum drawdown
	if metrics.PortfolioDrawdown > rm.Config.MaxDrawdown {
		return fmt.Errorf("portfolio drawdown %.2f%% exceeds maximum allowed %.2f%%",
			metrics.PortfolioDrawdown*100, rm.Config.MaxDrawdown*100)
	}

	// Check total exposure
	if metrics.TotalExposure > rm.Config.TotalCapital {
		return fmt.Errorf("total exposure %.2f exceeds capital %.2f",
			metrics.TotalExposure, rm.Config.TotalCapital)
	}

	return nil
}

// CalculateRiskMetrics calculates current risk metrics
func (rm *RiskManager) CalculateRiskMetrics() *RiskMetrics {
	totalExposure := rm.GetTotalExposure()
	portfolioDrawdown := rm.CalculatePortfolioDrawdown()
	volatility := rm.CalculatePortfolioVolatility()
	correlationRisk := rm.CalculateCorrelationRisk()

	return &RiskMetrics{
		TotalExposure:     totalExposure,
		PortfolioDrawdown: portfolioDrawdown,
		Volatility:        volatility,
		CorrelationRisk:   correlationRisk,
	}
}

// GetTotalExposure calculates total portfolio exposure
func (rm *RiskManager) GetTotalExposure() float64 {
	total := 0.0
	for _, pos := range rm.Positions {
		total += pos.CurrentSize * pos.CurrentPrice
	}
	return total
}

// CalculatePortfolioDrawdown calculates portfolio drawdown
func (rm *RiskManager) CalculatePortfolioDrawdown() float64 {
	totalPnL := 0.0
	totalValue := 0.0

	for _, pos := range rm.Positions {
		totalPnL += pos.UnrealizedPnL
		totalValue += pos.CurrentSize * pos.EntryPrice
	}

	if totalValue == 0 {
		return 0
	}

	return totalPnL / totalValue
}

// CalculatePortfolioVolatility calculates portfolio volatility
func (rm *RiskManager) CalculatePortfolioVolatility() float64 {
	// Simplified calculation - in practice would use covariance matrix
	totalVolatility := 0.0
	count := 0

	for _, _ = range rm.Positions {
		// Use a proxy for individual position volatility
		// In practice, this would come from market data analysis
		positionVolatility := 0.02 // 2% as example
		totalVolatility += positionVolatility
		count++
	}

	if count == 0 {
		return 0
	}

	return totalVolatility / float64(count)
}

// CalculateCorrelationRisk calculates correlation risk across positions
func (rm *RiskManager) CalculateCorrelationRisk() float64 {
	// Simplified calculation - in practice would use correlation matrix
	// Higher correlation = higher risk (less diversification)

	if len(rm.Positions) <= 1 {
		return 0
	}

	// Assume average correlation of 0.3 for crypto assets
	// In practice, this would be calculated from historical data
	return 0.3
}

// UpdatePosition updates position risk metrics
func (rm *RiskManager) UpdatePosition(symbol string, position bybit.Position) {
	size, _ := position.Size.Float64()
	avgPrice, _ := position.AvgPrice.Float64()
	unrealizedPnL, _ := position.UnrealisedPnl.Float64()

	// Calculate stop-loss and take-profit levels
	stopLossLevel := avgPrice * (1 - rm.Config.StopLossPercent/100)
	takeProfitLevel := avgPrice * (1 + rm.Config.TakeProfitPercent/100)

	// Get existing position data to preserve peak value and trailing stop
	existingPos, exists := rm.Positions[symbol]
	peakValue := existingPos.PeakValue
	trailingStopLevel := existingPos.TrailingStopLevel
	isTrailingStopSet := existingPos.IsTrailingStopSet

	// Calculate current position value
	currentValue := size*avgPrice + unrealizedPnL

	// Update peak value if current value is higher
	if !exists || currentValue > peakValue {
		peakValue = currentValue
		// Update trailing stop level when new peak is reached
		if exists && isTrailingStopSet {
			// Move trailing stop up by the same percentage as the peak increase
			trailingStopLevel = avgPrice * (1 - rm.Config.StopLossPercent/100)
		}
	}

	rm.Positions[symbol] = PositionRisk{
		Symbol:            symbol,
		CurrentSize:       size,
		EntryPrice:        avgPrice,
		CurrentPrice:      avgPrice, // Would use current market price
		UnrealizedPnL:     unrealizedPnL,
		MaxDrawdown:       0, // Would track historical drawdown
		CorrelationRisk:   0, // Would calculate based on other positions
		StopLossLevel:     stopLossLevel,
		TakeProfitLevel:   takeProfitLevel,
		PeakValue:         peakValue,
		TrailingStopLevel: trailingStopLevel,
		IsTrailingStopSet: isTrailingStopSet,
	}
}

// SetTrailingStop sets a trailing stop for a position
func (rm *RiskManager) SetTrailingStop(symbol string, currentPrice float64) {
	pos, exists := rm.Positions[symbol]
	if !exists {
		return
	}

	// Set trailing stop at the stop-loss level initially
	pos.TrailingStopLevel = currentPrice * (1 - rm.Config.StopLossPercent/100)
	pos.IsTrailingStopSet = true
	rm.Positions[symbol] = pos
}

// CheckStopLossTakeProfit checks if any positions have hit stop-loss or take-profit levels
func (rm *RiskManager) CheckStopLossTakeProfit(currentPrices map[string]float64) []string {
	var actions []string

	for symbol, pos := range rm.Positions {
		currentPrice, exists := currentPrices[symbol]
		if !exists {
			continue
		}

		// Update current price
		pos.CurrentPrice = currentPrice
		rm.Positions[symbol] = pos

		// Check for long positions
		if pos.CurrentSize > 0 {
			// Check trailing stop
			if pos.IsTrailingStopSet && currentPrice <= pos.TrailingStopLevel {
				actions = append(actions, fmt.Sprintf("TRAILING_STOP: Close long position for %s at %.4f (trailing stop level: %.4f)",
					symbol, currentPrice, pos.TrailingStopLevel))
			} else if currentPrice <= pos.StopLossLevel {
				// Check stop-loss (price dropped below stop-loss level)
				actions = append(actions, fmt.Sprintf("STOP_LOSS: Close long position for %s at %.4f (stop-loss level: %.4f)",
					symbol, currentPrice, pos.StopLossLevel))
			} else if currentPrice >= pos.TakeProfitLevel {
				// Check take-profit (price rose above take-profit level)
				actions = append(actions, fmt.Sprintf("TAKE_PROFIT: Close long position for %s at %.4f (take-profit level: %.4f)",
					symbol, currentPrice, pos.TakeProfitLevel))
			} else if pos.IsTrailingStopSet && currentPrice > pos.PeakValue {
				// Update trailing stop if price increased and trailing stop is set
				// Move trailing stop up to maintain the same distance from peak
				newTrailingStop := currentPrice * (1 - rm.Config.StopLossPercent/100)
				if newTrailingStop > pos.TrailingStopLevel {
					pos.TrailingStopLevel = newTrailingStop
					pos.PeakValue = currentPrice
					rm.Positions[symbol] = pos
				}
			}
		}
	}

	return actions
}

// CheckSymbolDrawdown checks if any symbol has exceeded its maximum drawdown limit
func (rm *RiskManager) CheckSymbolDrawdown() []string {
	var actions []string

	for symbol, pos := range rm.Positions {
		if pos.PeakValue > 0 {
			currentValue := pos.CurrentSize*pos.CurrentPrice + pos.UnrealizedPnL
			drawdown := (pos.PeakValue - currentValue) / pos.PeakValue

			// Check if drawdown exceeds the configured maximum (use same as portfolio for now)
			if drawdown > rm.Config.MaxDrawdown {
				actions = append(actions, fmt.Sprintf("MAX_DRAWDOWN_EXCEEDED: %s drawdown %.2f%% exceeds limit %.2f%%",
					symbol, drawdown*100, rm.Config.MaxDrawdown*100))
			}
		}
	}

	return actions
}

// GetRiskReport generates a risk report
func (rm *RiskManager) GetRiskReport() string {
	metrics := rm.CalculateRiskMetrics()

	report := fmt.Sprintf("Risk Report:\n")
	report += fmt.Sprintf("  Total Exposure: $%.2f (%.1f%% of capital)\n",
		metrics.TotalExposure, (metrics.TotalExposure/rm.Config.TotalCapital)*100)
	report += fmt.Sprintf("  Portfolio Drawdown: %.2f%%\n", metrics.PortfolioDrawdown*100)
	report += fmt.Sprintf("  Portfolio Volatility: %.2f%%\n", metrics.Volatility*100)
	report += fmt.Sprintf("  Correlation Risk: %.2f\n", metrics.CorrelationRisk)

	// Add stop-loss and take-profit information
	report += fmt.Sprintf("  Stop-Loss Level: %.2f%%\n", rm.Config.StopLossPercent)
	report += fmt.Sprintf("  Take-Profit Level: %.2f%%\n", rm.Config.TakeProfitPercent)

	// Add symbol drawdown information
	report += fmt.Sprintf("  Symbol Drawdown Limits: %.2f%%\n", rm.Config.MaxDrawdown*100)

	if rm.ShouldStopTrading() {
		report += "  WARNING: Trading should be stopped due to excessive risk!\n"
	}

	return report
}

// ShouldStopTrading checks if trading should be stopped due to risk limits
func (rm *RiskManager) ShouldStopTrading() bool {
	metrics := rm.CalculateRiskMetrics()

	// Stop if drawdown exceeds 2x the configured maximum
	if metrics.PortfolioDrawdown > rm.Config.MaxDrawdown*2 {
		return true
	}

	// Stop if exposure exceeds 1.5x capital
	if metrics.TotalExposure > rm.Config.TotalCapital*1.5 {
		return true
	}

	return false
}
