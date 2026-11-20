package infrastructure

import (
	"strings"
	"testing"

	"mangle-knowledge-graph/internal/graph"
	"mangle-knowledge-graph/internal/inventory"
)

func TestInfrastructureImpact(t *testing.T) {
	containers := []inventory.Container{
		{Name: "database"},
		{Name: "backend"},
		{Name: "frontend"},
	}

	deps := []Dependency{
		{Service: "backend", Requires: "database", Evidence: "test configuration"},
		{Service: "frontend", Requires: "backend", Evidence: "test configuration"},
	}

	facts, err := BuildGraphFacts(containers, deps)
	if err != nil {
		t.Fatal(err)
	}

	g, err := graph.Load(strings.NewReader(facts))
	if err != nil {
		t.Fatal(err)
	}

	affected, err := g.Query("impact", "id_6461746162617365")
	if err != nil {
		t.Fatal(err)
	}

	if len(affected) != 2 {
		t.Fatalf("expected 2 affected services, got %d: %v", len(affected), affected)
	}
}

func TestUnknownDependencyRejected(t *testing.T) {
	containers := []inventory.Container{{Name: "frontend"}}
	deps := []Dependency{
		{Service: "frontend", Requires: "missing-db", Evidence: "test"},
	}

	if _, err := BuildGraphFacts(containers, deps); err == nil {
		t.Fatal("expected unknown dependency rejection")
	}
}

func TestMissingEvidenceRejected(t *testing.T) {
	containers := []inventory.Container{
		{Name: "frontend"},
		{Name: "backend"},
	}

	deps := []Dependency{{Service: "frontend", Requires: "backend"}}

	if _, err := BuildGraphFacts(containers, deps); err == nil {
		t.Fatal("expected missing evidence rejection")
	}
}
