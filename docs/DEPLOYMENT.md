# Deployment Guide

This guide covers deploying the dinnerclub application in production using Make, Docker, and Docker Compose.

## Deployment Options

1. **Docker Compose** (Recommended for single-server deployments)
2. **Docker with external database**
3. **Fly.io** (Cloud deployment)
4. **Kubernetes** (For larger deployments)

## Prerequisites

- Docker and Docker Compose
- Make
- Production server with appropriate resources
- Domain name and SSL certificate (recommended)
- PostgreSQL database (if not using Docker Compose database)

## Docker Compose Deployment (Recommended)

### 1. Server Setup

```bash
# Update system
sudo apt update && sudo apt upgrade -y

# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Install Docker Compose
sudo apt install docker-compose-plugin

# Add user to docker group
sudo usermod -aG docker $USER
newgrp docker

# Install Make
sudo apt install make
```

### 2. Application Setup

```bash
# Clone repository
git clone https://github.com/toozej/dinnerclub.git
cd dinnerclub

# Create production configuration
cp .env.sample app.env
```

### 3. Configure Environment

Edit `app.env` for production:

```bash
# Production configuration
LOG_LEVEL=info
POSTGRES_HOSTNAME=dinnerclub_db
POSTGRES_PORT=5432
POSTGRES_USER=dinnerclub
POSTGRES_PASSWORD=secure_production_password_here
POSTGRES_DB=dinnerclub
CITY_CODE=PDX
GIN_MODE=release
JWT_SECRET=very-secure-production-jwt-secret-key-with-high-entropy-32-chars
JWT_REFRESH_TOKEN_SECRET=different-secure-refresh-token-secret-key-for-production
SESSION_SECRET=secure-production-session-secret-key-32-chars-minimum
CSRF_SECRET=secure-32-byte-csrf-production-key
REFERRAL_CODE=your_production_referral_code
RATE_LIMIT_ENABLED=true
RATE_LIMIT_REQUESTS=60
RATE_LIMIT_PERIOD=1m
RATE_LIMIT_AUTH_REQUESTS=5
RATE_LIMIT_AUTH_PERIOD=1m
```

Configure database environment in `db.env`:

```bash
POSTGRES_HOSTNAME=dinnerclub_db
POSTGRES_USER=dinnerclub
POSTGRES_PASSWORD=secure_production_password_here
POSTGRES_DB=dinnerclub
POSTGRES_PORT=5432
```

### 4. Configure Reverse Proxy (Traefik)

Update `docker-compose.yml` with your domain:

```yaml
services:
  dinnerclub_app:
    # ... other configuration
    labels:
      - "traefik.http.routers.dinnerclub.rule=Host(`your-domain.com`)"
      - "traefik.http.routers.dinnerclub-secure.rule=Host(`your-domain.com`)"
```

### 5. Deploy

```bash
# Build and start services
make build
docker compose up -d

# Check status
docker compose ps
docker compose logs -f dinnerclub_app
```

### 6. SSL Certificate Setup

If using Traefik with Let's Encrypt:

```yaml
# Add to docker-compose.yml
labels:
  - "traefik.http.routers.dinnerclub-secure.tls.certresolver=letsencrypt"
```

## Docker with External Database

### 1. Database Setup

Set up PostgreSQL on a separate server or use a managed database service:

```sql
-- Create database and user
CREATE ROLE dinnerclub WITH LOGIN PASSWORD 'secure_password';
CREATE DATABASE dinnerclub;
GRANT ALL PRIVILEGES ON DATABASE dinnerclub TO dinnerclub;
\c dinnerclub
GRANT ALL ON SCHEMA public TO dinnerclub;
```

### 2. Application Configuration

```bash
# app.env with external database
DATABASE_URL=postgres://dinnerclub:secure_password@db.example.com:5432/dinnerclub?sslmode=require
```

### 3. Docker Run

```bash
# Build image
make build

# Run container
docker run -d \
  --name dinnerclub \
  --restart unless-stopped \
  -p 8080:8080 \
  --env-file app.env \
  toozej/dinnerclub:latest
```

## Fly.io Deployment

### 1. Install Fly CLI

```bash
# Install flyctl
curl -L https://fly.io/install.sh | sh
```

### 2. Setup Fly Application

```bash
# Login to Fly.io
flyctl auth login

# Create app
flyctl apps create your-dinnerclub-app

# Set up PostgreSQL
flyctl postgres create --name your-dinnerclub-db --region your-region
```

### 3. Configure Secrets

