package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/aliasfoxkde/Atheon/core"
)

// SecurityHeaderCheck checks a URL for missing security headers
type SecurityHeaderCheck struct {
	URL         string
	StatusCode  int
	Headers     http.Header
	Issues      []string
}

// SecurityIssue represents a detected security issue
type SecurityIssue struct {
	URL       string
	Issue     string
	Severity  string
	Recommendation string
}

// NetworkScanner scans URLs for security issues
type NetworkScanner struct {
	client *http.Client
}

// NewNetworkScanner creates a new network scanner
func NewNetworkScanner() *NetworkScanner {
	return &NetworkScanner{
		client: &http.Client{
			Timeout: 15 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		},
	}
}

// ScanURL checks a single URL for security issues
func (s *NetworkScanner) ScanURL(ctx context.Context, targetURL string) ([]SecurityIssue, error) {
	var issues []SecurityIssue

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check security headers
	issues = append(issues, checkSecurityHeaders(targetURL, resp.Header)...)

	// Check for exposed secrets in response (basic check)
	if core.IsSensitiveURL(targetURL) {
		issues = append(issues, SecurityIssue{
			URL:       targetURL,
			Issue:     "Potentially sensitive URL pattern detected",
			Severity:  "low",
			Recommendation: "Verify this URL does not expose sensitive data",
		})
	}

	return issues, nil
}

// checkSecurityHeaders checks for missing security headers
func checkSecurityHeaders(targetURL string, headers http.Header) []SecurityIssue {
	var issues []SecurityIssue

	// Required security headers (OWASP recommendations)
	required := map[string]string{
		"Strict-Transport-Security": "HSTS header missing - enforce HTTPS",
		"Content-Security-Policy":   "CSP header missing - prevents XSS/injection",
		"X-Content-Type-Options":    "X-Content-Type-Options missing",
		"X-Frame-Options":          "X-Frame-Options missing - prevents clickjacking",
		"X-XSS-Protection":         "X-XSS-Protection missing",
	}

	for header, recommendation := range required {
		if headers.Get(header) == "" {
			issues = append(issues, SecurityIssue{
				URL:       targetURL,
				Issue:     fmt.Sprintf("Missing security header: %s", header),
				Severity:  "medium",
				Recommendation: recommendation,
			})
		}
	}

	// Check for sensitive headers exposure
	if headers.Get("Authorization") != "" {
		issues = append(issues, SecurityIssue{
			URL:       targetURL,
			Issue:     "Authorization header present in response",
			Severity:  "high",
			Recommendation: "Ensure authorization is properly scoped",
		})
	}

	return issues
}

// ScanTargets scans multiple URLs
func (s *NetworkScanner) ScanTargets(ctx context.Context, targets []string) map[string][]SecurityIssue {
	results := make(map[string][]SecurityIssue)

	for _, target := range targets {
		issues, err := s.ScanURL(ctx, target)
		if err != nil {
			results[target] = []SecurityIssue{{
				URL:       target,
				Issue:     fmt.Sprintf("Scan failed: %v", err),
				Severity:  "error",
			}}
			continue
		}
		results[target] = issues
	}

	return results
}

// PrintReport prints a formatted security report
func PrintReport(results map[string][]SecurityIssue) {
	fmt.Print("\n=== Network Security Scan Report ===\n\n")

	for url, issues := range results {
		fmt.Printf("Target: %s\n", url)
		if len(issues) == 0 {
			fmt.Println("  ✓ No issues detected")
		} else {
			for _, issue := range issues {
				fmt.Printf("  [%s] %s\n", issue.Severity, issue.Issue)
				fmt.Printf("       → %s\n", issue.Recommendation)
			}
		}
		fmt.Println()
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Atheon Network Security Scanner")
		fmt.Println("\nUsage: atheon-network-scan <url1> [url2] ...")
		fmt.Println("\nExample: atheon-network-scan https://example.com https://api.example.com")
		os.Exit(1)
	}

	targets := os.Args[1:]
	scanner := NewNetworkScanner()
	ctx := context.Background()

	results := scanner.ScanTargets(ctx, targets)
	PrintReport(results)

	// Check if any issues found
	hasIssues := false
	for _, issues := range results {
		if len(issues) > 0 {
			hasIssues = true
			break
		}
	}

	if hasIssues {
		os.Exit(1)
	}
}
