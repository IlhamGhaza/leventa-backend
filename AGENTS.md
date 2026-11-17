# AI Agent Instructions

## Project identity

This repository contains the backend for a role-based Property Management System. It is an internal system. It supports public property discovery for guests (contact via WhatsApp only), marketing listing management, and admin management of co-listing, ownership, deals and commission.

Read `README.md` and the relevant `docs/*.md` before coding.

## Non-negotiable technical decisions

- Go + Fiber.
- PostgreSQL + GORM.
- JWT access and refresh tokens.
- `golang-migrate` for schema migrations.
- Swagger/OpenAPI documentation.
- Modular monolith.
- Local filesystem storage behind an interface for the initial implementation.
- No Redis, Docker, Kubernetes, microservices, Kafka, RabbitMQ, Elasticsearch, GraphQL, CQRS, or Event Sourcing.

## Domain vocabulary

- **Marketing** is the product-facing term for the Sales role.
- **Primary Marketing** is the marketing user who created and owns a property listing.
- **Co-listing Marketing** is an additional marketing user assigned to help market a listing. Co-listing does not transfer ownership or grant edit rights.
- **Guest** (customer) has no account; browses public listings and contacts Marketing via WhatsApp. See D-026.
- **Admin** manages the system and co-listing assignments, but does not edit property listing details or owner contact information under the agreed policy.

## Mandatory authorization principles

- Determine the actor and role from the verified authentication context, never from client-supplied role/user identifiers.
- Enforce role permission and record ownership in use cases/repositories, not only in route middleware.
- Primary Marketing can edit only listings where they are the primary owner.
- Co-listing Marketing is read-only for allowed listing fields.
- A property has one primary marketing owner and no more than two co-listing users (three marketing users total).
- Admin can add/remove/replace co-listing members but cannot edit property details or owner data through this capability.
- Owner contact details and exact address are private fields. Return only fields allowed for the actor.
- Marketing can access only permitted listing records.
- Record access must be tested for cross-user IDOR attempts.

## Implementation workflow

1. Inspect existing code and report relevant findings.
2. Read the applicable requirements and decision documents.
3. State a short implementation plan for multi-file or high-impact tasks.
4. Implement the smallest complete vertical slice.
5. Add/update migrations and seed data when schema or demo data changes.
6. Add unit/integration tests for business rules and authorization.
7. Update Swagger/OpenAPI and README when API behavior changes.
8. Run formatting, vetting, and relevant tests.
9. Summarize changes and any unresolved decisions.

## Code quality

- Idiomatic Go; use `context.Context` across I/O boundaries.
- Small, focused interfaces.
- Explicit dependency injection.
- No unnecessary abstractions, global mutable state, or hidden side effects.
- Parameterized database queries.
- Consistent error mapping and response envelopes.
- Use transactions for multi-step business operations that must succeed or fail together.
- Do not silently swallow errors.
- Avoid breaking API changes unless explicitly requested.

## Security

- Hash passwords with bcrypt or Argon2id.
- Store refresh-token hashes, not raw refresh tokens.
- Rotate and revoke refresh tokens.
- Validate uploads by size, MIME type, and safe generated filenames.
- Prevent path traversal.
- Never log passwords, tokens, owner phone numbers, or other secrets.
- Apply rate limiting to authentication endpoints.
- Use secure CORS configuration and security headers.
- Audit sensitive administrative and ownership actions.

## Completion report

Include:
- What changed.
- Important design decisions.
- Migrations added.
- Tests run and results.
- Known limitations or decisions requiring user input.
