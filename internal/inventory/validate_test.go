package inventory

import "testing"

func TestValidateFacts(t *testing.T) {
	containers := []Container{
		{
			Name:     "ai-qdrant",
			Image:    "qdrant/qdrant:v1.15.5",
			State:    "running",
			Networks: []string{"ai_default"},
		},
	}

	facts, err := GenerateFacts(containers)
	if err != nil {
		t.Fatal(err)
	}

	if err := ValidateFacts(facts); err != nil {
		t.Fatalf("valid inventory rejected: %v", err)
	}
}

func TestValidateFactsRejectsInvalidSyntax(t *testing.T) {
	if err := ValidateFacts("container("); err == nil {
		t.Fatal("expected invalid Mangle syntax to be rejected")
	}
}
