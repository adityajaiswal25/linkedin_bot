package messaging

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"linkedin-automation/pkg/config"
	"linkedin-automation/pkg/database"
	"linkedin-automation/pkg/logger"
	"linkedin-automation/pkg/stealth"
)

// Messaging handles LinkedIn messaging
type Messaging struct {
	config  *config.Config
	page    *rod.Page
	stealth *stealth.Stealth
	db      *database.DB
}

// NewMessaging creates a new messaging instance
func NewMessaging(cfg *config.Config, page *rod.Page, st *stealth.Stealth, db *database.DB) *Messaging {
	return &Messaging{
		config:  cfg,
		page:    page,
		stealth: st,
		db:      db,
	}
}

// SendMessage sends a message to a profile
func (m *Messaging) SendMessage(profileURL string, message string) error {
	// Check if already sent
	hasMessage, err := m.db.HasMessage(profileURL)
	if err != nil {
		logger.Warn("Failed to check message status", map[string]interface{}{"error": err.Error()})
	}
	if hasMessage {
		logger.Info("Message already sent", map[string]interface{}{"profile_url": profileURL})
		return fmt.Errorf("message already sent")
	}

	logger.Info("Sending message", map[string]interface{}{"profile_url": profileURL})

	// Navigate to profile
	if err := m.page.Navigate(profileURL); err != nil {
		return fmt.Errorf("failed to navigate to profile: %w", err)
	}

	m.page.MustWaitLoad()
	m.stealth.RandomDelay()

	// Find message button
	messageButton, err := m.findMessageButton()
	if err != nil {
		return fmt.Errorf("failed to find message button: %w", err)
	}

	// Human-like interaction
	box, _ := messageButton.Shape()
	if err := m.stealth.HumanMouseMove(box.X+box.Width/2, box.Y+box.Height/2); err != nil {
		return fmt.Errorf("failed to move mouse to message button: %w", err)
	}

	m.stealth.RandomHover(messageButton)
	m.stealth.RandomDelay()

	// Click message button
	messageButton.MustClick()
	m.stealth.RandomDelay()

	// Wait for message modal/chat to open
	time.Sleep(2 * time.Second)

	// Find message input
	messageInput, err := m.findMessageInput()
	if err != nil {
		return fmt.Errorf("failed to find message input: %w", err)
	}

	// Type message
	box, _ = messageInput.Shape()
	if err := m.stealth.HumanMouseMove(box.X+box.Width/2, box.Y+box.Height/2); err != nil {
		return fmt.Errorf("failed to move mouse to message input: %w", err)
	}

	messageInput.MustClick()
	m.stealth.RandomDelay()

	// Type message with human-like typing
	if err := m.stealth.HumanType(message); err != nil {
		return fmt.Errorf("failed to type message: %w", err)
	}

	m.stealth.RandomDelay()

	// Find and click send button
	sendButton, err := m.findSendButton()
	if err != nil {
		return fmt.Errorf("failed to find send button: %w", err)
	}

	box, _ = sendButton.Shape()
	if err := m.stealth.HumanMouseMove(box.X+box.Width/2, box.Y+box.Height/2); err != nil {
		return fmt.Errorf("failed to move mouse to send button: %w", err)
	}

	m.stealth.RandomHover(sendButton)
	sendButton.MustClick()
	m.stealth.RandomDelay()

	// Wait for message to send
	time.Sleep(1 * time.Second)

	// Save to database
	profile, _ := m.db.GetProfileByURL(profileURL)
	profileID := int64(0)
	if profile != nil {
		profileID = profile.ID
	}

	msg := &database.Message{
		ProfileID:  profileID,
		ProfileURL: profileURL,
		Content:    message,
	}

	if err := m.db.AddMessage(msg); err != nil {
		logger.Warn("Failed to save message", map[string]interface{}{"error": err.Error()})
	}

	// Update daily stats
	if err := m.db.IncrementDailyMessages(time.Now()); err != nil {
		logger.Warn("Failed to increment daily messages", map[string]interface{}{"error": err.Error()})
	}

	// Apply cooldown
	m.stealth.MessageCooldown()

	logger.Info("Message sent", map[string]interface{}{"profile_url": profileURL})
	return nil
}

// findMessageButton finds the message button on the profile page
func (m *Messaging) findMessageButton() (*rod.Element, error) {
	selectors := []string{
		"button[aria-label*='Message']",
		"button:has-text('Message')",
		"a[href*='/messaging/']",
		"button.pvs-profile-actions__action",
	}

	for _, selector := range selectors {
		if !m.page.MustHas(selector) {
			continue
		}

		button, err := m.page.Element(selector)
		if err != nil {
			continue
		}

		text := strings.ToLower(button.MustText())
		if strings.Contains(text, "message") {
			return button, nil
		}

		// Check href for messaging link
		href, _ := button.Attribute("href")
		if href != nil && strings.Contains(*href, "messaging") {
			return button, nil
		}
	}

	return nil, fmt.Errorf("message button not found")
}

