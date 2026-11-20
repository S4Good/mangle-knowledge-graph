package inventory

import "testing"

func TestParseDockerCLI(t *testing.T) {
	data := []byte(`{"ID":"abc123","Names":"ai-qdrant","Image":"qdrant/qdrant:v1.15.5","State":"running","Networks":"ai_default,bridge"}` + "\n")

	containers, err := ParseDockerCLI(data)
	if err != nil {
		t.Fatal(err)
	}

	if len(containers) != 1 {
		t.Fatalf("expected 1 container, got %d", len(containers))
	}

	got := containers[0]

	if got.ID != "abc123" {
		t.Errorf("unexpected ID: %s", got.ID)
	}

	if got.Name != "ai-qdrant" {
		t.Errorf("unexpected name: %s", got.Name)
	}

	if got.Image != "qdrant/qdrant:v1.15.5" {
		t.Errorf("unexpected image: %s", got.Image)
	}

	if got.State != "running" {
		t.Errorf("unexpected state: %s", got.State)
	}

	if len(got.Networks) != 2 ||
		got.Networks[0] != "ai_default" ||
		got.Networks[1] != "bridge" {
		t.Errorf("unexpected networks: %v", got.Networks)
	}
}
