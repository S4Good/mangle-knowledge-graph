package infrastructure

import (
	"strings"
	"testing"
)

func TestParseResources(t *testing.T) {
	data := []byte(`[
		{"id":"ollama-ssh-tunnel","type":"ssh_tunnel","host":"homelab"},
		{"id":"macbook-ollama","type":"service","host":"macbook"}
	]`)

	resources, err := ParseResources(data)
	if err != nil {
		t.Fatalf("ParseResources failed: %v", err)
	}

	if len(resources) != 2 {
		t.Fatalf("expected 2 resources, got %d", len(resources))
	}

	if resources[0].Type != "ssh_tunnel" {
		t.Errorf("unexpected type: %q", resources[0].Type)
	}

	if resources[1].Host != "macbook" {
		t.Errorf("unexpected host: %q", resources[1].Host)
	}
}

func TestParseResourcesInvalidJSON(t *testing.T) {
	_, err := ParseResources([]byte(`{invalid`))
	if err == nil {
		t.Fatal("expected invalid JSON error")
	}
}

func TestBuildResourceGraphFactsCrossHost(t *testing.T) {
	resources := []Resource{
		{ID: "ai-open-webui", Type: "container", Host: "homelab"},
		{ID: "ollama-ssh-tunnel", Type: "ssh_tunnel", Host: "homelab"},
		{ID: "macbook-ollama", Type: "service", Host: "macbook"},
	}

	dependencies := []Dependency{
		{
			Service:  "ai-open-webui",
			Requires: "ollama-ssh-tunnel",
			Evidence: "Configured Ollama endpoint",
		},
		{
			Service:  "ollama-ssh-tunnel",
			Requires: "macbook-ollama",
			Evidence: "Verified SSH remote forwarding",
		},
	}

	facts, err := BuildResourceGraphFacts(nil, resources, dependencies)
	if err != nil {
		t.Fatalf("BuildResourceGraphFacts failed: %v", err)
	}

	expected := []string{
		"technology(" + EncodeName("ai-open-webui") + ").",
		"technology(" + EncodeName("ollama-ssh-tunnel") + ").",
		"technology(" + EncodeName("macbook-ollama") + ").",
		"prerequisite(" + EncodeName("ai-open-webui") + ", " +
			EncodeName("ollama-ssh-tunnel") + ").",
		"prerequisite(" + EncodeName("ollama-ssh-tunnel") + ", " +
			EncodeName("macbook-ollama") + ").",
	}

	for _, want := range expected {
		if !strings.Contains(facts, want) {
			t.Errorf("missing fact %q", want)
		}
	}
}

func TestBuildResourceGraphFactsValidation(t *testing.T) {
	valid := []Resource{
		{ID: "service-a", Type: "service", Host: "homelab"},
		{ID: "service-b", Type: "service", Host: "macbook"},
	}

	tests := []struct {
		name         string
		resources    []Resource
		dependencies []Dependency
	}{
		{
			name: "duplicate resource",
			resources: []Resource{
				valid[0],
				valid[0],
			},
		},
		{
			name: "invalid resource type",
			resources: []Resource{
				{ID: "invalid", Type: "unknown", Host: "homelab"},
			},
		},
		{
			name: "empty host",
			resources: []Resource{
				{ID: "invalid", Type: "service", Host: ""},
			},
		},
		{
			name:      "unknown dependency",
			resources: valid,
			dependencies: []Dependency{
				{
					Service:  "service-a",
					Requires: "missing-service",
					Evidence: "Test evidence",
				},
			},
		},
		{
			name:      "self dependency",
			resources: valid,
			dependencies: []Dependency{
				{
					Service:  "service-a",
					Requires: "service-a",
					Evidence: "Test evidence",
				},
			},
		},
		{
			name:      "missing evidence",
			resources: valid,
			dependencies: []Dependency{
				{
					Service:  "service-a",
					Requires: "service-b",
					Evidence: "",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := BuildResourceGraphFacts(
				nil,
				tt.resources,
				tt.dependencies,
			)
			if err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}
