# Configuration Guide

This document describes all configuration options available for the dinnerclub application.

## Configuration Sources

The application loads configuration from multiple sources in the following order of precedence:

1. Environment variables
2. `.env` files (app.env, db.env)
3. Default values (where applicable)

## Environment Files

### app.env

The main application configuration file. Copy from `.env.sample` and customize:

```bash
cp .env.sample app.env
```

### db.env

Database-specific configuration for Docker Compose environments.

## Configuration Options

### Required Configuration

These settings must be provided for the application to start:

#### LOG_LEVEL
- **Type**: String
- **Required**: Yes
- **Default**: None
- **Values**: `debug`, `info`, `warn`, `error`, `fatal`, `panic`
- **Description**: Sets the logging level for the application

```bash
LOG_LEVEL=info
```

#### CITY_CODE
- **Type**: String
- **Required**: Yes
- **Default**: None
- **Description**: Three-letter city code used for branding and identification
- **Example**: `PDX` (Portland), `NYC` (New York), `LAX` (Los Angeles)

```bash
CITY_CODE=PDX
```

#### GIN_MODE
- **Type**: String
- **Required**: Yes
- **Default**: None
- **Values**: `debug`, `release`, `test`
- **Description**: Sets the Gin web framework mode

```bash
# Development
GIN_MODE=debug

# Production
GIN_MODE=release
```

#### JWT_SECRET
- **Type**: String
- **Required**: Yes
- **Default**: None
- **Description**: Secret key for signing JWT access tokens (minimum 32 characters recommended)
- **Security**: Use a cryptographically secure random string

```bash
JWT_SECRET=your-very-secure-jwt-secret-key-here-32-chars-minimum
```

#### JWT_REFRESH_TOKEN_SECRET
- **Type**: String
- **Required**: Yes
- **Default**: None
- **Description**: Secret key for signing JWT refresh tokens (should be different from JWT_SECRET)

```bash
JWT_REFRESH_TOKEN_SECRET=your-different-refresh-token-secret-key-here
```

#### SESSION_SECRET
- **Type**: String
- **Required**: Yes
- **Default**: None
- **Description**: Secret key for session management (minimum 32 characters recommended)

```bash
SESSION_SECRET=your-session-secret-key-here-32-chars-minimum
```

#### CSRF_SECRET
- **Type**: String
- **Required**: Yes
- **Default**: None
- **Description**: Secret key for CSRF protection (must be exactly 32 bytes)

```bash
CSRF_SECRET=your-32-byte-csrf-secret-key-here
```

#### REFERRAL_CODE
- **Type**: String
- **Required**: Yes
- **Default**: None
- **Description**: Required code for user registration

```bash
REFERRAL_CODE=your-referral-code-here
```

### Database Configuration

#### Option 1: Individual Database Settings

```bash
POSTGRES_HOSTNAME=localhost
POSTGRES_PORT=5432
POSTGRES_USER=dinnerclub
POSTGRES_PASSWORD=your-secure-password
POSTGRES_DB=dinnerclub
```

#### Option 2: Database URL (takes precedence if provided)

```bash
DATABASE_URL=postgres://username:password@hostname:port/database?sslmode=disable
```

**Database Configuration Details:**

##### POSTGRES_HOSTNAME
- **Type**: String
- **Required**: No (if DATABASE_URL provided)
- **Default**: None
- **Description**: PostgreSQL server hostname or IP address

##### POSTGRES_PORT
- **Type**: Integer
- **Required**: No (if DATABASE_URL provided)
- **Default**: None
- **Description**: PostgreSQL server port

##### POSTGRES_USER
- **Type**: String
- **Required**: No (if DATABASE_URL provided)
- **Default**: None
- **Description**: PostgreSQL username

##### POSTGRES_PASSWORD
- **Type**: String
- **Required**: No (if DATABASE_URL provided)
- **Default**: None
- **Description**: PostgreSQL password

##### POSTGRES_DB
- **Type**: String
- **Required**: No (if DATABASE_URL provided)
- **Default**: None
- **Description**: PostgreSQL database name

### Rate Limiting Configuration

#### RATE_LIMIT_ENABLED
- **Type**: Boolean
- **Required**: No
- **Default**: `true`
- **Description**: Enable or disable rate limiting globally

```bash
RATE_LIMIT_ENABLED=true
```

#### RATE_LIMIT_REQUESTS
- **Type**: Integer
- **Required**: No
- **Default**: `100`
- **Description**: Number of requests allowed per period for general endpoints

