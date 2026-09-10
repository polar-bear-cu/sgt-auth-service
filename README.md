# Subglutee Project - Auth Service

Google OAuth2 + JWT

- REST `:8084` - frontend/gateway (`/api/v1/auth/...`)
- gRPC client - user-service (`FindOrCreateUser`)
- Postgres - เก็บ refresh token

### Migrations

golang-migrate, ไฟล์คู่ `up`/`down` ใน `migrations/`

```terminal
make migrate-create name=add_something
make migrate-up
make migrate-down
```

แก้ schema = migration ใหม่เสมอ ห้ามแก้ไฟล์ที่ merge ไปแล้ว
