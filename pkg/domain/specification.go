package domain

// Specification interface for generic criteria.
type Specification[T any] interface {
	IsSatisfiedBy(entity T) bool
}

// AndSpecification combines two specs with AND logic.
type AndSpecification[T any] struct {
	Left  Specification[T]
	Right Specification[T]
}

func (s AndSpecification[T]) IsSatisfiedBy(entity T) bool {
	return s.Left.IsSatisfiedBy(entity) && s.Right.IsSatisfiedBy(entity)
}

// And creates a new AND specification.
func And[T any](left, right Specification[T]) Specification[T] {
	return AndSpecification[T]{Left: left, Right: right}
}

// OrSpecification combines two specs with OR logic.
type OrSpecification[T any] struct {
	Left  Specification[T]
	Right Specification[T]
}

func (s OrSpecification[T]) IsSatisfiedBy(entity T) bool {
	return s.Left.IsSatisfiedBy(entity) || s.Right.IsSatisfiedBy(entity)
}

// Or creates a new OR specification.
func Or[T any](left, right Specification[T]) Specification[T] {
	return OrSpecification[T]{Left: left, Right: right}
}

// NotSpecification negates a spec.
type NotSpecification[T any] struct {
	Spec Specification[T]
}

func (s NotSpecification[T]) IsSatisfiedBy(entity T) bool {
	return !s.Spec.IsSatisfiedBy(entity)
}

// Not creates a new NOT specification.
func Not[T any](spec Specification[T]) Specification[T] {
	return NotSpecification[T]{Spec: spec}
}
