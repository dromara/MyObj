package internal

import "strings"

// ParseOAuthAccessCredential parses access_token|refresh_token credential string.
func ParseOAuthAccessCredential(raw string) (accessToken, refreshToken string) {
	parts := strings.SplitN(strings.TrimSpace(raw), "|", 2)
	if len(parts) == 0 {
		return "", ""
	}
	accessToken = parts[0]
	if len(parts) > 1 {
		refreshToken = parts[1]
	}
	return accessToken, refreshToken
}

// FormatOAuthAccessCredential serializes OAuth session credential.
func FormatOAuthAccessCredential(accessToken, refreshToken string) string {
	accessToken = strings.TrimSpace(accessToken)
	refreshToken = strings.TrimSpace(refreshToken)
	if accessToken == "" {
		return ""
	}
	if refreshToken != "" {
		return accessToken + "|" + refreshToken
	}
	return accessToken
}
