# Technology Stack

## Core Technologies

- **Language**: Go 1.25
- **CLI Framework**: Cobra (spf13/cobra)
- **Configuration**: Viper (spf13/viper) with YAML support
- **Web Scraping**: goquery (PuerkitoBio/goquery)
- **Notifications**: nikoksr/notify library
- **Logging**: logrus (sirupsen/logrus)
- **Containerization**: Docker with distroless base images

## Build System

- **Build Tool**: Make + Go toolchain
- **Release**: GoReleaser for cross-platform builds
- **Container Registry**: Docker Hub, GitHub Container Registry, Quay.io
- **Package Managers**: Homebrew, APT, RPM, APK
- **Security**: Cosign for artifact signing

## Development Tools

- **Live Reload**: Air for development iteration
- **Pre-commit**: Hooks for code quality (golangci-lint, goimports, gosec, staticcheck)
- **Testing**: Go test with race detection and coverage
- **Profiling**: Built-in CPU and memory profiling support

## Common Commands

### Development
```bash
make local                    # Full local development workflow
make local-build             # Build binary locally
make local-run               # Run with config.yaml
make local-test              # Run tests with coverage
make local-iterate           # Live reload development
```

### Docker
```bash
make build                   # Build Docker image
make run                     # Run containerized
make up                      # Docker Compose up
make down                    # Docker Compose down
```

### Release & Quality
```bash
make pre-commit              # Run all quality checks
make local-release           # Create release artifacts
make local-sign              # Sign artifacts with Cosign
make local-verify            # Verify signatures
```

### Utilities
```bash
make clean                   # Remove build artifacts
make docs                    # Serve Go documentation
make diagrams                # Generate architecture diagrams
make update-golang-version   # Update Go version across repo
```

## Configuration

- **Environment**: variables for backward compatibility
- **CLI Flags**: Basic options like debug mode and config file path