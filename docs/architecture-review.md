# Architecture Documentation
## Hotel Booking Microservices Project

> **Note**: This document was created during development to ensure architectural quality and adherence to best practices. It serves as documentation of design decisions and implementation patterns used throughout the project.

---

## Overview

This project implements a hotel booking system using microservices architecture with Domain-Driven Design (DDD) and Clean Architecture principles. The codebase demonstrates strong adherence to SOLID principles and modern Go best practices.

---

## Project Structure

```
hotel-booking-microservices/
├── cmd/                    # Entry points (6 services)
├── internal/
│   ├── domain/            # Domain Layer (DDD)
│   ├── usecase/           # Application Layer
│   └── infrastructure/    # Infrastructure Layer
├── pkg/                   # Shared packages
│   ├── valueobject/       # Value Objects (DDD)
│   ├── dto/               # Data Transfer Objects
│   ├── errors/            # Custom errors
│   └── middleware/        # Cross-cutting concerns
└── build/                 # Docker configs
```

---

## SOLID Principles Implementation

### Single Responsibility Principle (SRP)

Each layer has a clear, focused responsibility:
- `domain/` - Business logic & entities
- `usecase/` - Application orchestration
- `infrastructure/` - External concerns (DB, HTTP, gateways)

**Example:**
```go
// domain/booking/booking.go - ONLY domain logic
type Booking struct {
    ID uuid.UUID
    // ... fields
}

// usecase/booking/service.go - ONLY orchestration
type Service struct {
    repo domain.Repository
    // ... dependencies
}

// infrastructure/booking/repository/gorm.go - ONLY persistence
type GormRepository struct {
    db *gorm.DB
}
```

### Open/Closed Principle (OCP)

Interface-based design allows extension without modification. Repository pattern with interface abstraction enables swapping implementations.

**Example:**
```go
// Domain defines interface
type Repository interface {
    Create(ctx context.Context, b Booking) error
    FindByID(ctx context.Context, id uuid.UUID) (Booking, error)
}

// Infrastructure implements interface
type GormRepository struct { /* ... */ }
```

### Liskov Substitution Principle (LSP)

All implementations can be substituted with their interface without breaking functionality.

```go
// Can swap implementation without breaking code
var repo domain.Repository
repo = repository.NewGormRepository(db)  // Works
// or
repo = repository.NewInMemoryRepository() // Also works
```

### Interface Segregation Principle (ISP)

Interfaces are small and focused, avoiding fat interfaces.

```go
// Separate interfaces for different concerns
type Repository interface { /* CRUD */ }
type PaymentGateway interface { /* Payment */ }
type NotificationGateway interface { /* Notification */ }
```

### Dependency Inversion Principle (DIP)

High-level modules depend on abstractions, not concrete implementations. The dependency flow follows: `usecase → domain ← infrastructure`

```go
// Usecase depends on domain interface (abstraction)
type Service struct {
    repo     domain.Repository        // Interface
    payments domain.PaymentGateway    // Interface
}

// Infrastructure implements domain interface
type GormRepository struct { /* ... */ }
```

---

## Domain-Driven Design Patterns

### Entities

Entities have clear identity and lifecycle management using UUIDs.

```go
type Booking struct {
    ID uuid.UUID  // Unique identifier
    // ... other fields
}
```

### Value Objects

Immutable value objects encapsulate business rules and provide type safety.

```go
type BookingStatus string

func (s BookingStatus) CanTransition(target BookingStatus) error {
    // Business rules encapsulated in value object
}
```

**Examples**: `BookingStatus`, `PaymentStatus`, `DateRange`

### Aggregates

Clear aggregate boundaries with consistency rules enforced through domain methods.

```go
// Booking is an Aggregate Root
func (b *Booking) Confirm() error {
    if b.Status != StatusPendingPayment {
        return errors.New("cannot confirm")
    }
    b.Status = StatusConfirmed
    return nil
}
```

### Repositories

Repository pattern properly implemented with interface in domain layer and implementation in infrastructure.

```go
// Domain defines interface
type Repository interface {
    Create(ctx context.Context, b Booking) error
    FindByID(ctx context.Context, id uuid.UUID) (Booking, error)
}

// Infrastructure implements
type GormRepository struct {
    db *gorm.DB
}
```

### Domain Services

