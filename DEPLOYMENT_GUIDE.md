# Deployment Guide

## Overview

This guide covers deploying the Go Authentication System in various environments, from development to production.

## Prerequisites

- Docker and Docker Compose
- Go 1.21+ (for local development)
- PostgreSQL 15+
- Redis 7+

## Development Setup

### 1. Clone and Setup

```bash
git clone <repository-url>
cd go-auth-system-1
```

### 2. Environment Configuration

Create a `.env` file:

```bash
# Application Configuration
PORT=8080
GIN_MODE=debug

# Database Configuration
DATABASE_URL=postgres://user:password@localhost:5432/auth_db?sslmode=disable

# Redis Configuration
REDIS_URL=redis://localhost:6379

# JWT Configuration
JWT_SECRET=your-super-secret-jwt-key-change-in-production-must-be-at-least-32-characters

# Email Configuration
EMAIL_SERVICE=smtp
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password
```

### 3. Start Services

```bash
# Start all services
docker-compose up -d

# Check service status
docker-compose ps

# View logs
docker-compose logs -f app
```

### 4. Run Tests

```bash
# Run all tests
go test ./tests/...

# Run tests with coverage
go test -cover ./tests/...

# Run specific test
go test -run TestAuthTestSuite ./tests/
```

### 5. Manual Testing

```bash
# Register a new user
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -H "X-CSRF-Token: your-csrf-token" \
  -d '{
    "email": "test@example.com",
    "password": "TestPassword123!",
    "first_name": "Test",
    "last_name": "User"
  }'

# Login
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -H "X-CSRF-Token: your-csrf-token" \
  -d '{
    "email": "test@example.com",
    "password": "TestPassword123!"
  }'

# Health check
curl http://localhost:8080/health
```

## Production Deployment

### 1. Environment Setup

Create production environment file:

```bash
# Production Configuration
PORT=8080
GIN_MODE=release

# Database Configuration (use connection pooling)
DATABASE_URL=postgres://user:password@db:5432/auth_db?sslmode=require&pool_max_conns=25&pool_min_conns=5

# Redis Configuration
REDIS_URL=redis://cache:6379

# JWT Configuration (use strong secret)
JWT_SECRET=your-production-jwt-secret-must-be-at-least-32-characters-long-and-random

# Email Configuration
SMTP_HOST=smtp.yourdomain.com
SMTP_PORT=587
SMTP_USERNAME=noreply@yourdomain.com
SMTP_PASSWORD=your-smtp-password

# Security Configuration
CSRF_SECRET=your-production-csrf-secret-key
ALLOWED_ORIGINS=https://yourdomain.com,https://www.yourdomain.com
```

### 2. Docker Production Build

```bash
# Build production image
docker build -f Dockerfile.production -t auth-system:latest .

# Or use docker-compose for production
docker-compose -f docker-compose.production.yml up -d
```

### 3. Database Setup

```bash
# Run migrations
docker-compose exec app ./main migrate

# Or manually run migrations
docker-compose exec db psql -U user -d auth_db -f /docker-entrypoint-initdb.d/001_create_users_table.up.sql
```

### 4. SSL/TLS Configuration

#### Option 1: Nginx Reverse Proxy (Recommended)

