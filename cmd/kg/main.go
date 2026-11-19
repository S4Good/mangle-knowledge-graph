package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"codeberg.org/TauCeti/mangle-go/provenance"
	"mangle-knowledge-graph/internal/graph"
)

const usage = `Usage: kg [-data facts.mg] [-why] <command> [technology] [prerequisite]

Commands:
  list                    List known technologies
  deps TECHNOLOGY         Direct and transitive prerequisites
  impact TECHNOLOGY       Technologies that depend on TECHNOLOGY
  explain TECH PREREQ     Explain a dependency with Mangle proof trees

Flags must precede the command. Names may include or omit the leading /.
Without -data, the embedded demonstration graph is used.
-why prints up to three proofs for every deps or impact result.
`

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(args []string, out, errOut io.Writer) error {
	fs := flag.NewFlagSet("kg", flag.ContinueOnError)
	fs.SetOutput(errOut)
	data := fs.String("data", "", "Mangle facts file (replaces demo)")
	why := fs.Bool("why", false, "explain each result")
	fs.Usage = func() { fmt.Fprint(errOut, usage) }
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	args = fs.Args()
	if len(args) == 0 {
		fs.Usage()
		return errors.New("command required")
	}
	want := map[string]int{"list": 1, "deps": 2, "impact": 2, "explain": 3}
	n, ok := want[args[0]]
	if !ok || len(args) != n {
		return errors.New("invalid command or argument count; use -help")
	}
	if *why && (args[0] == "list" || args[0] == "explain") {
		return errors.New("-why requires deps or impact")
	}
	var input io.Reader = strings.NewReader(graph.Demo)
	if *data != "" {
		f, err := os.Open(*data)
		if err != nil {
			return fmt.Errorf("open data: %w", err)
		}
		defer f.Close()
		input = f
	}
	g, err := graph.Load(input)
	if err != nil {
		return err
	}
	switch args[0] {
	case "list":
		for _, name := range g.Technologies() {
			if _, err := fmt.Fprintln(out, name); err != nil {
				return err
			}
		}
	case "explain":
		proofs, err := g.Explain("deps", args[1], args[2])
		if errors.Is(err, provenance.ErrNoProof) {
			_, err = fmt.Fprintln(out, "No dependency found in this graph.")
			return err
		}
		if err != nil {
			return err
		}
		return provenance.Print(out, proofs)
	default:
		facts, err := g.Query(args[0], args[1])
		if err != nil {
			return err
		}
		if len(facts) == 0 {
			_, err = fmt.Fprintln(out, "No results.")
			return err
		}
		for _, fact := range facts {
			if _, err := fmt.Fprintln(out, fact.String()); err != nil {
				return err
			}
			if *why {
				proofs, err := g.Explain(args[0], args[1], fact.Args[1].String())
				if err != nil {
					return err
				}
				if err := provenance.Print(out, proofs); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
