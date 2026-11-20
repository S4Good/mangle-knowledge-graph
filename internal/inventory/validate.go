package inventory

import (
	_ "embed"
	"fmt"
	"reflect"

	"codeberg.org/TauCeti/mangle-go/ast"
	"strings"

	"codeberg.org/TauCeti/mangle-go/analysis"
	"codeberg.org/TauCeti/mangle-go/engine"
	"codeberg.org/TauCeti/mangle-go/factstore"
	"codeberg.org/TauCeti/mangle-go/parse"
)

//go:embed rules.mg
var inventoryRules string

// ValidateFacts parses, analyzes and evaluates generated inventory facts.
func ValidateFacts(facts string) error {
	// Parse input facts separately so they cannot introduce new rules
	// or declarations into the trusted inventory program.
	factUnit, err := parse.Unit(strings.NewReader(facts))
	if err != nil {
		return fmt.Errorf("parse inventory facts: %w", err)
	}

	ruleUnit, err := parse.Unit(strings.NewReader(inventoryRules))
	if err != nil {
		return fmt.Errorf("parse inventory declarations: %w", err)
	}

	if !reflect.DeepEqual(factUnit.Decls, ruleUnit.Decls[:1]) {
		return fmt.Errorf("inventory input must contain facts only")
	}

	allowed := map[string]int{
		"container":      1,
		"docker_network": 1,
		"connected_to":   2,
	}

	for i, clause := range factUnit.Clauses {
		if clause.Premises != nil || clause.Transform != nil || clause.HeadTime != nil {
			return fmt.Errorf("inventory entry %d must be a fact", i+1)
		}

		arity, ok := allowed[clause.Head.Predicate.Symbol]
		if !ok || len(clause.Head.Args) != arity {
			return fmt.Errorf("inventory entry %d has an unexpected predicate or arity", i+1)
		}

		for _, arg := range clause.Head.Args {
			constant, ok := arg.(ast.Constant)
			if !ok || constant.Type != ast.NameType {
				return fmt.Errorf("inventory entry %d requires ground name constants", i+1)
			}
		}
	}

	ruleUnit.Clauses = append(ruleUnit.Clauses, factUnit.Clauses...)
	unit := ruleUnit
	if err != nil {
		return fmt.Errorf("parse inventory facts: %w", err)
	}

	program, err := analysis.AnalyzeOneUnit(unit, nil)
	if err != nil {
		return fmt.Errorf("analyze inventory facts: %w", err)
	}

	store := factstore.NewSimpleInMemoryStore()

	if err := engine.EvalProgram(program, store); err != nil {
		return fmt.Errorf("evaluate inventory facts: %w", err)
	}

	return nil
}
