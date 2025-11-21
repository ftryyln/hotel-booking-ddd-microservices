# Project Structure Analysis
## Hotel Booking Microservices

Dokumen ini menjelaskan struktur file dan folder dalam project ini, serta tujuan dari setiap komponen utama. Struktur project mengikuti standar **Golang Standard Project Layout** yang disesuaikan dengan **Domain-Driven Design (DDD)**.

---

## 📂 Root Directory

| File/Folder | Deskripsi |
|-------------|-----------|
| `cmd/` | Entry points untuk setiap microservice. Setiap subfolder berisi `main.go`. |
| `internal/` | Private application code. Kode di sini tidak bisa di-import oleh project lain. |
| `pkg/` | Public library code. Kode yang bisa dishare antar service (shared kernel). |
| `build/` | Dockerfile untuk setiap service. |
| `config/` | File konfigurasi (misal: routes untuk API Gateway). |
| `docs/` | Dokumentasi project (Swagger, Analysis, Glossary). |
| `migrations/` | SQL scripts untuk database schema dan seeding. |
| `go.mod` | Definisi module dan dependencies. |
| `docker-compose.yml` | Definisi orchestrasi container untuk menjalankan semua service. |
| `Makefile` | Shortcut commands untuk build, test, dan run. |

---

## 🏗️ `cmd/` - Application Entry Points

Setiap folder di sini adalah satu microservice yang bisa dicompile menjadi binary terpisah.

- `cmd/auth-service/` - Service Autentikasi (JWT, Login, Register)
- `cmd/booking-service/` - Service Pemesanan (Core business logic)
- `cmd/hotel-service/` - Service Inventory Hotel (Kamar, Tipe Kamar)
- `cmd/payment-service/` - Service Pembayaran (Integrasi Payment Gateway)
- `cmd/notification-service/` - Service Notifikasi (Email/Log dispatcher)
- `cmd/api-gateway/` - API Gateway (Reverse proxy, Rate limiting)

---

## 🧠 `internal/` - Core Application Logic

Struktur di dalam `internal/` mengikuti pola **DDD Layers**:

### 1. `internal/domain/` (Domain Layer)
Berisi **Enterprise Business Rules**. Layer ini murni Go struct dan interface, **tanpa dependencies** ke database atau HTTP.

- **Entities**: Struct dengan identity (misal: `Booking`, `Hotel`).
- **Value Objects**: Struct immutable tanpa identity (misal: `Money`, `DateRange`).
- **Repository Interfaces**: Contract untuk persistence (misal: `BookingRepository`).
- **Domain Events**: Definisi event (misal: `BookingCreated`).
- **Domain Services**: Logic kompleks yang melibatkan multiple entities (misal: `PricingService`).

### 2. `internal/usecase/` (Application Layer)
Berisi **Application Business Rules**. Layer ini mengorkestrasi domain objects untuk menjalankan use case tertentu.

- **Service**: Implementasi use case (misal: `CreateBooking`, `ConfirmPayment`).
- **Assembler**: Mapping dari DTO ke Domain Model dan sebaliknya.

### 3. `internal/infrastructure/` (Infrastructure Layer)
Berisi implementasi teknis detail.

- **Repository Implementation**: Code yang akses database (GORM implementation).
- **HTTP Handlers**: Code yang handle request/response HTTP.
- **Gateways**: Client untuk external services (misal: Payment Gateway Client).

---

## 📦 `pkg/` - Shared Libraries

Kode yang digunakan oleh banyak service.

- `pkg/config/` - Helper untuk load env vars.
- `pkg/domain/` - Shared domain interfaces (misal: `DomainEvent`, `Specification`).
- `pkg/dto/` - Data Transfer Objects (struct untuk JSON request/response).
- `pkg/errors/` - Standardized error types.
- `pkg/logger/` - Structured logging setup (Zap).
- `pkg/middleware/` - HTTP middlewares (Auth, Logging, Recovery).
- `pkg/valueobject/` - Shared value objects (Money, DateRange).

---

## 📜 Key Files Breakdown

### Booking Service (`internal/domain/booking/`)
- `booking.go`: Aggregate Root `Booking`. Berisi logic state changes (`Confirm`, `Cancel`).
- `events.go`: Definisi domain events (`BookingCreated`, `BookingConfirmed`).
- `pricing_service.go`: Domain service untuk kalkulasi harga kompleks.
- `specifications.go`: Specification pattern untuk query filtering.

### Booking Infrastructure (`internal/infrastructure/booking/`)
- `repository/gorm.go`: Implementasi repository pakai GORM.
- `repository/factory.go`: Factory pattern untuk membuat repository.
- `http/handler.go`: HTTP endpoints handler.

---

## 🗃️ Database Migrations (`migrations/`)

- `001_init.sql`: Skema database awal (tabel users, hotels, bookings, dll).
- `002_seed_data.sql`: Data dummy untuk testing (Admin user, Hotel contoh).

---

## 📝 Documentation (`docs/`)

- `solid-ddd-analysis.md`: Analisis mendalam tentang penerapan SOLID dan DDD.
- `glossary.md`: Kamus istilah domain (Ubiquitous Language).
- `swagger/`: Generated API documentation.

---

## 💡 Summary

Struktur ini didesain untuk:
1.  **Separation of Concerns**: Memisahkan logic bisnis dari detail teknis.
2.  **Testability**: Mudah ditest karena dependencies menggunakan interface.
3.  **Scalability**: Mudah menambah fitur tanpa merusak fitur lain.
4.  **Maintainability**: Kode terorganisir dengan jelas, mudah dinavigasi.
