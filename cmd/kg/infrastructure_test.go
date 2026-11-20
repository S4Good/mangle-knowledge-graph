package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInfrastructureCLI(t *testing.T) {
	dir := t.TempDir()
	inventoryPath := filepath.Join(dir, "containers.jsonl")
	depsPath := filepath.Join(dir, "dependencies.json")

	inventoryData := `{"ID":"1","Names":"database","Image":"postgres","State":"running","Networks":"test"}
{"ID":"2","Names":"backend","Image":"backend","State":"running","Networks":"test"}
{"ID":"3","Names":"frontend","Image":"frontend","State":"running","Networks":"test"}
`

	depsData := `[
		{"service":"backend","requires":"database","evidence":"synthetic test"},
		{"service":"frontend","requires":"backend","evidence":"synthetic test"}
	]`

	if err := os.WriteFile(inventoryPath, []byte(inventoryData), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(depsPath, []byte(depsData), 0600); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name     string
		command  []string
		expected []string
	}{
		{
			name:     "impact",
			command:  []string{"infra-impact", "database"},
			expected: []string{"backend", "frontend"},
		},
		{
			name:     "dependencies",
			command:  []string{"infra-deps", "frontend"},
			expected: []string{"backend", "database"},
		},
		{
			name:     "explanation",
			command:  []string{"infra-explain", "frontend", "database"},
			expected: []string{"proof 1 of 1:", "prerequisite("},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := []string{
				"-input", inventoryPath,
				"-dependencies", depsPath,
			}
			args = append(args, tt.command...)

			var out, errOut bytes.Buffer
			if err := run(args, &out, &errOut); err != nil {
				t.Fatalf("run failed: %v; stderr: %s", err, errOut.String())
			}

			for _, expected := range tt.expected {
				if !strings.Contains(out.String(), expected) {
					t.Fatalf("expected %q in output:\n%s", expected, out.String())
				}
			}
		})
	}
}
