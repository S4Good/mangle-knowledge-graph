package infrastructure

import (
	"encoding/json"
	"fmt"
	"strings"

	"mangle-knowledge-graph/internal/inventory"
)

type Dependency struct {
	Service  string `json:"service"`
	Requires string `json:"requires"`
	Evidence string `json:"evidence"`
}

func BuildGraphFacts(containers []inventory.Container, dependencies []Dependency) (string, error) {
	known := make(map[string]bool)
	var facts strings.Builder

	for _, c := range containers {
		if c.Name == "" {
			return "", fmt.Errorf("container has an empty name")
		}
		if known[c.Name] {
			return "", fmt.Errorf("duplicate container %q", c.Name)
		}
		known[c.Name] = true
		fmt.Fprintf(&facts, "technology(/id_%x).\n", []byte(c.Name))
	}

	for _, d := range dependencies {
		if !known[d.Service] || !known[d.Requires] {
			return "", fmt.Errorf("dependency references unknown container: %q -> %q", d.Service, d.Requires)
		}
		if d.Service == d.Requires {
			return "", fmt.Errorf("self-dependency is not allowed: %q", d.Service)
		}
		if strings.TrimSpace(d.Evidence) == "" {
			return "", fmt.Errorf("dependency %q -> %q requires evidence", d.Service, d.Requires)
		}

		fmt.Fprintf(&facts, "prerequisite(/id_%x, /id_%x).\n",
			[]byte(d.Service), []byte(d.Requires))
	}

	return facts.String(), nil
}

func ParseDependencies(data []byte) ([]Dependency, error) {
	var deps []Dependency
	if err := json.Unmarshal(data, &deps); err != nil {
		return nil, fmt.Errorf("parse dependencies: %w", err)
	}
	return deps, nil
}
