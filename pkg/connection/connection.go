package connection

import (
	"fmt"
	"strings"
	"time"

	"linkedin-automation/pkg/config"
	"linkedin-automation/pkg/database"
	"linkedin-automation/pkg/logger"
	stealthpkg "linkedin-automation/pkg/stealth"

	"github.com/go-rod/rod"
)

// Connection handles connection requests
type Connection struct {
	config  *config.Config
	page    *rod.Page
	stealth *stealthpkg.Stealth
	db      *database.DB
}

// ConnectionRequest represents a connection request
type ConnectionRequest struct {
	ID        int       `json:"id"`
	ProfileID int       `json:"profile_id"`
	Note      string    `json:"note"`
	Status    string    `json:"status"` // pending, sent, accepted, declined
	SentAt    time.Time `json:"sent_at"`
}

// NewConnection creates a new connection instance
func NewConnection(cfg *config.Config, page *rod.Page, stealth *stealthpkg.Stealth, db *database.DB) *Connection {
	return &Connection{
		config:  cfg,
		page:    page,
		stealth: stealth,
		db:      db,
	}
}

// SendConnectionRequests sends connection requests to profiles
func (c *Connection) SendConnectionRequests() error {
	logger.Warn("Connection requests functionality not fully implemented", nil)
	logger.Info("Connection operations placeholder", nil)

	// Get profiles that haven't been contacted yet
	profiles, err := c.getUncontactedProfiles()
	if err != nil {
		logger.Warn("Failed to get uncontacted profiles", map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}

	if len(profiles) == 0 {
		logger.Info("No uncontacted profiles found", nil)
		return nil
	}

	// Check daily limit
	sentToday, err := c.getConnectionsSentToday()
	if err != nil {
		logger.Warn("Failed to check daily limit", map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}

	remaining := c.config.Connections.DailyLimit - sentToday
	if remaining <= 0 {
		logger.Warn("Daily connection limit reached", nil)
		return fmt.Errorf("daily limit reached")
	}

	// Send connection requests
	sent := 0
	for _, profile := range profiles {
		if sent >= remaining {
			break
		}

		if err := c.sendConnectionRequest(profile); err != nil {
			logger.Warn("Failed to send connection request", map[string]interface{}{
				"profile_url": profile.URL,
				"error":       err.Error(),
			})
			continue
		}

		sent++
		logger.Info("Connection request sent", map[string]interface{}{
			"profile_url": profile.URL,
		})

		// Apply cooldown
		c.stealth.RandomDelay()
		time.Sleep(time.Duration(c.config.Stealth.RateLimiting.ConnectionCooldown) * time.Millisecond)
	}

	logger.Info("Connection requests completed", map[string]interface{}{
		"sent": sent,
	})

	return nil
}

func (c *Connection) sendConnectionRequest(profile database.Profile) error {
	// Navigate to profile
	if err := c.page.Navigate(profile.URL); err != nil {
		return fmt.Errorf("failed to navigate to profile: %w", err)
	}

	c.page.MustWaitLoad()

	// Scroll to load the connect button
	c.stealth.ScrollHumanLike(500)
	c.stealth.RandomDelay()

	// Find connect button
	connectBtn := c.page.MustElement("button[aria-label*='Connect']")
	if connectBtn == nil {
		return fmt.Errorf("connect button not found")
	}

	// Click connect button
	c.stealth.HumanClick(connectBtn)

	// Wait for modal
	c.page.MustElement("div[data-test-modal]").MustWaitVisible()

	// Check if "Send without note" is available
	sendWithoutNoteBtn := c.page.MustElements("button[aria-label='Send without a note']")
	if len(sendWithoutNoteBtn) > 0 {
		c.stealth.HumanClick(sendWithoutNoteBtn[0])
	} else {
		// Add a note
		addNoteBtn := c.page.MustElement("button[aria-label='Add a note']")
		if addNoteBtn != nil {
			c.stealth.HumanClick(addNoteBtn)

			// Wait for note textarea
			noteTextarea := c.page.MustElement("textarea[name='message']")
			noteTextarea.MustWaitVisible()

			// Generate personalized note
			note := c.generatePersonalizedNote(profile)

			// Type the note
			c.stealth.HumanType(noteTextarea, note)

			// Click send
			sendBtn := c.page.MustElement("button[aria-label='Send invitation']")
			c.stealth.HumanClick(sendBtn)
		}
	}

	// Save to database
	return c.saveConnectionRequest(profile.ID, "sent")
}

func (c *Connection) generatePersonalizedNote(profile database.Profile) string {
	note := c.config.Connections.DefaultNote

	// Replace placeholders
	note = strings.ReplaceAll(note, "{name}", c.extractFirstName(profile.Name))

	return note
}

func (c *Connection) extractFirstName(fullName string) string {
	parts := strings.Split(fullName, " ")
	if len(parts) > 0 {
		return parts[0]
	}
	return fullName
}

func (c *Connection) getUncontactedProfiles() ([]database.Profile, error) {
	rows, err := c.db.Query(`
		SELECT id, url, name, headline, location, found_at
		FROM profiles
		WHERE id NOT IN (
			SELECT profile_id FROM connection_requests WHERE status IN ('sent', 'accepted')
		)
		LIMIT 100
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var profiles []database.Profile
	for rows.Next() {
		var p database.Profile
		err := rows.Scan(&p.ID, &p.URL, &p.Name, &p.Headline, &p.Location, &p.FoundAt)
		if err != nil {
			continue
		}
		profiles = append(profiles, p)
	}

	return profiles, nil
}

func (c *Connection) getConnectionsSentToday() (int, error) {
	var count int
	err := c.db.QueryRow(`
		SELECT COUNT(*) FROM connection_requests
		WHERE DATE(sent_at) = DATE('now')
	`).Scan(&count)

	return count, err
}

func (c *Connection) saveConnectionRequest(profileID int, status string) error {
	_, err := c.db.Exec(`
		INSERT INTO connection_requests (profile_id, status, sent_at)
		VALUES (?, ?, ?)
	`, profileID, status, time.Now())

	return err
}
