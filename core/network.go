package core

import (
	"net/url"
	"strings"
)

// IsSensitiveURL checks if a URL pattern may contain sensitive data
func IsSensitiveURL(rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}

	// Check path for sensitive patterns
	sensitivePaths := []string{
		"/api/key",
		"/api/token",
		"/api/secret",
		"/admin",
		"/config",
		"/credentials",
		"/.env",
		"/token",
		"/auth",
		"/private",
		"/internal",
		"/v1/secrets",
		"/v2/secrets",
	}

	pathLower := strings.ToLower(u.Path)
	for _, pattern := range sensitivePaths {
		if strings.Contains(pathLower, pattern) {
			return true
		}
	}

	// Check host for sensitive patterns
	hostLower := strings.ToLower(u.Host)
	sensitiveHosts := []string{
		"secret",
		"api-key",
		"api_key",
		"credentials",
		"vault",
	}

	for _, pattern := range sensitiveHosts {
		if strings.Contains(hostLower, pattern) {
			return true
		}
	}

	return false
}
