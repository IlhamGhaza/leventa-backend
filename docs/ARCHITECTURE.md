# Backend Architecture

## 1. Architecture style

Use a pragmatic modular monolith with clear boundaries. Do not introduce microservices or distributed infrastructure.

Dependency direction:

`HTTP Handler → Use Case → Repository Interface → PostgreSQL Repository`

Supporting services (JWT, password hashing, storage, notification, audit) are injected into use cases as needed.

## 2. Suggested project structure

```text
cmd/
  server/
    main.go
internal/
  config/
  domain/
    entity/
    repository/
    valueobject/
  usecase/
    auth/
    user/
    property/
    owner/
    deal/
    commission/
    masterdata/
    audit/
    colisting/
  delivery/
    http/
      handler/
      middleware/
      request/
      response/
  repository/
    postgres/
  service/
    auth/
    notification/
    audit/
  storage/
  routes/
pkg/
  logger/
  response/
  validator/
database/
  migrations/
  seeds/
docs/
```

Adapt names to the existing repository if a sound structure already exists. Do not move files merely to match this example.

### Implemented layout (differs slightly from the suggestion above)

- `internal/usecase/` is a single package (not one package per module); each use case file is named after its module.
- Authorization lives in `internal/usecase` (`Actor`, `propertyAccess`); handlers only build the `Actor` from the verified token.
- `repository.Transactor` runs multi-step operations in one transaction; repositories pick the transaction up from `ctx`.
- Response projections (`PublicProperty`, `InternalProperty`, `PrimaryProperty`, `OwnerView`) live in `internal/usecase/property_dto.go`.
- OpenAPI spec: `docs/openapi.yaml`, embedded by `docs/embed.go` and served at `/openapi.yaml` and `/swagger`.

## 3. Responsibilities

### Handler
- Parse HTTP input.
- Validate request shape.
- Read authenticated actor from context.
- Call use cases.
- Map results/errors to HTTP responses.
- Must not contain business rules or raw SQL.

### Use case
- Own business rules and authorization decisions.
- Validate ownership and state transitions.
- Coordinate repositories and services.
- Use transactions for atomic workflows.

### Repository
- Encapsulate database access.
- Support context cancellation.
- Provide scoped queries where needed.
- Never return private fields to a response merely because they are present in a model.

### Domain
- Entities and domain-level types.
- Avoid HTTP/Fiber dependencies in domain logic.

## 4. Cross-cutting services

- Authentication/token service.
- Password hasher.
- File storage interface.
- Audit logger.
- Clock abstraction only if needed for deterministic tests.

## 5. Concurrency and background work

Use goroutines only for justified, bounded tasks such as non-critical notification delivery or thumbnail processing. Do not create a complex worker queue. Ensure cancellation, error handling, and lifecycle management.

## 6. Storage

Start with local filesystem storage behind an interface:
- `Upload`
- `Delete`
- `GetURL`

Generate safe filenames. Validate MIME type and size. Keep stored files outside executable/source directories where appropriate. Design the interface so cloud storage can be added later without forcing it now.

## 7. Database and transactions

- PostgreSQL.
- UUID primary keys.
- `created_at`, `updated_at` timestamps.
- Soft delete only where it makes business sense.
- Use database constraints for invariants where possible.
- Use transactions for deal finalization, co-listing assignment changes, and other multi-record operations.

## 8. Avoid overengineering

Do not add Redis, Docker, Kubernetes, Kafka, RabbitMQ, Elasticsearch, GraphQL, CQRS, Event Sourcing, or generic plugin systems. Prefer straightforward Go code and PostgreSQL capabilities.
