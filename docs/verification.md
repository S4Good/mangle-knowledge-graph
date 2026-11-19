# Verification record

Verified locally on 2026-11-19. This records observed checks, not a claim of a hosted GitHub Actions run.

## Upstream selection

The direct dependency is **`codeberg.org/TauCeti/mangle-go v0.7.0`**, pinned in `go.mod` with module checksums in `go.sum`. The Go module proxy identifies the tag with commit `7cd2543ac1817f8b92d42efe3c6ef82dededbeeb`. The maintained repository's README identifies Codeberg as its home and links the official documentation.

## Official references checked

| Requirement | Official reference | Implementation |
| --- | --- | --- |
| Facts, variables, `:-`, joins and recursive rules | [Getting to know Datalog](https://mangle.readthedocs.io/en/latest/datalog.html) | `rules.mg`: base and recursive dependency rules |
| `/name` constants | [Basic types](https://mangle.readthedocs.io/en/latest/basictypes.html) and tagged `ast/ast.go` | Named technologies; AST validation |
| Predicate declarations | [Declarations](https://mangle.readthedocs.io/en/latest/declarations.html) and tagged documentation | `Decl` statements with `/name` bounds |
| Parsing, analysis and evaluation | [v0.7.0 parser](https://codeberg.org/TauCeti/mangle-go/src/tag/v0.7.0/parse/parse.go), [analyzer](https://codeberg.org/TauCeti/mangle-go/src/tag/v0.7.0/analysis/validation.go), [evaluator](https://codeberg.org/TauCeti/mangle-go/src/tag/v0.7.0/engine/seminaivebottomup.go) | `parse.Unit`, `analysis.AnalyzeOneUnit`, `engine.EvalProgram` |
| Explainable recursion | [Provenance documentation](https://mangle.readthedocs.io/en/latest/provenance.html) and [tagged Go API](https://codeberg.org/TauCeti/mangle-go/src/tag/v0.7.0/provenance/provenance.go) | `provenance.Explain` and `provenance.Print` |

The Datalog and provenance web pages were read live. API signatures, declaration syntax, name types, and the evaluation pattern were additionally checked in the downloaded v0.7.0 module's source, documentation, and upstream tests. Tagged source is the version-specific authority when the live documentation changes.

## Results

- Go 1.27.4 on macOS arm64: `go test ./...`, `go test -race -cover ./...`, `go vet ./...`, `go mod verify`, and binary build passed.
- Minimum supported Go 1.25.0 on macOS arm64: race tests passed. The initial static-analysis process exited with status 137; `GOMAXPROCS=2 go vet -p=1 ./...` then passed.
- Statement coverage: CLI **86.7%**, graph package **93.2%**.
- All 64 directed three-node graphs without self-edges matched an independent transitive-closure oracle for both dependency and impact queries.
- Sample CLI results matched the README. An actual proof traced Helm to Linux through Kubernetes, Docker, and containers.
- The compiled binary listed embedded technologies successfully from `/tmp`, outside the repository.
- The first test run exposed the parser's implicit unnamed package declaration. Input validation was corrected to permit only that default declaration, and the test suite passed afterward.

The GitHub Actions workflow is prepared but has not run on GitHub. Publication requires the owner's approval.
