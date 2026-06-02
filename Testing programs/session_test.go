package main

// tests for session token generation, session creation, retrieval, deletion, account lockout, and cookie helpers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// token generation

func TestGenerateToken_Length(t *testing.T) {
	token, err := generateToken()
	if err != nil {
		t.Fatalf("generateToken() error: %v", err)
	}
	// Base64 URL-encoded 32 bytes produces 44 characters
	if len(token) < 40 {
		t.Errorf("token too short: got %d chars, want >= 40", len(token))
	}
}

func TestGenerateToken_Unique(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		token, err := generateToken()
		if err != nil {
			t.Fatalf("generateToken() error on iteration %d: %v", i, err)
		}
		if seen[token] {
			t.Errorf("duplicate token generated on iteration %d: %s", i, token)
		}
		seen[token] = true
	}
}

// session creation and retrieval
func TestCreateSession_Stores(t *testing.T) {
	// get a user ID from the seeded database
	var userID int
	err := db.QueryRow("SELECT id FROM users WHERE username = 'jsmith'").Scan(&userID)
	if err != nil {
		t.Fatalf("could not find test user: %v", err)
	}

	session, err := createSession(userID)
	if err != nil {
		t.Fatalf("createSession() error: %v", err)
	}

	if session.Token == "" {
		t.Error("session.Token is empty")
	}
	if session.CSRFToken == "" {
		t.Error("session.CSRFToken is empty")
	}
	if session.UserID != userID {
		t.Errorf("session.UserID = %d, want %d", session.UserID, userID)
	}
	if session.ExpiresAt.Before(time.Now()) {
		t.Error("session.ExpiresAt is in the past")
	}

	// confirm the row exists in the database
	var count int
	db.QueryRow("SELECT COUNT(*) FROM sessions WHERE token = ?", session.Token).Scan(&count)
	if count != 1 {
		t.Errorf("expected 1 session row, got %d", count)
	}

	// cleanup
	deleteSession(session.Token)
}

func TestGetSession_ValidToken(t *testing.T) {
	var userID int
	db.QueryRow("SELECT id FROM users WHERE username = 'emilyr'").Scan(&userID)

	created, err := createSession(userID)
	if err != nil {
		t.Fatalf("createSession() error: %v", err)
	}
	defer deleteSession(created.Token)

	//  build a fake request carrying the session cookie.
	req := httptest.NewRequest(http.MethodGet, "/record", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: created.Token})

	session, user, err := getSession(req)
	if err != nil {
		t.Fatalf("getSession() error: %v", err)
	}
	if session == nil {
		t.Fatal("getSession() returned nil session for valid token")
	}
	if user == nil {
		t.Fatal("getSession() returned nil user for valid token")
	}
	if user.Username != "emilyr" {
		t.Errorf("user.Username = %q, want %q", user.Username, "emilyr")
	}
}

func TestGetSession_NoCooke(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/record", nil)
	// no cookie set.
	session, user, err := getSession(req)
	if err != nil {
		t.Fatalf("getSession() unexpected error: %v", err)
	}
	if session != nil || user != nil {
		t.Error("expected nil session and user when no cookie present")
	}
}

func TestGetSession_InvalidToken(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/record", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: "totallyinvalidtoken"})

	session, user, err := getSession(req)
	if err != nil {
		t.Fatalf("getSession() unexpected error: %v", err)
	}
	if session != nil || user != nil {
		t.Error("expected nil session and user for invalid token")
	}
}

// session deletion
func TestDeleteSession_Removes(t *testing.T) {
	var userID int
	db.QueryRow("SELECT id FROM users WHERE username = 'mbrown'").Scan(&userID)

	session, _ := createSession(userID)
	deleteSession(session.Token)

	var count int
	db.QueryRow("SELECT COUNT(*) FROM sessions WHERE token = ?", session.Token).Scan(&count)
	if count != 0 {
		t.Errorf("expected 0 session rows after deletion, got %d", count)
	}
}

// cookie hellpers
func TestSetSessionCookie_Flags(t *testing.T) {
	rr := httptest.NewRecorder()
	setSessionCookie(rr, "testtoken123")

	cookies := rr.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("no cookie set in response")
	}

	cookie := cookies[0]
	if cookie.Name != "session" {
		t.Errorf("cookie.Name = %q, want %q", cookie.Name, "session")
	}
	if cookie.Value != "testtoken123" {
		t.Errorf("cookie.Value = %q, want %q", cookie.Value, "testtoken123")
	}
	if !cookie.HttpOnly {
		t.Error("cookie.HttpOnly = false, want true")
	}
	if cookie.SameSite != http.SameSiteStrictMode {
		t.Error("cookie.SameSite != SameSiteStrictMode")
	}
}

func TestClearSessionCookie(t *testing.T) {
	rr := httptest.NewRecorder()
	clearSessionCookie(rr)

	cookies := rr.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("no cookie set in clear response")
	}
	cookie := cookies[0]
	if cookie.MaxAge != -1 {
		t.Errorf("cookie.MaxAge = %d, want -1 to expire immediately", cookie.MaxAge)
	}
}

// account lockout
func TestAccountLockout_LocksAfterFiveFailures(t *testing.T) {
	resetLockout("jsmith")
	defer resetLockout("jsmith")

	for i := 0; i < maxFailedAttempts; i++ {
		locked, _ := isLockedOut("jsmith")
		if locked {
			t.Fatalf("account locked early after %d attempts", i)
		}
		recordFailedAttempt("jsmith")
	}

	locked, msg := isLockedOut("jsmith")
	if !locked {
		t.Error("account should be locked after max failed attempts")
	}
	if msg == "" {
		t.Error("lockout message should not be empty")
	}
}

func TestAccountLockout_ResetsOnSuccess(t *testing.T) {
	resetLockout("jsmith")
	defer resetLockout("jsmith")

	// record failures
	recordFailedAttempt("jsmith")
	recordFailedAttempt("jsmith")

	// simulate successful login resetting the counter
	resetFailedAttempts("jsmith")

	var attempts int
	db.QueryRow("SELECT failed_attempts FROM users WHERE username = 'jsmith'").Scan(&attempts)
	if attempts != 0 {
		t.Errorf("failed_attempts = %d after reset, want 0", attempts)
	}
}

func TestAccountLockout_UnknownUser(t *testing.T) {
	// should not panic or error for a username that doesn't exist
	locked, _ := isLockedOut("doesnotexist")
	if locked {
		t.Error("non-existent user should not be locked")
	}
}
