package auth

import (
	"fmt"
	"time"

	"linkedin-automation/pkg/config"
	"linkedin-automation/pkg/logger"
	stealthpkg "linkedin-automation/pkg/stealth"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	rodstealth "github.com/go-rod/stealth"
)

// Auth handles LinkedIn authentication
type Auth struct {
	cfg     *config.Config
	page    *rod.Page
	browser *rod.Browser
	stealth *stealthpkg.Stealth
}

// NewAuth creates a new authentication instance
func NewAuth(cfg *config.Config) (*Auth, error) {
	return &Auth{cfg: cfg}, nil
}

// Login performs LinkedIn login
func (a *Auth) Login() error {
	// Launch browser
	l := launcher.New().
		Headless(a.cfg.Browser.Headless).
		Set("disable-blink-features", "AutomationControlled").
		Set("disable-web-security", "true").
		Set("disable-features", "VizDisplayCompositor")

	url, err := l.Launch()
	if err != nil {
		return fmt.Errorf("failed to launch browser: %w", err)
	}

	browser := rod.New().ControlURL(url).MustConnect()
	page := browser.MustPage().Timeout(time.Duration(a.cfg.Browser.Timeout) * time.Millisecond)
	page.MustSetViewport(a.cfg.Browser.Viewport.Width, a.cfg.Browser.Viewport.Height, 1, false)

	// Apply stealth mode
	page = rodstealth.MustPage(page)

	// Initialize stealth instance
	a.stealth = stealthpkg.NewStealth(a.cfg, page)

	// Apply stealth techniques
	if err := a.stealth.Apply(); err != nil {
		return fmt.Errorf("failed to apply stealth: %w", err)
	}

	a.browser = browser
	a.page = page

	// Navigate to LinkedIn login
	if err := page.Navigate(a.cfg.LinkedIn.BaseURL + "/login"); err != nil {
		return fmt.Errorf("failed to navigate to login page: %w", err)
	}

	// Wait for page load
	page.MustWaitLoad()

	// Check for existing session
	if a.isLoggedIn() {
		logger.Info("Already logged in", nil)
		return nil
	}

	// Fill login form
	if err := a.fillLoginForm(); err != nil {
		return fmt.Errorf("failed to fill login form: %w", err)
	}

	// Wait for potential security checkpoints or feed
	if a.hasSecurityCheckpoint() {
		logger.Warn("Security checkpoint detected - manual intervention required", nil)
		waitUntil := time.Now().Add(5 * time.Minute)
		for time.Now().Before(waitUntil) {
			time.Sleep(3 * time.Second)
			if a.isLoggedIn() {
				break
			}
		}
		if !a.isLoggedIn() {
			return fmt.Errorf("login timeout or failed")
		}
	} else {
		// Wait for successful login
		waitUntil := time.Now().Add(20 * time.Second)
		for time.Now().Before(waitUntil) {
			if a.isLoggedIn() {
				break
			}
			time.Sleep(500 * time.Millisecond)
		}
		if !a.isLoggedIn() {
			return fmt.Errorf("login failed")
		}
	}

	logger.Info("Login successful", nil)
	return nil
}

// GetPage returns the authenticated page
func (a *Auth) GetPage() *rod.Page {
	return a.page
}

// GetStealth returns the stealth instance
func (a *Auth) GetStealth() *stealthpkg.Stealth {
	return a.stealth
}

// Close closes the browser
func (a *Auth) Close() error {
	if a.browser != nil {
		return a.browser.Close()
	}
	return nil
}

func (a *Auth) isLoggedIn() bool {
	// Check if we're on the feed page or have the feed URL
	currentURL := a.page.MustInfo().URL
	return currentURL == a.cfg.LinkedIn.BaseURL+"/feed" ||
		currentURL == a.cfg.LinkedIn.BaseURL+"/feed/" ||
		a.page.MustHas("div[data-control-name=\"feed_out_of_network\"]") ||
		a.page.MustHas("div[data-control-name=\"feed_reconnect\"]")
}

func (a *Auth) fillLoginForm() error {
	// Wait for login form
	a.page.MustElement("input[name=\"session_key\"]").MustWaitVisible()

	// Type email with human-like behavior
	emailEl := a.page.MustElement("input[name=\"session_key\"]")
	a.stealth.HumanType(emailEl, a.cfg.LinkedIn.Email)

	// Type password
	passwordEl := a.page.MustElement("input[name=\"session_password\"]")
	a.stealth.HumanType(passwordEl, a.cfg.LinkedIn.Password)

	// Click sign in button
	signInBtn := a.page.MustElement("button[type=\"submit\"]")
	a.stealth.HumanClick(signInBtn)

	return nil
}

func (a *Auth) hasSecurityCheckpoint() bool {
	// Check for 2FA input
	if a.page.MustHas("input[name=\"pin\"]") {
		return true
	}

	// Check for captcha
	if a.page.MustHas("#captcha-internal") || a.page.MustHas(".captcha") {
		return true
	}

	return false
}
