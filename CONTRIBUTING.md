# Contributing to Seocheon

Thank you for your interest in contributing to Seocheon! This document provides guidelines and information for contributors.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Coding Conventions](#coding-conventions)
- [Commit Messages](#commit-messages)
- [Pull Request Process](#pull-request-process)
- [Testing Requirements](#testing-requirements)
- [Reporting Issues](#reporting-issues)

## Code of Conduct

By participating in this project, you agree to maintain a respectful and inclusive environment for everyone.

## Getting Started

1. Fork the repository
2. Clone your fork: `git clone https://github.com/YOUR_USERNAME/seocheon.git`
3. Add upstream remote: `git remote add upstream https://github.com/seocheon/seocheon.git`
4. Create a feature branch: `git checkout -b feat/your-feature`
5. Make your changes, commit, and push
6. Open a Pull Request against `main`

## Development Setup

### Prerequisites

- Go >= 1.24.1
- Make
- Ignite CLI (for protobuf code generation)
- Docker & Docker Compose (for testnet)

### Build

```bash
make install          # Build and install binary
make proto-gen        # Generate protobuf code
make lint             # Run linter
```

### Test

```bash
make test             # Full test suite (go vet + govulncheck + unit)
make test-unit        # Unit tests only (30 min timeout)
make test-race        # Race condition detection
make test-cover       # Coverage report

# E2E tests (~5.5 min, in-memory testnet)
go test -v -timeout 20m -count=1 ./tests/e2e/...
```

## Coding Conventions

### Go / Cosmos SDK Patterns

Seocheon follows standard Cosmos SDK module conventions. Key patterns:

#### Module Directory Structure

```
x/{module}/
├── ante/           # AnteHandler decorators
├── client/cli/     # Custom CLI commands
├── keeper/         # Business logic
├── module/         # AppModule, AutoCLI, Depinject
└── types/          # Errors, events, params, store keys
```

#### Msg Handlers — One File Per Message

Each message handler gets its own file and corresponding test file:

```
keeper/msg_server_{action}.go
keeper/msg_server_{action}_test.go
```

#### Query Handlers — One File Per Query

```
keeper/query_{subject}.go
keeper/query_{subject}_test.go
```

#### Error Definitions

Define errors in `types/errors.go` using `errors.Register(ModuleName, code, msg)`:

- `x/node` error codes: 1100+
- `x/activity` error codes: 1200+

#### Event Definitions

Define events in `types/events.go` with `EventType*` and `AttributeKey*` constants.

#### AnteHandler Decorators

Place in `ante/` directory with the signature:

```go
func (d Decorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error)
```

#### Collections

Use Cosmos SDK Collections for state management:

```go
collections.Map[K, V]
collections.Item[T]
collections.KeySet[K]
```

#### ABCI — EndBlocker

Place epoch/window boundary logic in `keeper/abci.go`:

```go
func (k Keeper) EndBlocker(ctx context.Context) error
```

#### Depinject

Use `module/depinject.go` with the pattern:

```go
func ProvideModule(in ModuleInputs) ModuleOutputs
```

#### AutoCLI

Define in `module/autocli.go`. For complex flags, use `Skip: true` and implement custom CLI in `client/cli/`.

#### Protobuf

Files go in `proto/seocheon/{module}/v1/*.proto` with:

```protobuf
option go_package = "seocheon/x/{module}/types";
```

#### Testing

Use mock keeper structs with factory functions:

```go
func newMockAuthKeeper() *mockAuthKeeper { ... }
```

### General Guidelines

- Keep functions focused and small
- Use meaningful variable and function names
- Add comments only where the logic is not self-evident
- Do not add unnecessary abstractions for one-time operations
- Prefer standard library solutions over external dependencies

## Commit Messages

We follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

### Types

| Type | Description |
|------|-------------|
| `feat` | A new feature |
| `fix` | A bug fix |
| `docs` | Documentation changes |
| `refactor` | Code refactoring (no feature or fix) |
| `test` | Adding or updating tests |
| `chore` | Maintenance tasks |
| `ci` | CI/CD changes |
| `perf` | Performance improvements |

### Scopes

Common scopes: `node`, `activity`, `app`, `cli`, `proto`, `e2e`, `docker`

### Examples

```
feat(activity): add global uniqueness check for activity_hash
fix(node): handle edge case in feegrant quota reset
docs(blockchain): update tokenomics section
refactor(activity): extract reward calculation to separate function
test(e2e): add epoch transition reward distribution test
```

## Pull Request Process

1. **Branch naming**: Use `feat/`, `fix/`, `docs/`, `refactor/`, `test/` prefixes
2. **Keep PRs focused**: One feature or fix per PR
3. **Update documentation**: If your change affects behavior, update relevant docs
4. **Add tests**: All new features and bug fixes should include tests
5. **Pass CI**: Ensure all checks pass before requesting review
6. **Write a clear description**: Explain what, why, and how
7. **Link issues**: Reference related issues with `Closes #123` or `Relates to #123`

### PR Checklist

- [ ] Code follows the project's coding conventions
- [ ] Tests added/updated for the changes
- [ ] Documentation updated if needed
- [ ] `make test` passes locally
- [ ] `make lint` passes with no new warnings
- [ ] Commit messages follow conventional commits format
- [ ] No prohibited terminology used (see below)

## Testing Requirements

### Unit Tests

- All new keeper methods must have corresponding test files
- Test both success and error paths
- Use mock keepers for dependency isolation
- Minimum coverage: aim for critical path coverage

### E2E Tests

- New features affecting on-chain behavior should include E2E tests
- E2E tests use `app.TestNetworkFixture` for in-memory testnet
- Run with: `go test -v -timeout 20m -count=1 ./tests/e2e/...`

### Running Tests Before Submitting

```bash
# Minimum required before PR
make test
make lint

# Recommended
go test -v -timeout 20m -count=1 ./tests/e2e/...
```

## Reporting Issues

### Bug Reports

Use the [Bug Report template](.github/ISSUE_TEMPLATE/bug_report.md) and include:

- Clear description of the bug
- Steps to reproduce
- Expected vs actual behavior
- Environment details (Go version, OS, etc.)
- Relevant logs or error messages

### Feature Requests

Use the [Feature Request template](.github/ISSUE_TEMPLATE/feature_request.md) and include:

- Clear description of the proposed feature
- Use case and motivation
- Proposed implementation approach (if any)

## Terminology Guidelines

Seocheon has specific terminology requirements. Please use the correct expressions:

| Use | Do Not Use |
|-----|-----------|
| "The ecosystem determines value" | Any financial/securities terminology |
| "Ecological circulation" | Terms implying trading or speculation |
| "Contribution resources" | Terms implying revenue or returns |

For the complete list of prohibited terms, refer to the project's CLAUDE.md.

## Questions?

If you have questions about contributing, please open an issue with the `question` label.
