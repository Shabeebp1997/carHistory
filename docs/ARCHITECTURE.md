# CarHistory — Architecture Overview

This document explains the structure, design decisions, and deployment setup of the CarHistory backend so new developers can get up to speed quickly.

---

## 1. Overview

CarHistory is a **vehicle maintenance tracker**. Users can register, add vehicles, and log maintenance, service, insurance, PUC, tyres, batteries, odometer readings, fuel, expenses, reminders, and documents. Admins can view analytics and manage users.

The backend is a stateless REST API written in Go. It connects to PostgreSQL for persistence and is designed to run as a container-friendly web service.

---

## 2. Tech Stack

| Layer | Technology |
|---|---|
| Language | Go 1.22 |
| Web Framework | Gin |
| Database | PostgreSQL 14+ |
| DB Driver / Mapper | sqlx + lib/pq |
| Migrations | Embedded SQL files, applied automatically on startup |
| Authentication | JWT (access + refresh tokens) |
| Password Hashing | bcrypt |
| Validation | go-playground/validator |
| Background Jobs | robfig/cron |
| Email | SMTP via gomail |
| File Storage | Local filesystem (configurable to cloud later) |

---

## 3. High-Level Architecture

```
Frontend / Postman / Mobile App
        |
        | HTTPS
        v
+---------------------+
|   Render Web Service |
|   Go + Gin API       |
+---------------------+
        |
        | SQL over SSL
        v
+---------------------+
|   PostgreSQL         |
|   (Supabase)         |
+---------------------+
```

The API is **stateless**. All state lives in Postgres. Uploaded files live on the container disk by default, so for production file persistence you should move to object storage (S3, Cloudflare R2, etc.).

---

## 4. Project Structure

```
carHistory/
├── cmd/api/main.go              # Application entry point
├── internal/
│   ├── config/                  # Environment-based configuration
│   ├── database/                # DB connection, migrations, seeding
│   ├── middleware/              # Auth, admin, vehicle-access middleware
│   ├── models/                  # Request/response and DB models
│   ├── repository/              # Database access layer (SQL queries)
│   ├── services/                # Business logic (auth, users, vehicles, email, sms)
│   ├── handlers/                # HTTP handlers (Gin controllers)
│   ├── router/                  # Route registration and middleware wiring
│   └── worker/                  # Background cron jobs (reminders)
├── pkg/
│   ├── jwt/                     # JWT token creation and validation
│   ├── response/                # Standard JSON response helpers
│   └── utils/                   # General utilities
├── migrations/                  # SQL schema files, embedded into the binary
├── docs/                        # Documentation
├── uploads/                     # Local file storage (created at runtime)
├── .env.example                 # Required environment variables
├── go.mod                       # Go module definition
└── README.md                    # Local setup instructions
```

---

## 5. Request Lifecycle

A typical request flows through these layers:

```
HTTP Request
    |
    v
Gin Router (internal/router/router.go)
    |
    v
Middleware (CORS, Auth, Admin, VehicleAccess)
    |
    v
Handler (internal/handlers/*.go)
    |-- validates request body / query params
    |-- calls services or repositories
    |
    v
Repository (internal/repository/*.go)
    |-- executes SQL via sqlx
    |
    v
PostgreSQL
```

Handlers never talk to the database directly. They always go through repositories or services.

---

## 6. Layer Responsibilities

### `config`
- Loads `.env` variables at startup.
- Provides typed config structs (`DBConfig`, `JWTConfig`, etc.).
- Adds sensible defaults for local development.

### `database`
- Opens the PostgreSQL connection pool.
- Runs embedded migrations automatically on startup.
- Seeds the admin user from `ADMIN_EMAIL` / `ADMIN_PASSWORD` if configured.

### `models`
- Defines data structures used for:
  - Database rows (`User`, `Vehicle`, etc.)
  - API request bodies (`RegisterRequest`, `CreateVehicleRequest`, etc.)
  - API response shapes

### `repository`
- Contains all SQL.
- Each repository owns one domain area (users, vehicles, maintenance, etc.).
- Uses parameterized queries to prevent SQL injection.
- Returns errors that handlers translate into HTTP responses.

### `services`
- Holds business logic that spans multiple repositories or external systems.
- `AuthService`: register, login, logout, token refresh, password reset.
- `EmailService` / `SMSService`: send notifications.
- `UserService` / `VehicleService`: domain-specific workflows.

### `handlers`
- Gin HTTP handlers.
- Bind JSON, validate input, call services/repositories, return responses.
- Do not contain business rules.

### `middleware`
- Reusable request interceptors.
- `auth`: validates JWT access tokens.
- `admin`: ensures the user has `is_admin = true`.
- `vehicle_access`: ensures the user is a member of the requested vehicle.

### `worker`
- Cron-based background worker.
- Currently sends reminder notifications for upcoming maintenance/insurance renewals.

### `pkg/response`
- Standard JSON envelope: `{ "success": bool, "data": ..., "message": ..., "error": ... }`.
- Automatically converts nil slices to empty arrays so clients never see `null` for list responses.
- Logs internal errors.

