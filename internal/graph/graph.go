// Package graph evaluates a small technology knowledge graph with Mangle.
package graph

import (
	_ "embed"
	"fmt"
	"io"
	"reflect"
	"sort"
	"strings"

	"codeberg.org/TauCeti/mangle-go/analysis"
	"codeberg.org/TauCeti/mangle-go/ast"
	"codeberg.org/TauCeti/mangle-go/engine"
	"codeberg.org/TauCeti/mangle-go/factstore"
	"codeberg.org/TauCeti/mangle-go/parse"
	"codeberg.org/TauCeti/mangle-go/provenance"
)

//go:embed rules.mg
var rules string

//go:embed technologies.mg
var Demo string

// Graph is an evaluated snapshot. Construct a new graph to change its facts.
type Graph struct {
	program      *analysis.ProgramInfo
	store        factstore.SimpleInMemoryStore
	technologies map[string]bool
}

// Load accepts only technology/1 and prerequisite/2 ground name facts.
// Keeping rules fixed preserves the meaning of queries and their explanations.
func Load(r io.Reader) (*Graph, error) {
	unit, err := parse.Unit(r)
	if err != nil {
		return nil, fmt.Errorf("parse facts: %w", err)
	}
	ruleUnit, err := parse.Unit(strings.NewReader(rules))
	if err != nil {
		return nil, fmt.Errorf("parse embedded rules: %w", err)
	}
	// The parser inserts an unnamed package declaration in every source unit.
	if !reflect.DeepEqual(unit.Decls, ruleUnit.Decls[:1]) {
		return nil, fmt.Errorf("data must contain facts only, not declarations")
	}
	names := map[string]bool{}
	for i, c := range unit.Clauses {
		if c.Premises != nil || c.Transform != nil || c.HeadTime != nil {
			return nil, fmt.Errorf("fact %d: rules and temporal annotations are not allowed", i+1)
		}
		want := 0
		switch c.Head.Predicate.Symbol {
		case "technology":
			want = 1
		case "prerequisite":
			want = 2
		}
		if want == 0 || len(c.Head.Args) != want {
			return nil, fmt.Errorf("fact %d: expected technology/1 or prerequisite/2", i+1)
		}
		for _, arg := range c.Head.Args {
			v, ok := arg.(ast.Constant)
			if !ok || v.Type != ast.NameType {
				return nil, fmt.Errorf("fact %d: arguments must be ground /name constants", i+1)
			}
		}
		if want == 1 {
			names[c.Head.Args[0].(ast.Constant).Symbol] = true
		}
	}
	for _, c := range unit.Clauses {
		if c.Head.Predicate.Symbol == "prerequisite" {
			for _, arg := range c.Head.Args {
				if !names[arg.(ast.Constant).Symbol] {
					return nil, fmt.Errorf("undeclared technology %s in %s", arg, c.Head)
				}
			}
		}
	}
	ruleUnit.Clauses = append(ruleUnit.Clauses, unit.Clauses...)
	program, err := analysis.AnalyzeOneUnit(ruleUnit, nil)
	if err != nil {
		return nil, fmt.Errorf("analyze: %w", err)
	}
	store := factstore.NewSimpleInMemoryStore()
	if err := engine.EvalProgram(program, store); err != nil {
		return nil, fmt.Errorf("evaluate: %w", err)
	}
	return &Graph{program: program, store: store, technologies: names}, nil
}

func (g *Graph) Technologies() []string {
	result := make([]string, 0, len(g.technologies))
	for name := range g.technologies {
		result = append(result, name)
	}
	sort.Strings(result)
	return result
}

func (g *Graph) name(s string) (ast.Constant, error) {
	if !strings.HasPrefix(s, "/") {
		s = "/" + s
	}
	if !g.technologies[s] {
		return ast.Constant{}, fmt.Errorf("unknown technology %q (use list)", s)
	}
	return ast.Name(s)
}

// Query returns sorted Mangle-derived facts. Dependencies include self only
// when a nonempty cycle exists; impact always excludes the source itself.
func (g *Graph) Query(mode, technology string) ([]ast.Atom, error) {
	pred, err := predicate(mode)
	if err != nil {
		return nil, err
	}
	name, err := g.name(technology)
	if err != nil {
		return nil, err
	}
	result := []ast.Atom{}
	err = g.store.GetFacts(ast.NewAtom(pred, name, ast.Variable{Symbol: "Result"}), func(a ast.Atom) error {
		result = append(result, a)
		return nil
	})
	sort.Slice(result, func(i, j int) bool { return result[i].String() < result[j].String() })
	return result, err
}

func predicate(mode string) (string, error) {
	switch mode {
	case "deps":
		return "depends_on", nil
	case "impact":
		return "impacted_by", nil
	}
	return "", fmt.Errorf("unknown query %q: use deps or impact", mode)
}

// Explain uses upstream provenance, including variable bindings and EDB leaves.
// At most three alternatives and 64 levels are returned; partial trees are marked.
func (g *Graph) Explain(mode, source, target string) ([]*provenance.ProofNode, error) {
	pred, err := predicate(mode)
	if err != nil {
		return nil, err
	}
	a, err := g.name(source)
	if err != nil {
		return nil, err
	}
	b, err := g.name(target)
	if err != nil {
		return nil, err
	}
	return provenance.Explain(g.program, g.store, ast.NewAtom(pred, a, b), provenance.Options{MaxProofs: 3, MaxDepth: 64})
}
