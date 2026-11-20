package inventory

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// ParseDockerCLI parses Docker's JSON Lines listing format.
func ParseDockerCLI(data []byte) ([]Container, error) {
	var containers []Container

	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(make([]byte, 64*1024), 10*1024*1024)

	for scanner.Scan() {
		if len(bytes.TrimSpace(scanner.Bytes())) == 0 {
			continue
		}

		var item struct {
			ID       string `json:"ID"`
			Names    string `json:"Names"`
			Image    string `json:"Image"`
			State    string `json:"State"`
			Networks string `json:"Networks"`
		}

		if err := json.Unmarshal(scanner.Bytes(), &item); err != nil {
			return nil, fmt.Errorf("parse Docker CLI line: %w", err)
		}

		var networks []string
		for _, network := range strings.Split(item.Networks, ",") {
			network = strings.TrimSpace(network)
			if network != "" {
				networks = append(networks, network)
			}
		}
		sort.Strings(networks)

		containers = append(containers, Container{
			ID:       item.ID,
			Name:     item.Names,
			Image:    item.Image,
			State:    item.State,
			Networks: networks,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read Docker CLI output: %w", err)
	}

	return containers, nil
}
