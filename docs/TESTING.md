# Testing Strategy

## 1. General

Use Go unit tests for domain/use-case logic and integration tests for PostgreSQL behavior. Use test fixtures/factories with isolated data. Tests must not rely on production data.

## 2. Authentication

- No register endpoint exists; no customer role can be created.
- Login does not accept a role selector.
- Wrong password fails.
- Expired/invalid access token fails.
- Refresh token rotates correctly.
- Revoked refresh token cannot be reused.
- Logout revokes the refresh token/session as designed.
- Password hashes are never returned.

## 3. Listing ownership

- Marketing A creates a property; A becomes primary owner.
- Marketing A can edit own listing.
- Marketing B cannot edit A's listing.
- Marketing B cannot delete/archive A's listing.
- Client-supplied `primary_marketing_id` cannot override authenticated actor.
- Co-listing assignment does not grant edit permission.
- Admin cannot edit property content through co-listing management.
- Ownership transfer cannot occur through ordinary co-listing endpoints.

## 4. Co-listing

- Admin can add an eligible Marketing user.
- Non-admin cannot add/remove/replace co-listing users.
- Duplicate assignment is rejected.
- Inactive or non-Marketing account cannot be assigned.
- One primary plus two co-listing users succeeds.
- A fourth marketing assignment is rejected.
- Concurrent assignment attempts cannot exceed the maximum.
- Primary cannot be removed through co-listing removal endpoint.
- Co-listing changes create audit records.

## 5. Owner and address privacy

For each role, verify response payloads and direct endpoint access:
- Primary Marketing can read owner details for own listing.
- Primary Marketing cannot read owner details for another listing.
- Co-listing Marketing cannot read owner details.
- Other Marketing cannot read owner details.
- Guest (no token) cannot read owner details.
- Admin cannot read owner details through ordinary endpoints.
- Restricted address fields are absent for unauthorized actors.
- Exact coordinates are not exposed where they would reveal protected location.
- No owner PII is written to application logs or audit metadata.

## 6. Guest access

- Unauthenticated callers can read only public property endpoints and receive the public projection (no owner data, no exact address, owner phone replaced by Marketing contact).
- Every other endpoint returns 401 without a token.

## 6b. Ownership documents

- Upload requires the matching `cert_*` flag; a type that does not apply to the property type is rejected.
- Only the primary marketing uploads/deletes; co-listing reads and downloads; other marketing, Admin and guests are denied.
- Content is sniffed (HTML named `.pdf` is rejected); replace removes the old file; cross-listing document IDs return 404.
- A flag cannot be cleared while its document exists.

## 7. Deal and commission

- Only Admin can record a deal or cancelled; Marketing gets 403.
- Sale sets property to `sold`; rent sets it to `rented`; buyer/tenant name is required for a Deal.
- Cancelled sets property to `archived`, creates no commission, and the property no longer appears in public or Marketing lists.
- Recording an outcome on an already sold/rented/archived property is rejected.
- Failure at any step rolls back deal, property and commission changes.
- Commission uses the default percentage; only Admin overrides it while `pending`.
- Property with a deal cannot be hard-deleted.

## 8. Search and WhatsApp

- Price min/max filters work.
- Location filters work.
- Land/building area filters work.
- Pagination and sort allowlist work.
- Public availability results exclude Sold/Rented.
- Public filter contract does not offer Sold/Rented.
- Historical Sold/Rented listing can display status badge.
- WhatsApp URL uses marketing user's number, never owner's.
- Message text is correctly URL-encoded.
- Email contact action is absent from the property contact contract.

## 8b. Logging

- A returned 500 is logged with its real cause and `request_id`; the response body stays generic.
- The request log line carries the real status of failed requests, not 200.

## 8c. Abuse protection

- A burst beyond the limit returns 429 with `Retry-After`.
- Search longer than 100 characters, more than 20 facility IDs or a page deeper than offset 10 000 returns 422.

## 9. Quality checks

Run:
```bash
gofmt -w .
go vet ./...
go test ./...
```

Also verify migrations from an empty local PostgreSQL database and confirm Swagger matches implemented endpoints.