```bash
RATE_LIMIT_REQUESTS=100
```

#### RATE_LIMIT_PERIOD
- **Type**: Duration String
- **Required**: No
- **Default**: `1m`
- **Description**: Time period for general rate limiting
- **Format**: Go duration format (`1s`, `1m`, `1h`, `30s`, `5m`)

```bash
RATE_LIMIT_PERIOD=1m
```

#### RATE_LIMIT_AUTH_REQUESTS
- **Type**: Integer
- **Required**: No
- **Default**: `10`
- **Description**: Number of requests allowed per period for authentication endpoints

```bash
RATE_LIMIT_AUTH_REQUESTS=10
```

#### RATE_LIMIT_AUTH_PERIOD
- **Type**: Duration String
- **Required**: No
- **Default**: `1m`
- **Description**: Time period for authentication rate limiting

```bash
RATE_LIMIT_AUTH_PERIOD=1m
```

## Configuration Examples

### Development Configuration

```bash
# app.env for development
LOG_LEVEL=debug
POSTGRES_HOSTNAME=localhost
POSTGRES_PORT=5432
POSTGRES_USER=dinnerclub
POSTGRES_PASSWORD=devpassword123
POSTGRES_DB=dinnerclub_dev
CITY_CODE=DEV
GIN_MODE=debug
JWT_SECRET=dev-jwt-secret-key-for-development-only-32-chars
JWT_REFRESH_TOKEN_SECRET=dev-refresh-secret-key-for-development-only
SESSION_SECRET=dev-session-secret-key-for-development-32-chars
CSRF_SECRET=dev-csrf-secret-32-byte-key-here
REFERRAL_CODE=devcode123
RATE_LIMIT_ENABLED=false
```

### Production Configuration

```bash
# app.env for production
LOG_LEVEL=info
DATABASE_URL=postgres://dinnerclub:secure_prod_password@prod-db:5432/dinnerclub?sslmode=require
CITY_CODE=PDX
GIN_MODE=release
JWT_SECRET=very-secure-production-jwt-secret-key-with-high-entropy
JWT_REFRESH_TOKEN_SECRET=different-secure-refresh-token-secret-key
SESSION_SECRET=secure-production-session-secret-key-32-chars
CSRF_SECRET=secure-32-byte-csrf-production-key
REFERRAL_CODE=prod_referral_2024
RATE_LIMIT_ENABLED=true
RATE_LIMIT_REQUESTS=60
RATE_LIMIT_PERIOD=1m
RATE_LIMIT_AUTH_REQUESTS=5
RATE_LIMIT_AUTH_PERIOD=1m
```

### Docker Compose Configuration

```bash
# app.env for Docker Compose
LOG_LEVEL=info
POSTGRES_HOSTNAME=dinnerclub_db
POSTGRES_PORT=5432
POSTGRES_USER=dinnerclub
POSTGRES_PASSWORD=docker_compose_password
POSTGRES_DB=dinnerclub
CITY_CODE=PDX
GIN_MODE=debug
JWT_SECRET=docker-jwt-secret-key-for-compose-environment
JWT_REFRESH_TOKEN_SECRET=docker-refresh-secret-key-for-compose
SESSION_SECRET=docker-session-secret-key-for-compose-env
CSRF_SECRET=docker-csrf-secret-32-byte-key-here
REFERRAL_CODE=dockercode
RATE_LIMIT_ENABLED=true
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_PERIOD=1m
RATE_LIMIT_AUTH_REQUESTS=10
RATE_LIMIT_AUTH_PERIOD=1m
```

### Testing Configuration

```bash
# app.env for testing
LOG_LEVEL=error
POSTGRES_HOSTNAME=localhost
POSTGRES_PORT=5432
POSTGRES_USER=dinnerclub_test
POSTGRES_PASSWORD=testpassword
POSTGRES_DB=dinnerclub_test
CITY_CODE=TST
GIN_MODE=test
JWT_SECRET=test-jwt-secret-key-for-testing-only-32-chars
JWT_REFRESH_TOKEN_SECRET=test-refresh-secret-key-for-testing-only
SESSION_SECRET=test-session-secret-key-for-testing-32-chars
CSRF_SECRET=test-csrf-secret-32-byte-key-here
REFERRAL_CODE=testcode
RATE_LIMIT_ENABLED=false
```

## Security Considerations

### Secret Generation

Generate secure secrets using:

