package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Browser     BrowserConfig    `yaml:"browser"`
	LinkedIn    LinkedInConfig   `yaml:"linkedin"`
	Search      SearchConfig     `yaml:"search"`
	Connections ConnectionConfig `yaml:"connections"`
	Messaging   MessagingConfig  `yaml:"messaging"`
	Stealth     StealthConfig    `yaml:"stealth"`
	Database    DatabaseConfig   `yaml:"database"`
	Logging     LoggingConfig    `yaml:"logging"`
}

type BrowserConfig struct {
	Headless bool           `yaml:"headless"`
	Timeout  int            `yaml:"timeout"`
	Viewport ViewportConfig `yaml:"viewport"`
}

type ViewportConfig struct {
	Width  int `yaml:"width"`
	Height int `yaml:"height"`
}

type LinkedInConfig struct {
	Email    string `yaml:"email"`
	Password string `yaml:"password"`
	BaseURL  string `yaml:"base_url"`
}

type SearchConfig struct {
	MaxResults      int `yaml:"max_results"`
	ResultsPerPage  int `yaml:"results_per_page"`
	PaginationDelay int `yaml:"pagination_delay"`
}

type ConnectionConfig struct {
	DailyLimit  int    `yaml:"daily_limit"`
	MinDelay    int    `yaml:"min_delay"`
	MaxDelay    int    `yaml:"max_delay"`
	DefaultNote string `yaml:"default_note"`
}

type MessagingConfig struct {
	Enabled          bool     `yaml:"enabled"`
	FollowUpDelay    int      `yaml:"follow_up_delay"`
	MessageTemplates []string `yaml:"message_templates"`
}

type StealthConfig struct {
	MouseMovement MouseMovementConfig `yaml:"mouse_movement"`
	Timing        TimingConfig        `yaml:"timing"`
	Fingerprint   FingerprintConfig   `yaml:"fingerprint"`
	Scrolling     ScrollingConfig     `yaml:"scrolling"`
	Typing        TypingConfig        `yaml:"typing"`
	Hovering      HoveringConfig      `yaml:"hovering"`
	Scheduling    SchedulingConfig    `yaml:"scheduling"`
	RateLimiting  RateLimitingConfig  `yaml:"rate_limiting"`
}

type MouseMovementConfig struct {
	Enabled              bool    `yaml:"enabled"`
	BezierCurves         bool    `yaml:"bezier_curves"`
	OvershootProbability float64 `yaml:"overshoot_probability"`
	MicroCorrections     bool    `yaml:"micro_corrections"`
}

type TimingConfig struct {
	Enabled             bool    `yaml:"enabled"`
	MinThinkTime        int     `yaml:"min_think_time"`
	MaxThinkTime        int     `yaml:"max_think_time"`
	ScrollSpeedVariance float64 `yaml:"scroll_speed_variance"`
}

type FingerprintConfig struct {
	Enabled              bool `yaml:"enabled"`
	RandomizeUserAgent   bool `yaml:"randomize_user_agent"`
	RandomizeViewport    bool `yaml:"randomize_viewport"`
	DisableWebdriverFlag bool `yaml:"disable_webdriver_flag"`
}

type ScrollingConfig struct {
	Enabled               bool    `yaml:"enabled"`
	VariableSpeed         bool    `yaml:"variable_speed"`
	ScrollBackProbability float64 `yaml:"scroll_back_probability"`
}

type TypingConfig struct {
	Enabled           bool    `yaml:"enabled"`
	MinKeystrokeDelay int     `yaml:"min_keystroke_delay"`
	MaxKeystrokeDelay int     `yaml:"max_keystroke_delay"`
	TypoProbability   float64 `yaml:"typo_probability"`
}

type HoveringConfig struct {
	Enabled          bool    `yaml:"enabled"`
	HoverProbability float64 `yaml:"hover_probability"`
	HoverDurationMin int     `yaml:"hover_duration_min"`
	HoverDurationMax int     `yaml:"hover_duration_max"`
}

type SchedulingConfig struct {
	Enabled           bool    `yaml:"enabled"`
	BusinessHoursOnly bool    `yaml:"business_hours_only"`
	StartHour         int     `yaml:"start_hour"`
	EndHour           int     `yaml:"end_hour"`
	BreakProbability  float64 `yaml:"break_probability"`
}

type RateLimitingConfig struct {
	Enabled            bool `yaml:"enabled"`
	ConnectionCooldown int  `yaml:"connection_cooldown"`
	MessageCooldown    int  `yaml:"message_cooldown"`
}

type DatabaseConfig struct {
	Path string `yaml:"path"`
}

type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
	Output string `yaml:"output"`
}

// LoadConfig loads configuration from YAML file and environment variables
func LoadConfig(configPath string) (*Config, error) {
	// Load environment variables from .env file if it exists
	_ = godotenv.Load()

	// Read YAML config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Override with environment variables if set
	if email := os.Getenv("LINKEDIN_EMAIL"); email != "" {
		cfg.LinkedIn.Email = email
	}
	if password := os.Getenv("LINKEDIN_PASSWORD"); password != "" {
		cfg.LinkedIn.Password = password
	}
	if headless := os.Getenv("LINKEDIN_HEADLESS"); headless != "" {
		if val, err := strconv.ParseBool(headless); err == nil {
			cfg.Browser.Headless = val
		}
	}
	if dailyLimit := os.Getenv("LINKEDIN_DAILY_LIMIT"); dailyLimit != "" {
		if val, err := strconv.Atoi(dailyLimit); err == nil {
			cfg.Connections.DailyLimit = val
		}
	}

	return &cfg, nil
}
