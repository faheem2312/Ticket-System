# Ticket System — Backend Intern Assignment

A REST API ticket system built in Go with JWT authentication, ownership-based access control, and in-memory storage.

---

## Tech Stack

- **Go 1.21**
- **gorilla/mux** — routing
- **golang-jwt/jwt** — JWT auth
- **bcrypt** — password hashing
- **In-memory store** (thread-safe, `sync.RWMutex`)

---

## Local Run (without Docker)

```bash
git clone <your-repo-url>
cd ticket-system

cp .env.example .env   # set JWT_SECRET

go mod tidy
go run ./cmd/main.go
```

Server runs on `http://localhost:8080`.

---

## Docker Run

```bash
docker build -t ticket-system .
docker run -p 8080:8080 -e JWT_SECRET=your-secret ticket-system
```

---

## Health Check

```bash
curl http://localhost:8080/health
# {"status":"ok"}
```

---

## API Reference

### Auth

#### Register
```
POST /auth/register
Content-Type: application/json

{"email": "user@example.com", "password": "secret123"}
```

#### Login
```
POST /auth/login
Content-Type: application/json

{"email": "user@example.com", "password": "secret123"}
```
Returns `{"token": "<jwt>"}`. Use this token in all protected requests.

---

### Tickets (all require `Authorization: Bearer <token>`)

#### Create Ticket
```
POST /tickets
{"title": "Bug in login", "description": "Users can't log in"}
```

#### List My Tickets
```
GET /tickets
```

#### Get Ticket by ID
```
GET /tickets/{id}
```

#### Update Ticket Status
```
PATCH /tickets/{id}/status
{"status": "in_progress"}
```

**Status flow:** `open` → `in_progress` → `closed`
Closed tickets cannot be reopened.

---

## Deployment

Deployed URL: **https://your-deployed-url.com**  
Health check: **https://your-deployed-url.com/health**

### Recommended free platforms

| Platform | Notes |
|----------|-------|
| [Render](https://render.com) | Free tier, Docker support, easy setup |
| [Railway](https://railway.app) | Free tier, auto-detects Go |
| [Fly.io](https://fly.io) | Free tier, `flyctl deploy` |

---

## Assumptions

- Storage is in-memory; data resets on restart. Swap `store.New()` for a DB-backed store if persistence is needed.
- JWT tokens expire after 24 hours.
- `JWT_SECRET` defaults to a placeholder if not set — always set it in production.
- Ownership check: users can only view/update their own tickets. Other users' tickets return 404 (not 403) to avoid leaking existence.
