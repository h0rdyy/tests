package handlers

import (
	"regexp"
	"strings"
)

var nonDigit = regexp.MustCompile(`\D`)

// normalizePhone converts common Russian phone formats to +7XXXXXXXXXX.
func normalizePhone(raw string) (string, bool) {
	raw = strings.TrimSpace(raw)
	digits := nonDigit.ReplaceAllString(raw, "")
	if len(digits) < 10 {
		return "", false
	}
	if len(digits) == 10 {
		return "+7" + digits, true
	}
	if len(digits) == 11 && digits[0] == '8' {
		return "+7" + digits[1:], true
	}
	if len(digits) == 11 && digits[0] == '7' {
		return "+" + digits, true
	}
	if strings.HasPrefix(raw, "+") {
		return "+" + digits, true
	}
	return "", false
}
