package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLI(t *testing.T) {
	for _, tc := range []struct {
		args     []string
		contains string
		fail     bool
	}{
		{[]string{"list"}, "/git\n", false},
		{[]string{"deps", "helm"}, "depends_on(/helm,/linux)", false},
		{[]string{"impact", "linux"}, "impacted_by(/linux,/helm)", false},
		{[]string{"-why", "impact", "linux"}, "[EDB]", false},
		{[]string{"explain", "helm", "linux"}, "by rule", false},
		{[]string{"explain", "linux", "helm"}, "No dependency found", false},
		{[]string{"deps", "git"}, "No results.", false},
		{[]string{"-help"}, "", false},
		{nil, "", true}, {[]string{"bad"}, "", true},
		{[]string{"deps"}, "", true}, {[]string{"list", "extra"}, "", true},
		{[]string{"deps", "missing"}, "", true},
		{[]string{"-why", "list"}, "", true},
		{[]string{"-bad"}, "", true},
		{[]string{"-data", "does-not-exist.mg", "list"}, "", true},
	} {
		t.Run(strings.Join(tc.args, " "), func(t *testing.T) {
			var out, stderr bytes.Buffer
			err := run(tc.args, &out, &stderr)
			if (err != nil) != tc.fail {
				t.Fatalf("err=%v output=%s", err, &out)
			}
			if !strings.Contains(out.String(), tc.contains) {
				t.Fatalf("missing %q in %s", tc.contains, &out)
			}
		})
	}
}

func TestCustomData(t *testing.T) {
	p := filepath.Join(t.TempDir(), "facts.mg")
	if err := os.WriteFile(p, []byte("technology(/custom)."), 0600); err != nil {
		t.Fatal(err)
	}
	var out, stderr bytes.Buffer
	if err := run([]string{"-data", p, "list"}, &out, &stderr); err != nil {
		t.Fatal(err)
	}
	if out.String() != "/custom\n" {
		t.Fatal(out.String())
	}
	if err := os.WriteFile(p, []byte("broken"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := run([]string{"-data", p, "list"}, &out, &stderr); err == nil {
		t.Fatal("accepted broken data")
	}
}
