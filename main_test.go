package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"testing"
)

func TestJSONFlagPrintsFindings(t *testing.T) {
	file := t.TempDir() + "/config.txt"
	if err := os.WriteFile(file, []byte("token=sk-abcdefghijklmnopqrstuvwxyz\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	out, err := exec.Command("go", "run", ".", "--json", "--file", file).Output()
	if err == nil || !json.Valid(out) || !bytes.Contains(out, []byte(`"pattern":"openai-api-key"`)) {
		t.Fatalf("unexpected output/error: %s / %v", out, err)
	}
}
