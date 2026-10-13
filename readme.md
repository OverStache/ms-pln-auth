# MS-PLN-AUTH

A simple authentication microservice built with Go, Gin framework, and PostgreSQL following Clean Architecture principles.

## Features

- User registration
- User login with session token
- User logout (token invalidation)
- Token-based authentication middleware
- Clean Architecture structure

## Project Structure

```
ms-pln-auth/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── ddl/
│   └── init.sql                 # Database schema
├── internal/
│   ├── delivery/
│   │   └── http/
│   │       ├── handler/         # HTTP handlers
│   │       ├── middleware/      # Auth middleware
│   │       └── router/          # Route definitions
│   ├── domain/
│   │   ├── entity/              # Business entities
│   │   └── repository/          # Repository interfaces
│   ├── infrastructure/
│   │   ├── config/              # Configuration
│   │   └── database/            # Database connection
│   ├── repository/
│   │   └── postgres/            # PostgreSQL implementations
│   └── usecase/                 # Business logic
├── .env.example                 # Environment variables template
├── go.mod
├── Makefile
└── readme.md
```

## Prerequisites

- Go 1.21+
- PostgreSQL 12+

## Setup

### 1. Create the database

```bash
# Create database
psql -h localhost -U postgres -c "CREATE DATABASE auth_db;"

# Or use Makefile
make db-create
```

### 2. Run the DDL

```bash
psql -h localhost -U postgres -d auth_db -f ddl/init.sql

# Or use Makefile
make db-setup
```

### 3. Configure environment variables

```bash
cp .env.example .env
# Edit .env with your configuration
```

### 4. Install dependencies

```bash
go mod download
# Or
make deps
```

### 5. Run the application

```bash
go run ./cmd/server
# Or
make run
```

The server will start on port 8080 by default.

## API Endpoints

### Health Check
```
GET /health
```

### Authentication

#### Register
```
POST /api/v1/auth/register
Content-Type: application/json

{
    "username": "johndoe",
    "email": "john@example.com",
    "password": "password123"
}
```

#### Login
```
POST /api/v1/auth/login
Content-Type: application/json

{
    "username": "johndoe",
    "password": "password123"
}
```

Response:
```json
{
    "token": "your-session-token",
    "expires_at": "2024-01-02T15:04:05Z",
    "user": {
        "id": 1,
        "username": "johndoe",
        "email": "john@example.com"
    }
}
```

#### Logout (Protected)
```
POST /api/v1/auth/logout
Authorization: Bearer <token>
```

#### Get Current User (Protected)
```
GET /api/v1/auth/me
Authorization: Bearer <token>
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| SERVER_PORT | Server port | 8080 |
| GIN_MODE | Gin mode (debug/release) | debug |
| DB_HOST | Database host | localhost |
| DB_PORT | Database port | 5432 |
| DB_USER | Database user | postgres |
| DB_PASSWORD | Database password | postgres |
| DB_NAME | Database name | auth_db |
| DB_SSLMODE | SSL mode | disable |
| AUTH_TOKEN_DURATION | Token validity duration | 24h |
| AUTH_BCRYPT_COST | Bcrypt cost factor | 10 |

## Testing with cURL

```bash
# Register a new user
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","email":"test@example.com","password":"testpass123"}'

# Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"testpass123"}'

# Get current user (replace TOKEN with actual token)
curl http://localhost:8080/api/v1/auth/me \
  -H "Authorization: Bearer TOKEN"

# Logout
curl -X POST http://localhost:8080/api/v1/auth/logout \
  -H "Authorization: Bearer TOKEN"
```

## License

MIT
