# Development Guide

This guide covers how to set up and run the dinnerclub application in development mode.

## Prerequisites

- Go 1.25 or later
- Docker and Docker Compose
- Make
- PostgreSQL (if running locally without Docker)

## Quick Start with Docker Compose

The easiest way to get started is using Docker Compose for the development environment:

```bash
# Start the development environment
make up

# View logs
make logs

# Stop the environment
make down
```

This will:
- Build the application container
- Start a PostgreSQL database container
- Run database migrations automatically
- Expose the application on http://localhost:8080

## Local Development Setup

### 1. Environment Configuration

Copy the sample environment file and configure it:

```bash
cp .env.sample app.env
```

Edit `app.env` with your configuration. See [CONFIGURATION.md](CONFIGURATION.md) for detailed configuration options.

### 2. Database Setup

#### Option A: Using Docker for Database Only

```bash
# Start only the database container
docker compose -f docker-compose-dev.yml up -d dinnerclub_db
```

#### Option B: Local PostgreSQL Installation

Install PostgreSQL locally and create the database:

```sql
CREATE ROLE dinnerclub WITH LOGIN PASSWORD 'your_password' CREATEDB;
CREATE DATABASE dinnerclub;
GRANT ALL PRIVILEGES ON DATABASE dinnerclub TO dinnerclub;
```

Update your `app.env` file with the correct database connection details.

### 3. Install Dependencies

```bash
# Update Go module dependencies
make local-update-deps

# Vendor dependencies
make local-vendor
```

### 4. Build and Run

```bash
# Run all checks and build
make local

# Or run individual steps:
make local-vet      # Run go vet
make local-test     # Run tests
make local-build    # Build binary
make local-run      # Run the application
```

The application will be available at http://localhost:8080

## Development Workflow

### Running Tests

```bash
# Run all tests
make local-test

# Run tests with coverage
make local-cover

# Run specific test packages
go test ./internal/controllers/...
go test ./pkg/authentication/...
```

### Code Quality Checks

```bash
# Install pre-commit hooks
make pre-commit-install

# Run all pre-commit checks
make pre-commit-run

# Run individual checks
make local-vet
```

### Hot Reloading

For development with hot reloading, you can use Air:

```bash
# Install Air (done automatically with make pre-commit-install)
go install github.com/air-verse/air@latest

# Run with hot reloading
air
```

### Database Operations

#### Migrations

Database migrations run automatically when the application starts. The migration system is located in `internal/models/migration.go`.

#### Database Console Access

```bash
# Access database console (when using Docker Compose)
docker exec -it dinnerclub_db psql -U dinnerclub -d dinnerclub

# Or using local PostgreSQL
psql -U dinnerclub -d dinnerclub
```

#### Reset Database

```bash
# Stop containers and remove volumes
make down
docker volume rm dinnerclub_dinnerclub_db

# Restart
make up
```

## Development Tools

### Make Targets

The project includes comprehensive Make targets for development:

```bash
make help                    # Show all available targets
make local                   # Full local development workflow
make local-build            # Build binary locally
make local-run              # Run locally built binary
make local-test             # Run tests
make local-cover            # View test coverage
make up                     # Start Docker Compose development environment
make down                   # Stop Docker Compose environment
make rebuild                # Rebuild and restart app container
make logs                   # View application logs
make pre-commit-install     # Install development tools
make pre-commit-run         # Run code quality checks
```

### Debugging

#### Application Logs

```bash
# View logs in Docker Compose
make logs

# View logs for specific service
docker compose -f docker-compose-dev.yml logs -f dinnerclub_app
```

#### Debug Mode

Set the following in your `app.env` for debug mode:

```bash
LOG_LEVEL=debug
GIN_MODE=debug
```

#### Database Query Logging

GORM query logging is controlled by the `LOG_LEVEL` environment variable. Set to `debug` to see all SQL queries.

## Project Structure

```
dinnerclub/
├── cmd/dinnerclub/          # CLI application entry point
├── internal/                # Private application code
│   ├── controllers/         # HTTP request handlers
│   ├── models/             # Database models and migrations
│   ├── routers/            # Route definitions
│   ├── services/           # Business logic services
│   ├── helpers/            # Template helpers
│   ├── aggregation/        # Data aggregation services
│   ├── rating/             # Rating system
│   ├── search/             # Search functionality
│   └── validation/         # Input validation
├── pkg/                    # Public/reusable packages
│   ├── authentication/     # JWT authentication
│   ├── config/            # Configuration management
│   ├── database/          # Database connection
│   ├── csrf/              # CSRF protection
│   ├── ratelimit/         # Rate limiting
│   └── session/           # Session management
├── templates/             # HTML templates
├── assets/               # Static assets (CSS, JS, images)
├── docs/                 # Documentation
└── scripts/              # Utility scripts
```

## Common Development Tasks

### Adding New Features

1. Create models in `internal/models/`
2. Add database migrations to `internal/models/migration.go`
3. Create controllers in `internal/controllers/`
4. Add routes in `internal/routers/`
5. Create templates in `templates/`
6. Add tests for all components

### Adding New API Endpoints

1. Define handler function in appropriate controller
2. Add route in `internal/routers/public.go` or `private.go`
3. Add validation if needed
4. Write tests
5. Update API documentation

### Modifying Templates

Templates are located in the `templates/` directory and use Go's `html/template` package. Template helpers are defined in `internal/helpers/template_helpers.go`.

### Working with the Database

- Models use GORM ORM
- Database configuration is in `pkg/database/`
- Migrations are in `internal/models/migration.go`
- Connection string is built from environment variables

## Troubleshooting

### Common Issues

#### Port Already in Use

```bash
# Find process using port 8080
lsof -i :8080

# Kill the process
kill -9 <PID>
```

#### Database Connection Issues

1. Check database is running: `docker ps`
2. Verify connection details in `app.env`
3. Check database logs: `docker logs dinnerclub_db`

#### Build Issues

```bash
# Clean and rebuild
make clean
make local-build

# Update dependencies
make local-update-deps
```

#### Permission Issues with Docker

```bash
# Fix Docker permissions (Linux)
sudo usermod -aG docker $USER
newgrp docker
```

### Getting Help

- Check the logs: `make logs`
- Run tests to verify setup: `make local-test`
- Verify configuration: Check `app.env` against `CONFIGURATION.md`
- Check database connectivity: Use database console commands above

## Performance Considerations

### Development Performance

- Use `make local-run` for faster iteration (no Docker rebuild)
- Use Air for hot reloading during development
- Consider using local PostgreSQL for faster database operations

### Database Performance

- Database indexes are created automatically during migration
- Use `LOG_LEVEL=debug` to monitor query performance
- Consider using connection pooling settings in production

## Security in Development

Even in development, the application includes security features:

- CSRF protection (can be disabled by removing `CSRF_SECRET`)
- Rate limiting (can be disabled with `RATE_LIMIT_ENABLED=false`)
- JWT authentication with secure defaults
- Input validation and sanitization

For development convenience, you may want to:
- Use longer JWT expiration times
- Disable rate limiting
- Use debug log levels
- Use less strict CSRF settings