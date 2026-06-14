# Task Manager

A full-stack task management application: a **Go** REST API backed by **PostgreSQL**, with a **Next.js** frontend. Users sign up, log in, and manage their own tasks with filtering, search, sorting, pagination, and real-time updates.

> Built for the Rival Full-Stack Developer Assessment.

---

## Table of contents

- [Tech stack](#tech-stack)
- [Features](#features)
- [Project structure](#project-structure)
- [Quick start (Docker)](#quick-start-docker)
- [Manual setup](#manual-setup)
- [Environment variables](#environment-variables)
- [API reference](#api-reference)
- [Testing](#testing)
- [Deployment](#deployment)
- [Assumptions & trade-offs](#assumptions--trade-offs)

---

## Tech stack

| Layer    | Choice                                                              |
| -------- | ------------------------------------------------------------------ |
| Frontend | Next.js 16 (App Router), React 19, TypeScript, Tailwind CSS v4     |
| Backend  | Go 1.22, chi router, pgx (PostgreSQL driver), JWT auth             |
| Database | PostgreSQL 16 (Neon-compatible)                                    |
| Auth     | JWT (HS256) + bcrypt password hashing                              |
| Infra    | Docker / docker-compose, GitHub Actions CI                        |

The assessment lists Go as the preferred backend, which is what this project uses.

## Features

### Core requirements

- **CRUD REST API** for tasks (`title`, `description`, `status`, `priority`, `due date`) with input validation, proper HTTP status codes, and a consistent error envelope.
- **Auth**: signup/login with JWT, bcrypt-hashed passwords, protected task routes. Users can only see and modify **their own** tasks. Session persists across refresh (token in `localStorage`, re-validated on load via `/api/auth/me`).
- **Frontend**: task list with status filter + pagination, create/edit form with client-side validation, complete/delete from the UI, graceful loading/empty/error states, responsive (mobile + desktop).
- **Search & sort**: search by title; sort by due date, priority, or created date; filters + search + sort all compose together.
- **Tests**: Go unit tests for auth (bcrypt + JWT), sorting logic, and validation.

### Bonus features included

- ✅ **Role-based access** — an `admin` role can view tasks across all users (scope toggle in the UI).
- ✅ **Real-time updates** — task changes stream to the client over **SSE** (`/api/tasks/stream`).
- ✅ **Optimistic UI** — completing and deleting tasks update instantly and roll back on failure.
- ✅ **Activity log** — every create/update is recorded (`/api/tasks/:id/activity`).
- ✅ **Dockerized setup** — `docker compose up` runs DB + API + frontend.
- ✅ **CI pipeline** — GitHub Actions runs backend tests and frontend lint/build on push.
- ✅ **Dark mode** — theme toggle with a persisted, no-flash preference.

## Project structure

```
taskmanager/
├── backend/                 # Go REST API
│   ├── cmd/server/          # main entrypoint
│   ├── internal/
│   │   ├── auth/            # JWT + bcrypt
│   │   ├── config/         # env loading
│   │   ├── database/       # pgx pool + migration runner
│   │   ├── events/         # in-memory SSE hub
│   │   ├── httpx/          # JSON + error helpers
│   │   ├── middleware/     # JWT auth middleware
│   │   ├── models/         # domain types
│   │   ├── server/         # router wiring
│   │   ├── tasks/          # task repository + handlers
│   │   ├── users/          # user repository + auth handlers
│   │   └── validate/       # request validation
│   ├── migrations/         # SQL (embedded, applied on startup)
│   └── Dockerfile
├── frontend/                # Next.js app
│   └── src/
│       ├── app/            # routes: /, /login, /signup, /tasks
│       ├── components/     # UI components
│       └── lib/            # API client, auth/theme context, helpers
├── .github/workflows/ci.yml
└── docker-compose.yml
```

## Quick start (Docker)

The fastest way to run everything locally — no Go, Node, or Postgres needed, just Docker:

```bash
docker compose up --build
```

Then open **http://localhost:3000**. The API is at **http://localhost:8080**, Postgres on `5432`. Migrations run automatically on backend startup.

## Manual setup

### Prerequisites

- Go 1.22+
- Node.js 20+
- A PostgreSQL database. The easiest is a free [Neon](https://neon.tech) project; any Postgres 14+ works.

### 1. Backend

```bash
cd backend
cp .env.example .env
# Edit .env: set DATABASE_URL to your Neon/Postgres string and a strong JWT_SECRET.
go run ./cmd/server
```

The server applies migrations on startup and listens on `http://localhost:8080`. Verify with:

```bash
curl http://localhost:8080/health   # {"status":"ok"}
```

> **Neon tip:** the connection string must end with `?sslmode=require`.

### 2. Frontend

```bash
cd frontend
cp .env.example .env.local
# NEXT_PUBLIC_API_URL should point at the backend (default http://localhost:8080)
npm install
npm run dev
```

Open **http://localhost:3000**, sign up, and start adding tasks.

### Creating an admin user

New accounts are created with the `user` role. To grant admin (to test cross-user viewing), run one SQL statement against your database:

```sql
UPDATE users SET role = 'admin' WHERE email = 'you@example.com';
```

Log out and back in so the new role is reflected in your token.

## Environment variables

### Backend (`backend/.env`)

| Variable        | Required | Default                 | Description                                   |
| --------------- | -------- | ----------------------- | --------------------------------------------- |
| `DATABASE_URL`  | ✅       | —                       | PostgreSQL connection string                  |
| `JWT_SECRET`    | ✅       | —                       | Secret for signing JWTs (use 32+ random bytes)|
| `PORT`          | —        | `8080`                  | Port to listen on                             |
| `JWT_TTL_HOURS` | —        | `72`                    | Token lifetime in hours                       |
| `CORS_ORIGINS`  | —        | `http://localhost:3000` | Comma-separated allowed frontend origins      |

### Frontend (`frontend/.env.local`)

| Variable              | Required | Default                 | Description              |
| --------------------- | -------- | ----------------------- | ------------------------ |
| `NEXT_PUBLIC_API_URL` | ✅       | `http://localhost:8080` | Base URL of the Go API   |

See `.env.example` in each folder for a copy-paste template.

## API reference

Base URL: `/api`. All task routes require `Authorization: Bearer <token>`.

| Method   | Endpoint               | Description                                       |
| -------- | ---------------------- | ------------------------------------------------- |
| `POST`   | `/auth/signup`         | Create account, returns `{ token, user }`         |
| `POST`   | `/auth/login`          | Log in, returns `{ token, user }`                 |
| `GET`    | `/auth/me`             | Current user (used to restore session)            |
| `POST`   | `/tasks`               | Create a task                                     |
| `GET`    | `/tasks`               | List tasks (filter/search/sort/paginate)          |
| `GET`    | `/tasks/:id`           | Fetch one task                                    |
| `PATCH`  | `/tasks/:id`           | Partial update                                    |
| `DELETE` | `/tasks/:id`           | Delete a task                                     |
| `GET`    | `/tasks/:id/activity`  | Change history for a task                         |
| `GET`    | `/tasks/stream`        | SSE stream of task changes for the current user   |

### List query parameters

| Param      | Values                                 | Notes                                  |
| ---------- | -------------------------------------- | -------------------------------------- |
| `status`   | `todo` `in_progress` `done`            | Filter by status                       |
| `search`   | any string                             | Case-insensitive title match           |
| `sort`     | `created_at` `due_date` `priority`     | Default `created_at`                   |
| `order`    | `asc` `desc`                           | Default `desc`                         |
| `page`     | integer ≥ 1                            | Default `1`                            |
| `pageSize` | integer (max 100)                      | Default `10`                           |
| `scope`    | `all`                                  | Admin only — list across all users     |

### Error format

Every error returns a consistent envelope:

```json
{ "error": { "message": "validation failed", "details": { "title": "is required" } } }
```

### Example

```bash
# Sign up
curl -X POST http://localhost:8080/api/auth/signup \
  -H "Content-Type: application/json" \
  -d '{"email":"demo@example.com","password":"password123"}'

# Create a task (use the returned token)
curl -X POST http://localhost:8080/api/tasks \
  -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
  -d '{"title":"Ship the assessment","priority":"high","dueDate":"2026-06-20T00:00:00Z"}'
```

## Testing

```bash
cd backend
go test ./...
```

Covers password hashing, JWT issue/verify (including tampered and expired tokens), the priority/due-date sort logic, and request validation. CI runs these on every push along with `go vet`, plus the frontend lint and production build.

## Deployment

The app is split so each part deploys independently:

- **Database** — Neon (managed Postgres).
- **Backend** — any container/host platform (Render, Railway, Fly.io). Build with the included `backend/Dockerfile`; set `DATABASE_URL`, `JWT_SECRET`, and `CORS_ORIGINS` (your deployed frontend URL).
- **Frontend** — Vercel. Set `NEXT_PUBLIC_API_URL` to the deployed backend URL.

Remember to add the deployed frontend origin to the backend's `CORS_ORIGINS`.

## Assumptions & trade-offs

- **Token storage in `localStorage`.** Simple and survives refresh. The trade-off is XSS exposure vs. httpOnly cookies; given a token-based SPA talking to a separate API origin, `localStorage` keeps CORS/CSRF simple. For production I'd consider httpOnly refresh cookies + short-lived access tokens.
- **Migrations run on startup** from embedded, idempotent SQL. Zero-dependency and great for this scope; a larger project would use versioned migrations (golang-migrate) with explicit up/down files.
- **SSE over WebSockets.** Task updates are one-directional (server→client), so SSE is lighter and needs no extra protocol. The hub is in-memory, so it's single-instance; multi-instance would need Redis/NATS pub/sub.
- **Real-time refetches** the current page on an event rather than surgically merging, which keeps sort/pagination correct at the cost of an extra request.
- **Admin role** is assigned via SQL, not a self-serve UI — appropriate for an internal role.
- **`due_date` is a calendar date** captured at UTC midnight; time-of-day was out of scope.
- Passwords are capped at 72 bytes (bcrypt's limit) and require a minimum of 8 characters.
```
