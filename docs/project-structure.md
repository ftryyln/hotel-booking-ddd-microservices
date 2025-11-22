# Project Structure Analysis
## Hotel Booking Microservices

This document explains the file and folder structure of this project, as well as the purpose of each major component. The project structure follows the **Golang Standard Project Layout** adapted for **Domain-Driven Design (DDD)**.

---

## 📂 Root Directory

| File/Folder | Description |
|-------------|-------------|
| `cmd/` | Entry points for each microservice. Each subfolder contains a `main.go`. |
| `internal/` | Private application code. Code here cannot be imported by other projects. |
| `pkg/` | Public library code. Code shared between services (shared kernel). |
| `build/` | Dockerfiles for each service. |
| `config/` | Configuration files (e.g., routes for API Gateway). |
| `docs/` | Project documentation (Swagger, Analysis, Glossary). |
| `migrations/` | SQL scripts for database schema and seeding. |
| `go.mod` | Module definition and dependencies. |
| `docker-compose.yml` | Container orchestration definition to run all services. |
| `Makefile` | Shortcut commands for build, test, and run. |

---

## 🏗️ `cmd/` - Application Entry Points

Each folder here represents a single microservice that can be compiled into a separate binary.

- `cmd/auth-service/` - Authentication Service (JWT, Login, Register)
- `cmd/booking-service/` - Booking Service (Core business logic)
- `cmd/hotel-service/` - Hotel Inventory Service (Hotels, Room Types)
- `cmd/payment-service/` - Payment Service (Payment Gateway Integration)
- `cmd/notification-service/` - Notification Service (Email/Log dispatcher)
- `cmd/api-gateway/` - API Gateway (Reverse proxy, Rate limiting)

---

## 🧠 `internal/` - Core Application Logic

The structure within `internal/` follows **DDD Layers**:

### 1. `internal/domain/` (Domain Layer)
Contains **Enterprise Business Rules**. This layer consists purely of Go structs and interfaces, **without dependencies** on the database or HTTP.

- **Entities**: Structs with identity (e.g., `Booking`, `Hotel`).
- **Value Objects**: Immutable structs without identity (e.g., `Money`, `DateRange`).
- **Repository Interfaces**: Contracts for persistence (e.g., `BookingRepository`).
- **Domain Events**: Event definitions (e.g., `BookingCreated`).
- **Domain Services**: Complex logic involving multiple entities (e.g., `PricingService`).

### 2. `internal/usecase/` (Application Layer)
Contains **Application Business Rules**. This layer orchestrates domain objects to execute specific use cases.

- **Service**: Use case implementation (e.g., `CreateBooking`, `ConfirmPayment`).
- **Assembler**: Mapping from DTO to Domain Model and vice versa.

### 3. `internal/infrastructure/` (Infrastructure Layer)
Contains detailed technical implementations.

- **Repository Implementation**: Code accessing the database (GORM implementation).
- **HTTP Handlers**: Code handling HTTP requests/responses.
- **Gateways**: Clients for external services (e.g., Payment Gateway Client).

---

## 📦 `pkg/` - Shared Libraries

Code used by multiple services.

- `pkg/config/` - Helpers for loading env vars.
- `pkg/domain/` - Shared domain interfaces (e.g., `DomainEvent`, `Specification`).
- `pkg/dto/` - Data Transfer Objects (structs for JSON request/response).
- `pkg/errors/` - Standardized error types.
- `pkg/logger/` - Structured logging setup (Zap).
- `pkg/middleware/` - HTTP middlewares (Auth, Logging, Recovery).
- `pkg/valueobject/` - Shared value objects (Money, DateRange).

---

## 📜 Key Files Breakdown

### Booking Service (`internal/domain/booking/`)
- `booking.go`: Aggregate Root `Booking`. Contains state change logic (`Confirm`, `Cancel`).
- `events.go`: Domain event definitions (`BookingCreated`, `BookingConfirmed`).
- `pricing_service.go`: Domain service for complex price calculations.
- `specifications.go`: Specification pattern for query filtering.

### Booking Infrastructure (`internal/infrastructure/booking/`)
- `repository/gorm.go`: Repository implementation using GORM.
- `repository/factory.go`: Factory pattern for creating repositories.
- `http/handler.go`: HTTP endpoints handler.
- `worker/scheduler.go`: Background CronJob for auto-checkout.

---

## 🗃️ Database Migrations (`migrations/`)

- `001_init.sql`: Initial database schema (users, hotels, bookings tables, etc.).
- `002_seed_data.sql`: Dummy data for testing (Admin user, sample Hotel).

---

## 📝 Documentation (`docs/`)

- `solid-ddd-analysis.md`: In-depth analysis of SOLID and DDD implementation.
- `glossary.md`: Dictionary of domain terms (Ubiquitous Language).
- `swagger/`: Generated API documentation.

---

## 💡 Summary

This structure is designed for:
1.  **Separation of Concerns**: Separating business logic from technical details.
2.  **Testability**: Easy to test because dependencies use interfaces.
3.  **Scalability**: Easy to add features without breaking others.
4.  **Maintainability**: Code is clearly organized and easy to navigate.
