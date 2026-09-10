# Subglutee Project - Auth Service

Google OAuth2 login -> ออก JWT (access) + refresh token

- REST `:8084` - frontend/gateway (`/api/v1/auth/...`)
- gRPC client -> user-service `FindOrCreateUser` (ตอน callback)
- Postgres - เก็บ refresh token (hash)

### Structure

```
routes/         map path -> controller
controllers/    HTTP <-> DTO, state cookie
usecases/       oauth flow, token issue/refresh/revoke
repositories/   refresh_tokens db
clients/        gRPC client -> user-service (proto adapter)
tokens/         JWT sign/parse
dtos/           request/response struct
models/         db struct
config/         env loader + db pool
migrations/     sql migration
```

### Prerequisite

- Go 1.26
- Docker + Docker Compose
- `make` - `winget install ezwinports.make`
- tools:

```terminal
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
go install github.com/evilmartians/lefthook@latest
```

### Google OAuth setup (ครั้งเดียว)

1. Google Cloud Console -> OAuth consent screen (External) -> เพิ่ม test users
2. Credentials -> OAuth client ID -> Web application
3. Authorized redirect URIs: `http://localhost:8084/api/v1/auth/google/callback`
4. เอา Client ID / Secret ใส่ `.env`

### Setup

```terminal
git clone https://github.com/polar-bear-cu/sgt-auth-service.git
cd sgt-auth-service
lefthook install
cp .env.example .env
go mod download
make compose-up
make migrate-up
```

user-service ต้องรันด้วย (gRPC `:50052`) — callback ถึงจะสำเร็จ

### Useful Commands

Check `Makefile`

### Migrations

golang-migrate, ไฟล์คู่ `up`/`down` ใน `migrations/`

```terminal
make migrate-create name=add_something
make migrate-up
make migrate-down
```

แก้ schema = migration ใหม่เสมอ ห้ามแก้ไฟล์ที่ merge ไปแล้ว

### Dev tools

- pgweb: `localhost:8085` - ดู local db

#### ถ้าแก้ proto พร้อม service นี้

```terminal
cd ..
go work use ./sgt-proto ./sgt-auth-service
```
