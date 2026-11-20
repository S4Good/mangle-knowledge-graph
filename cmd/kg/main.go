package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"mangle-knowledge-graph/internal/graph"
	"mangle-knowledge-graph/internal/infrastructure"
	"mangle-knowledge-graph/internal/inventory"

	"codeberg.org/TauCeti/mangle-go/provenance"
)

const usage = `Usage: kg [flags] <command> [arguments]

Commands:
  inventory               Convert Docker CLI JSON Lines into Mangle facts
  infra-list              List infrastructure resources
  infra-deps RESOURCE     Show direct and transitive dependencies
  infra-impact RESOURCE   Show potentially affected resources
  infra-explain A B       Explain a dependency using Mangle proofs
  list                    List known technologies
  deps TECHNOLOGY         Direct and transitive prerequisites
  impact TECHNOLOGY       Technologies that depend on TECHNOLOGY
  explain TECH PREREQ     Explain a dependency with Mangle proof trees

Flags:
  -data FILE              Mangle facts file
  -why                    Show dependency proofs
  -input FILE              Docker CLI JSON Lines inventory
  -output FILE             Generated Mangle facts output
  -dependencies FILE       Dependency JSON file
  -resources FILE          Additional infrastructure resources JSON file

Inventory requires -input and -output.
Infrastructure commands require -dependencies and at least one of -input or -resources.
Flags must precede the command.
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
	inputPath := fs.String("input", "", "Docker CLI JSON Lines input")
	outputPath := fs.String("output", "", "Generated Mangle facts output")
	depsPath := fs.String("dependencies", "", "Verified dependency JSON file")
	resourcesPath := fs.String("resources", "", "Additional infrastructure resources JSON file")

	fs.Usage = func() {
		fmt.Fprint(errOut, usage)
	}

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

	// Infrastructure dependency analysis.
	if strings.HasPrefix(args[0], "infra-") {
		if *depsPath == "" || (*inputPath == "" && *resourcesPath == "") {
			return errors.New(
				"infrastructure commands require -dependencies and at least one of -input or -resources",
			)
		}

		if *data != "" || *outputPath != "" || *why {
			return errors.New(
				"infrastructure commands do not accept -data, -output or -why",
			)
		}

		expected := map[string]int{
			"infra-list":    1,
			"infra-deps":    2,
			"infra-impact":  2,
			"infra-explain": 3,
		}

		count, ok := expected[args[0]]
		if !ok || len(args) != count {
			return errors.New(
				"invalid infrastructure command or argument count",
			)
		}

		// Load Docker containers when an inventory is provided.
		var containers []inventory.Container

		if *inputPath != "" {
			raw, err := os.ReadFile(*inputPath)
			if err != nil {
				return fmt.Errorf(
					"read infrastructure inventory: %w",
					err,
				)
			}

			containers, err = inventory.ParseDockerCLI(raw)
			if err != nil {
				return err
			}
		}

		// Load additional infrastructure resources.
		var resources []infrastructure.Resource

		if *resourcesPath != "" {
			raw, err := os.ReadFile(*resourcesPath)
			if err != nil {
				return fmt.Errorf("read resources: %w", err)
			}

			resources, err = infrastructure.ParseResources(raw)
			if err != nil {
				return err
			}
		}

		if len(containers) == 0 && len(resources) == 0 {
			return errors.New(
				"infrastructure inventory contains no resources",
			)
		}

		// Load explicitly declared dependencies.
		depData, err := os.ReadFile(*depsPath)
		if err != nil {
			return fmt.Errorf("read dependencies: %w", err)
		}

		dependencies, err := infrastructure.ParseDependencies(depData)
		if err != nil {
			return err
		}

		// Generate Mangle facts from Docker and external resources.
		facts, err := infrastructure.BuildResourceGraphFacts(
			containers,
			resources,
			dependencies,
		)
		if err != nil {
			return err
		}

		g, err := graph.Load(strings.NewReader(facts))
		if err != nil {
			return err
		}

		switch args[0] {
		case "infra-list":
			for _, name := range g.Technologies() {
				decoded, err := infrastructure.DecodeName(name)
				if err != nil {
					decoded = name
				}

				if _, err := fmt.Fprintln(out, decoded); err != nil {
					return err
				}
			}
			return nil

		case "infra-explain":
			proofs, err := g.Explain(
				"deps",
				infrastructure.EncodeName(args[1]),
				infrastructure.EncodeName(args[2]),
			)

			if errors.Is(err, provenance.ErrNoProof) {
				_, err = fmt.Fprintln(
					out,
					"No verified dependency found.",
				)
				return err
			}

			if err != nil {
				return err
			}

			return provenance.Print(out, proofs)

		default:
			mode := strings.TrimPrefix(args[0], "infra-")

			results, err := g.Query(
				mode,
				infrastructure.EncodeName(args[1]),
			)
			if err != nil {
				return err
			}

			if len(results) == 0 {
				_, err = fmt.Fprintln(
					out,
					"No verified dependencies found.",
				)
				return err
			}

			for _, result := range results {
				output := result.String()

				if len(result.Args) == 2 {
					decoded, err := infrastructure.DecodeName(
						result.Args[1].String(),
					)
					if err == nil {
						output = decoded
					}
				}

				if _, err := fmt.Fprintln(out, output); err != nil {
					return err
				}
			}

			return nil
		}
	}

	// Docker inventory conversion.
	if args[0] == "inventory" {
		if len(args) != 1 || *inputPath == "" || *outputPath == "" {
			return errors.New(
				"inventory requires -input and -output",
			)
		}

		if *data != "" || *why || *depsPath != "" || *resourcesPath != "" {
			return errors.New(
				"inventory does not accept -data, -why, -dependencies or -resources",
			)
		}

		raw, err := os.ReadFile(*inputPath)
		if err != nil {
			return fmt.Errorf("read inventory: %w", err)
		}

		containers, err := inventory.ParseDockerCLI(raw)
		if err != nil {
			return err
		}

		if len(containers) == 0 {
			return errors.New(
				"inventory contains no containers",
			)
		}

		facts, err := inventory.GenerateFacts(containers)
		if err != nil {
			return err
		}

		if err := inventory.ValidateFacts(facts); err != nil {
			return fmt.Errorf(
				"validate generated inventory: %w",
				err,
			)
		}

		if err := os.WriteFile(
			*outputPath,
			[]byte(facts),
			0600,
		); err != nil {
			return fmt.Errorf("write inventory: %w", err)
		}

		_, err = fmt.Fprintf(
			out,
			"Generated Mangle facts for %d containers: %s\n",
			len(containers),
			*outputPath,
		)
		return err
	}

	// Reject infrastructure flags for standard graph commands.
	if *inputPath != "" ||
		*outputPath != "" ||
		*depsPath != "" ||
		*resourcesPath != "" {
		return errors.New(
			"-input, -output, -dependencies and -resources are not supported for this command",
		)
	}

	// Standard technology dependency analysis.
	want := map[string]int{
		"list":    1,
		"deps":    2,
		"impact":  2,
		"explain": 3,
	}

	n, ok := want[args[0]]
	if !ok || len(args) != n {
		return errors.New(
			"invalid command or argument count; use -help",
		)
	}

	if *why && (args[0] == "list" || args[0] == "explain") {
		return errors.New(
			"-why requires deps or impact",
		)
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
		proofs, err := g.Explain(
			"deps",
			args[1],
			args[2],
		)

		if errors.Is(err, provenance.ErrNoProof) {
			_, err = fmt.Fprintln(
				out,
				"No dependency found in this graph.",
			)
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
			if _, err := fmt.Fprintln(
				out,
				fact.String(),
			); err != nil {
				return err
			}

			if *why {
				proofs, err := g.Explain(
					args[0],
					args[1],
					fact.Args[1].String(),
				)
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