// findMessageInput finds the message input field
func (m *Messaging) findMessageInput() (*rod.Element, error) {
	selectors := []string{
		"div[contenteditable='true'][role='textbox']",
		"textarea[placeholder*='message']",
		"div[data-placeholder*='message']",
		"div.msg-form__contenteditable",
		"div[aria-label*='message']",
	}

	for _, selector := range selectors {
		if !m.page.MustHas(selector) {
			continue
		}

		input, err := m.page.Element(selector)
		if err != nil {
			continue
		}

		return input, nil
	}

	return nil, fmt.Errorf("message input not found")
}

// findSendButton finds the send button
func (m *Messaging) findSendButton() (*rod.Element, error) {
	selectors := []string{
		"button[aria-label*='Send']",
		"button:has-text('Send')",
		"button.msg-form__send-button",
		"button[type='submit']",
	}

	for _, selector := range selectors {
		if !m.page.MustHas(selector) {
			continue
		}

		button, err := m.page.Element(selector)
		if err != nil {
			continue
		}

		text := strings.ToLower(button.MustText())
		if strings.Contains(text, "send") {
			return button, nil
		}

		// Check if it's a submit button in message form
		disabled, _ := button.Attribute("disabled")
		if disabled == nil || *disabled != "true" {
			return button, nil
		}
	}

	return nil, fmt.Errorf("send button not found")
}

// SendFollowUpMessages sends follow-up messages to newly accepted connections
func (m *Messaging) SendFollowUpMessages() error {
	if !m.config.Messaging.Enabled {
		return nil
	}

	logger.Info("Checking for newly accepted connections", nil)

	// Get pending connections
	pendingConnections, err := m.db.GetPendingConnections()
	if err != nil {
		return fmt.Errorf("failed to get pending connections: %w", err)
	}

	// Check each connection to see if it was accepted
	for _, conn := range pendingConnections {
		// Check if enough time has passed since connection request
		timeSinceRequest := time.Since(conn.SentAt)
		if timeSinceRequest < m.config.Messaging.FollowUpDelay {
			continue
		}

		// Navigate to profile to check status
		if err := m.page.Navigate(conn.ProfileURL); err != nil {
			logger.Warn("Failed to navigate to profile", map[string]interface{}{
				"profile_url": conn.ProfileURL,
				"error":       err.Error(),
			})
			continue
		}

		m.page.MustWaitLoad()
		m.stealth.RandomDelay()

		// Check if connection was accepted (message button should be available)
		if m.page.MustHas("button[aria-label*='Message']") {
			// Connection accepted, send follow-up message
			message := m.getFollowUpMessage(conn.ProfileURL)

			if err := m.SendMessage(conn.ProfileURL, message); err != nil {
				logger.Warn("Failed to send follow-up message", map[string]interface{}{
					"profile_url": conn.ProfileURL,
					"error":       err.Error(),
				})
				continue
			}

			// Update connection status
			if err := m.db.UpdateConnectionRequestStatus(conn.ProfileURL, "accepted"); err != nil {
				logger.Warn("Failed to update connection status", map[string]interface{}{"error": err.Error()})
			}

			logger.Info("Follow-up message sent", map[string]interface{}{"profile_url": conn.ProfileURL})
		}
	}

	return nil
}

// getFollowUpMessage generates a follow-up message from templates
func (m *Messaging) getFollowUpMessage(profileURL string) string {
	templates := m.config.Messaging.MessageTemplates
	if len(templates) == 0 {
		return "Hi! Thanks for connecting. I'd love to learn more about your work."
	}

	// Select random template
	template := templates[0] // In a real implementation, you'd randomize this

	// Personalize template
	profile, err := m.db.GetProfileByURL(profileURL)
	if err == nil && profile != nil {
		template = strings.ReplaceAll(template, "{name}", profile.Name)
		template = strings.ReplaceAll(template, "{title}", profile.Title)
		template = strings.ReplaceAll(template, "{company}", profile.Company)
		template = strings.ReplaceAll(template, "{location}", profile.Location)
		template = strings.ReplaceAll(template, "{industry}", "your industry") // Could be extracted from profile
	}

	return template
}

// SendBulkMessages sends messages to multiple profiles
func (m *Messaging) SendBulkMessages(profiles []string, messageTemplate string) error {
	successCount := 0
	for _, profileURL := range profiles {
		// Personalize message
		message := m.personalizeMessage(messageTemplate, profileURL)

		if err := m.SendMessage(profileURL, message); err != nil {
			logger.Warn("Failed to send message", map[string]interface{}{
				"profile_url": profileURL,
				"error":       err.Error(),
			})
			continue
		}

		successCount++
	}

	logger.Info("Bulk messages completed", map[string]interface{}{
		"total":   len(profiles),
		"success": successCount,
	})

	return nil
}

// personalizeMessage personalizes a message template
func (m *Messaging) personalizeMessage(template, profileURL string) string {
	message := template

	// Get profile from database
	profile, err := m.db.GetProfileByURL(profileURL)
	if err == nil && profile != nil {
		message = strings.ReplaceAll(message, "{name}", profile.Name)
		message = strings.ReplaceAll(message, "{title}", profile.Title)
		message = strings.ReplaceAll(message, "{company}", profile.Company)
		message = strings.ReplaceAll(message, "{location}", profile.Location)
		message = strings.ReplaceAll(message, "{industry}", "your industry")
	}

	return message
}

