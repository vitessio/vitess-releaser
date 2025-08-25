# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

### Build and Test
- `make build` - Builds local vitess-releaser binary for development 
- `make install` - Installs the latest released version from GitHub
- `make test` - Builds and runs a test release command (RC release)
- `go build -o vitess-releaser ./main.go` - Manual build command

### Linting and Formatting
- `make lint` - Run golangci-lint to check code quality
- `make lint-fix` - Run golangci-lint and automatically fix issues where possible
- `make fmt` - Format code using gofumpt and goimports

### Running the Tool
- **Production**: `vitess-releaser --date="YYYY-MM-DD" --live --release=XX [--rc=N] [--vtop-release=X.X]`
- **Development/Testing**: `vitess-releaser --date="YYYY-MM-DD" --release=XX [--rc=N]` (uses personal forks)

### Authentication
Set `VITESS_RELEASER_GH_TOKEN` environment variable with GitHub Personal Access Token (repo and org:read permissions) to avoid PR review bottlenecks.

## Architecture Overview

This is a Go CLI tool for automating Vitess and Vitess Operator releases. The application uses:

- **CLI Framework**: Cobra for command-line interface
- **TUI Framework**: Bubbletea with Lipgloss for terminal user interface
- **State Management**: Context-based state passing throughout the application

### Core Package Structure

- **`main.go`**: Entry point with version checking and panic recovery
- **`go/cmd/`**: Command-line interface and flag parsing
- **`go/interactive/`**: Terminal UI components and menu systems
- **`go/releaser/`**: Core release automation logic
  - **`code_freeze/`**: Code freeze preparation tasks
  - **`pre_release/`**: Pre-release tasks (release notes, PRs)
  - **`release/`**: Release execution (tagging, artifacts, milestones) 
  - **`post_release/`**: Post-release cleanup tasks
  - **`github/`**: GitHub API interactions (issues, PRs, milestones, releases)
  - **`git/`**: Git operations and repository management
  - **`slack/`**: Slack integration for announcements

### Release Flow Phases

1. **Prerequisites**: General checks, Slack announcements, blog post drafts
2. **Code Freeze**: Branch creation, milestone management, version bumping
3. **Pre-Release**: Release PR creation, release notes generation
4. **Release**: Tagging, artifact publishing, milestone closing
5. **Post-Release**: Issue closure, cleanup tasks

### Key Types

- **`releaser.State`**: Global application state passed via context
- **`releaser.ReleaseInformation`**: Release metadata for Vitess and VitessOperator
- **Interactive UI**: Menu-driven interface with step tracking

## Local Development Setup

For testing without affecting production repositories:

1. Fork `vitessio/vitess` and clone to `vitess/` subdirectory
2. Fork `planetscale/vitess-operator` and clone to `vitess-operator/` subdirectory  
3. Create required labels on your Vitess fork: `Component: General`, `Type: Release`, `Do Not Merge`
4. Run without `--live` flag to use personal forks

## Important Constants

- Current tool version defined in `go/cmd/cmd.go` as `VERSION` constant
- Vitess version file path: `./go/vt/servenv/version.go` (in Vitess repo)
- Release branches follow pattern: `release-XX.0` for Vitess, `release-X.XX` for VitessOperator

## Release Types

- **RC Release**: `--rc=N` flag for release candidates
- **GA Release**: No RC flag for general availability 
- **Patch Release**: Determined by SNAPSHOT version on release branch
- **VitessOperator**: Optional parallel release with `--vtop-release` flag