// Package app contains use cases: orchestration, transaction boundaries, and
// coordination of domain logic via ports (interfaces). It depends on domain
// types and port interfaces; concrete adapters are injected from outside.
package app

// Application is the composition root for application-layer services. Add
// fields and constructors when you introduce use cases and ports.
type Application struct{}
