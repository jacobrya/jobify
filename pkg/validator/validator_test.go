package validator

import (
	"bytes"
	"net/http/httptest"
	"testing"
)

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		email string
		want  bool
	}{
		{"user@example.com", true},
		{"user+tag@sub.domain.org", true},
		{"first.last@company.io", true},
		{"invalid", false},
		{"@", false},
		{"a@b", false},
		{"", false},
		{"user@", false},
		{"@example.com", false},
	}
	for _, tt := range tests {
		t.Run(tt.email, func(t *testing.T) {
			if got := ValidateEmail(tt.email); got != tt.want {
				t.Errorf("ValidateEmail(%q) = %v, want %v", tt.email, got, tt.want)
			}
		})
	}
}

func TestValidatePassword(t *testing.T) {
	if !ValidatePassword("123456") {
		t.Error("6-char password should be valid")
	}
	if !ValidatePassword("longpassword123") {
		t.Error("long password should be valid")
	}
	if ValidatePassword("12345") {
		t.Error("5-char password should be invalid")
	}
	if ValidatePassword("") {
		t.Error("empty password should be invalid")
	}
}

func TestDecode_Valid(t *testing.T) {
	body := bytes.NewBufferString(`{"email":"x@example.com"}`)
	r := httptest.NewRequest("POST", "/", body)

	var out struct {
		Email string `json:"email"`
	}
	if err := Decode(r, &out); err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if out.Email != "x@example.com" {
		t.Errorf("got %q, want %q", out.Email, "x@example.com")
	}
}

func TestDecode_UnknownField(t *testing.T) {
	body := bytes.NewBufferString(`{"unknown_field":"value"}`)
	r := httptest.NewRequest("POST", "/", body)

	var out struct{}
	if err := Decode(r, &out); err == nil {
		t.Error("expected error for unknown field")
	}
}

func TestDecode_InvalidJSON(t *testing.T) {
	body := bytes.NewBufferString(`not json`)
	r := httptest.NewRequest("POST", "/", body)

	var out struct{}
	if err := Decode(r, &out); err == nil {
		t.Error("expected error for invalid JSON")
	}
}
