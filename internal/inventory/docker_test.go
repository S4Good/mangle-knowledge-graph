package inventory

import (
    "os"
    "testing"
)

func TestContainerModel(t *testing.T) {
    container := Container{
        ID:       "abc123",
        Name:     "ai-qdrant",
        Image:    "qdrant/qdrant:v1.15.5",
        State:    "running",
        Networks: []string{"ai_default"},
    }

    if container.Name != "ai-qdrant" {
        t.Errorf("unexpected container name: %s", container.Name)
    }

    if len(container.Networks) != 1 {
        t.Errorf("expected 1 network, got %d", len(container.Networks))
    }
}

func TestParseContainers(t *testing.T) {
    data, err := os.ReadFile("testdata/containers.json")
    if err != nil {
        t.Fatal(err)
    }

    containers, err := ParseContainers(data)
    if err != nil {
        t.Fatal(err)
    }

    if len(containers) != 1 {
        t.Fatalf("expected 1 container, got %d", len(containers))
    }

    got := containers[0]

    if got.Name != "ai-qdrant" {
        t.Errorf("unexpected name: %s", got.Name)
    }

    if got.Image != "qdrant/qdrant:v1.15.5" {
        t.Errorf("unexpected image: %s", got.Image)
    }

    if got.State != "running" {
        t.Errorf("unexpected state: %s", got.State)
    }

    if len(got.Networks) != 1 || got.Networks[0] != "ai_default" {
        t.Errorf("unexpected networks: %v", got.Networks)
    }
}
