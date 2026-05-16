# Project Structure

## Directory Layout

The project follows Go standard project layout conventions:

```
├── cmd/                     # Main applications
│   ├── project-name-here/   # CLI application entry point
│   └── diagrams/           # Architecture diagram generator
├── internal/               # Private application code
│   ├── package-name-here/ 
├── pkg/                    # Public library code
│   ├── config/            # Configuration management
│   ├── man/               # Manual page generation
│   └── version/           # Version information
├── scripts/               # Build and utility scripts
├── .kiro/                 # Kiro AI assistant configuration
└── main.go               # Application entry point
```

## Package Organization

### `cmd/` - Application Entry Points
- **`cmd/project-name-here/`**: Main CLI application with Cobra commands
- **`cmd/diagrams/`**: Utility for generating architecture diagrams

### `internal/` - Private Application Logic
- **`internal/package-name-here/`**: internal private application logic here, and in other internal/ files

### `pkg/` - Reusable Library Code
- **`pkg/config/`**: Configuration loading, validation, legacy migration
- **`pkg/version/`**: Build-time version information
- **`pkg/man/`**: Manual page generation for documentation

## Naming Conventions

- **Packages**: Lowercase, single word when possible
- **Files**: Snake_case for multi-word files, lowercase for single word
- **Types**: PascalCase (exported), camelCase (unexported)
- **Functions**: PascalCase (exported), camelCase (unexported)
- **Constants**: PascalCase or SCREAMING_SNAKE_CASE for package-level

## Configuration Files

- **`config.yaml`**: Runtime configuration (not in repo)
- **`config.example.yaml`**: Example configuration template
- **`.env`**: Environment variables (not in repo, use `.env.sample`)
- **`docker-compose.yml`**: Container orchestration
- **`.goreleaser.yml`**: Release automation configuration

## Build Artifacts

- **`out/`**: Compiled binaries (gitignored)
- **`vendor/`**: Go module dependencies (optional)
- **`c.out`**: Test coverage reports
- **`profiles/`**: Performance profiling data

## Key Files

- **`main.go`**: Application entry point, delegates to cmd package
- **`Makefile`**: Primary build automation and development workflow
- **`Dockerfile`**: Multi-stage container build
- **`.air.toml`**: Live reload configuration for development