package cloudsync

import "myobj/src/pkg/cloudsync/internal"

// FormatOAuthAccessCredential serializes OAuth session credential (access|refresh).
func FormatOAuthAccessCredential(accessToken, refreshToken string) string {
	return internal.FormatOAuthAccessCredential(accessToken, refreshToken)
}

// ParseOAuthAccessCredential parses OAuth session credential.
func ParseOAuthAccessCredential(raw string) (accessToken, refreshToken string) {
	return internal.ParseOAuthAccessCredential(raw)
}
