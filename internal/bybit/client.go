package bybit

import (
	"context"
	"fmt"
	"time"

	"github.com/hirokisan/bybit/v2"
	"github.com/shopspring/decimal"
)

// Client wraps the Bybit API client
type Client struct {
	bybitClient *bybit.Client
}

// NewClient creates a new Bybit client
func NewClient(apiKey, apiSecret string, testnet bool) *Client {
	var client *bybit.Client

	if testnet {
		client = bybit.NewTestClient().Client
		// Set the correct base URL for testnet
		client.WithBaseURL(bybit.TestNetBaseURL)
	} else {
		client = bybit.NewClient()
		// Set the correct base URL for mainnet
		client.WithBaseURL(bybit.MainNetBaseURL)
	}

	// Set authentication
	client.WithAuth(apiKey, apiSecret)

	return &Client{
		bybitClient: client,
	}
}

// GetTopCoins fetches the top traded coins on Bybit
func (c *Client) GetTopCoins(ctx context.Context, limit int) ([]string, error) {
	// For now, return a fixed list of top coins
	// In a real implementation, you would fetch this from the API
	topCoins := []string{"BTCUSDT", "ETHUSDT", "SOLUSDT", "XRPUSDT", "ADAUSDT", "DOGEUSDT"}

	if limit < len(topCoins) {
		return topCoins[:limit], nil
	}

	return topCoins, nil
}

// GetMarketData fetches market data for a symbol
func (c *Client) GetMarketData(ctx context.Context, symbol string) (*MarketData, error) {
	// Try using V5 API instead
	limit := 100
	param := bybit.V5GetKlineParam{
		Category: "spot",
		Symbol:   bybit.SymbolV5(symbol),
		Interval: "5",
		Limit:    &limit,
	}

	resp, err := c.bybitClient.V5().Market().GetKline(param)
	if err != nil {
		return nil, fmt.Errorf("failed to get kline data via V5 API: %w", err)
	}

	// Convert kline data to our format
	klineData := make([]KlineData, 0, len(resp.Result.List))
	for _, k := range resp.Result.List {
		open, _ := decimal.NewFromString(k.Open)
		high, _ := decimal.NewFromString(k.High)
		low, _ := decimal.NewFromString(k.Low)
		close, _ := decimal.NewFromString(k.Close)
		volume, _ := decimal.NewFromString(k.Volume)

		startTime, _ := time.Parse("2006-01-02 15:04:05", k.StartTime)

		klineData = append(klineData, KlineData{
			Open:      open,
			High:      high,
			Low:       low,
			Close:     close,
			Volume:    volume,
			Timestamp: startTime,
		})
	}

	return &MarketData{
		Symbol:    symbol,
		Timestamp: time.Now(),
		Kline:     klineData,
	}, nil
}

// PlaceOrder places a new order
func (c *Client) PlaceOrder(ctx context.Context, order Order) error {
	// Convert order side
	var side bybit.Side
	if order.Side == "BUY" {
		side = "Buy"
	} else {
		side = "Sell"
	}

	// Convert order type
	var orderType bybit.OrderTypeSpot
	if order.Type == "MARKET" {
		orderType = bybit.OrderTypeSpotMarket
	} else {
		orderType = bybit.OrderTypeSpotLimit
	}

	// Convert quantity to float64
	quantity, _ := order.Quantity.Float64()

	// Create order request
	symbolSpot := bybit.SymbolSpot(order.Symbol)
	req := bybit.SpotPostOrderParam{
		Symbol: symbolSpot,
		Qty:    quantity,
		Side:   side,
		Type:   orderType,
	}

	if order.Type == "LIMIT" {
		price, _ := order.Price.Float64()
		req.Price = &price
	}

	// Place the order
	_, err := c.bybitClient.Spot().V1().SpotPostOrder(req)
	if err != nil {
		return fmt.Errorf("failed to place order: %w", err)
	}

	return nil
}

// CancelOrder cancels an existing order
func (c *Client) CancelOrder(ctx context.Context, symbol, orderID string) error {
	req := bybit.SpotDeleteOrderParam{
		OrderID: &orderID,
	}

	_, err := c.bybitClient.Spot().V1().SpotDeleteOrder(req)
	if err != nil {
		return fmt.Errorf("failed to cancel order: %w", err)
	}

	return nil
}

// GetPositions gets current positions (for spot, this would be account balances)
func (c *Client) GetPositions(ctx context.Context, symbol string) ([]Position, error) {
	// For spot trading, we'll get account balances
	account, err := c.bybitClient.Spot().V1().SpotGetWalletBalance()
	if err != nil {
		return nil, fmt.Errorf("failed to get account info: %w", err)
	}

	// Find the base and quote currencies from the symbol
	// e.g., BTCUSDT -> BTC and USDT
	var baseCurrency, quoteCurrency string
	if len(symbol) > 5 && symbol[len(symbol)-4:] == "USDT" {
		baseCurrency = symbol[:len(symbol)-4]
		quoteCurrency = "USDT"
	} else {
		// Simplified - in practice you'd parse this more carefully
		baseCurrency = "BTC"
		quoteCurrency = "USDT"
	}

	positions := make([]Position, 0, 2)

	// Look for base currency balance
	for _, balance := range account.Result.Balances {
		if balance.Coin == baseCurrency {
			free, _ := decimal.NewFromString(balance.Free)
			locked, _ := decimal.NewFromString(balance.Locked)
			total := free.Add(locked)

			positions = append(positions, Position{
				Symbol:        symbol,
				Side:          "LONG", // Simplified
				Size:          total,
				AvgPrice:      decimal.Zero, // Would need to calculate from trade history
				UnrealisedPnl: decimal.Zero, // Would need to calculate
			})
		}

		if balance.Coin == quoteCurrency {
			free, _ := decimal.NewFromString(balance.Free)
			locked, _ := decimal.NewFromString(balance.Locked)
			total := free.Add(locked)

			positions = append(positions, Position{
				Symbol:        symbol,
				Side:          "CASH", // Simplified
				Size:          total,
				AvgPrice:      decimal.Zero,
				UnrealisedPnl: decimal.Zero,
			})
		}
	}

	return positions, nil
}