1. Update `nginx.conf` with your domain
2. Obtain SSL certificates (Let's Encrypt)
3. Update nginx configuration with SSL settings
4. Start nginx service

```bash
# Generate SSL certificates with Let's Encrypt
certbot --nginx -d yourdomain.com

# Update nginx.conf with SSL configuration
# Uncomment HTTPS server block
```

#### Option 2: Application-Level SSL

```bash
# Add SSL configuration to application
# Update docker-compose.production.yml with SSL certificates
```

### 5. Monitoring and Logging

#### Application Logs

```bash
# View application logs
docker-compose logs -f app

# View all service logs
docker-compose logs -f
```

#### Database Monitoring

```bash
# Connect to database
docker-compose exec db psql -U user -d auth_db

# Check database size
docker-compose exec db psql -U user -d auth_db -c "SELECT pg_size_pretty(pg_database_size('auth_db'));"

# Check active connections
docker-compose exec db psql -U user -d auth_db -c "SELECT count(*) FROM pg_stat_activity;"
```

#### Redis Monitoring

```bash
# Connect to Redis
docker-compose exec cache redis-cli

# Check Redis info
docker-compose exec cache redis-cli info

# Monitor Redis commands
docker-compose exec cache redis-cli monitor
```

### 6. Backup Strategy

#### Database Backup

```bash
# Create backup script
cat > backup.sh << 'EOF'
#!/bin/bash
BACKUP_DIR="/backups"
DATE=$(date +%Y%m%d_%H%M%S)
docker-compose exec -T db pg_dump -U user auth_db > $BACKUP_DIR/auth_db_$DATE.sql
gzip $BACKUP_DIR/auth_db_$DATE.sql
EOF

chmod +x backup.sh

# Schedule daily backups
echo "0 2 * * * /path/to/backup.sh" | crontab -
```

#### Redis Backup

```bash
# Redis persistence is enabled by default
# Check Redis persistence settings
docker-compose exec cache redis-cli config get save
```

### 7. Scaling Considerations

#### Horizontal Scaling

```yaml
# docker-compose.scale.yml
version: '3.8'
services:
  app:
    deploy:
      replicas: 3
    environment:
      - REDIS_URL=redis://redis-cluster:6379
      - DATABASE_URL=postgres://user:password@postgres-cluster:5432/auth_db
```

#### Load Balancer Configuration

```nginx
upstream auth_backend {
    server app1:8080;
    server app2:8080;
    server app3:8080;
}
```

### 8. Security Hardening

#### Container Security

```bash
# Run containers as non-root user
# Use read-only filesystems where possible
# Limit container capabilities
# Regular security updates
```

#### Network Security

```bash
# Use internal networks
# Implement firewall rules
# Use VPN for database access
# Enable TLS everywhere
```

#### Application Security

```bash
# Regular security audits
# Dependency vulnerability scanning
# Secret rotation
# Access logging and monitoring
```

## Kubernetes Deployment

### 1. Create Kubernetes Manifests

```yaml
# k8s/namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: auth-system
```

```yaml
# k8s/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: auth-config
  namespace: auth-system
data:
  PORT: "8080"
  GIN_MODE: "release"
  SMTP_HOST: "smtp.yourdomain.com"
  SMTP_PORT: "587"
```

```yaml
# k8s/secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: auth-secrets
  namespace: auth-system
type: Opaque
data:
  DATABASE_URL: <base64-encoded-database-url>
  REDIS_URL: <base64-encoded-redis-url>
  JWT_SECRET: <base64-encoded-jwt-secret>
  SMTP_USERNAME: <base64-encoded-smtp-username>
  SMTP_PASSWORD: <base64-encoded-smtp-password>
```

```yaml
# k8s/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: auth-app
  namespace: auth-system
spec:
  replicas: 3
  selector:
    matchLabels:
      app: auth-app
  template:
    metadata:
      labels:
        app: auth-app
    spec:
      containers:
      - name: auth-app
        image: auth-system:latest
        ports:
        - containerPort: 8080
        envFrom:
        - configMapRef:
            name: auth-config
        - secretRef:
            name: auth-secrets
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

```yaml
# k8s/service.yaml
apiVersion: v1
kind: Service
metadata:
  name: auth-service
  namespace: auth-system
spec:
  selector:
    app: auth-app
  ports:
  - port: 80
    targetPort: 8080
  type: LoadBalancer
```

### 2. Deploy to Kubernetes

```bash
# Apply all manifests
kubectl apply -f k8s/

# Check deployment status
kubectl get pods -n auth-system

# Check service
kubectl get svc -n auth-system

# View logs
kubectl logs -f deployment/auth-app -n auth-system
```

## CI/CD Pipeline

### GitHub Actions Example

```yaml
# .github/workflows/ci-cd.yml
name: CI/CD Pipeline

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v3
      with:
        go-version: 1.21
    
    - name: Run tests
      run: go test ./tests/...
    
    - name: Run security scan
      run: |
        go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
        gosec ./...

  build:
    needs: test
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    
    - name: Build Docker image
      run: docker build -f Dockerfile.production -t auth-system:${{ github.sha }} .
    
    - name: Push to registry
      run: |
        echo ${{ secrets.DOCKER_PASSWORD }} | docker login -u ${{ secrets.DOCKER_USERNAME }} --password-stdin
        docker push auth-system:${{ github.sha }}

  deploy:
    needs: build
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'
    steps:
    - name: Deploy to production
      run: |
        # Add deployment commands here
        echo "Deploying to production..."
```

## Troubleshooting

### Common Issues

#### 1. Database Connection Issues

```bash
# Check database connectivity
docker-compose exec app nc -zv db 5432

# Check database logs
docker-compose logs db

# Reset database
docker-compose down -v
docker-compose up -d
```

#### 2. Redis Connection Issues

```bash
# Check Redis connectivity
docker-compose exec app nc -zv cache 6379

# Check Redis logs
docker-compose logs cache

# Test Redis commands
docker-compose exec cache redis-cli ping
```

#### 3. Application Issues

```bash
# Check application logs
docker-compose logs app

# Check application health
curl http://localhost:8080/health

# Restart application
docker-compose restart app
```

#### 4. Migration Issues

```bash
# Check migration status
docker-compose exec app ./main migrate status

# Rollback last migration
docker-compose exec app ./main migrate rollback

# Force migration
docker-compose exec app ./main migrate force <version>
```

### Performance Optimization

#### Database Optimization

```sql
-- Add indexes for frequently queried columns
CREATE INDEX CONCURRENTLY idx_users_email ON users(email);
CREATE INDEX CONCURRENTLY idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX CONCURRENTLY idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);

-- Analyze query performance
EXPLAIN ANALYZE SELECT * FROM users WHERE email = 'test@example.com';
```

#### Redis Optimization

```bash
# Configure Redis memory policy
docker-compose exec cache redis-cli config set maxmemory-policy allkeys-lru

# Monitor Redis memory usage
docker-compose exec cache redis-cli info memory
```

#### Application Optimization

```bash
# Enable Go profiling
export GODEBUG=pprof=1

# Monitor application metrics
curl http://localhost:8080/debug/pprof/
```

## Maintenance

### Regular Tasks

1. **Security Updates**: Monthly dependency updates
2. **Backup Verification**: Weekly backup restoration tests
3. **Performance Monitoring**: Daily performance metrics review
4. **Log Analysis**: Weekly security log analysis
5. **Certificate Renewal**: Automated SSL certificate renewal

### Monitoring Checklist

- [ ] Application health checks passing
- [ ] Database connection pool healthy
- [ ] Redis memory usage within limits
- [ ] SSL certificates valid
- [ ] Security logs clean
- [ ] Performance metrics normal
- [ ] Backup jobs successful

## Support

For issues and questions:
1. Check the troubleshooting section
2. Review application logs
3. Check GitHub issues
4. Contact the development team
