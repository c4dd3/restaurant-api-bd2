# Reserva Inteligente de Restaurantes — API REST

API REST desarrollada en **Go (Gin)** con autenticación **JWT** en un **servicio separado**, base de datos **PostgreSQL**, y contenedorización con **Docker**.

---

## Arquitectura de servicios

```
┌─────────────────────────────────────────────────────┐
│                  Docker Compose                     │
│                                                     │
│  ┌──────────────┐   ┌──────────────┐   ┌─────────┐ │
│  │ auth-service │   │     api      │   │   db    │ │
│  │   :8081      │   │   :8080      │   │  :5432  │ │
│  │              │   │              │   │         │ │
│  │ POST         │   │ GET  /users  │   │ Postgres│ │
│  │ /auth/       │   │ POST /rest.. │   │   16    │ │
│  │   register   │   │ POST /menus  │   │         │ │
│  │   login      │   │ POST /reserv │   │         │ │
│  └──────┬───────┘   └──────┬───────┘   └────┬────┘ │
│         └──────────────────┴────────────────┘      │
│                   comparten JWT_SECRET              │
└─────────────────────────────────────────────────────┘
```

El `auth-service` emite tokens JWT. La `api` los valida usando el mismo `JWT_SECRET` — sin llamadas HTTP entre servicios.

---

## Tecnologías

| Herramienta | Uso |
|---|---|
| Go 1.22 + Gin | Framework HTTP |
| PostgreSQL 16 | Base de datos relacional |
| JWT (HS256) | Autenticación stateless |
| Docker + Compose | Contenedorización |
| bcrypt | Hash de contraseñas |

---

## Estructura del Proyecto

```
restaurant-api/
├── cmd/
│   ├── api/          # Entrypoint del servicio API
│   └── auth/         # Entrypoint del servicio de autenticación
├── internal/
│   ├── auth/         # Lógica JWT (GenerateToken, ValidateToken)
│   ├── authrouter/   # Router del auth-service (reutilizable en tests)
│   ├── handlers/     # Handlers HTTP + interfaces de repositorio
│   ├── middleware/   # Auth middleware y AdminOnly
│   ├── models/       # Structs y DTOs
│   ├── repository/   # Acceso a PostgreSQL
│   └── router/       # Router del API service
├── tests/            # Tests unitarios e integración
├── Dockerfile        # Imagen del API service
├── Dockerfile.auth   # Imagen del auth-service
├── docker-compose.yml
└── .env.example
```

---

## Cómo correr el proyecto

### Con Docker Compose (recomendado)

```bash
cp .env.example .env
docker compose up --build
```

| Servicio | URL |
|---|---|
| Auth service | http://localhost:8081 |
| API service | http://localhost:8080 |

---

## Endpoints

### Auth Service (`localhost:8081`) — público

| Método | Endpoint | Descripción |
|---|---|---|
| `POST` | `/auth/register` | Registro de usuario |
| `POST` | `/auth/login` | Login y obtención de JWT |

### API Service (`localhost:8080`) — requiere JWT

| Método | Endpoint | Descripción | Rol |
|---|---|---|---|
| `GET` | `/users/me` | Perfil del usuario autenticado | todos |
| `PUT` | `/users/:id` | Actualizar usuario | propio/admin |
| `DELETE` | `/users/:id` | Eliminar usuario | propio/admin |
| `POST` | `/restaurants` | Registrar restaurante | admin |
| `GET` | `/restaurants` | Listar restaurantes | todos |
| `POST` | `/menus` | Crear menú | admin |
| `GET` | `/menus/:id` | Ver menú | todos |
| `PUT` | `/menus/:id` | Actualizar menú | admin |
| `DELETE` | `/menus/:id` | Eliminar menú | admin |
| `POST` | `/reservations` | Crear reserva | todos |
| `DELETE` | `/reservations/:id` | Cancelar reserva | propio/admin |
| `POST` | `/orders` | Realizar pedido | todos |
| `GET` | `/orders/:id` | Ver pedido | propio/admin |

---

## Ejemplos de uso

```bash
# 1. Registrar (auth-service)
curl -X POST http://localhost:8081/auth/register \
  -H "Content-Type: application/json" \
  -d '{"name":"Ana","email":"ana@test.com","password":"secret123","role":"client"}'

# 2. Login (auth-service)
curl -X POST http://localhost:8081/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"ana@test.com","password":"secret123"}'

# 3. Usar el token en la API
curl http://localhost:8080/restaurants \
  -H "Authorization: Bearer <TOKEN>"
```

---

## Correr pruebas

```bash
# Unitarias (sin BD)
go test ./tests/... -run "TestJWT|TestOrder|TestMenu|TestUser|TestReservation|TestMiddleware|TestRouter" -v

# Auth service
go test ./cmd/auth/... -v

# API service
go test ./cmd/api/... -v

# Integración (requiere BD)
TEST_DB_HOST=localhost TEST_DB_NAME=restaurant_test go test ./tests/... -v -count=1

# Coverage total
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

---

## Variables de entorno

| Variable | Default | Descripción |
|---|---|---|
| `DB_HOST` | `localhost` | Host de PostgreSQL |
| `DB_PORT` | `5432` | Puerto de PostgreSQL |
| `DB_USER` | `postgres` | Usuario de BD |
| `DB_PASSWORD` | `postgres` | Contraseña de BD |
| `DB_NAME` | `restaurant_db` | Nombre de la BD |
| `JWT_SECRET` | *(ver .env)* | Secreto compartido entre servicios |
| `PORT` | `8080` | Puerto del API service |
| `AUTH_PORT` | `8081` | Puerto del auth-service |
| `GIN_MODE` | `debug` | Modo Gin (`debug`/`release`) |
