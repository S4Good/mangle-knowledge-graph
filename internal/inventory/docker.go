package inventory

import (
    "encoding/json"
    "fmt"
    "sort"
    "strings"
)

// Container represents selected, non-sensitive Docker inventory data.
type Container struct {
    ID       string
    Name     string
    Image    string
    State    string
    Networks []string
}

// ParseContainers converts a Docker container-list API response
// into inventory records.
func ParseContainers(data []byte) ([]Container, error) {
    var raw []struct {
        ID    string   `json:"Id"`
        Names []string `json:"Names"`
        Image string   `json:"Image"`
        State string   `json:"State"`
        NetworkSettings struct {
            Networks map[string]json.RawMessage `json:"Networks"`
        } `json:"NetworkSettings"`
    }

    if err := json.Unmarshal(data, &raw); err != nil {
        return nil, fmt.Errorf("parse Docker containers: %w", err)
    }

    containers := make([]Container, 0, len(raw))

    for _, item := range raw {
        name := ""
        if len(item.Names) > 0 {
            name = strings.TrimPrefix(item.Names[0], "/")
        }

        networks := make([]string, 0, len(item.NetworkSettings.Networks))
        for network := range item.NetworkSettings.Networks {
            networks = append(networks, network)
        }
        sort.Strings(networks)

        containers = append(containers, Container{
            ID:       item.ID,
            Name:     name,
            Image:    item.Image,
            State:    item.State,
            Networks: networks,
        })
    }

    return containers, nil
}
