package inventory

import "testing"

func TestMangleNameCollision(t *testing.T) {
	first := mangleName("ai-qdrant")
	second := mangleName("ai_qdrant")

	if first == second {
		t.Fatalf(
			"different Docker names produced the same Mangle identifier: %s",
			first,
		)
	}
}
