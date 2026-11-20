package infrastructure

import (
	"encoding/hex"
	"fmt"
	"strings"
)

func EncodeName(name string) string {
	return "/id_" + hex.EncodeToString([]byte(name))
}

func DecodeName(encoded string) (string, error) {
	value := strings.TrimPrefix(encoded, "/")
	if !strings.HasPrefix(value, "id_") {
		return "", fmt.Errorf("invalid encoded name %q", encoded)
	}

	raw, err := hex.DecodeString(strings.TrimPrefix(value, "id_"))
	if err != nil {
		return "", fmt.Errorf("decode name: %w", err)
	}
	return string(raw), nil
}
