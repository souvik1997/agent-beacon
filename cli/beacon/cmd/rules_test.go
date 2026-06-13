package cmd

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

const validRuleYAML = `
id: cli-test-rule
version: 1
title: CLI test rule
severity: low
status: experimental
posture: detect
match: 'e.event.action == "file.read"'
emit:
  reason: ok
tests:
  - name: p
    verdict: match
    events:
      - event: { action: file.read }
`

func writeRuleFile(t *testing.T, dir, name, body string) string {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func newCmd() (*cobra.Command, *bytes.Buffer) {
	buf := &bytes.Buffer{}
	c := &cobra.Command{}
	c.SetOut(buf)
	c.SetErr(buf)
	return c, buf
}

func TestRulesLintSuccess(t *testing.T) {
	dir := t.TempDir()
	writeRuleFile(t, dir, "ok.rule.yaml", validRuleYAML)
	cmd, buf := newCmd()
	if err := lintRulesPath(cmd, dir); err != nil {
		t.Fatalf("lint returned error: %v\n%s", err, buf.String())
	}
	if !strings.Contains(buf.String(), "ok   cli-test-rule") {
		t.Fatalf("expected ok line, got: %s", buf.String())
	}
}

func TestRulesLintFailsOnFixtureMismatch(t *testing.T) {
	dir := t.TempDir()
	bad := strings.Replace(validRuleYAML, "action: file.read }\n", "action: tool.invoked }\n", 1)
	writeRuleFile(t, dir, "bad.rule.yaml", bad)
	cmd, buf := newCmd()
	if err := lintRulesPath(cmd, dir); err == nil {
		t.Fatalf("expected lint failure; output:\n%s", buf.String())
	}
	if !strings.Contains(buf.String(), "FAIL") {
		t.Fatalf("expected FAIL line, got: %s", buf.String())
	}
}

func TestRulesLintEmptyDir(t *testing.T) {
	cmd, _ := newCmd()
	if err := lintRulesPath(cmd, t.TempDir()); err == nil {
		t.Fatalf("expected error for directory with no rule files")
	}
}

func makeTarGz(t *testing.T, entries map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	for name, body := range entries {
		if err := tw.WriteHeader(&tar.Header{Name: name, Mode: 0o644, Size: int64(len(body)), Typeflag: tar.TypeReg}); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write([]byte(body)); err != nil {
			t.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestExtractRuleTarballNeutralizesTraversal(t *testing.T) {
	dir := t.TempDir()
	// One legit entry, one path-traversal entry, and a nested-dir entry.
	data := makeTarGz(t, map[string]string{
		"ok.rule.yaml":                   "id: ok",
		"../../../../tmp/evil.rule.yaml": "id: evil",
		"pack/sub/nested.rule.yaml":      "id: nested",
		"notes.txt":                      "ignored",
	})
	if err := extractRuleTarball(bytes.NewReader(data), dir); err != nil {
		t.Fatalf("extract: %v", err)
	}
	// All .rule.yaml entries land flattened inside dir; nothing escapes.
	for _, name := range []string{"ok.rule.yaml", "evil.rule.yaml", "nested.rule.yaml"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Errorf("expected %s inside dest dir: %v", name, err)
		}
	}
	// The traversal target must NOT exist outside the dest dir.
	escaped := filepath.Join(filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(dir)))), "tmp", "evil.rule.yaml")
	if _, err := os.Stat(escaped); err == nil {
		t.Fatalf("traversal entry escaped the destination: %s", escaped)
	}
	// Non-rule files are ignored.
	if _, err := os.Stat(filepath.Join(dir, "notes.txt")); err == nil {
		t.Errorf("notes.txt should not be extracted")
	}
}

func TestLoadRuleFilesSingleFile(t *testing.T) {
	dir := t.TempDir()
	p := writeRuleFile(t, dir, "ok.rule.yaml", validRuleYAML)
	rules, err := loadRuleFiles(p)
	if err != nil {
		t.Fatalf("loadRuleFiles single: %v", err)
	}
	if len(rules) != 1 || rules[0].ID != "cli-test-rule" {
		t.Fatalf("unexpected: %+v", rules)
	}
}
