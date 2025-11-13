package config

import (
	"os"
	"strconv"
)

// Config holds all configuration parameters for the trading bot
type Config struct {
	BybitAPIKey        string
	BybitAPISecret     string
	Testnet            bool
	TotalCapital       float64
	MaxPositionPerCoin float64
	RebalanceMinutes   int
	BaseOrderSize      float64
	RiskPerTrade       float64
	MaxDrawdown        float64
	VolatilityLookback int
	TrendPeriod        int
	MomentumPeriod     int
	// Stop-loss and take-profit settings
	StopLossPercent   float64
	TakeProfitPercent float64
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	cfg := &Config{
		BybitAPIKey:    os.Getenv("BYBIT_API_KEY"),
		BybitAPISecret: os.Getenv("BYBIT_API_SECRET"),
		Testnet:        os.Getenv("TESTNET") == "true",
	}

	if val, err := strconv.ParseFloat(os.Getenv("TOTAL_CAPITAL"), 64); err == nil {
		cfg.TotalCapital = val
	}

	if val, err := strconv.ParseFloat(os.Getenv("MAX_POSITION_PER_COIN"), 64); err == nil {
		cfg.MaxPositionPerCoin = val
	}

	if val, err := strconv.Atoi(os.Getenv("REBALANCE_MINUTES")); err == nil {
		cfg.RebalanceMinutes = val
	}

	if val, err := strconv.ParseFloat(os.Getenv("BASE_ORDER_SIZE"), 64); err == nil {
		cfg.BaseOrderSize = val
	}

	if val, err := strconv.ParseFloat(os.Getenv("RISK_PER_TRADE"), 64); err == nil {
		cfg.RiskPerTrade = val
	}

	if val, err := strconv.ParseFloat(os.Getenv("MAX_DRAWDOWN"), 64); err == nil {
		cfg.MaxDrawdown = val
	}

	if val, err := strconv.Atoi(os.Getenv("VOLATILITY_LOOKBACK")); err == nil {
		cfg.VolatilityLookback = val
	}

	if val, err := strconv.Atoi(os.Getenv("TREND_PERIOD")); err == nil {
		cfg.TrendPeriod = val
	}

	if val, err := strconv.Atoi(os.Getenv("MOMENTUM_PERIOD")); err == nil {
		cfg.MomentumPeriod = val
	}

	// Load stop-loss and take-profit settings
	if val, err := strconv.ParseFloat(os.Getenv("STOP_LOSS_PERCENT"), 64); err == nil {
		cfg.StopLossPercent = val
	} else {
		cfg.StopLossPercent = 2.0 // Default 2% stop-loss
	}

	if val, err := strconv.ParseFloat(os.Getenv("TAKE_PROFIT_PERCENT"), 64); err == nil {
		cfg.TakeProfitPercent = val
	} else {
		cfg.TakeProfitPercent = 5.0 // Default 5% take-profit
	}

	return cfg, nil
}
