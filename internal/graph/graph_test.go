package graph

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"codeberg.org/TauCeti/mangle-go/provenance"
)

func load(t *testing.T, facts string) *Graph {
	t.Helper()
	g, err := Load(strings.NewReader(facts))
	if err != nil {
		t.Fatal(err)
	}
	return g
}

func results(t *testing.T, g *Graph, mode, name string) []string {
	t.Helper()
	facts, err := g.Query(mode, name)
	if err != nil {
		t.Fatal(err)
	}
	got := []string{}
	for _, f := range facts {
		got = append(got, f.Args[1].String())
	}
	return got
}

func TestDemo(t *testing.T) {
	g := load(t, Demo)
	for _, tc := range []struct {
		mode, name string
		want       []string
	}{
		{"deps", "helm", []string{"/containers", "/docker", "/kubernetes", "/linux", "/networking"}},
		{"impact", "linux", []string{"/containers", "/docker", "/helm", "/kubernetes"}},
		{"deps", "git", []string{}}, {"impact", "helm", []string{}},
	} {
		t.Run(tc.mode+tc.name, func(t *testing.T) {
			if got := results(t, g, tc.mode, tc.name); !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
		})
	}
	if len(g.Technologies()) != 7 {
		t.Fatal(g.Technologies())
	}
}

// Independent reachability oracle over every directed, loop-free 3-node graph.
// Includes disconnected graphs and cycles.
func TestAllThreeNodeGraphs(t *testing.T) {
	edges := [][2]int{{0, 1}, {0, 2}, {1, 0}, {1, 2}, {2, 0}, {2, 1}}
	for mask := 0; mask < 64; mask++ {
		var data strings.Builder
		data.WriteString("technology(/a). technology(/b). technology(/c).\n")
		reach := [3][3]bool{}
		for bit, e := range edges {
			if mask&(1<<bit) != 0 {
				reach[e[0]][e[1]] = true
				fmt.Fprintf(&data, "prerequisite(/%c, /%c).\n", 'a'+e[0], 'a'+e[1])
			}
		}
		for k := 0; k < 3; k++ {
			for i := 0; i < 3; i++ {
				for j := 0; j < 3; j++ {
					reach[i][j] = reach[i][j] || (reach[i][k] && reach[k][j])
				}
			}
		}
		g := load(t, data.String())
		for i := 0; i < 3; i++ {
			for _, mode := range []string{"deps", "impact"} {
				want := []string{}
				for j := 0; j < 3; j++ {
					if (mode == "deps" && reach[i][j]) || (mode == "impact" && i != j && reach[j][i]) {
						want = append(want, fmt.Sprintf("/%c", 'a'+j))
					}
				}
				got := results(t, g, mode, fmt.Sprintf("/%c", 'a'+i))
				if !reflect.DeepEqual(got, want) {
					t.Fatalf("mask %d %s %d: got %v want %v", mask, mode, i, got, want)
				}
			}
		}
	}
}

func TestProofs(t *testing.T) {
	g := load(t, `technology(/a). technology(/b). technology(/c). technology(/d).
 prerequisite(/a,/b). prerequisite(/a,/c). prerequisite(/b,/d). prerequisite(/c,/d).`)
	proofs, err := g.Explain("deps", "a", "d")
	if err != nil {
		t.Fatal(err)
	}
	if len(proofs) != 2 {
		t.Fatalf("want two alternative proofs, got %d", len(proofs))
	}
	var check func(*provenance.ProofNode)
	check = func(p *provenance.ProofNode) {
		if p.Partial {
			t.Fatal("unexpected partial proof")
		}
		if p.Kind == provenance.KindEDB {
			if p.Fact.Predicate.Symbol != "prerequisite" || !g.store.Contains(p.Fact) {
				t.Fatalf("unsupported leaf: %v", p.Fact)
			}
		} else if p.Rule == nil || len(p.Bindings) == 0 || len(p.Premises) == 0 {
			t.Fatalf("missing rule evidence: %+v", p)
		}
		for _, child := range p.Premises {
			check(child)
		}
	}
	for _, p := range proofs {
		check(p)
	}
	impact, err := g.Explain("impact", "d", "a")
	if err != nil {
		t.Fatal(err)
	}
	if impact[0].Fact.Predicate.Symbol != "impacted_by" {
		t.Fatal("wrong impact proof")
	}
	if _, err := g.Explain("deps", "d", "a"); !errors.Is(err, provenance.ErrNoProof) {
		t.Fatalf("want no proof, got %v", err)
	}
}

func TestCyclesDuplicatesAndEmpty(t *testing.T) {
	g := load(t, `technology(/a). technology(/b).
 prerequisite(/a,/b). prerequisite(/b,/a). prerequisite(/a,/b).`)
	if got := results(t, g, "deps", "a"); !reflect.DeepEqual(got, []string{"/a", "/b"}) {
		t.Fatal(got)
	}
	if got := results(t, g, "impact", "a"); !reflect.DeepEqual(got, []string{"/b"}) {
		t.Fatal(got)
	}
	if _, err := g.Explain("deps", "a", "a"); err != nil {
		t.Fatal(err)
	}
	self := load(t, "technology(/a). prerequisite(/a,/a).")
	if len(results(t, self, "deps", "a")) != 1 || len(results(t, self, "impact", "a")) != 0 {
		t.Fatal("self loop semantics")
	}
	if len(load(t, "").Technologies()) != 0 {
		t.Fatal("empty graph")
	}
}

func TestRejectInvalidData(t *testing.T) {
	for _, data := range []string{
		"technology(/a)", "technology(X).", `technology("a").`,
		"technology(/a,/b).", "other(/a).", "depends_on(/a,/b).",
		"technology(/a). prerequisite(/a,/b).", "Decl technology(T).",
		"technology(/a). technology(X) :- technology(X).",
	} {
		t.Run(data, func(t *testing.T) {
			if _, err := Load(strings.NewReader(data)); err == nil {
				t.Fatal("accepted invalid data")
			}
		})
	}
	g := load(t, Demo)
	if _, err := g.Query("deps", "missing"); err == nil {
		t.Fatal("accepted unknown technology")
	}
	if _, err := g.Query("bad", "linux"); err == nil {
		t.Fatal("accepted unknown mode")
	}
	if _, err := g.Explain("deps", "linux", "missing"); err == nil {
		t.Fatal("accepted unknown proof target")
	}
}
