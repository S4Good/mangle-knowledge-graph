package infrastructure

import (
	"strings"
	"testing"
)

func TestNameRoundTrip(t *testing.T) {
	names := []string{
		"ai-qdrant",
		"ai-open-webui",
		"lume-postgis",
		"container_with_underscores",
	}

	for _, name := range names {
		encoded := EncodeName(name)

		decoded, err := DecodeName(encoded)
		if err != nil {
			t.Fatalf("decode %q: %v", name, err)
		}

		if decoded != name {
			t.Fatalf("expected %q, got %q", name, decoded)
		}
	}
}

func TestInvalidEncodedName(t *testing.T) {
	_, err := DecodeName("/invalid")
	if err == nil {
		t.Fatal("expected invalid name error")
	}
}

func TestDistinctNamesRemainDistinct(t *testing.T) {
	a := EncodeName("ai-qdrant")
	b := EncodeName("ai_qdrant")

	if a == b {
		t.Fatal("different Docker names produced identical identifiers")
	}

	if !strings.HasPrefix(a, "/id_") {
		t.Fatal("expected Mangle name identifier")
	}
}
