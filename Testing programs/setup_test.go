package main

// initialises the shared test environment for all test files.
// uses an in-memory SQLite database so tests never touch srms.db.
// all test files in package main share the global db and tmpl variables set up here by TestMain

import (
	"database/sql"
	"html/template"
	"os"
	"testing"
)

// TestMain runs once before all tests.
// mirrors what main() does at startup
func TestMain(m *testing.M) {
	var err error

	// open an in memory SQLite database so tests are isolated from srms.db
	db, err = sql.Open("sqlite3", ":memory:")
	if err != nil {
		panic("TestMain: failed to open in-memory database: " + err.Error())
	}
	db.Exec("PRAGMA foreign_keys=ON")

	// create the same schema as initDB() in database.go
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id              INTEGER PRIMARY KEY AUTOINCREMENT,
		username        TEXT    UNIQUE NOT NULL,
		password_hash   TEXT    NOT NULL,
		role            TEXT    NOT NULL DEFAULT 'user',
		failed_attempts INTEGER NOT NULL DEFAULT 0,
		locked_until    DATETIME,
		created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	CREATE TABLE IF NOT EXISTS sessions (
		token      TEXT PRIMARY KEY,
		user_id    INTEGER NOT NULL,
		csrf_token TEXT    NOT NULL,
		expires_at DATETIME NOT NULL,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);
	CREATE TABLE IF NOT EXISTS medical_records (
		id                INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id           INTEGER NOT NULL UNIQUE,
		full_name         TEXT    NOT NULL,
		date_of_birth     TEXT    NOT NULL,
		blood_type        TEXT    NOT NULL DEFAULT '',
		allergies         TEXT    NOT NULL DEFAULT '',
		medications       TEXT    NOT NULL DEFAULT '',
		phone             TEXT    NOT NULL DEFAULT '',
		emergency_contact TEXT    NOT NULL DEFAULT '',
		gp_name           TEXT    NOT NULL DEFAULT '',
		notes             TEXT    NOT NULL DEFAULT '',
		last_updated_by   TEXT    NOT NULL DEFAULT 'system',
		last_updated_at   DATETIME,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);`

	if _, err := db.Exec(schema); err != nil {
		panic("TestMain: schema creation failed: " + err.Error())
	}

	// seed test accounts, thes same as seedData() in database.go
	seedData()

	// Parse HTML templates (relative path works because go test runs from
	// the package directory, i.e. the project root).
	tmpl = template.Must(template.ParseGlob("templates/*.html"))

	// run all tests and exit with their combined result code.
	code := m.Run()

	db.Close()
	os.Exit(code)
}

// resetLockout is a test helper that clears the lockout state for a username so lockout tests don't bleed into each other
func resetLockout(username string) {
	db.Exec("UPDATE users SET failed_attempts = 0, locked_until = NULL WHERE username = ?", username)
}

// createTestSession is a test helper that creates a real session for a user
// and returns the session token, used by handler tests that need authentication
func createTestSession(username string) (string, string, error) {
	var userID int
	err := db.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userID)
	if err != nil {
		return "", "", err
	}
	session, err := createSession(userID)
	if err != nil {
		return "", "", err
	}
	return session.Token, session.CSRFToken, nil
}
