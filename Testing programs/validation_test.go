package main

// unit tests for all input validation logic in validation.go.
// These tests do not touch the database or HTTP layer

import "testing"

func TestValidatePhone_Valid(t *testing.T) {
	valid := []string{
		"07700 900001",
		"+447700900001",
		"(0113) 496-0000",
		"01134960000",
		"", // empty is allowed — field is optional
	}
	for _, v := range valid {
		if !validatePhone(v) {
			t.Errorf("validatePhone(%q) = false, want true", v)
		}
	}
}

func TestValidatePhone_Invalid(t *testing.T) {
	invalid := []string{
		"<script>alert(1)</script>",
		"DROP TABLE users",
		"abc",
		"077 900 ABCD",
		"123",                           // too short (under 7 chars)
		"01134960000123456789012345678", // too long (over 25 chars)
	}
	for _, v := range invalid {
		if validatePhone(v) {
			t.Errorf("validatePhone(%q) = true, want false", v)
		}
	}
}

// name regex

func TestNameRegex_Valid(t *testing.T) {
	valid := []string{
		"John Smith",
		"Emily Roberts",
		"O'Brien",
		"Mary-Jane",
		"Dr. House",
		"A", // single character
	}
	for _, v := range valid {
		if !nameRegex.MatchString(v) {
			t.Errorf("nameRegex.MatchString(%q) = false, want true", v)
		}
	}
}

func TestNameRegex_Invalid(t *testing.T) {
	invalid := []string{
		"John123",
		"<script>",
		"'; DROP TABLE",
		"",
	}
	for _, v := range invalid {
		if nameRegex.MatchString(v) {
			t.Errorf("nameRegex.MatchString(%q) = true, want false", v)
		}
	}
}

// date of birth regex (YYYY-MM-DD)
func TestDOBRegex_Valid(t *testing.T) {
	valid := []string{
		"1985-03-15",
		"2000-12-31",
		"1900-01-01",
	}
	for _, v := range valid {
		if !dobRegex.MatchString(v) {
			t.Errorf("dobRegex.MatchString(%q) = false, want true", v)
		}
	}
}

func TestDOBRegex_Invalid(t *testing.T) {
	invalid := []string{
		"15/03/1985",
		"1985/03/15",
		"March 15 1985",
		"85-03-15",
		"",
		"<script>",
	}
	for _, v := range invalid {
		if dobRegex.MatchString(v) {
			t.Errorf("dobRegex.MatchString(%q) = true, want false", v)
		}
	}
}

// blood type regex
func TestBloodRegex_Valid(t *testing.T) {
	valid := []string{"A+", "A-", "B+", "B-", "AB+", "AB-", "O+", "O-"}
	for _, v := range valid {
		if !bloodRegex.MatchString(v) {
			t.Errorf("bloodRegex.MatchString(%q) = false, want true", v)
		}
	}
}

func TestBloodRegex_Invalid(t *testing.T) {
	invalid := []string{
		"C+", "X-", "AB", "a+", "A", "+", "",
		"<script>", "O+ extra",
	}
	for _, v := range invalid {
		if bloodRegex.MatchString(v) {
			t.Errorf("bloodRegex.MatchString(%q) = true, want false", v)
		}
	}
}

// username regex
func TestUsernameRegex_Valid(t *testing.T) {
	valid := []string{
		"admin", "jsmith", "user_123", "ABC", "a",
	}
	for _, v := range valid {
		if !usernameRegex.MatchString(v) {
			t.Errorf("usernameRegex.MatchString(%q) = false, want true", v)
		}
	}
}

func TestUsernameRegex_Invalid(t *testing.T) {
	invalid := []string{
		"admin'--",
		"user name", // space
		"user@domain",
		"<script>",
		"'; DROP TABLE",
		"",
	}
	for _, v := range invalid {
		if usernameRegex.MatchString(v) {
			t.Errorf("usernameRegex.MatchString(%q) = true, want false", v)
		}
	}
}

// validationErrors helper
func TestValidationErrors_Empty(t *testing.T) {
	var ve validationErrors
	ve.check(true, "this should not appear")
	if len(ve) != 0 {
		t.Errorf("expected 0 errors, got %d", len(ve))
	}
}

func TestValidationErrors_Collects(t *testing.T) {
	var ve validationErrors
	ve.check(false, "error one")
	ve.check(false, "error two")
	ve.check(true, "this passes")
	if len(ve) != 2 {
		t.Errorf("expected 2 errors, got %d: %v", len(ve), ve)
	}
}
