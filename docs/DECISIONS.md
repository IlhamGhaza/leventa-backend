# Architecture and Product Decisions

This document records decisions already agreed. Do not silently reverse them.

## D-001 — Backend stack
Go + Fiber + PostgreSQL + GORM. Use `golang-migrate`, JWT, validation, and Swagger/OpenAPI.

## D-002 — Deployment/development simplicity
No Redis and no Docker. No Kubernetes or distributed infrastructure. Use local PostgreSQL and local filesystem storage initially.

## D-003 — Architecture
Modular monolith with pragmatic layered/Clean Architecture boundaries.

## D-004 — Authentication
One universal login. Login accepts email/password only. Role comes from the backend account record. There is no self-registration; Admin creates all accounts (D-026).

## D-005 — Roles
Persisted role values: `sales`, `admin` (`customer` removed by D-026, `staff` removed by D-027). Product UI may call the `sales` role “Marketing”.

## D-006 — Not an e-commerce marketplace
Customers are guests who browse the public catalog and contact Marketing via WhatsApp. No checkout, online payment, reservation, or customer selection of a Marketing user.

## D-007 — Property primary ownership
The Marketing user who creates a listing becomes its primary owner. Only that primary owner can edit listing content and owner information.

## D-008 — Co-listing
A property can have up to three Marketing users total: one primary and up to two co-listing users. Only Admin manages co-listing membership. Co-listing grants read-only access to permitted listing data, not edit rights or owner-data access.

## D-009 — Admin listing restriction
Admin can add/remove/replace co-listing users but cannot edit property listing details or owner information through this feature. Any future exceptional administrative override requires a separately approved, explicit, audited policy.

## D-010 — Owner privacy
Owner identity and contact details are visible only to the primary Marketing owner of the listing. Co-listing, other Marketing, guests, and ordinary Admin workflows cannot see owner PII.

## D-011 — Exact address privacy
Block, cluster/internal detail, street/house number, floor and room identifiers are private. Only primary Marketing can access the complete address under the current policy. General location can be shown according to listing visibility.

## D-012 — WhatsApp contact
Remove email contact action from the property contact experience. WhatsApp contact uses the representing Marketing user's own WhatsApp number. Never use the owner's number.

## D-013 — Property status presentation
Sold/Rented listings can remain visible with a small badge/banner, but must not be public availability filter options or shown as available.

## D-014 — (obsolete, D-027) Lead assignment vs listing ownership
A Marketing user assigned to a lead is not automatically the primary owner of the property listing. `leads.sales_id` and `properties.primary_marketing_id` represent different relationships.

## D-015 — Historical records
Do not hard-delete properties involved in deals. Archive/deactivate and preserve relationships.

## D-016 — Commission policy
Commission uses a server-configured default percentage (`COMMISSION_DEFAULT_PERCENTAGE`, default 2.5), rounded to 2 decimals. Marketing cannot supply a percentage. Admin may override the percentage and set status (`pending` → `paid`/`cancelled`) only while the commission is `pending`.

## D-017 — (obsolete, D-027) Staff scope
Staff is operational only: units, tenants, property operations (without owner data or exact address) and the property dashboard. Staff has no access to leads, deals, commissions, sales reports, users, or co-listing. Per-staff permission tables are not implemented.

## D-018 — (superseded by D-027) Deal finalization
Superseded: only Admin records Deal or Cancelled (D-027). Co-listing membership alone does not grant this.

## D-019 — Fiber v3
The HTTP layer uses Fiber v3. Fiber v3 treats an empty CORS origin list as "allow all"; the application therefore denies cross-origin access explicitly when `CORS_ALLOWED_ORIGINS` is empty.

## D-020 — Owner per listing
Owner records are created with the listing (nested `owner` object) and reachable only through `/properties/{id}/owner` by the primary marketing user. There is no owner list endpoint and no reuse of owners across listings (resolves the open question below).

