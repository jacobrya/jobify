package hasher

import "testing"

func TestBcryptHasher_HashAndCompare(t *testing.T) {
	h := New()

	hash, err := h.Hash("mypassword")
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}
	if hash == "" {
		t.Fatal("expected non-empty hash")
	}
	if hash == "mypassword" {
		t.Fatal("hash must differ from plaintext")
	}

	if err := h.Compare("mypassword", hash); err != nil {
		t.Errorf("Compare correct password: %v", err)
	}
	if err := h.Compare("wrongpassword", hash); err == nil {
		t.Error("Compare wrong password: expected error")
	}
}

func TestBcryptHasher_SaltedDifferently(t *testing.T) {
	h := New()
	h1, _ := h.Hash("same")
	h2, _ := h.Hash("same")
	if h1 == h2 {
		t.Error("bcrypt must produce different hashes for same input due to random salt")
	}
}

func TestBcryptHasher_EmptyPassword(t *testing.T) {
	h := New()
	hash, err := h.Hash("")
	if err != nil {
		t.Fatalf("Hash empty: %v", err)
	}
	if err := h.Compare("", hash); err != nil {
		t.Errorf("Compare empty password: %v", err)
	}
	if err := h.Compare("notempty", hash); err == nil {
		t.Error("Compare non-empty against empty hash: expected error")
	}
}