```bash
# Set application secrets
flyctl secrets set \
  JWT_SECRET="your-jwt-secret" \
  JWT_REFRESH_TOKEN_SECRET="your-refresh-secret" \
  SESSION_SECRET="your-session-secret" \
  CSRF_SECRET="your-csrf-secret" \
  POSTGRES_PASSWORD="your-db-password"
```

### 4. Deploy

```bash
# Deploy using Make
make deploy

# Or deploy manually
flyctl deploy
```

### 5. Database Setup

```bash
# Connect to database and set up user
make deploy-psql-console

# In PostgreSQL console:
CREATE ROLE your_app_name WITH LOGIN PASSWORD 'your_password' CREATEDB;
CREATE DATABASE your_app_name;
GRANT ALL PRIVILEGES ON DATABASE your_app_name TO your_app_name;
```

## Kubernetes Deployment

### 1. Create Namespace

```yaml
# namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: dinnerclub
```

### 2. Create Secrets

```yaml
# secrets.yaml
apiVersion: v1
kind: Secret
metadata:
  name: dinnerclub-secrets
  namespace: dinnerclub
type: Opaque
data:
  jwt-secret: <base64-encoded-jwt-secret>
  jwt-refresh-secret: <base64-encoded-refresh-secret>
  session-secret: <base64-encoded-session-secret>
  csrf-secret: <base64-encoded-csrf-secret>
  postgres-password: <base64-encoded-db-password>
```

### 3. Create ConfigMap

```yaml
# configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: dinnerclub-config
  namespace: dinnerclub
data:
  LOG_LEVEL: "info"
  CITY_CODE: "PDX"
  GIN_MODE: "release"
  POSTGRES_HOSTNAME: "postgres-service"
  POSTGRES_PORT: "5432"
  POSTGRES_USER: "dinnerclub"
  POSTGRES_DB: "dinnerclub"
  REFERRAL_CODE: "your-referral-code"
  RATE_LIMIT_ENABLED: "true"
  RATE_LIMIT_REQUESTS: "60"
  RATE_LIMIT_PERIOD: "1m"
  RATE_LIMIT_AUTH_REQUESTS: "5"
  RATE_LIMIT_AUTH_PERIOD: "1m"
```

### 4. Create Deployment

```yaml
# deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dinnerclub
  namespace: dinnerclub
spec:
  replicas: 3
  selector:
    matchLabels:
      app: dinnerclub
  template:
    metadata:
      labels:
        app: dinnerclub
    spec:
      containers:
      - name: dinnerclub
        image: toozej/dinnerclub:latest
        ports:
        - containerPort: 8080
        envFrom:
        - configMapRef:
            name: dinnerclub-config
        env:
        - name: JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: dinnerclub-secrets
              key: jwt-secret
        - name: JWT_REFRESH_TOKEN_SECRET
          valueFrom:
            secretKeyRef:
              name: dinnerclub-secrets
              key: jwt-refresh-secret
        - name: SESSION_SECRET
          valueFrom:
            secretKeyRef:
              name: dinnerclub-secrets
              key: session-secret
        - name: CSRF_SECRET
          valueFrom:
            secretKeyRef:
              name: dinnerclub-secrets
              key: csrf-secret
        - name: POSTGRES_PASSWORD
          valueFrom:
            secretKeyRef:
              name: dinnerclub-secrets
              key: postgres-password
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
```

### 5. Create Service

```yaml
# service.yaml
apiVersion: v1
kind: Service
metadata:
  name: dinnerclub-service
  namespace: dinnerclub
spec:
  selector:
    app: dinnerclub
  ports:
  - protocol: TCP
    port: 80
    targetPort: 8080
  type: ClusterIP
```

### 6. Create Ingress

```yaml
# ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: dinnerclub-ingress
  namespace: dinnerclub
  annotations:
    kubernetes.io/ingress.class: nginx
    cert-manager.io/cluster-issuer: letsencrypt-prod
spec:
  tls:
  - hosts:
    - your-domain.com
    secretName: dinnerclub-tls
  rules:
  - host: your-domain.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: dinnerclub-service
            port:
              number: 80
```

### 7. Deploy to Kubernetes

```bash
kubectl apply -f namespace.yaml
kubectl apply -f secrets.yaml
kubectl apply -f configmap.yaml
kubectl apply -f deployment.yaml
kubectl apply -f service.yaml
kubectl apply -f ingress.yaml
```

## Production Checklist

### Security

- [ ] Use strong, unique secrets for all keys
- [ ] Enable HTTPS/TLS encryption
- [ ] Configure firewall rules
- [ ] Use `GIN_MODE=release`
- [ ] Enable rate limiting
- [ ] Use secure database passwords
- [ ] Enable database SSL connections
- [ ] Regular security updates

### Performance

