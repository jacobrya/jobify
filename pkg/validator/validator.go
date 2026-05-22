package validator

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
)

var emailRe = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

func Decode(r *http.Request, dst any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return fmt.Errorf("invalid request body: %w", err)
	}
	return nil
}

func ValidateEmail(email string) bool {
	return emailRe.MatchString(email)
}

func ValidatePassword(password string) bool {
	return len(password) >= 6
}
