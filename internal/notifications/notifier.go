package notifications

import (
	"fmt"
	"log"
	"net/smtp"
	"os"
	"strings"
)

// Notifier handles sending notifications
type Notifier struct {
	EmailConfig    *EmailConfig
	TelegramConfig *TelegramConfig
}

// EmailConfig holds email configuration
type EmailConfig struct {
	SMTPHost      string
	SMTPPort      string
	SenderEmail   string
	SenderPass    string
	ReceiverEmail string
}

// TelegramConfig holds Telegram configuration
type TelegramConfig struct {
	BotToken string
	ChatID   string
}

// TradeAlert represents a trade alert
type TradeAlert struct {
	Symbol     string
	Action     string
	Quantity   float64
	Price      float64
	Strategy   string
	Confidence float64
	Reason     string
	Timestamp  string
}

// NewNotifier creates a new Notifier
func NewNotifier() *Notifier {
	// Load email configuration from environment variables
	emailConfig := &EmailConfig{
		SMTPHost:      os.Getenv("SMTP_HOST"),
		SMTPPort:      os.Getenv("SMTP_PORT"),
		SenderEmail:   os.Getenv("SENDER_EMAIL"),
		SenderPass:    os.Getenv("SENDER_PASS"),
		ReceiverEmail: os.Getenv("RECEIVER_EMAIL"),
	}

	// Load Telegram configuration from environment variables
	telegramConfig := &TelegramConfig{
		BotToken: os.Getenv("TELEGRAM_BOT_TOKEN"),
		ChatID:   os.Getenv("TELEGRAM_CHAT_ID"),
	}

	return &Notifier{
		EmailConfig:    emailConfig,
		TelegramConfig: telegramConfig,
	}
}

// SendTradeAlert sends a trade alert via email and/or Telegram
func (n *Notifier) SendTradeAlert(alert TradeAlert) error {
	// Send email alert if configured
	if n.EmailConfig.SenderEmail != "" && n.EmailConfig.ReceiverEmail != "" {
		if err := n.sendEmailAlert(alert); err != nil {
			log.Printf("Warning: Failed to send email alert: %v", err)
		}
	}

	// Send Telegram alert if configured
	if n.TelegramConfig.BotToken != "" && n.TelegramConfig.ChatID != "" {
		if err := n.sendTelegramAlert(alert); err != nil {
			log.Printf("Warning: Failed to send Telegram alert: %v", err)
		}
	}

	return nil
}

// sendEmailAlert sends an email alert
func (n *Notifier) sendEmailAlert(alert TradeAlert) error {
	// Check if email is configured
	if n.EmailConfig.SMTPHost == "" || n.EmailConfig.SenderEmail == "" || n.EmailConfig.SenderPass == "" {
		return fmt.Errorf("email not properly configured")
	}

	// Compose email
	subject := fmt.Sprintf("Trade Alert: %s %s", alert.Symbol, alert.Action)
	body := fmt.Sprintf(`
Trade Alert Details:
-------------------
Symbol: %s
Action: %s
Quantity: %.4f
Price: $%.4f
Strategy: %s
Confidence: %.2f%%
Reason: %s
Timestamp: %s
`, alert.Symbol, alert.Action, alert.Quantity, alert.Price, alert.Strategy, alert.Confidence*100, alert.Reason, alert.Timestamp)

	// Compose the full message
	message := fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s",
		n.EmailConfig.ReceiverEmail, subject, body)

	// Connect to SMTP server
	auth := smtp.PlainAuth("", n.EmailConfig.SenderEmail, n.EmailConfig.SenderPass, n.EmailConfig.SMTPHost)
	addr := n.EmailConfig.SMTPHost + ":" + n.EmailConfig.SMTPPort

	// Send email
	err := smtp.SendMail(addr, auth, n.EmailConfig.SenderEmail, []string{n.EmailConfig.ReceiverEmail}, []byte(message))
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	log.Printf("Email alert sent for %s %s", alert.Symbol, alert.Action)
	return nil
}

// sendTelegramAlert sends a Telegram alert
func (n *Notifier) sendTelegramAlert(alert TradeAlert) error {
	// This is a simplified implementation
	// In a real implementation, you would make an HTTP request to the Telegram Bot API
	message := fmt.Sprintf(`
🔔 *Trade Alert*
Symbol: %s
Action: %s
Quantity: %.4f
Price: $%.4f
Strategy: %s
Confidence: %.2f%%
Reason: %s
`, alert.Symbol, alert.Action, alert.Quantity, alert.Price, alert.Strategy, alert.Confidence*100, alert.Reason)

	// Log the message (in a real implementation, you would send it to Telegram)
	log.Printf("Telegram alert prepared: %s", strings.ReplaceAll(message, "\n", " | "))
	log.Printf("Telegram alert would be sent to chat %s with bot token %s...",
		n.TelegramConfig.ChatID, n.TelegramConfig.BotToken[:10]+"...")

	return nil
}

// SendEmergencyStopAlert sends an emergency stop alert
func (n *Notifier) SendEmergencyStopAlert(reason string) error {
	// Send email alert if configured
	if n.EmailConfig.SenderEmail != "" && n.EmailConfig.ReceiverEmail != "" {
		subject := "🚨 Emergency Stop Alert"
		body := fmt.Sprintf("The trading bot has been stopped due to: %s", reason)
		message := fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s",
			n.EmailConfig.ReceiverEmail, subject, body)

		auth := smtp.PlainAuth("", n.EmailConfig.SenderEmail, n.EmailConfig.SenderPass, n.EmailConfig.SMTPHost)
		addr := n.EmailConfig.SMTPHost + ":" + n.EmailConfig.SMTPPort

		err := smtp.SendMail(addr, auth, n.EmailConfig.SenderEmail, []string{n.EmailConfig.ReceiverEmail}, []byte(message))
		if err != nil {
			log.Printf("Warning: Failed to send emergency stop email: %v", err)
		}
	}

	// Send Telegram alert if configured
	if n.TelegramConfig.BotToken != "" && n.TelegramConfig.ChatID != "" {
		message := fmt.Sprintf("🚨 *Emergency Stop Alert*\nThe trading bot has been stopped due to: %s", reason)
		log.Printf("Emergency stop Telegram alert prepared: %s", strings.ReplaceAll(message, "\n", " | "))
	}

	return nil
}
