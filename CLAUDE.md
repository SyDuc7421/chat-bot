# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Start dependencies (MySQL + Redis)
docker-compose up -d

# Run the server (generates Swagger docs first)
make run

# Build binary to tmp/main
make build

# Run all tests
make test

# Run tests with coverage
make test-cover

# Run a single test
go test -v ./controllers -run TestRegister

# Code quality checks (fmt + tidy + vet)
make checks

# Generate/update Swagger docs
make swagger

# Pre-push validation
make pre-push
```

Swagger UI is available at `http://localhost:8080/swagger/index.html` when the server is running.

## Architecture

**Entry point:** `main.go` loads `.env`, initializes config, connects MySQL (with auto-migrate) and Redis, then starts the Gin server via `routes.SetupRouter()`.

**Request flow:** Gin router → CORS middleware → `AuthMiddleware` (JWT validation + Redis session check) → controller handler → service (for LLM calls)

**Auth pattern:** Login generates a JWT access token (15 min) + refresh token (7 days). A session record is stored in Redis under `session:<sessionID>` with the user ID. `AuthMiddleware` validates the JWT *and* checks that the Redis session still exists — logout is implemented by deleting the Redis key.

**Message creation with AI:** When `POST /messages` is called with `role: "user"`, the controller fetches the last 10 messages for context, calls `services.GetChatbotResponse()` (OpenAI `gpt-4o-mini`), and automatically saves the assistant response as a second message.

**Global singletons:** `database.DB` (*gorm.DB), `database.Redis` (*redis.Client), `config.App` (*Config) — set at startup and used directly by controllers.

**Test setup:** Tests use an in-memory SQLite DB (`controllers/setup_test.go`). `GetTestRouter()` returns a Gin router with middleware that injects `userID=1`, bypassing real JWT auth.

## Key Packages

| Path | Purpose |
|------|---------|
| `config/` | Env-based config loading |
| `database/` | MySQL & Redis connection setup |
| `routes/` | Route definitions, CORS, middleware wiring |
| `controllers/` | HTTP handlers + tests |
| `middleware/` | JWT + Redis session auth middleware |
| `models/` | GORM models: User, Conversation, Message |
| `services/` | OpenAI LLM integration |
| `dtos/` | Request/response structs with validation tags |
| `utils/` | JWT generation/validation, bcrypt password helpers |

## Environment Variables

Copy `.env.example` to `.env`. Key variables:

- `PORT` — server port (default `8080`)
- `DB_*` — MySQL connection (host, port, user, password, name)
- `REDIS_ADDR`, `REDIS_PASSWORD`, `REDIS_DB` — Redis connection
- `JWT_SECRET_KEY` — secret for signing JWTs
- `OPENAI_API_KEY` — OpenAI API key for LLM responses
- `FRONTEND_BASE_URL` — added to CORS allowed origins (alongside `localhost:3000` and `localhost:5173`)
