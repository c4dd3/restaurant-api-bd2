# Reserva Inteligente de Restaurantes — API REST

API REST desarrollada en **Go (Gin)** con autenticación **JWT**, base de datos **PostgreSQL**, y contenedorización con **Docker**.

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
├── cmd/api/          # Punto de entrada (main.go)
├── internal/
│   ├── auth/         # Servicio JWT
│   ├── handlers/     # Manejadores HTTP (auth, users, restaurants, menus, reservations, orders)
│   ├── middleware/   # Middleware de autenticación y autorización
│   ├── models/       # Estructuras de datos y DTOs
│   ├── repository/   # Acceso a base de datos
│   └── router/       # Configuración de rutas
├── tests/            # Tests unitarios e integración
├── Dockerfile
├── docker-compose.yml
└── .env.example
```

---

## Cómo correr el proyecto

### 1. Con Docker Compose (recomendado)

```bash
# Clonar el repositorio
git clone <repo-url>
cd restaurant-api

# Copiar variables de entorno
cp .env.example .env

# Levantar servicios
docker compose up --build
```

La API quedará disponible en `http://localhost:8080`.

### 2. Local (sin Docker)

Requisitos: Go 1.22+, PostgreSQL corriendo.

```bash
cp .env.example .env
# editar .env con sus credenciales de BD

go mod download
go run ./cmd/api
```

---

## Variables de Entorno

| Variable | Default | Descripción |
|---|---|---|
| `DB_HOST` | `localhost` | Host de PostgreSQL |
| `DB_PORT` | `5432` | Puerto de PostgreSQL |
| `DB_USER` | `postgres` | Usuario de BD |
| `DB_PASSWORD` | `postgres` | Contraseña de BD |
| `DB_NAME` | `restaurant_db` | Nombre de la BD |
| `JWT_SECRET` | *(ver .env)* | Llave secreta para firmar tokens |
| `PORT` | `8080` | Puerto del servidor |
| `GIN_MODE` | `debug` | Modo de Gin (`debug`/`release`) |

---

## Endpoints

### Autenticación (públicos)

| Método | Endpoint | Descripción |
|---|---|---|
| `POST` | `/auth/register` | Registro de usuario |
| `POST` | `/auth/login` | Login y obtención de JWT |

### Usuarios (requiere JWT)

| Método | Endpoint | Descripción |
|---|---|---|
| `GET` | `/users/me` | Perfil del usuario autenticado |
| `PUT` | `/users/:id` | Actualizar usuario |
| `DELETE` | `/users/:id` | Eliminar usuario |

### Restaurantes (requiere JWT)

| Método | Endpoint | Descripción | Rol |
|---|---|---|---|
| `POST` | `/restaurants` | Registrar restaurante | admin |
| `GET` | `/restaurants` | Listar restaurantes | todos |

### Menús (requiere JWT)

| Método | Endpoint | Descripción | Rol |
|---|---|---|---|
| `POST` | `/menus` | Crear menú | admin |
| `GET` | `/menus/:id` | Ver menú | todos |
| `PUT` | `/menus/:id` | Actualizar menú | admin |
| `DELETE` | `/menus/:id` | Eliminar menú | admin |

### Reservas (requiere JWT)

| Método | Endpoint | Descripción |
|---|---|---|
| `POST` | `/reservations` | Crear reserva |
| `DELETE` | `/reservations/:id` | Cancelar reserva |

### Pedidos (requiere JWT)

| Método | Endpoint | Descripción |
|---|---|---|
| `POST` | `/orders` | Realizar pedido |
| `GET` | `/orders/:id` | Ver detalle de pedido |

---

## Ejemplos de uso (curl)

```bash
# Registrar usuario
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"name":"Ana","email":"ana@test.com","password":"secret123","role":"client"}'

# Login
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"ana@test.com","password":"secret123"}'

# Crear restaurante (admin)
curl -X POST http://localhost:8080/restaurants \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{"name":"La Piazza","address":"Calle 5","phone":"2222-3333","capacity":60}'

# Crear menú
curl -X POST http://localhost:8080/menus \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "restaurant_id":"<RESTAURANT_ID>",
    "name":"Almuerzo",
    "items":[{"name":"Pasta","price":9.99,"available":true}]
  }'

# Hacer reserva
curl -X POST http://localhost:8080/reservations \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{"restaurant_id":"<ID>","date":"2026-04-01T19:00:00Z","party_size":4}'

# Hacer pedido
curl -X POST http://localhost:8080/orders \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "restaurant_id":"<ID>",
    "items":[{"menu_item_id":"<ITEM_ID>","quantity":2}],
    "pickup":false
  }'
```

---

## Correr Pruebas

```bash
# Unitarias (sin BD)
go test ./tests/... -run TestJWT -v
go test ./tests/... -run TestOrder -v
go test ./tests/... -run TestModel -v

# Integración (requiere BD)
TEST_DB_HOST=localhost TEST_DB_NAME=restaurant_test go test ./tests/... -v -count=1

# Coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

---

## Roles y Permisos

- **client**: puede registrarse, ver restaurantes/menús, hacer reservas y pedidos.
- **admin**: todo lo anterior + crear/editar restaurantes y menús.

Los tokens JWT expiran en **24 horas**.
