package httpx

import "strings"

func Required(value string) bool {
	return strings.TrimSpace(value) != ""
}

// ValidEmail does a basic sanity check: something@something.something
func ValidEmail(email string) bool {
	at := strings.Index(email, "@")
	if at < 1 {
		return false
	}
	domain := email[at+1:]
	dot := strings.LastIndex(domain, ".")
	return dot >= 1 && dot < len(domain)-1
}

// ValidPassword enforces a minimum length of 8 and a maximum of 72 characters.
// The 72-character cap prevents bcrypt from doing expensive work on oversized inputs.
func ValidPassword(password string) (ok bool, reason string) {
	if len(password) < 8 {
		return false, "password must be at least 8 characters"
	}
	if len(password) > 72 {
		return false, "password must not exceed 72 characters"
	}
	return true, ""
}
