package core

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// TestScanFileCanceled verifies that ScanFile returns ctx.Err() when
// the context is already canceled before the read.
func TestScanFileCanceled(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), "c.txt")
	if err := os.WriteFile(tmp, []byte("AKIAIOSFODNN7EXAMPLE"), 0o600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, _, err := ScanFile(ctx, tmp)
	if err == nil {
		t.Error("expected ctx.Err() from canceled ScanFile")
	}
}

// TestScanDirCanceledBefore verifies that ScanDir honors a pre-canceled
// context during the walk phase.
func TestScanDirCanceledBefore(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, _, err := ScanDir(ctx, t.TempDir(), ScanOpts{})
	if err == nil {
		t.Error("expected ctx.Err() from canceled ScanDir")
	}
}

// TestScanStringCanceled verifies ScanString exits early when ctx is
// already canceled.
func TestScanStringCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	findings := ScanString(ctx, "AKIAIOSFODNN7EXAMPLE\nAKIAanother\n", "test")
	if len(findings) != 0 {
		t.Errorf("expected no findings on canceled context, got %d", len(findings))
	}
}

// TestScanEnvCanceled verifies ScanEnv exits early on canceled context.
func TestScanEnvCanceled(t *testing.T) {
	t.Setenv("ATHEON_TEST_SCAN_ENV_CANCELED", "AKIAIOSFODNN7EXAMPLE")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	findings := ScanEnv(ctx)
	if len(findings) != 0 {
		t.Errorf("expected no findings on canceled ScanEnv, got %d", len(findings))
	}
}

// TestScanLinesCanceledMidStream exercises the ctx check inside scanLines
// when the input has many lines and ctx is canceled before scanning
// completes.
func TestScanLinesCanceledMidStream(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	// Build content large enough that ctx cancellation is observable.
	content := ""
	for i := 0; i < 5000; i++ {
		content += "AKIAIOSFODNN7EXAMPLE\n"
	}
	cancel()
	findings := scanLines(ctx, content, "test")
	// Cancellation should stop the loop early; we expect fewer than 5000.
	if len(findings) >= 5000 {
		t.Errorf("expected early termination on canceled ctx, got %d findings", len(findings))
	}
}

// TestScanEnvCanceledMidStream exercises the ctx check inside scanEnv
// with a hand-crafted env list.
func TestScanEnvCanceledMidStream(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	envs := []string{
		"KEY1=AKIAIOSFODNN7EXAMPLE",
		"KEY2=AKIAanother",
		"KEY3=AKIAthird",
	}
	cancel()
	findings := scanEnv(ctx, envs)
	if len(findings) != 0 {
		t.Errorf("expected no findings on canceled scanEnv, got %d", len(findings))
	}
}

// TestDownloadBundleCanceled verifies DownloadBundle returns ctx.Err()
// when its context is canceled before the HTTP request.
func TestDownloadBundleCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := DownloadBundle(ctx, false)
	if err == nil {
		t.Error("expected error from canceled DownloadBundle")
	}
}
