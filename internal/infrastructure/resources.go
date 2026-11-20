package infrastructure

import (
	"encoding/json"
	"fmt"
	"strings"

	"mangle-knowledge-graph/internal/inventory"
)

type Resource struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Host string `json:"host"`
}

func ParseResources(data []byte) ([]Resource, error) {
	var resources []Resource
	if err := json.Unmarshal(data, &resources); err != nil {
		return nil, fmt.Errorf("parse resources: %w", err)
	}
	return resources, nil
}

func BuildResourceGraphFacts(
	containers []inventory.Container,
	resources []Resource,
	dependencies []Dependency,
) (string, error) {
	known := make(map[string]bool)
	var facts strings.Builder

	add := func(id string) error {
		if strings.TrimSpace(id) == "" {
			return fmt.Errorf("resource has an empty ID")
		}
		if known[id] {
			return fmt.Errorf("duplicate resource %q", id)
		}
		known[id] = true
		fmt.Fprintf(&facts, "technology(/id_%x).\n", []byte(id))
		return nil
	}

	for _, c := range containers {
		if err := add(c.Name); err != nil {
			return "", err
		}
	}

	validTypes := map[string]bool{
		"container":  true,
		"service":    true,
		"ssh_tunnel": true,
	}

	for _, r := range resources {
		if !validTypes[r.Type] {
			return "", fmt.Errorf("invalid resource type %q", r.Type)
		}
		if strings.TrimSpace(r.Host) == "" {
			return "", fmt.Errorf("resource %q has an empty host", r.ID)
		}
		if err := add(r.ID); err != nil {
			return "", err
		}
	}

	for _, d := range dependencies {
		if !known[d.Service] || !known[d.Requires] {
			return "", fmt.Errorf(
				"dependency references unknown resource: %q -> %q",
				d.Service, d.Requires,
			)
		}
		if d.Service == d.Requires {
			return "", fmt.Errorf("self-dependency is not allowed: %q", d.Service)
		}
		if strings.TrimSpace(d.Evidence) == "" {
			return "", fmt.Errorf(
				"dependency %q -> %q requires evidence",
				d.Service, d.Requires,
			)
		}
		fmt.Fprintf(
			&facts,
			"prerequisite(/id_%x, /id_%x).\n",
			[]byte(d.Service),
			[]byte(d.Requires),
		)
	}

	return facts.String(), nil
}