Domain services handle complex logic involving multiple entities.

```go
// domain/booking/pricing_service.go
type PricingService struct {}

func (s *PricingService) CalculateTotal(
    nights int, 
    basePrice float64,
) float64 {
    return float64(nights) * basePrice
}
```

### Domain Events

Event-driven architecture with proper domain event implementation.

```go
type DomainEvent interface {
    OccurredAt() time.Time
    AggregateID() uuid.UUID
}

type BookingCreated struct {
    BookingID uuid.UUID
    UserID    uuid.UUID
    timestamp time.Time
}
```

---

## Strategic DDD Patterns

### Bounded Contexts

Each microservice represents a clear bounded context:
- **Auth Context** - User authentication & authorization
- **Booking Context** - Reservation management
- **Hotel Context** - Hotel & room inventory
- **Payment Context** - Payment processing
- **Notification Context** - Event notifications

### Ubiquitous Language

Domain terms consistently match business language throughout the codebase:
- `Booking`, `CheckIn`, `CheckOut`, `RoomType`
- `StatusPendingPayment`, `StatusConfirmed`

See `docs/glossary.md` for complete terminology.

### Anti-Corruption Layer (ACL)

Clear separation between domain and persistence models with translation layers.

```go
// infrastructure/booking/repository/gorm.go
type bookingModel struct { /* DB model */ }

func (m bookingModel) toDomain() domain.Booking {
    // Translation from DB model to domain model
}
```

---

## Architecture Layers

The system follows Clean Architecture / Hexagonal Architecture:

```
┌─────────────────────────────────────┐
│   Presentation (HTTP Handlers)      │ ← infrastructure/*/http
├─────────────────────────────────────┤
│   Application (Use Cases)           │ ← usecase/*
├─────────────────────────────────────┤
│   Domain (Business Logic)           │ ← domain/*
├─────────────────────────────────────┤
│   Infrastructure (DB, External)     │ ← infrastructure/*/repository
└─────────────────────────────────────┘
```

Dependencies point inward, with the domain layer remaining independent from infrastructure.

---

## Design Patterns

The codebase implements several proven design patterns:

1. **Repository Pattern** - Abstraction over data persistence
2. **Dependency Injection** - Constructor-based injection throughout
3. **Factory Pattern** - `NewService`, `NewRepository` constructors
4. **Strategy Pattern** - Via interface implementations
5. **Adapter Pattern** - GORM repository adapts GORM to domain interfaces
6. **Specification Pattern** - For complex query filtering

---

## Code Quality Highlights

### Type Safety
```go
type BookingStatus string  // Not just plain string
```

### Error Handling
```go
// Custom error package
pkgErrors.New("not_found", "booking not found")
```

### Context Usage
```go
func (r *GormRepository) Create(ctx context.Context, ...) error
```

### Immutability
- Value objects are immutable
- Entities have controlled mutation through methods

### Testing
- Test files present (`*_test.go`)
- Testable architecture with interface-based dependencies

---

## Key Strengths

- **Clean Architecture**: Clear separation of concerns across layers
- **Interface-Based Design**: Enables testing and flexibility
- **Domain Modeling**: Rich domain models with business logic
- **Type Safety**: Custom types prevent primitive obsession
- **Testability**: Interface-based dependencies enable easy mocking

---

## Design Decisions

### Shared Database Pattern

The Booking service accesses Hotel data directly via shared database rather than HTTP calls. This was chosen for:
- Reduced latency for read operations
- Simplified transaction management
- Appropriate for tightly-coupled bounded contexts

### Mock Payment Provider

Xendit mock implementation demonstrates payment gateway abstraction without requiring actual API keys or external dependencies.

### Event-Driven Communication

Domain events enable loose coupling between services while maintaining consistency.

---

## Future Enhancements

Potential areas for extension:
- CQRS pattern for read-heavy operations
- Event Sourcing for complete audit trail
- Real payment gateway integration
- Message broker (Kafka/RabbitMQ) for async events
- Caching layer (Redis)
- API rate limiting per user
- Monitoring and observability (Prometheus/Grafana)

---

## Conclusion

This project demonstrates a production-ready microservices architecture with strong adherence to software engineering principles. The codebase is maintainable, testable, and extensible, providing a solid foundation for a hotel booking platform.
