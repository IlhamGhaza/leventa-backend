# Database Design

This is a logical schema guide, not a mandate to create every table exactly as written. Inspect the existing schema before adding or changing migrations.

## 1. Core tables

### users
- `id`
- `name`
- `email` (unique)
- `phone`
- `whatsapp_number` (marketing contact number; normalize format)
- `password_hash`
- `role` (`sales`, `admin`)
- `avatar`
- `status`
- `created_at`, `updated_at`, optional `deleted_at`

Do not expose `password_hash`.

### refresh_tokens
- `id`
- `user_id`
- `token_hash`
- `expires_at`
- `revoked_at`
- `created_at`

Store only token hashes.

### owners
- `id`
- `name`
- `email`
- `phone`
- `private_address`
- `created_at`, `updated_at`, optional `deleted_at`

Owner data must be returned only through authorized primary-marketing use cases. Consider encryption at rest for particularly sensitive fields if the project threat model requires it; never log these values.

### properties
Suggested fields:
- `id`
- `code` (unique)
- `primary_marketing_id` (FK users)
- `owner_id` (FK owners)
- `title`
- `type` (house, apartment, shop, factory, ...)
- `transaction_type` (`sale`, `rent`)
- `status` (`draft`, `available`, `reserved`, `maintenance`, `sold`, `rented`, `archived`; `archived` also means cancelled)
- `price`
- `rental_price` where applicable
- `description` (text)
- `province_id`, `city_id`, `district_id`
- `general_location`
- `block_number` (restricted)
- `cluster_name` / `cluster_detail` (restricted as appropriate)
- `street_name` (restricted)
- `house_number` (restricted)
- `floor_number` (restricted)
- `room_number` (restricted)
- `latitude`, `longitude` (do not expose precise coordinates if they reveal private location)
- `land_area`
- `building_area`
- `postal_code`, `sub_district` (kelurahan)
- `short_description`, `area_label`
- `maintenance_fee` (monthly)
- `orientation`
- `cert_shm`, `cert_hgb`, `cert_hgu`, `cert_shmsrs`, `cert_imb`, `cert_pbb`, `cert_girik`, `cert_slf`, `cert_other` (booleans)
- `floor_count`
- `bedrooms`, `bathrooms` (totals, used for filtering)
- rooms: `living_rooms`, `dining_rooms`, `master_bedrooms`, `child_bedrooms`, `other_bedrooms`, `master_bathrooms`, `child_bathrooms`, `other_bathrooms`, `family_rooms`, `study_rooms`, `kitchens`
- `electricity_capacity`, `air_conditioners`, `tv_provider`, `water_source`, `furnishing_status`
- outdoor: `has_pool`, `has_garage`, `has_front_yard`, `has_back_yard`, `has_balcony`
- distances in meters: `dist_school`, `dist_bus_terminal`, `dist_mrt_lrt`, `dist_toll`, `dist_hospital`
- custom extras (JSONB arrays, not filterable, max 10 each): `extra_outdoor` (`["Rooftop"]`), `extra_distances` (`[{"name":"Pasar","meters":800}]`)
- `thumbnail`: copy of the cover image URL; at most one `property_images.is_thumbnail` per property (partial unique index)
- `building_condition`, `parking_capacity`
- timestamps and optional `deleted_at`

The exact private address may be normalized into a separate table if that simplifies field-level access. Keep public/general location distinct from exact address.

### property_documents
- `id`, `property_id`, `doc_type` (`shm`, `hgb`, `hgu`, `shmsrs`, `imb`, `pbb`, `girik`, `slf`, `other`)
- `storage_path` (private file, never returned), `original_name`, `mime_type`, `size_bytes`
- `uploaded_by`, `created_at`
- Unique `(property_id, doc_type)`. Allowed only when the matching `cert_*` flag is true and the type applies to the property type (D-029).

### property_images
- `id`
- `property_id`
- `url` or storage key
- `sort_order`
- `is_thumbnail`
- timestamps

### property_marketing_assignments
- `id`
- `property_id`
- `marketing_id`
- `assignment_type` (`primary`, `co_listing`)
- `assigned_by`
- `assigned_at`

Constraints:
- Unique `(property_id, marketing_id)`.
- Exactly one primary assignment per property (enforce through transaction and, where practical, a partial unique index).
- At most two co-listing assignments.
- At most three total marketing assignments.
- Primary assignment must agree with `properties.primary_marketing_id`.
- Only active users with role `sales` can be assigned.
- Co-listing rows are created/removed/replaced only by Admin.
- Consider a PostgreSQL trigger or transactional locking if application-level checks alone cannot protect concurrent writes. Prefer a clear transaction with row locking and constraints over a complex mechanism.

### removed tables (D-026, D-027)
`units`, `tenants`, `favorites`, `inquiries`, `leads`, `lead_follow_ups`, `surveys` and `notifications` are not part of the schema; migration 000023 drops them.

### deals
Recorded by Admin only. `lead_id` is dropped; the deal hangs on the property.
- `id`
- `property_id`
- `sales_id` (primary Marketing at deal time)
- `recorded_by` (Admin)
- `outcome` (`deal`, `cancelled`)
- `type` (`sale`, `rent`; null when cancelled)
- `buyer_name` (required when `outcome` = `deal`)
- `price`
- `deal_date`
- `notes`
- timestamps

### commissions
- `id`
- `deal_id`
- `sales_id`
- `percentage`
- `amount`
- `status` (`pending`, `paid`, `cancelled`)
- timestamps

### audit_logs
- `id`
- `user_id`
- `action`
- `module`
- `entity_id`
- `metadata` (JSONB; never include secrets or owner PII)
- `ip_address`
- `user_agent`
- `created_at`

### master data
Property types, facilities, provinces, cities, districts, and status catalogs may be database-backed where admin CRUD is required. Use enums/static reference data for genuinely fixed values.

## 2. Indexing

Consider indexes for:
- `properties.primary_marketing_id`
- `properties.status`
- `properties.transaction_type`
- `properties.price`
- location foreign keys
- `property_marketing_assignments.property_id`
- `property_marketing_assignments.marketing_id`

Use composite indexes based on actual query patterns and verify with query plans when needed.

## 3. Migrations

- Use versioned `golang-migrate` migrations.
- Every schema change must include an up migration and a safe down migration where possible.
- Never edit an already-applied migration in a shared environment; add a new migration.
- Add constraints and indexes deliberately.
- Seed demo data separately from schema migrations.
