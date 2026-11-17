# Implementation Plan

Implement incrementally. Inspect the repository first and adapt this plan to existing code rather than duplicating completed work.

## Phase 1 — Foundation
- Go module and application entry point.
- Configuration and environment validation.
- PostgreSQL connection and GORM setup.
- Migrations.
- Fiber server, routes, health endpoint.
- Logger, recovery, request ID, common response/error handling.

## Phase 2 — Authentication and users
- User entity/repository/use cases.
- Password hashing.
- Login/register/me.
- JWT access token and refresh-token rotation/revocation.
- RBAC middleware.
- No self-registration; Admin-only account management (D-026).
- Demo seed accounts.

## Phase 3 — Master data and property foundation
- Property types and location master data.
- Owners and private owner information.
- Property CRUD.
- Property images and local storage abstraction.
- Structured specifications and flexible description.
- Public/general versus restricted address fields.
- Primary Marketing ownership set from authenticated actor.

## Phase 4 — Property authorization and co-listing
- `property_marketing_assignments`.
- Primary owner and co-listing DTO projections.
- Admin co-listing add/remove/replace.
- Maximum three total assignments.
- Transactional validation and audit logs.
- Owner PII and exact address privacy tests.

## Phase 5 — Public discovery (guests)
- Property listing/detail/search.
- Price/location/LT/LB filters.
- Sold/Rented badge data while excluding those statuses from customer availability filters.
- WhatsApp contact data for the appropriate marketing user.
- Remove email contact action from the relevant contract.

## Phase 6 — Leads and Marketing workflow
- Manual lead creation by Marketing/Admin from WhatsApp conversations (D-026).
- Admin assignment/reassignment.
- Marketing assigned-lead views.
- Follow-up activities.
- Survey scheduling and status.
- Lost reason validation.
- Ownership checks for every lead operation.

## Phase 7 — Deals and commissions
- Deal won/lost flow.
- Transactional finalization.
- Property status update.
- Commission calculation/record creation.
- Rollback tests.

## Phase 8 — Units, tenants, notifications
- Units CRUD.
- Tenant records and rental period management.
- DB-backed in-app notifications.
- Read/unread operations.

## Phase 9 — Admin, reports, audit
- User and role management.
- Staff management and predefined permissions.
- Reports and filters.
- Master data CRUD where required.
- Audit log views.
- Ensure Admin's listing and owner-data restrictions remain intact.

## Phase 10 — Hardening and delivery
- Swagger/OpenAPI completeness.
- Unit and integration tests.
- Security and IDOR review.
- Pagination/index review.
- README setup instructions without Docker.
- `gofmt`, `go vet`, and `go test ./...`.
- Verify migrations on a clean local PostgreSQL database.

## Implementation discipline

Do not generate all phases in one unreviewed code dump. Finish and verify each phase before proceeding. Keep the API contract aligned with Flutter screens.

## Status

All ten phases are implemented. Verification: `gofmt`, `go vet ./...`, `go test ./...` (integration tests with `TEST_DATABASE_URL`), clean-database migration up/down, OpenAPI routes cross-checked against `internal/routes/routes.go`.
Not implemented (explicitly out of scope until decided): exceptional audited Admin access to owner PII.

## D-026 / D-027 / D-028 (simplified internal system) — implemented
- Roles are `sales` and `admin` only; `POST /auth/register`, favorites, inquiries, leads, follow-ups, surveys, units, tenants, notifications and reports are removed from code, routes, OpenAPI, Postman, seeds and tests.
- Migration `000023_internal_scope_deals_cancel`: deals hang on the property (`recorded_by`, `outcome`, `buyer_name`, one outcome per property); out-of-scope tables are dropped; `customer` and `staff` users are deleted; `users.role` is limited to `sales`/`admin`. The down migration recreates the dropped tables empty.
- Migration `000024_property_details`: listing fields from the reference page.
- Admin endpoints `POST /properties/{id}/deal` and `POST /properties/{id}/cancel` (cancel archives the property; archived listings are excluded from every list).
- Property types: `house`, `apartment`, `shop_house` (ruko), `factory`.
- Applying migration 000023 to an existing database is destructive (drops tables, deletes customer/staff accounts): back up first.
