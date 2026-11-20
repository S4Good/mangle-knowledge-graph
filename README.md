# Mangle Knowledge Graph

A small Go CLI that answers **what a technology depends on, what depends on it, and why**. The maintained [Mangle Go implementation](https://codeberg.org/TauCeti/mangle-go) evaluates recursive rules and produces proof trees; Go handles validated input and the command line.

Built as a focused portfolio project: no database server, API keys, generated answers, or external services at runtime.

## Try it

Requires Go 1.25 or newer. Dependencies are pinned in `go.mod` and `go.sum`; the first build requires internet access. Run from the repository root:

```sh
go test ./...
go build -o bin/kg ./cmd/kg
./bin/kg deps helm
./bin/kg impact linux
./bin/kg explain helm linux
./bin/kg -why impact networking
```

`deps helm` prints:

```text
depends_on(/helm,/containers)
depends_on(/helm,/docker)
depends_on(/helm,/kubernetes)
depends_on(/helm,/linux)
depends_on(/helm,/networking)
```

`impact linux` prints:

```text
impacted_by(/linux,/containers)
impacted_by(/linux,/docker)
impacted_by(/linux,/helm)
impacted_by(/linux,/kubernetes)
```

The explanation for Helm's Linux prerequisite traces the chain through Kubernetes, Docker, and containers. It includes the applied Mangle rules, variable bindings, and original `prerequisite` facts marked `[EDB]`. Explanations come from Mangle's provenance package, not from text templates or a second graph algorithm.

## What the project demonstrates

- Genuine Mangle name constants, typed predicate declarations, joins, recursive inference, and inequality.
- Transitive dependency closure evaluated to a fixed point by Mangle.
- Downstream impact expressed as a Mangle rule over that closure.
- Explainable results using upstream proof trees, including alternative derivations.
- A tested Go integration with embedded rules and data, strict input validation, stable query ordering, and CI.

The default data is an **illustrative learning graph**, not a claim that every team must learn these technologies in this order. For example, `/git` is intentionally isolated. Impact means a relationship follows from the supplied facts; it does not predict actual outages or runtime vulnerabilities.

## The logic

The complete rule set is in [`internal/graph/rules.mg`](internal/graph/rules.mg). An edge points from a technology to its prerequisite:

```prolog
prerequisite(/docker, /containers).
prerequisite(/containers, /linux).
```

The inference rules are:

```prolog
depends_on(T, P) :- prerequisite(T, P).
depends_on(T, P) :- prerequisite(T, Intermediate), depends_on(Intermediate, P).
impacted_by(P, T) :- depends_on(T, P), T != P.
```

This derives `depends_on(/docker,/linux)` and `impacted_by(/linux,/docker)` without storing either as an input fact.

## Commands

| Command | Result |
| --- | --- |
| `kg list` | All declared technologies, sorted |
| `kg deps TECHNOLOGY` | Direct and transitive prerequisites, sorted |
| `kg impact TECHNOLOGY` | All downstream dependents, sorted |
| `kg explain TECHNOLOGY PREREQUISITE` | Up to three dependency proofs |
| `kg -why deps TECHNOLOGY` | Prerequisites and their proofs |
| `kg -why impact TECHNOLOGY` | Downstream results and their proofs |

Use `-help` for usage. Flags precede the command. Both `helm` and `/helm` are accepted. Invalid arguments, unknown technologies, and invalid input exit with status 1. A valid query with no matches prints a clear message and exits successfully.

## Use your own graph

Save this as `my-graph.mg`:

```prolog
technology(/service).
technology(/database).
technology(/storage).
prerequisite(/service, /database).
prerequisite(/database, /storage).
```

```sh
./bin/kg -data my-graph.mg deps service
./bin/kg -data my-graph.mg -why impact storage
```

Custom data replaces the embedded sample. It must contain only ground `technology/1` and `prerequisite/2` facts with `/name` arguments. Declare both endpoints of every edge. Comments and duplicate facts are accepted. Rules, declarations, temporal annotations, unknown predicates, and variables are rejected.

Cycles are allowed: a nonempty path from a technology back to itself produces a self-dependency. Impact excludes the queried technology itself. An isolated technology is valid and returns no relationships. An empty file is a valid empty graph.

## Real-world infrastructure dependency analysis

This project also demonstrates how Google Mangle can analyze dependencies in a real infrastructure environment.

The application combines Docker inventory data, explicitly declared infrastructure resources, and evidence-backed dependency relationships.

### Architecture

```text
Docker inventory (JSONL)     External resources (JSON)
          |                           |
          +-------------+-------------+
                        |
              Dependencies (JSON)
                        |
                        v
                 Go validation
                        |
                        v
                 Mangle engine
                        |
             +----------+----------+
             |          |          |
             v          v          v
         Dependencies  Impact   Explanations
                        |
                        v
                    Go CLI
```

### Real-world HomeLab example

The project was tested against a Docker-based Ubuntu HomeLab and a separate MacBook running Ollama.

The following dependency chain was verified through application configuration, SSH forwarding, and API connectivity:

```text
Open WebUI (Docker / Ubuntu HomeLab)
             |
             v
SSH remote-forwarding tunnel
             |
             v
Ollama (MacBook)
```

The graph models configured dependencies, not guaranteed service failures.

### External resources

Infrastructure resources outside Docker can be declared in JSON:

```json
[
  {
    "id": "ollama-ssh-tunnel",
    "type": "ssh_tunnel",
    "host": "homelab"
  },
  {
    "id": "macbook-ollama",
    "type": "service",
    "host": "macbook"
  }
]
```

### Evidence-backed dependencies

Dependencies are declared separately from the Docker inventory:

```json
[
  {
    "service": "ai-open-webui",
    "requires": "ollama-ssh-tunnel",
    "evidence": "Configured Ollama endpoint uses SSH forwarding"
  },
  {
    "service": "ollama-ssh-tunnel",
    "requires": "macbook-ollama",
    "evidence": "SSH forwarding and Ollama API connectivity verified"
  }
]
```

The `ai-open-webui` container is supplied by the Docker inventory.

### Infrastructure CLI commands

Assuming the Docker inventory and JSON files are available locally:

```bash
go run ./cmd/kg -input containers.jsonl -resources resources.json -dependencies dependencies.json infra-list

go run ./cmd/kg -input containers.jsonl -resources resources.json -dependencies dependencies.json infra-deps ai-open-webui

go run ./cmd/kg -input containers.jsonl -resources resources.json -dependencies dependencies.json infra-impact macbook-ollama

go run ./cmd/kg -input containers.jsonl -resources resources.json -dependencies dependencies.json infra-explain ai-open-webui macbook-ollama
```

The verified dependency query returns:

```text
macbook-ollama
ollama-ssh-tunnel
```

The reverse-impact query for `macbook-ollama` returns:

```text
ai-open-webui
ollama-ssh-tunnel
```

These transitive relationships are inferred by Mangle rather than manually stored as additional dependency edges.

### Engineering considerations

- Docker inventory and dependency evidence are kept separate.
- Shared Docker networks do not automatically imply dependencies.
- Input validation rejects invalid resources and unknown dependency endpoints.
- Cross-host relationships can be represented without requiring all services to run on one machine.
- Mangle provenance provides explanations for inferred relationships.
- The analysis uses snapshots and does not continuously monitor service availability.
- Impact results represent potential dependency consequences, not confirmed outages.
- Sensitive infrastructure data should not be committed to public repositories.

This demonstrates practical declarative reasoning, Go integration, infrastructure modeling, and explainable dependency analysis.

## Verify

```sh
go mod verify
go vet ./...
go test -race -cover ./...
go build -o bin/kg ./cmd/kg
```

Tests cover the sample, all 64 directed three-node graphs without self-edges against an independent reachability oracle, explicit self-loops, cycles, duplicates, disconnected nodes, alternative proofs, supporting facts, custom files, and CLI errors. GitHub Actions runs formatting, module verification, vet, race tests, build, and CLI smoke checks on Linux. It runs when this project is published and pushed; no hosted run is claimed here.

See [architecture and tradeoffs](docs/architecture.md) and [documentation verification](docs/verification.md).

## Scope

This is an in-memory snapshot engine for small graphs, not a graph database. Each invocation reloads and evaluates the graph. The rules use a finite Datalog fragment; they do not generate new values recursively. Proof enumeration is capped at three alternatives and depth 64, with truncated proofs marked partial. There is no persistence, incremental update API, arbitrary rule execution, or shortest-path guarantee for displayed explanations.

## License

MIT for this project's code. Mangle and other dependencies retain their own licenses.

**For non-technical readers:** [What problem does this project solve?](README_NON_TECHNICAL.md)