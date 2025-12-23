package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

// DB wraps the database connection
type DB struct {
	conn *sql.DB
}

// Profile represents a LinkedIn profile
type Profile struct {
	ID          int64
	URL         string
	Name        string
	Headline    string
	Title       string
	Company     string
	Location    string
	FoundAt     time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// ConnectionRequest represents a sent connection request
type ConnectionRequest struct {
	ID          int64
	ProfileID   int64
	ProfileURL  string
	Note        string
	Status      string // "pending", "accepted", "rejected"
	SentAt      time.Time
	AcceptedAt  *time.Time
}

// Message represents a sent message
type Message struct {
	ID          int64
	ProfileID   int64
	ProfileURL  string
	Content     string
	SentAt      time.Time
}

// DailyStats tracks daily activity limits
type DailyStats struct {
	Date           time.Time
	ConnectionsSent int
	MessagesSent    int
}

// NewDB creates a new database connection
func NewDB(path string) (*DB, error) {
	conn, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db := &DB{conn: conn}
	if err := db.init(); err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	return db, nil
}

// init creates the necessary database tables
func (db *DB) init() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS profiles (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			url TEXT UNIQUE NOT NULL,
			name TEXT,
			headline TEXT,
			title TEXT,
			company TEXT,
			location TEXT,
			found_at DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS connection_requests (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			profile_id INTEGER,
			profile_url TEXT NOT NULL,
			note TEXT,
			status TEXT DEFAULT 'pending',
			sent_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			accepted_at DATETIME,
			FOREIGN KEY (profile_id) REFERENCES profiles(id)
		)`,
		`CREATE TABLE IF NOT EXISTS messages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			profile_id INTEGER,
			profile_url TEXT NOT NULL,
			content TEXT NOT NULL,
			sent_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (profile_id) REFERENCES profiles(id)
		)`,
		`CREATE TABLE IF NOT EXISTS daily_stats (
			date DATE PRIMARY KEY,
			connections_sent INTEGER DEFAULT 0,
			messages_sent INTEGER DEFAULT 0
		)`,
		`CREATE INDEX IF NOT EXISTS idx_profiles_url ON profiles(url)`,
		`CREATE INDEX IF NOT EXISTS idx_connection_requests_profile_url ON connection_requests(profile_url)`,
		`CREATE INDEX IF NOT EXISTS idx_connection_requests_status ON connection_requests(status)`,
		`CREATE INDEX IF NOT EXISTS idx_messages_profile_url ON messages(profile_url)`,
		`CREATE INDEX IF NOT EXISTS idx_daily_stats_date ON daily_stats(date)`,
	}

	for _, query := range queries {
		if _, err := db.conn.Exec(query); err != nil {
			return fmt.Errorf("failed to execute query: %w", err)
		}
	}

	return nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.conn.Close()
}

// Exec is a passthrough to the underlying database Exec
func (db *DB) Exec(query string, args ...interface{}) (sql.Result, error) {
	return db.conn.Exec(query, args...)
}

// QueryRow is a passthrough to the underlying database QueryRow
func (db *DB) QueryRow(query string, args ...interface{}) *sql.Row {
	return db.conn.QueryRow(query, args...)
}

// Query is a passthrough to the underlying database Query
func (db *DB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return db.conn.Query(query, args...)
}

// AddProfile adds a new profile to the database
func (db *DB) AddProfile(profile *Profile) error {
	query := `INSERT OR IGNORE INTO profiles (url, name, headline, title, company, location, found_at) 
	          VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := db.conn.Exec(query, profile.URL, profile.Name, profile.Headline, profile.Title, profile.Company, profile.Location, profile.FoundAt)
	return err
}

