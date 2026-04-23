# Contributing

## Prerequisites

This project uses [mise](https://mise.jdx.dev) to manage tool versions. Install it, then run:

```bash
mise install
```

This installs the pinned versions of Go, golangci-lint, task, goreleaser, and lefthook.

Set up git hooks:

```bash
lefthook install
```

This runs the linter before every commit and the tests before every push.

## Building

```bash
task build        # builds to bin/jfmt
go build ./cmd/jfmt  # or directly with go
```

## Testing

```bash
task test         # run all tests with race detection and coverage
task bench        # run benchmarks
task fuzz         # run fuzz tests for 30s
```

## Linting

```bash
task lint         # run golangci-lint
task lint-fix     # run golangci-lint with auto-fix
```

## Commit conventions

This project follows [Conventional Commits](https://www.conventionalcommits.org/). Every commit message must have the form:

```
<type>(<scope>): <subject>
```

Allowed types: `feat`, `fix`, `docs`, `refactor`, `test`, `chore`, `ci`, `build`, `perf`, `revert`.

Use imperative mood, keep the subject under 72 characters, and do not end it with a period. No commit body is required for straightforward changes.

Examples:

```
feat: add shell completion subcommand
fix: handle empty input on stdin correctly
docs: add contributing guide
ci: bump golangci-lint to 2.11.4
```

## Pull requests

- Keep PRs focused on a single concern.
- Make sure `task lint` and `task test` pass locally before opening a PR.
- Commit history should be clean: one logical change per commit.
- PR titles must also follow the Conventional Commits format, as they feed into the automated changelog.

## Use of AI tools

AI assistance is welcome, but you remain fully responsible for every line you submit. Reviewers will hold AI-assisted code to the same standard as hand-written code, and "the AI wrote it" is not an explanation for a bug or a shortcut past review.

Before opening a PR with AI-assisted code:

- Read and understand every line. If you cannot explain a change, do not submit it.
- Run the full test and lint suite yourself. Do not rely on the AI to verify correctness.
- Do not regenerate code wholesale in response to review feedback. Iterate on the specific concern raised.

Good uses: exploring unfamiliar APIs, drafting boilerplate, understanding existing code, generating test cases.

Avoid: generating finished logic you have not reviewed, using AI output to substitute for understanding a subsystem you are modifying.

## Code style

The linter enforces formatting via `gofumpt` and import ordering via `gci`. Run `task lint-fix` to apply automatic fixes. For anything the linter does not catch, prefer clarity over brevity and avoid unnecessary abstractions.
