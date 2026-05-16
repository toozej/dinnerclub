# dinnerclub

![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/toozej/dinnerclub)
[![Go Report Card](https://goreportcard.com/badge/github.com/toozej/dinnerclub)](https://goreportcard.com/report/github.com/toozej/dinnerclub)
![GitHub Actions Workflow Status](https://img.shields.io/github/actions/workflow/status/toozej/dinnerclub/cicd.yaml)
![Docker Pulls](https://img.shields.io/docker/pulls/toozej/dinnerclub)
![GitHub Downloads (all assets, all releases)](https://img.shields.io/github/downloads/toozej/dinnerclub/total)

Web application to facilitate reviewing and organizing get-togethers at restaurants

Uses [toozej/golang-starter](https://github.com/toozej/golang-starter) as a base

## Security Features

### CSRF Protection
The application implements Cross-Site Request Forgery (CSRF) protection using the `filippo.io/csrf/gorilla` package via the `pkg/csrf` middleware. This protects all forms from CSRF attacks by:

- Generating unique CSRF tokens for each user session
- Validating CSRF tokens on all POST requests
- Automatically including CSRF tokens in all forms via the `CSRFInput` template helper
- Using secure cookie settings (HttpOnly, SameSite=Strict)

#### Configuration
CSRF protection requires the `CSRF_SECRET` environment variable to be set. This should be a random 32-byte string in production.

Example in `.env` file:
```
CSRF_SECRET=your-random-32-byte-secret-here
```

#### Template Usage
All forms automatically include CSRF tokens using the `CSRFInput` template helper:
```html
<form action="/some/endpoint" method="POST">
    {{ CSRFInput . }}
    <!-- form fields -->
</form>
```

### Rate Limiting
The application implements rate limiting using the `github.com/JGLTechnologies/gin-rate-limit` package to protect against abuse and ensure fair usage. Rate limiting is applied at two levels:

- **General Rate Limiting**: Applied to all endpoints to prevent general abuse
- **Authentication Rate Limiting**: Stricter limits applied to authentication endpoints (login, register, token refresh) to prevent brute force attacks

#### Configuration
Rate limiting can be configured using the following environment variables:

```bash
# Enable/disable rate limiting (default: true)
RATE_LIMIT_ENABLED=true

# General rate limiting (default: 100 requests per minute)
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_PERIOD=1m

# Authentication rate limiting (default: 10 requests per minute)
RATE_LIMIT_AUTH_REQUESTS=10
RATE_LIMIT_AUTH_PERIOD=1m
```

#### Rate Limiting Behavior
- Rate limits are applied per client IP address
- When limits are exceeded, the server returns HTTP 429 (Too Many Requests)
- Authentication endpoints have stricter limits to prevent brute force attacks
- Rate limiting can be disabled by setting `RATE_LIMIT_ENABLED=false` (not recommended for production)

#### Period Format
The period values support Go duration format:
- `1s` = 1 second
- `1m` = 1 minute  
- `1h` = 1 hour
- `30s` = 30 seconds
- `5m` = 5 minutes

## Deployment via Fly.io
### First time setup
#### Postgres Database
- fly postgres create
    - app name: ${CITYCODE}dinnerclub-db
    - single node
    - don't scale to zero after one hour
    - add connection info to deploy.env, specifically DATABASE_URL
    - create dinnerclub user and database
        - `flyctl proxy 5434:5432 -a $(DEPLOY_APPNAME)-db &`
        - `PGPASSWORD=$(DEPLOY_POSTGRES_PASSWORD) psql -h localhost -p 5434 -U postgres`
            - `CREATE ROLE ${CITYCODE}dinnerclub WITH LOGIN PASSWORD 'passwordhere' CREATEDB;`
            - `CREATE DATABASE ${CITYCODE}dinnerclub;`
            - `GRANT ALL PRIVILEGES ON DATABASE ${CITYCODE}dinnerclub TO ${CITYCODE}dinnerclub;`
            - `\c ${CITYCODE}dinnerclub`
            - `GRANT ALL ON SCHEMA public TO ${CITYCODE}dinnerclub;`
            - `\q`
        - `pkill -15 -f 'flyctl proxy'`
