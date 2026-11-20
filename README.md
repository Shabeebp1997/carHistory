# Vehicle Maintenance Tracker — Backend

Go REST API backend for the Vehicle Maintenance Tracker application.

## Tech Stack

- **Language**: Go 1.22
- **Framework**: Gin
- **Database**: PostgreSQL (via sqlx + lib/pq)
- **Auth**: JWT (access + refresh tokens)
- **Background Jobs**: robfig/cron
- **Email**: SMTP via gomail

## Project Structure

```
carHistory/
├── cmd/api/main.go              # Entry point
├── internal/
│   ├── config/                  # Configuration loading
│   ├── database/                # DB connection
│   ├── middleware/              # Auth & vehicle access middleware
│   ├── models/                  # Data models
│   ├── repository/              # Database access layer
│   ├── services/                # Business logic
│   ├── handlers/                # HTTP request handlers
│   ├── router/                  # Route definitions
│   └── worker/                  # Background reminder worker
├── pkg/
│   ├── jwt/                     # JWT helpers
│   ├── response/                # Standard API response
│   └── utils/                   # Utility functions
└── migrations/                  # SQL migration files
```

## API Base Path

```
/api/v1/
```

## Setup

### 1. Prerequisites
- Go 1.22+
- PostgreSQL 14+

### 2. Environment
```bash
cp .env.example .env
# Edit .env with your configuration
```

### 3. Database
```bash
createdb carhistory
make migrate-up
```

### 4. Run
```bash
go mod tidy
make run
```

## API Endpoints

### Auth
| Method | Path | Description |
|--------|------|-------------|
| POST | /api/v1/auth/register | Register new user |
| POST | /api/v1/auth/login | Login |
| POST | /api/v1/auth/logout | Logout |
| POST | /api/v1/auth/refresh | Refresh access token |
| POST | /api/v1/auth/forgot-password | Request password reset |
| POST | /api/v1/auth/reset-password | Reset password |

### Users
| Method | Path | Description |
|--------|------|-------------|
| GET | /api/v1/users/me | Get current user profile |
| PUT | /api/v1/users/me | Update profile |
| PUT | /api/v1/users/me/password | Change password |

### Vehicles
| Method | Path | Description |
|--------|------|-------------|
| GET | /api/v1/vehicles | List all vehicles |
| POST | /api/v1/vehicles | Add vehicle |
| GET | /api/v1/vehicles/:id | Get vehicle |
| PUT | /api/v1/vehicles/:id | Update vehicle |
| DELETE | /api/v1/vehicles/:id | Archive vehicle |
| GET | /api/v1/vehicles/:id/members | List members |
| POST | /api/v1/vehicles/:id/invitations | Invite user |
| DELETE | /api/v1/vehicles/:id/members/:userId | Remove member |

### Maintenance, Services, Insurance, PUC, Tyres, Battery, Odometer
Each resource supports CRUD under `/api/v1/vehicles/:id/<resource>`.

### Reminders & Notifications
| Method | Path | Description |
|--------|------|-------------|
| GET | /api/v1/vehicles/:id/reminders | List reminders |
| POST | /api/v1/vehicles/:id/reminders | Create reminder |
| GET | /api/v1/notifications | Get user notifications |

### Documents & Expenses
Full CRUD under `/api/v1/vehicles/:id/documents` and `/api/v1/vehicles/:id/expenses`.

### Dashboard & Reports
| Method | Path | Description |
|--------|------|-------------|
| GET | /api/v1/dashboard | User dashboard |
| GET | /api/v1/vehicles/:id/reports/summary | Vehicle summary |
| GET | /api/v1/vehicles/:id/reports/expenses | Expense report |

## Vehicle Roles

| Role | Permissions |
|------|-------------|
| owner | Full access including sharing management |
| editor | Add/edit records, upload documents |
| viewer | View only |

## Background Worker

The reminder worker runs on a configurable cron schedule (default: 2:00 AM daily).
It checks for upcoming and overdue reminders and sends email/SMS notifications.
