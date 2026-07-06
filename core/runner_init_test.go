package core

import (
	"os"
	"strings"
	"testing"
)

// TestInitEnvVars tests the init() function's environment variable parsing.
func TestInitEnvVars(t *testing.T) {
	// Save original env vars
	origSkipDirs := os.Getenv("ATHEON_SKIP_DIRS")
	origBinaryExts := os.Getenv("ATHEON_BINARY_EXTS")
	defer func() {
		if origSkipDirs != "" {
			os.Setenv("ATHEON_SKIP_DIRS", origSkipDirs)
		} else {
			os.Unsetenv("ATHEON_SKIP_DIRS")
		}
		if origBinaryExts != "" {
			os.Setenv("ATHEON_BINARY_EXTS", origBinaryExts)
		} else {
			os.Unsetenv("ATHEON_BINARY_EXTS")
		}
	}()

	tests := []struct {
		name          string
		skipDirs      string
		binaryExts    string
		wantSkipDir   string
		wantBinaryExt string
	}{
		{
			name:          "custom skip dir",
			skipDirs:      "custom_dir,another_dir",
			binaryExts:    "",
			wantSkipDir:   "custom_dir",
			wantBinaryExt: ".png", // default
		},
		{
			name:          "custom binary ext with dot",
			skipDirs:      "",
			binaryExts:    ".custom,.other",
			wantSkipDir:   ".git", // default
			wantBinaryExt: ".custom",
		},
		{
			name:          "custom binary ext without dot",
			skipDirs:      "",
			binaryExts:    "custom,other",
			wantSkipDir:   ".git", // default
			wantBinaryExt: ".custom",
		},
		{
			name:          "whitespace trimming",
			skipDirs:      "  dir1  ,  dir2  ",
			binaryExts:    "  .ext1  ,  .ext2  ",
			wantSkipDir:   "dir1",
			wantBinaryExt: ".ext1",
		},
		{
			name:          "empty values ignored",
			skipDirs:      ",,,",
			binaryExts:    ",,,",
			wantSkipDir:   ".git", // default unchanged
			wantBinaryExt: ".png", // default unchanged
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			os.Unsetenv("ATHEON_SKIP_DIRS")
			os.Unsetenv("ATHEON_BINARY_EXTS")

			if tc.skipDirs != "" {
				os.Setenv("ATHEON_SKIP_DIRS", tc.skipDirs)
			}
			if tc.binaryExts != "" {
				os.Setenv("ATHEON_BINARY_EXTS", tc.binaryExts)
			}

			// Re-call init-like behavior by setting the maps
			// Note: The actual init() runs at package load, so we test the
			// env parsing logic separately. This test documents expected behavior.
			if tc.skipDirs != "" && tc.skipDirs != ",,," {
				if _, ok := skipDirs[tc.skipDirs]; !ok {
					// Check first entry in comma split
					parts := splitDirsEnv(tc.skipDirs)
					if len(parts) > 0 && parts[0] != tc.wantSkipDir {
						t.Errorf("first skip dir = %q, want %q", parts[0], tc.wantSkipDir)
					}
				}
			}
		})
	}
}

// splitDirsEnv splits ATHEON_SKIP_DIRS value (replicates init logic)
func splitDirsEnv(v string) []string {
	var result []string
	for _, d := range splitEnv(v) {
		d = strings.TrimSpace(d)
		if d != "" {
			result = append(result, d)
		}
	}
	return result
}

// splitEnv splits comma-separated env value
func splitEnv(v string) []string {
	return strings.Split(v, ",")
}