## D-021 — Co-listing view and status (supersedes the co-listing part of D-008 and D-011)
Co-listing marketing receives the same property view as the primary marketing (including exact address and coordinates) but never owner data, and cannot edit listing content or images. A co-listing marketing user may toggle the status between `available`, `reserved` and `maintenance` (not `draft`/`archived`, and not `sold`/`rented`, which come only from deal finalization). The exact address is therefore visible to the primary and co-listing marketing of that listing only.

## D-022 — Primary ownership transfer
Admin only, audited (`transferred_primary_ownership`). The previous primary becomes a co-listing member (needs a free co-listing slot if the new primary was not already in the team); owner data follows the listing. Marketing users request a transfer through the admin; there is no self-service transfer.

## D-023 — Public WhatsApp contact
The representing marketing user's WhatsApp contact is returned to everyone, including unauthenticated callers (`contact` in the property projection). The owner's number is never returned.

## D-024 — Historical listings
Default public lists contain only available/reserved listings. `history=true` returns only sold/rented; `include_history=true` returns both, and every item carries `is_available` and, when applicable, `status_badge`.

## D-025 — Authentication hardening
- Password policy: 8–72 bytes, at least one letter and one digit, a few very common passwords rejected (bcrypt ignores bytes beyond 72, so longer input is refused instead of truncated).
- Emails are case-insensitive (stored lower-case).
- Per-account lockout (no Redis, state in PostgreSQL): `MAX_LOGIN_ATTEMPTS` (5) failures lock the account for `LOCKOUT_MINUTES` (15); locked login returns 429 even with the right password; success resets the counter. Unknown emails take the same time as wrong passwords.
- Access tokens are short-lived (default 900 s). The middleware re-reads role and status from the database on every request, so role changes, deactivation and deletion take effect immediately; they also revoke all refresh tokens.
- At most `MAX_ACTIVE_SESSIONS` (5) active refresh tokens per user; reuse of a rotated token revokes all sessions.
- `go run ./cmd/server -seed` refuses to run when `APP_ENV=production`.
- Behind a reverse proxy set `TRUSTED_PROXIES`, otherwise per-IP rate limiting sees only the proxy. HSTS is sent when `APP_ENV=production`; TLS itself is terminated by the proxy.

