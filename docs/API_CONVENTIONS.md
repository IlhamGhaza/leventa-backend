# API Conventions

## 1. Base path

`/api/v1`

## 2. Response envelope

Success:
```json
{
  "success": true,
  "message": "Request completed",
  "data": {}
}
```

Error:
```json
{
  "success": false,
  "message": "Validation failed",
  "errors": []
}
```

Use consistent error codes and HTTP statuses:
- 200 OK
- 201 Created
- 204 No Content
- 400 Bad Request
- 401 Unauthorized
- 403 Forbidden
- 404 Not Found
- 409 Conflict
- 422 Unprocessable Entity
- 500 Internal Server Error

Do not expose stack traces, SQL errors, secrets, or internal implementation details.

## 3. Endpoint groups

Implement endpoints according to the product requirements and actual UI flows. Candidate groups:

- `/auth`
- `/users`
- `/properties`
- `/properties/:id/images`
- `/properties/:id/co-listings`
- `/owners`
- `/deals` — Admin only: record deal or cancelled for a property
- `/commissions`
- `/master-data`
- `/audit-logs`

Do not expose an endpoint solely because it appears in this list; implement only with defined use cases and authorization.

## 4. Property APIs

Candidate:
- `GET /properties` — public (no token, guests) or role-scoped listing search.
- `POST /properties` — authenticated Marketing creates a listing; primary owner comes from actor context.
- `GET /properties/:id` — return role-appropriate projection.
- `PATCH /properties/:id` — primary Marketing only under current policy.
- `DELETE /properties/:id` — primary owner only if business rules permit; archive is preferred where historical references exist.
- `POST /properties/:id/images`
- `DELETE /properties/:id/images/:imageId`

Never accept `primary_marketing_id` from an ordinary create/update request to choose ownership. Set it from the authenticated actor.

## 5. Co-listing APIs

Candidate:
- `GET /properties/:id/co-listings` — authorized internal view.
- `POST /properties/:id/co-listings` — Admin only.
- `DELETE /properties/:id/co-listings/:marketingId` — Admin only.
- `PUT /properties/:id/co-listings/:marketingId` or a dedicated replacement endpoint — Admin only.

Validate:
- Marketing user exists, active, and has `sales` role.
- No duplicate assignment.
- Maximum three total including primary.
- Cannot remove or replace primary through co-listing endpoints.
- Perform changes transactionally.
- Write audit log.

## 5b. Ownership document APIs

- `GET /properties/:id/documents` — metadata, primary and co-listing marketing.
- `POST /properties/:id/documents` — multipart `doc_type` + `file`, primary marketing; requires the matching `cert_*` flag to be true.
- `GET /properties/:id/documents/:docId/download` — attachment, primary and co-listing marketing; audited.
- `DELETE /properties/:id/documents/:docId` — primary marketing.

Files live outside the static `/uploads` route; never return storage paths.

## 6. Owner APIs

Owner endpoints must be scoped to primary Marketing ownership. Do not create a general owner listing endpoint that leaks all owner records to every internal user.

## 7. Search and pagination

Support query parameters such as:
- `min_price`, `max_price`
- `province_id`, `city_id`, `district_id`
- `min_land_area`, `max_land_area`
- `min_building_area`, `max_building_area`
- `page`, `limit`, `sort`

Marketing assignment filters are internal and role-scoped. Do not offer `sold`/`rented` as public availability filter values. The server must enforce availability semantics regardless of client input.

Validate ranges, maximum page size, sorting allowlist, and filter combinations. Query at the database level; do not load all rows into memory.

Abuse limits (D-031): every `/api/v1` request is rate limited per IP (429 with `Retry-After`); `search` ≤ 100 characters, `facility_ids` ≤ 20, `limit` ≤ 100, `(page-1)*limit` ≤ 10 000.

## 8. WhatsApp link

The API may return an authorized marketing contact number or a generated WhatsApp URL only in a role-appropriate response. The number must belong to the marketing user representing the listing, never the owner.

Prefilled message should include property title/code, asking price, and general location. URL-encode message text correctly. Remove email contact actions from the property UI/API contract where they are no longer needed.

## 9. DTO and privacy rules

Use separate response DTOs/projections for:
- Public (guest) property.
- Marketing primary property.
- Marketing co-listing property.
- Admin property (no owner data, no exact address).

Do not serialize full database entities directly. Exclude owner PII and restricted address fields unless the current actor is explicitly authorized.

## 10. OpenAPI

Every implemented endpoint must have:
- Request schema.
- Response schema.
- Authentication requirement.
- Role/ownership behavior.
- Validation and error responses.
Update Swagger whenever API behavior changes.
