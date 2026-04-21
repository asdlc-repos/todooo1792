package validation

import (
	"errors"
	"net/mail"
	"strings"
)

const (
	MaxTitleLen       = 200
	MaxDescriptionLen = 2000
	MaxCategoryLen    = 50
	MinPasswordLen    = 8
	MaxPasswordLen    = 200
)

func Email(s string) (string, error) {
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "" {
		return "", errors.New("email is required")
	}
	if _, err := mail.ParseAddress(s); err != nil {
		return "", errors.New("invalid email")
	}
	return s, nil
}

func Password(s string) error {
	if len(s) < MinPasswordLen {
		return errors.New("password must be at least 8 characters")
	}
	if len(s) > MaxPasswordLen {
		return errors.New("password too long")
	}
	return nil
}

func Title(s string) (string, error) {
	s = strings.TrimSpace(Sanitize(s))
	if s == "" {
		return "", errors.New("title is required")
	}
	if len(s) > MaxTitleLen {
		return "", errors.New("title must be 200 characters or fewer")
	}
	return s, nil
}

func Description(s string) (string, error) {
	s = Sanitize(s)
	if len(s) > MaxDescriptionLen {
		return "", errors.New("description must be 2000 characters or fewer")
	}
	return s, nil
}

func CategoryName(s string) (string, error) {
	s = strings.TrimSpace(Sanitize(s))
	if s == "" {
		return "", errors.New("category name is required")
	}
	if len(s) > MaxCategoryLen {
		return "", errors.New("category name must be 50 characters or fewer")
	}
	return s, nil
}

// Sanitize removes control characters and trims the string.
func Sanitize(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if r == '\t' || r == '\n' || r == '\r' || r >= 0x20 {
			b.WriteRune(r)
		}
	}
	return b.String()
}