## D-026 — Internal system, guests instead of customers (supersedes D-004/D-006 customer parts)
The system is internal. There is no `customer` role, no registration, no favorites, no inquiries and no customer notifications. Customers are anonymous guests who view the public catalog (owner data and exact address hidden; the owner's phone is replaced by the representing Marketing contact) and talk to Marketing on WhatsApp. No chat inside the system.

## D-027 — Simplified scope: listings, co-listing, deals, commission (supersedes D-017, D-018, lead parts of D-014)
Flow: the property owner calls a Marketing user or the office; that Marketing user creates the listing and is its primary owner. Marketing tells Admin about a deal outside the system. Admin only: manages co-listing, changes the primary Marketing, records Deal or Cancelled, manages commission.
Removed from scope: `staff` role, leads, lead follow-ups, surveys, units, tenants, reports dashboard, favorites, inquiries, in-app notifications. Removed from code, migrations (000023), OpenAPI and Postman.
**Status: confirmed by the product owner.** Answers: Admin has two records, Deal and Cancelled. Cancelled means the owner no longer sells, so the property becomes `archived` and is not shown in lists. A Deal records the buyer/tenant name. Staff is removed. Commission uses the default (D-016) for now.

## D-028 — Property description fields
Listing fields follow the reference page (see PRODUCT_REQUIREMENTS §7 and `docs/syahira_prop_desc.png`). The reference page shows the owner's phone; this system never does (D-010, D-023). The reference inquiry form is not implemented. Map pins are approximate.

## D-029 — Ownership documents per property type
Each `cert_*` flag on a property says the document exists. When a flag is `true`, the primary marketing can upload that document (PDF, JPEG or PNG, one per type; re-upload replaces it). Document types per property type: house (shm, hgb, imb, pbb, girik, other), apartment (shmsrs, hgb, imb, pbb, slf, other), shop_house (shm, hgb, imb, pbb, slf, other), factory (shm, hgb, hgu, imb, pbb, slf, other). A flag cannot be true for a type that does not apply, and cannot be cleared while its document is stored.
Files are private: stored under `DOCUMENT_PATH` (not a static route), never in public or co-listing projections as URLs. Primary and co-listing marketing may list and download them ("send" to the customer happens outside the system, e.g. WhatsApp); only the primary uploads or deletes. Admin and guests have no access. Uploads, downloads and deletions are audited.
**Status: confirmed (only people with access: primary write, primary + co-listing read, Admin none).**

## D-030 — Logging and troubleshooting
- Every request is logged as one JSON/console line: method, path (no query string), status, latency, IP, `request_id`, and `user_id`/`role` when authenticated. The status reflects errors returned by handlers.
- Every response carries `X-Request-ID`. A 5xx is logged a second time with the real cause (`error`), so support can search the log by the id a user reports.
- Clients only ever see "Internal server error"; causes, stack traces and SQL errors stay in the log. Panics are recovered and logged with a stack trace.
- `LOG_FILE` (optional) also appends JSON lines to a file; rotation is left to the operating system. Passwords, tokens, owner phone numbers and request bodies are never logged.
- Business actions are additionally stored in the `audit_logs` table (Admin: `GET /audit-logs`): listing changes, ownership transfer, co-listing changes, deal/cancel, commission changes, document upload/download/delete, user changes.

## D-031 — Abuse protection for anonymous access
Guests use the API without an account, so the API protects itself (in-memory, per process, no Redis):
- Every `/api/v1` request: `API_RATE_LIMIT_MAX` (300) per `API_RATE_LIMIT_WINDOW` (60 s) per IP, sliding window.
- Login and refresh: `RATE_LIMIT_MAX` (5) per `RATE_LIMIT_WINDOW` (60 s) per IP, plus per-account lockout after `MAX_LOGIN_ATTEMPTS` (D-025).
- Image and document uploads: `UPLOAD_RATE_LIMIT_MAX` (30) per `UPLOAD_RATE_LIMIT_WINDOW` (60 s) per IP, plus `UPLOAD_MAX_SIZE` and the request body limit.
- Property search: `search` at most 100 characters, at most 20 `facility_ids`, `limit` at most 100, and `(page-1)*limit` at most 10 000 (deeper pages must narrow the filters). Violations return 422.
- A blocked request returns 429 with `Retry-After` and is logged with the limiter name, IP and path.
- Behind a reverse proxy `TRUSTED_PROXIES` must be set, otherwise every client shares the proxy IP. Many users behind one office or mobile-carrier IP share one budget; raise the limits if that happens.
- Accounts are created only by Admin (`POST /users`); there is no self-registration endpoint to abuse.

## D-032 — Mobile scope
All roles use the mobile app: guests (default, no login), Marketing and Admin. The app opens as guest; staff log in from inside the app. Mobile screens of removed flows (customer registration, favorites, inquiries, customer notifications, leads, follow-ups, surveys) are removed.

## D-033 — Self-service WhatsApp number
Every logged-in user may change their own WhatsApp number (`PATCH /auth/me`, body `whatsapp_number`; empty clears it). Only that field; name, email, role and status stay Admin-managed. The user is taken from the token, the number is normalized like in `POST /users`, and the change is audited (`updated_own_whatsapp`). Admin can still set it via `PUT /users/{id}`.

## Open decisions — ask before implementing if needed
- ~~Commission policy~~ — resolved by D-016.
- ~~Owner reuse across listings~~ — resolved by D-020 (no reuse).
- ~~Primary ownership transfer~~ — D-022.
- ~~Staff access to owner PII~~ — no (D-017).
- ~~Ownership transfer~~ — D-022.
- ~~Visible co-listing fields~~ — D-021.
- ~~Customer WhatsApp before login~~ — D-023.
- ~~Sold/Rented in default search~~ — D-024.
- Exact definition of which co-listing property fields are visible (default: marketing-safe fields only).
- ~~D-027 open points~~ — resolved (see D-027). Remaining: final commission scheme (default used until the business decides).