// GetProfileByURL retrieves a profile by URL
func (db *DB) GetProfileByURL(url string) (*Profile, error) {
	query := `SELECT id, url, name, headline, title, company, location, found_at, created_at, updated_at 
	          FROM profiles WHERE url = ?`
	row := db.conn.QueryRow(query, url)

	var profile Profile
	err := row.Scan(&profile.ID, &profile.URL, &profile.Name, &profile.Headline, &profile.Title,
		&profile.Company, &profile.Location, &profile.FoundAt, &profile.CreatedAt, &profile.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &profile, nil
}

// AddConnectionRequest adds a new connection request
func (db *DB) AddConnectionRequest(req *ConnectionRequest) error {
	query := `INSERT INTO connection_requests (profile_id, profile_url, note, status) 
	          VALUES (?, ?, ?, ?)`
	_, err := db.conn.Exec(query, req.ProfileID, req.ProfileURL, req.Note, req.Status)
	return err
}

// HasConnectionRequest checks if a connection request was already sent
func (db *DB) HasConnectionRequest(profileURL string) (bool, error) {
	query := `SELECT COUNT(*) FROM connection_requests WHERE profile_url = ?`
	var count int
	err := db.conn.QueryRow(query, profileURL).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// UpdateConnectionRequestStatus updates the status of a connection request
func (db *DB) UpdateConnectionRequestStatus(profileURL string, status string) error {
	query := `UPDATE connection_requests SET status = ?, accepted_at = ? WHERE profile_url = ?`
	var acceptedAt *time.Time
	if status == "accepted" {
		now := time.Now()
		acceptedAt = &now
	}
	_, err := db.conn.Exec(query, status, acceptedAt, profileURL)
	return err
}

// GetPendingConnections returns all pending connection requests
func (db *DB) GetPendingConnections() ([]*ConnectionRequest, error) {
	query := `SELECT id, profile_id, profile_url, note, status, sent_at, accepted_at 
	          FROM connection_requests WHERE status = 'pending'`
	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []*ConnectionRequest
	for rows.Next() {
		var req ConnectionRequest
		err := rows.Scan(&req.ID, &req.ProfileID, &req.ProfileURL, &req.Note, 
			&req.Status, &req.SentAt, &req.AcceptedAt)
		if err != nil {
			return nil, err
		}
		requests = append(requests, &req)
	}

	return requests, nil
}

// AddMessage adds a new message
func (db *DB) AddMessage(msg *Message) error {
	query := `INSERT INTO messages (profile_id, profile_url, content) VALUES (?, ?, ?)`
	_, err := db.conn.Exec(query, msg.ProfileID, msg.ProfileURL, msg.Content)
	return err
}

// HasMessage checks if a message was already sent to a profile
func (db *DB) HasMessage(profileURL string) (bool, error) {
	query := `SELECT COUNT(*) FROM messages WHERE profile_url = ?`
	var count int
	err := db.conn.QueryRow(query, profileURL).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetDailyStats retrieves daily statistics for a given date
func (db *DB) GetDailyStats(date time.Time) (*DailyStats, error) {
	query := `SELECT date, connections_sent, messages_sent FROM daily_stats WHERE date = ?`
	row := db.conn.QueryRow(query, date.Format("2006-01-02"))

	var stats DailyStats
	err := row.Scan(&stats.Date, &stats.ConnectionsSent, &stats.MessagesSent)
	if err == sql.ErrNoRows {
		// Return zero stats if no record exists
		return &DailyStats{
			Date:            date,
			ConnectionsSent: 0,
			MessagesSent:    0,
		}, nil
	}
	if err != nil {
		return nil, err
	}

	return &stats, nil
}

// IncrementDailyConnections increments the daily connection count
func (db *DB) IncrementDailyConnections(date time.Time) error {
	query := `INSERT INTO daily_stats (date, connections_sent, messages_sent) 
	          VALUES (?, 1, 0)
	          ON CONFLICT(date) DO UPDATE SET connections_sent = connections_sent + 1`
	_, err := db.conn.Exec(query, date.Format("2006-01-02"))
	return err
}

// IncrementDailyMessages increments the daily message count
func (db *DB) IncrementDailyMessages(date time.Time) error {
	query := `INSERT INTO daily_stats (date, connections_sent, messages_sent) 
	          VALUES (?, 0, 1)
	          ON CONFLICT(date) DO UPDATE SET messages_sent = messages_sent + 1`
	_, err := db.conn.Exec(query, date.Format("2006-01-02"))
	return err
}

