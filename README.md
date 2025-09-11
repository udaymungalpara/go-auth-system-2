# Go Auth System

Production-ready authentication service built with Go, PostgreSQL, and Redis. It ships with JWT auth (access + refresh), email verification, password reset, rate limiting, strong validation, and Docker-first deployment.

## What you get

- JWT auth with refresh rotation and blacklist
- Secure password handling (bcrypt), account lockout, CSRF protection
- Rate limiting (per IP/user), security headers, audit logging
- Postgres + Redis integration, health checks, migrations
- Docker and Docker Compose ready, CI to build and push your image

---

## Quick start

If you just want it running locally using your prebuilt image on Docker Hub:

```bash
docker pull udaymungalpara/go-auth-system-1-app:latest
docker run -d --name go-auth \
  -p 8080:8080 \
  -e PORT=8080 \
  -e GIN_MODE=release \
  -e DATABASE_URL="postgres://<user>:<pass>@<db-host>:5432/<db-name>?sslmode=disable" \
  -e REDIS_URL="redis://<redis-host>:6379" \
  -e JWT_SECRET="<strong-32+-char-secret>" \
  -e EMAIL_SERVICE="smtp" \
  -e SMTP_HOST="<smtp-host>" \
  -e SMTP_PORT="587" \
  -e SMTP_USERNAME="<smtp-user>" \
  -e SMTP_PASSWORD="<smtp-pass>" \
  -e CSRF_SECRET="<csrf-secret>" \
  -e ALLOWED_ORIGINS="http://localhost" \
  udaymungalpara/go-auth-system-1-app:latest
```

Health check:

```bash
curl http://localhost:8080/health
```

---

## Run full stack with Docker Compose (App + Postgres + Redis)

1) In `docker-compose.production.yml`, set the app image:

```yaml
services:
  app:
    image: udaymungalpara/go-auth-system-1-app:latest
```

2) Create `.env.production` (example values):

```bash
PORT=8080
DB_NAME=auth_db
USER=auth_user
PASSWORD=auth_pass
JWT_SECRET=change-to-strong-32+-chars
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your-email@gmail.com
SMTP_PASS=your-app-password
CSRF_SECRET=change-me
ALLOWED_ORIGINS=http://localhost
```

3) Start:

```bash
docker compose -f docker-compose.production.yml --env-file .env.production up -d
```

---

## Build the image yourself (optional)

If you prefer building locally instead of pulling:

```bash
docker build -f Dockerfile.production -t auth-system:latest .
docker run -d --name go-auth -p 8080:8080 --env-file .env.production auth-system:latest
```

Or with compose (build on this machine):

```bash
docker compose -f docker-compose.production.yml up -d --build
```

---

## Deploy from GitHub (CI/CD)

This repo includes a GitHub Actions workflow at `.github/workflows/docker-image.yml` that builds with `Dockerfile.production` and pushes to Docker Hub.

1) Push your code to GitHub:

```bash
git init
git remote add origin https://github.com/<your-user>/<your-repo>.git
git add .
git commit -m "Initial commit"
git push -u origin main
```

2) In GitHub → Settings → Secrets and variables → Actions, add:

- DOCKERHUB_USERNAME
- DOCKERHUB_TOKEN (Docker Hub access token)

3) On each push to `main` (and on tags like `v1.2.3`), CI will build and push:

- `udaymungalpara/go-auth-system-1-app:latest` (default branch)
- `udaymungalpara/go-auth-system-1-app:sha-<short>` (per-commit)

Then deploy by pulling and running on your server:

```bash
docker pull udaymungalpara/go-auth-system-1-app:latest
docker run -d --name go-auth -p 8080:8080 --env-file .env.production udaymungalpara/go-auth-system-1-app:latest
```

---

## Environment variables

Copy `.env.example` to `.env` or `.env.production` and fill in the values.

- PORT (default 8080)
- GIN_MODE (debug|release)
- DATABASE_URL (e.g., `postgres://user:pass@db:5432/auth_db?sslmode=disable`)
- REDIS_URL (e.g., `redis://cache:6379`)
- JWT_SECRET (32+ chars, strong, random)
- EMAIL_SERVICE (e.g., `smtp`)
- SMTP_HOST, SMTP_PORT, SMTP_USERNAME, SMTP_PASSWORD
- CSRF_SECRET
- ALLOWED_ORIGINS (comma-separated, e.g., `http://localhost`)

Never commit real secrets. Use GitHub Secrets and your server’s secret storage.

---

## API quick checks

```bash
# Health
curl http://localhost:8080/health

# Register
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
```

---

## Development

```bash
go mod download
go run src/main.go

# With hot reload (requires air)
go install github.com/cosmtrek/air@latest
air
```

Run migrations via the app:

```bash
go run src/main.go migrate
go run src/main.go migrate rollback
go run src/main.go migrate status
```

---

## Notes & security

- Always use HTTPS in production (Nginx config is included).
- Use strong, random secrets for JWT and CSRF.
- Keep your Docker image up to date with security patches.

