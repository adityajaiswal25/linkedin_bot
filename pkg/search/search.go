package search

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"linkedin-automation/pkg/config"
	"linkedin-automation/pkg/database"
	"linkedin-automation/pkg/logger"
	stealthpkg "linkedin-automation/pkg/stealth"

	"github.com/go-rod/rod"
)

// Search handles LinkedIn profile search
type Search struct {
	config  *config.Config
	page    *rod.Page
	stealth *stealthpkg.Stealth
	db      *database.DB
}

// SearchParams represents search parameters
type SearchParams struct {
	JobTitle string
	Location string
	Keywords string
}

// Profile represents a LinkedIn profile
type Profile struct {
	ID         int       `json:"id"`
	URL        string    `json:"url"`
	Name       string    `json:"name"`
	Headline   string    `json:"headline"`
	Location   string    `json:"location"`
	Company    string    `json:"company"`
	JobTitle   string    `json:"job_title"`
	ProfilePic string    `json:"profile_pic"`
	FoundAt    time.Time `json:"found_at"`
}

// NewSearch creates a new search instance
func NewSearch(cfg *config.Config, page *rod.Page, stealth *stealthpkg.Stealth, db *database.DB) *Search {
	return &Search{
		config:  cfg,
		page:    page,
		stealth: stealth,
		db:      db,
	}
}

// SearchProfiles searches for profiles based on parameters
func (s *Search) SearchProfiles(params SearchParams) ([]Profile, error) {
	logger.Info("Starting profile search", map[string]interface{}{
		"job_title": params.JobTitle,
		"location":  params.Location,
		"keywords":  params.Keywords,
	})

	// Build search URL
	searchURL := s.buildSearchURL(params)

	// Navigate to search page
	if err := s.page.Navigate(searchURL); err != nil {
		return nil, fmt.Errorf("failed to navigate to search page: %w", err)
	}

	s.page.MustWaitLoad()

	// Scroll to load more results
	s.stealth.ScrollHumanLike(1000)
	s.stealth.RandomDelay()

	var profiles []Profile
	pageNum := 1

	for len(profiles) < s.config.Search.MaxResults {
		logger.Info("Processing search page", map[string]interface{}{
			"page":           pageNum,
			"profiles_found": len(profiles),
		})

		// Extract profiles from current page
		pageProfiles, err := s.extractProfilesFromPage()
		if err != nil {
			logger.Warn("Failed to extract profiles from page", map[string]interface{}{
				"error": err.Error(),
			})
			break
		}

		// Filter duplicates and save to database
		for _, profile := range pageProfiles {
			if len(profiles) >= s.config.Search.MaxResults {
				break
			}

			// Check if profile already exists
			if s.profileExists(profile.URL) {
				continue
			}

			profile.FoundAt = time.Now()
			profiles = append(profiles, profile)

			// Save to database
			if err := s.saveProfile(profile); err != nil {
				logger.Debug("Failed to save profile", map[string]interface{}{
					"url":   profile.URL,
					"error": err.Error(),
				})
			}
		}

		// Check if there's a next page
		if !s.hasNextPage() {
			break
		}

		// Go to next page
		if err := s.goToNextPage(); err != nil {
			logger.Warn("Failed to go to next page", map[string]interface{}{
				"error": err.Error(),
			})
			break
		}

		pageNum++
		time.Sleep(time.Duration(s.config.Search.PaginationDelay) * time.Millisecond)
	}

	logger.Info("Search completed", map[string]interface{}{
		"total_profiles": len(profiles),
	})

	return profiles, nil
}

func (s *Search) buildSearchURL(params SearchParams) string {
	baseURL := s.config.LinkedIn.BaseURL + "/search/results/people/"

	queryParams := url.Values{}

	if params.Keywords != "" {
		queryParams.Set("keywords", params.Keywords)
	}

	if params.JobTitle != "" {
		queryParams.Set("title", params.JobTitle)
	}

	if params.Location != "" {
		queryParams.Set("geoUrn", s.getLocationURN(params.Location))
	}

	queryParams.Set("origin", "SWITCH_SEARCH_VERTICAL")

	return baseURL + "?" + queryParams.Encode()
}

func (s *Search) getLocationURN(location string) string {
	// This would need a mapping of locations to LinkedIn URNs
	// For now, return a placeholder
	return "urn:li:geo:103644278" // United States
}

func (s *Search) extractProfilesFromPage() ([]Profile, error) {
	var profiles []Profile

	// Find profile cards
	profileCards := s.page.MustElements("div[data-chameleon-result-urn]")

	for _, card := range profileCards {
		profile, err := s.extractProfileFromCard(card)
		if err != nil {
			continue // Skip problematic profiles
		}

		profiles = append(profiles, profile)
	}

	return profiles, nil
}

func (s *Search) extractProfileFromCard(card *rod.Element) (Profile, error) {
	profile := Profile{}

	// Extract profile URL
	linkEl := card.MustElement("a[href*='/in/']")
	if linkEl == nil {
		return profile, fmt.Errorf("no profile link found")
	}

	href, err := linkEl.Attribute("href")
	if err != nil {
		return profile, err
	}

	// Clean up URL
	if strings.Contains(*href, "?") {
		profile.URL = strings.Split(*href, "?")[0]
	} else {
		profile.URL = *href
	}

	// Extract name
	nameEl := card.MustElement("span[aria-hidden='true']")
	if nameEl != nil {
		name, _ := nameEl.Text()
		profile.Name = strings.TrimSpace(name)
	}

	// Extract headline
	headlineEl := card.MustElement("div[data-anonymize='job-title']")
	if headlineEl != nil {
		headline, _ := headlineEl.Text()
		profile.Headline = strings.TrimSpace(headline)
	}

	// Extract location
	locationEl := card.MustElement("div[data-anonymize='location']")
	if locationEl != nil {
		location, _ := locationEl.Text()
		profile.Location = strings.TrimSpace(location)
	}

	return profile, nil
}

func (s *Search) hasNextPage() bool {
	nextBtn := s.page.MustElements("button[aria-label='Next']")
	return len(nextBtn) > 0 && nextBtn[0].MustVisible()
}

func (s *Search) goToNextPage() error {
	nextBtn := s.page.MustElement("button[aria-label='Next']")

	s.stealth.HumanClick(nextBtn.First)
	s.page.MustWaitLoad()

	return nil
}

func (s *Search) profileExists(url string) bool {
	// Check database for existing profile
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM profiles WHERE url = ?", url).Scan(&count)
	return err == nil && count > 0
}

func (s *Search) saveProfile(profile Profile) error {
	_, err := s.db.Exec(`
		INSERT INTO profiles (url, name, headline, location, found_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(url) DO NOTHING
	`, profile.URL, profile.Name, profile.Headline, profile.Location, profile.FoundAt)

	return err
}
