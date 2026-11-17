# Authentication, Roles, Ownership, and Privacy

## 1. Authentication

One universal login endpoint and screen. No role selector.

Endpoints:
- `POST /api/v1/auth/login`
- `POST /api/v1/auth/refresh`
- `POST /api/v1/auth/logout`
- `GET /api/v1/auth/me`
- `PATCH /api/v1/auth/me` (own `whatsapp_number` only, D-033)

Login request:
```json
{
  "email": "user@example.com",
  "password": "secret"
}
```

The backend looks up the account and returns the role. Flutter routes to the correct role experience using the authenticated user profile.

There is no self-registration. Customers are guests without an account. Only Admin creates `sales` and `admin` accounts.

## 2. Role values

Use stable backend values:
- `sales` (UI/product label: Marketing)
- `admin`

Do not rename the persisted `sales` role casually; if changing it, provide a migration and update all clients/contracts.

## 3. Permission matrix

| Capability | Guest (no account) | Marketing primary | Marketing co-listing | Admin |
|---|---:|---:|---:|---:|
| Browse public listings (owner data and exact address hidden) | Yes | Yes | Yes | Yes |
| Create listing | No | Yes | No | No |
| Edit listing details / images | No | Own primary listings | No | No |
| Toggle availability status | No | Own primary listings | Yes (D-021) | No |
| View owner contact | No | Own primary listings | No | No ordinary access |
| View exact private address | No | Own primary listings | Yes (D-021) | No ordinary access |
| Upload / delete ownership documents | No | Own primary listings | No | No |
| View / download ownership documents | No | Own primary listings | Yes | No |
| Manage co-listing assignments | No | No | No | Yes |
| Change primary Marketing (transfer) | No | No | No | Yes (audited) |
| Manage users | No | No | No | Yes |
| Record deal / cancel | No | No | No | Yes |
| Manage commission | No | View own | View own | Yes |

Guests have no token; every endpoint they reach is public, read-only and returns the public projection only. The Staff role is removed by D-027.

This matrix is a baseline. Do not silently broaden Admin access to private owner data.

## 4. Property ownership

Each property has exactly one `primary_marketing_id`.

- The user who creates a listing becomes its primary owner.
- Only that primary marketing user can edit listing content and owner details.
- Co-listing assignment does not transfer ownership.
- Admin's co-listing permission does not grant edit permission.
- Ownership transfer is a separate explicit workflow, not an ordinary co-listing operation.

## 5. Co-listing

Maximum three marketing users per property, including the primary owner:
- 1 primary marketing.
- 0–2 co-listing marketing users.

Only Admin manages co-listing membership. Co-listing marketing has read-only access to approved marketing fields.

> Update (D-021/D-022/D-023): co-listing marketing sees the primary view of its listings (including exact address, never owner data) and may toggle availability status; Admin can transfer the primary owner; the marketing WhatsApp contact is public. Where this file says exact address is "primary only", D-021 prevails.

## 6. Owner privacy

Owner PII:
- Full name.
- Phone.
- Email.
- Private address.
- Other personal contact information.

Access:
- Primary Marketing: own listing only.
- Co-listing Marketing: denied.
- Other Marketing: denied.
- Guest/public: denied.
- Admin: no ordinary access; exceptional access requires a separate explicit audited authorization path.

Use separate DTOs/query projections for each access scope. Do not serialize a full GORM model and hope the frontend hides fields.

## 7. Exact address privacy

Split general location from restricted address details.

General location may include province, city/regency, district, and broad neighborhood/area.

Restricted fields include:
- Block number.
- Internal cluster/block details.
- Street name and house number.
- Private floor identifier.
- Room number.

Only the primary marketing owner may access the complete address under the current policy. Do not return restricted fields to other marketing users, admin, or public endpoints. Co-listing marketing of that listing also sees it (D-021).

## 8. Authorization implementation

- Authentication middleware verifies token and attaches actor ID/role.
- Role middleware may reject entire route groups.
- Use cases must still enforce record-level ownership.
- Repository methods should support scoped reads/updates.
- Never trust actor IDs or roles supplied in request payloads.
- Return 403 for authenticated but unauthorized access; consider 404 for scoped resources where revealing existence would leak information. Apply consistently.

## 9. IDOR tests

Test attempts to access another user's:
- Property edit endpoint.
- Owner details.
- Exact address.
- Commission.
- Notification.