- [ ] Configure appropriate resource limits
- [ ] Set up database connection pooling
- [ ] Configure caching if needed
- [ ] Monitor resource usage
- [ ] Set up log rotation
- [ ] Configure backup strategy

### Monitoring

- [ ] Set up application monitoring
- [ ] Configure log aggregation
- [ ] Set up health checks
- [ ] Monitor database performance
- [ ] Set up alerting
- [ ] Monitor SSL certificate expiration

### Backup

- [ ] Database backup strategy
- [ ] Configuration backup
- [ ] Test restore procedures
- [ ] Document recovery procedures

## Monitoring and Maintenance

### Health Checks

The application provides a health endpoint:

```bash
curl http://your-domain.com/health
```

### Logs

```bash
# Docker Compose
docker compose logs -f dinnerclub_app

# Docker
docker logs -f dinnerclub

# Kubernetes
kubectl logs -f deployment/dinnerclub -n dinnerclub
```

### Database Backup

```bash
# Docker Compose
make deploy-backup

# Manual backup
docker exec dinnerclub_db pg_dump -U dinnerclub dinnerclub > backup.sql
```

### Updates

```bash
# Pull latest image
docker compose pull

# Restart services
docker compose up -d

# Or use Make target
make rebuild
```

### Scaling

#### Docker Compose Scaling

```bash
# Scale application containers
docker compose up -d --scale dinnerclub_app=3
```

#### Kubernetes Scaling

```bash
# Scale deployment
kubectl scale deployment dinnerclub --replicas=5 -n dinnerclub
```

## Troubleshooting

### Common Issues

#### Application Won't Start

1. Check configuration:
   ```bash
   docker compose logs dinnerclub_app
   ```

2. Verify environment variables:
   ```bash
   docker compose exec dinnerclub_app env | grep -E "(JWT|POSTGRES|CSRF)"
   ```

3. Test database connection:
   ```bash
   docker compose exec dinnerclub_db psql -U dinnerclub -d dinnerclub -c "SELECT 1;"
   ```

#### Database Connection Issues

1. Check database is running:
   ```bash
   docker compose ps dinnerclub_db
   ```

2. Verify database credentials:
   ```bash
   docker compose exec dinnerclub_db psql -U dinnerclub -d dinnerclub
   ```

3. Check network connectivity:
   ```bash
   docker compose exec dinnerclub_app nc -zv dinnerclub_db 5432
   ```

#### SSL/TLS Issues

1. Check certificate status:
   ```bash
   openssl s_client -connect your-domain.com:443 -servername your-domain.com
   ```

2. Verify Traefik configuration:
   ```bash
   docker compose logs traefik
   ```

#### Performance Issues

1. Monitor resource usage:
   ```bash
   docker stats
   ```

2. Check database performance:
   ```bash
   docker compose exec dinnerclub_db psql -U dinnerclub -d dinnerclub -c "SELECT * FROM pg_stat_activity;"
   ```

3. Review application logs for slow queries:
   ```bash
   docker compose logs dinnerclub_app | grep -i "slow\|timeout"
   ```

### Recovery Procedures

#### Database Recovery

```bash
# Stop application
docker compose stop dinnerclub_app

# Restore database
docker compose exec -T dinnerclub_db psql -U dinnerclub -d dinnerclub < backup.sql

# Start application
docker compose start dinnerclub_app
```

#### Configuration Recovery

```bash
# Restore configuration from backup
cp backup/app.env ./app.env

# Restart services
docker compose restart
```

## Security Considerations

### Network Security

- Use firewalls to restrict access
- Only expose necessary ports
- Use VPN for administrative access
- Enable fail2ban for SSH protection

### Application Security

- Regular security updates
- Monitor for vulnerabilities
- Use security scanning tools
- Implement proper logging and monitoring

### Database Security

- Use strong passwords
- Enable SSL connections
- Restrict database access
- Regular security patches
- Backup encryption

## Performance Optimization

### Database Optimization

- Regular VACUUM and ANALYZE
- Monitor query performance
- Optimize indexes
- Configure connection pooling

### Application Optimization

- Monitor memory usage
- Optimize Docker images
- Use multi-stage builds
- Configure resource limits

### Caching

Consider adding caching layers:
- Redis for session storage
- CDN for static assets
- Application-level caching

## Maintenance Schedule

### Daily

- Monitor application health
- Check error logs
- Verify backup completion

### Weekly

- Review performance metrics
- Check security logs
- Update dependencies if needed

### Monthly

- Security updates
- Database maintenance
- Review and rotate logs
- Test backup restoration

### Quarterly

- Security audit
- Performance review
- Capacity planning
- Disaster recovery testing