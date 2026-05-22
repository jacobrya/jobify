package jwt

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestManager_GenerateParse(t *testing.T) {
	m := NewManager("secret", time.Hour)
	uid := uuid.New()

	token, err := m.Generate(uid, "developer")
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}

	claims, err := m.Parse(token)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if claims.UserID != uid.String() {
		t.Errorf("user_id: got %q, want %q", claims.UserID, uid.String())
	}
	if claims.Role != "developer" {
		t.Errorf("role: got %q, want %q", claims.Role, "developer")
	}
}

func TestManager_Parse_InvalidToken(t *testing.T) {
	m := NewManager("secret", time.Hour)
	_, err := m.Parse("not.a.token")
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
}

func TestManager_Parse_WrongSecret(t *testing.T) {
	m1 := NewManager("secret1", time.Hour)
	m2 := NewManager("secret2", time.Hour)

	token, err := m1.Generate(uuid.New(), "admin")
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	_, err = m2.Parse(token)
	if err == nil {
		t.Fatal("expected error for wrong secret")
	}
}

func TestManager_Parse_Expired(t *testing.T) {
	m := NewManager("secret", -time.Second)
	token, err := m.Generate(uuid.New(), "developer")
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	_, err = m.Parse(token)
	if err == nil {
		t.Fatal("expected error for expired token")
	}
}

func TestManager_AdminRole(t *testing.T) {
	m := NewManager("supersecret", time.Hour)
	uid := uuid.New()

	token, err := m.Generate(uid, "admin")
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	claims, err := m.Parse(token)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if claims.Role != "admin" {
		t.Errorf("role: got %q, want %q", claims.Role, "admin")
	}
	if claims.UserID != uid.String() {
		t.Errorf("user_id: got %q, want %q", claims.UserID, uid.String())
	}
}
