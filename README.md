# Chirpy

A RESTful backend API built in Go that implements user authentication, authorization, JWT sessions, refresh tokens, webhooks, and CRUD operations for a microblogging platform.

This project was built to explore backend engineering concepts from first principles while learning Go and PostgreSQL.

---

## Features

### Authentication

- User registration
- User login
- Password hashing with bcrypt
- JWT access tokens
- Refresh token authentication
- Token revocation
- Secure Bearer token authentication

### Authorization

- Users can only create chirps as themselves
- Users can only update their own account
- Users can only delete their own chirps
- Protected endpoints using JWT middleware

### Chirps

- Create chirps
- Retrieve all chirps
- Retrieve a single chirp
- Delete chirps
- Filter chirps by author
- Sort ascending or descending by creation date
- Automatic profanity filtering

### User Management

- Register users
- Login
- Update email
- Update password
- Chirpy Red membership support

### Webhooks

- Polka webhook endpoint
- Idempotent webhook handling
- Automatic Chirpy Red upgrades

### Database

- PostgreSQL
- SQLC for type-safe queries
- Goose database migrations

---

## Tech Stack

- Go
- PostgreSQL
- SQLC
- Goose
- JWT
- bcrypt
- Standard net/http package

---

## API

### Users

```
POST   /api/users
PUT    /api/users
POST   /api/login
POST   /api/refresh
POST   /api/revoke
```

### Chirps

```
GET     /api/chirps
GET     /api/chirps/{id}
POST    /api/chirps
DELETE  /api/chirps/{id}
```

Supports

```
GET /api/chirps?author_id=<uuid>

GET /api/chirps?sort=asc

GET /api/chirps?sort=desc
```

### Webhooks

```
POST /api/polka/webhooks
```

---

## Security

- Passwords are never stored in plaintext
- Passwords are hashed using bcrypt
- JWT access tokens expire after one hour
- Refresh tokens expire after sixty days
- Refresh tokens can be revoked
- Authorization enforced on protected endpoints

---

## Project Structure

```
.
├── sql/
│   ├── schema/
│   └── queries/
│
├── internal/
│   ├── auth/
│   └── database/
│
├── main.go
├── sqlc.yaml
├── goose migrations
└── README.md
```

---

## Running

Clone the repository

```bash
git clone <repo>
cd chirpy
```

Create a `.env`

```env
DB_URL=postgres://...
SECRET=your-secret
PLATFORM=dev
```

Run migrations

```bash
goose postgres "$DB_URL" up
```

Generate SQLC code

```bash
sqlc generate
```

Start the server

```bash
go run .
```

---

## What I Learned

This project helped me gain hands-on experience with:

- REST API design
- HTTP routing
- JSON encoding/decoding
- Authentication vs Authorization
- JWT lifecycle
- Refresh token sessions
- Password hashing
- SQL schema design
- PostgreSQL
- Database migrations
- SQLC
- CRUD operations
- Webhooks
- Request validation
- Error handling
- Secure API development

---

## Future Improvements

- Pagination
- Rate limiting
- Caching
- Structured logging
- Docker support
- CI/CD
- Unit and integration tests
- OpenAPI/Swagger documentation
- Metrics and observability

---

Built with Go while learning backend engineering.
