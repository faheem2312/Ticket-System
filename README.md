# Ticket System — Backend Intern Assignment

A REST API ticket system built in Go with JWT authentication, ownership-based access control, and in-memory storage.

---

## Tech Stack

- **Go 1.21**
- **gorilla/mux** — routing
- **golang-jwt/jwt** — JWT auth
- **bcrypt** — password hashing
- **In-memory store** (thread-safe via `sync.RWMutex`)

---

## Local Run (without Docker)

```bash
git clone https://github.com/faheem2312/Ticket-System.git
cd Ticket-System

copy .env.example .env   # Windows
# cp .env.example .env   # Mac/Linux

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

## Deployed URLs

| | URL |
|---|---|
| **App** | https://ticket-system-k8ek.onrender.com |
| **Health check** | https://ticket-system-k8ek.onrender.com/health |

> Note: Render free tier spins down after inactivity. First request may take ~30 seconds to wake up.

---

## API Reference

### Auth

#### `POST /auth/register`
```json
{"email": "user@example.com", "password": "secret123"}
```
Response `201`:
```json
{"id": "...", "email": "user@example.com", "created_at": "..."}
```

#### `POST /auth/login`
```json
{"email": "user@example.com", "password": "secret123"}
```
Response `200`:
```json
{"token": "<jwt>"}
```

---

### Tickets

All ticket endpoints require the header:
```
Authorization: Bearer <token>
```

#### `POST /tickets` — Create ticket
```json
{"title": "Bug in login", "description": "Users can't log in"}
```

#### `GET /tickets` — List my tickets

#### `GET /tickets/{id}` — Get ticket by ID

#### `PATCH /tickets/{id}/status` — Update ticket status
```json
{"status": "in_progress"}
```

**Status flow:** `open` → `in_progress` → `closed`  
Closed tickets cannot be reopened.

---

## Running Tests

```bash
go test ./...
```

---

## Assumptions

- Storage is in-memory; data resets on restart.
- JWT tokens expire after 24 hours.
- `JWT_SECRET` defaults to a placeholder if not set — always override in production via environment variable.
- Ownership: users can only view/update their own tickets. Other users' tickets return `404` (not `403`) to avoid leaking ticket existence.
- Email is normalized (lowercased, trimmed) on register and login.
