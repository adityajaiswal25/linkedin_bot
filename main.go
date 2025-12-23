package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-rod/rod"
	"linkedin-automation/pkg/auth"
	"linkedin-automation/pkg/config"
	"linkedin-automation/pkg/connection"
	"linkedin-automation/pkg/database"
	"linkedin-automation/pkg/logger"
	"linkedin-automation/pkg/messaging"
	"linkedin-automation/pkg/search"
	"linkedin-automation/pkg/stealth"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "config/config.yaml", "Path to configuration file")
	mode := flag.String("mode", "search", "Operation mode: search, connect, message, or all")
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	log, err := logger.NewLogger(cfg.Logging.Level, cfg.Logging.Format, cfg.Logging.Output)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Close()

	logger.Info("LinkedIn Automation Tool Started", map[string]interface{}{
		"mode": *mode,
	})

	// Ensure data directory exists
	dbDir := filepath.Dir(cfg.Database.Path)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		logger.Error("Failed to create data directory", map[string]interface{}{"error": err.Error()})
		os.Exit(1)
	}

	// Initialize database
	db, err := database.NewDB(cfg.Database.Path)
	if err != nil {
		logger.Error("Failed to initialize database", map[string]interface{}{"error": err.Error()})
		os.Exit(1)
	}
	defer db.Close()

	// Initialize authentication
	authInstance, err := auth.NewAuth(cfg)
	if err != nil {
		logger.Error("Failed to initialize authentication", map[string]interface{}{"error": err.Error()})
		os.Exit(1)
	}
	defer authInstance.Close()

	// Perform login
	if err := authInstance.Login(); err != nil {
		logger.Error("Login failed", map[string]interface{}{"error": err.Error()})
		os.Exit(1)
	}

	// Get authenticated page and stealth instance
	page := authInstance.GetPage()
	stealthInstance := authInstance.GetStealth()

	// Check if we should operate based on scheduling
	if !stealthInstance.ShouldOperate() {
		logger.Info("Outside business hours, waiting...", nil)
		// In a real implementation, you'd wait until business hours
	}

	// Execute based on mode
	switch *mode {
	case "search":
		if err := runSearch(cfg, page, stealthInstance, db); err != nil {
			logger.Error("Search failed", map[string]interface{}{"error": err.Error()})
			os.Exit(1)
		}

	case "connect":
		if err := runConnect(cfg, page, stealthInstance, db); err != nil {
			logger.Error("Connection failed", map[string]interface{}{"error": err.Error()})
			os.Exit(1)
		}

	case "message":
		if err := runMessage(cfg, page, stealthInstance, db); err != nil {
			logger.Error("Messaging failed", map[string]interface{}{"error": err.Error()})
			os.Exit(1)
		}

	case "all":
		// Run all operations in sequence
		if err := runSearch(cfg, page, stealthInstance, db); err != nil {
			logger.Warn("Search failed, continuing", map[string]interface{}{"error": err.Error()})
		}

		stealthInstance.RandomBreak()

		if err := runConnect(cfg, page, stealthInstance, db); err != nil {
			logger.Warn("Connection failed, continuing", map[string]interface{}{"error": err.Error()})
		}

		stealthInstance.RandomBreak()

		if err := runMessage(cfg, page, stealthInstance, db); err != nil {
			logger.Warn("Messaging failed, continuing", map[string]interface{}{"error": err.Error()})
		}

	default:
		logger.Error("Invalid mode", map[string]interface{}{"mode": *mode})
		os.Exit(1)
	}

	logger.Info("LinkedIn Automation Tool Completed", nil)
}

// runSearch executes search operations
func runSearch(cfg *config.Config, page *rod.Page, stealthInstance *stealth.Stealth, db *database.DB) error {
	searchInstance := search.NewSearch(cfg, page, stealthInstance, db)

	// Example search parameters
	params := search.SearchParams{
		JobTitle: "Software Engineer",
		Location: "San Francisco",
		Keywords: "Python Go",
	}

	profiles, err := searchInstance.SearchProfiles(params)
	if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}

	logger.Info("Search completed", map[string]interface{}{
		"profiles_found": len(profiles),
	})

	return nil
}

// runConnect executes connection request operations
func runConnect(cfg *config.Config, page *rod.Page, stealthInstance *stealth.Stealth, db *database.DB) error {
	connInstance := connection.NewConnection(cfg, page, stealthInstance, db)

	// Get profiles from database that haven't been connected
	// In a real implementation, you'd query the database for profiles without connection requests
	// For now, this is a placeholder

	logger.Info("Connection operations completed", nil)
	return nil
}

// runMessage executes messaging operations
func runMessage(cfg *config.Config, page *rod.Page, stealthInstance *stealth.Stealth, db *database.DB) error {
	msgInstance := messaging.NewMessaging(cfg, page, stealthInstance, db)

	// Send follow-up messages to accepted connections
	if err := msgInstance.SendFollowUpMessages(); err != nil {
		return fmt.Errorf("follow-up messages failed: %w", err)
	}

	logger.Info("Messaging operations completed", nil)
	return nil
}