```bash
# Generate 32-byte random string (for CSRF_SECRET)
openssl rand -base64 32

# Generate random string for other secrets
openssl rand -base64 48

# Or use uuidgen for simpler secrets
uuidgen
```

### Production Security Checklist

- [ ] Use `GIN_MODE=release` in production
- [ ] Set `LOG_LEVEL=info` or `warn` in production
- [ ] Use strong, unique secrets for all secret keys
- [ ] Enable SSL/TLS for database connections (`sslmode=require`)
- [ ] Use environment variables or secure secret management
- [ ] Enable rate limiting (`RATE_LIMIT_ENABLED=true`)
- [ ] Use appropriate rate limits for your traffic
- [ ] Regularly rotate secret keys
- [ ] Use secure database passwords
- [ ] Restrict database access to application only

### Secret Management

#### Environment Variables

```bash
# Set in shell environment
export JWT_SECRET="your-secret-here"
export POSTGRES_PASSWORD="your-db-password"
```

#### Docker Secrets

```yaml
# docker-compose.yml
services:
  dinnerclub_app:
    secrets:
      - jwt_secret
      - db_password
    environment:
      JWT_SECRET_FILE: /run/secrets/jwt_secret
      POSTGRES_PASSWORD_FILE: /run/secrets/db_password

secrets:
  jwt_secret:
    file: ./secrets/jwt_secret.txt
  db_password:
    file: ./secrets/db_password.txt
```

#### Kubernetes Secrets

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: dinnerclub-secrets
type: Opaque
data:
  jwt-secret: <base64-encoded-secret>
  db-password: <base64-encoded-password>
```

## Configuration Validation

The application validates configuration on startup:

- Required fields must be present
- Secret keys must meet minimum length requirements
- Database connection is tested
- Invalid duration formats are rejected

### Common Configuration Errors

#### Missing Required Configuration

```
Error: unable to parse environment variables: required key LOG_LEVEL missing value
```

**Solution**: Ensure all required environment variables are set.

#### Invalid Duration Format

```
Error: invalid duration format for RATE_LIMIT_PERIOD
```

**Solution**: Use Go duration format (`1s`, `1m`, `1h`).

#### Database Connection Failed

```
Error: failed to connect to database
```

**Solution**: Verify database configuration and ensure database is running.

#### Weak Secrets

While not enforced, use strong secrets:
- Minimum 32 characters for JWT secrets
- Exactly 32 bytes for CSRF secret
- Use cryptographically secure random generation

## Environment-Specific Configuration

### Development

- Use `debug` log level for detailed logging
- Disable rate limiting for easier testing
- Use `debug` Gin mode for detailed error pages
- Use simple passwords for local development

### Staging

- Use `info` log level
- Enable rate limiting with relaxed limits
- Use `release` Gin mode
- Use production-like secrets (but not production secrets)

### Production

- Use `info` or `warn` log level
- Enable rate limiting with appropriate limits
- Use `release` Gin mode
- Use strong, unique secrets
- Enable database SSL
- Use secure secret management

## Configuration Loading Order

1. **Default values** (built into application)
2. **Configuration files** (app.env, db.env)
3. **Environment variables** (override file values)
4. **Command-line flags** (if any, highest priority)

## Troubleshooting Configuration

### Debug Configuration Loading

Set `LOG_LEVEL=debug` to see configuration loading details:

```bash
LOG_LEVEL=debug ./dinnerclub
```

### Verify Configuration

The application logs loaded configuration (secrets are masked):

```
INFO[0000] Loaded config: {LogLevel:info CityCode:PDX PostgresHostname:localhost ...}
```

### Test Database Connection

```bash
# Test database connection manually
psql -h $POSTGRES_HOSTNAME -p $POSTGRES_PORT -U $POSTGRES_USER -d $POSTGRES_DB
```

### Validate Secrets

Ensure secrets meet requirements:
- JWT secrets: minimum 32 characters
- CSRF secret: exactly 32 bytes
- Session secret: minimum 32 characters

## Configuration Best Practices

1. **Use environment-specific files**: `app.dev.env`, `app.prod.env`
2. **Never commit secrets**: Add `*.env` to `.gitignore`
3. **Use secret management**: Vault, AWS Secrets Manager, etc.
4. **Rotate secrets regularly**: Especially in production
5. **Validate configuration**: Test configuration changes in staging first
6. **Document custom settings**: Keep configuration documented for team
7. **Use strong defaults**: Set secure defaults where possible
8. **Monitor configuration**: Log configuration loading and validation