### `pkg/jwt`
- Creates and validates access/refresh tokens.

---

## 7. Database

### Migrations
Migrations live in `migrations/*.sql` and are embedded into the binary with `//go:embed`. They are applied automatically on startup in order by filename:

```
migrations/001_initial_schema.sql
migrations/002_additions.sql
```

To add a new migration, create `migrations/003_....sql` and restart the service.

### Key Tables
- `users` — accounts and admin flag
- `vehicles` — vehicle details, linked to `vehicle_types`
- `vehicle_members` — many-to-many linking users to vehicles with a role
- `maintenance_records`, `service_records`, `service_items` — maintenance history
- `insurance_records`, `puc_records`, `tyre_records`, `battery_records` — vehicle lifecycle
- `odometer_records`, `fuel_logs`, `expenses` — tracking data
- `reminders`, `notifications` — reminders and in-app notifications
- `documents`, `media` — file metadata
- `activity_logs` — audit trail

---

## 8. Authentication

1. User registers or logs in with email/phone + password.
2. `AuthService` verifies the password with bcrypt.
3. On success, the API returns an **access token** (short-lived) and a **refresh token** (long-lived).
4. Protected endpoints require `Authorization: Bearer <access_token>`.
5. Access tokens expire after `JWT_ACCESS_TOKEN_EXPIRY_MINUTES`.
6. Refresh tokens can be exchanged for a new access token at `POST /auth/refresh`.
7. Logout revokes the refresh token.

Admin accounts are created automatically on startup via `ADMIN_EMAIL` and `ADMIN_PASSWORD` environment variables.

---

## 9. File Storage

By default, uploaded files are stored on disk under `./uploads` and served at `/uploads/<path>`.

**Important for production:** Render free instances do not have persistent disk. Uploaded files are lost on redeploy or when the instance sleeps. For a real production deployment, switch to:

- AWS S3
- Cloudflare R2
- Supabase Storage

Update `STORAGE_TYPE`, `STORAGE_LOCAL_PATH`, and `STORAGE_BASE_URL` accordingly.

---

## 10. Background Jobs

A cron worker starts in `cmd/api/main.go` and checks for reminders. When a reminder is due, it can send email or SMS notifications. The schedule is controlled by `REMINDER_CRON_SCHEDULE` (default: daily at 02:00 UTC).

---

## 11. Deployment Architecture

Current production deployment uses **Render** for the Go backend and **Supabase** for PostgreSQL.

```
GitHub Repo
    |
    v
Render Web Service (Go build + run)
    |
    | shared pooler connection
    v
Supabase PostgreSQL
```

### Why Supabase pooler?
Supabase direct connections are IPv6-only on the free plan. Render free instances are IPv4-only, so the backend connects through Supabase's shared pooler on port `5432` with a username like `postgres.<project-ref>`.

### Environment Variables on Render
See `.env.example` for the full list. Critical ones for production:

```text
ENV=production
SERVER_PORT=10000

DB_HOST=aws-1-ap-northeast-1.pooler.supabase.com
DB_PORT=5432
DB_USER=postgres.<project-ref>
DB_PASSWORD=<strong password>
DB_NAME=postgres
DB_SSLMODE=require

JWT_SECRET=<random 32+ character string>
APP_BASE_URL=https://your-frontend.com
CORS_ORIGINS=https://your-frontend.com,https://another-frontend.com

ADMIN_EMAIL=admin@carhistory.app
ADMIN_PASSWORD=<strong password>
```

---

## 12. Adding a New Feature

To add a new domain feature (for example, `warranty_records`):

1. **Migration:** add the table in `migrations/003_warranty_records.sql`.
2. **Model:** add request/response structs in `internal/models/`.
3. **Repository:** add queries in `internal/repository/warranty.go`.
4. **Service:** add business logic in `internal/services/` if needed.
5. **Handler:** add HTTP handlers in `internal/handlers/warranty.go`.
6. **Router:** register routes in `internal/router/router.go`.
7. **Middleware:** apply `auth.AuthMiddleware()` and `vehicleaccess.RequireMember()` where appropriate.
8. **Docs:** update `docs/API_GUIDE.md` with new endpoints.

---

## 13. Common Issues

| Symptom | Likely Cause |
|---|---|
| `Database connected successfully` not in logs | Wrong `DB_HOST`, `DB_USER`, or `DB_PASSWORD` |
| `tenant/user not found` | Supabase pooler username missing `.project-ref` or wrong pooler host |
| `no such host` on DB host | Using direct Supabase host instead of shared pooler |
| `null` in list responses | Now handled by `pkg/response/response.go`; previously meant a nil slice |
| 500 with generic message | Check Render logs; real error is now logged |
| CORS errors from frontend | Add frontend URL to `CORS_ORIGINS` env variable |

---

## 14. Useful Commands

```bash
# Local development
cp .env.example .env
# edit .env
go run ./cmd/api

# Build production binary
go build -o carhistory ./cmd/api

# Run migrations manually (if needed)
# migrations are auto-applied on startup
```
