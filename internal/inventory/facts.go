package inventory

import (
	"fmt"
	"strings"
)

// mangleName converts a Docker name into a Mangle name constant.
func mangleName(name string) string {
	if name == "" {
		return ""
	}
	return fmt.Sprintf("/id_%x", []byte(name))
}

// GenerateFacts converts observed Docker metadata into Mangle facts.
func GenerateFacts(containers []Container) (string, error) {
	var facts strings.Builder
	seenNetworks := make(map[string]bool)

	for _, container := range containers {
		name := mangleName(container.Name)
		if name == "" {
			return "", fmt.Errorf("container has an empty name")
		}

		fmt.Fprintf(&facts, "container(%s).\n", name)

		for _, network := range container.Networks {
			networkName := mangleName(network)
			if networkName == "" {
				return "", fmt.Errorf("container %q has an empty network name", container.Name)
			}

			if !seenNetworks[networkName] {
				fmt.Fprintf(&facts, "docker_network(%s).\n", networkName)
				seenNetworks[networkName] = true
			}
			fmt.Fprintf(&facts, "connected_to(%s, %s).\n", name, networkName)
		}
	}

	return facts.String(), nil
}
