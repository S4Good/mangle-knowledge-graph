package inventory

import (
	"strings"
	"testing"
)

func TestValidateFactsRejectsUnexpectedInput(t *testing.T) {
	tests := []struct {
		name  string
		facts string
	}{
		{
			name:  "unknown predicate",
			facts: "unknown(/id_74657374).",
		},
		{
			name:  "wrong arity",
			facts: "container(/id_61, /id_62).",
		},
		{
			name:  "unexpected declaration",
			facts: "Decl unexpected(X) bound [/name].",
		},
		{
			name:  "rule instead of fact",
			facts: "container(X) :- docker_network(X).",
		},
		{
			name:  "string instead of name",
			facts: `container("test").`,
		},
		{
			name:  "variable instead of constant",
			facts: "container(X).",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateFacts(tt.facts); err == nil {
				t.Fatalf("expected rejection for %s", tt.name)
			}
		})
	}
}

func TestValidateFactsAcceptsGeneratedInventory(t *testing.T) {
	containers := []Container{
		{
			Name:     "ai-qdrant",
			Networks: []string{"ai_default"},
		},
		{
			Name:     "ai-open-webui",
			Networks: []string{"host"},
		},
	}

	facts, err := GenerateFacts(containers)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(facts, "connected_to(") {
		t.Fatal("expected network connection facts")
	}

	if err := ValidateFacts(facts); err != nil {
		t.Fatalf("generated inventory failed validation: %v", err)
	}
}
