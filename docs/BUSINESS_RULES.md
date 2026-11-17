# Business Rules

## 1. Account rules

1. No self-registration. Customers are guests without an account.
2. Login accepts email/password only. Backend resolves role from the database.
3. Only Admin creates Marketing (`sales`) and Admin accounts.
4. A user cannot change their own role.

## 2. Listing ownership rules

1. The Marketing user who creates a property becomes its primary owner. Exactly one per property.
2. Only primary Marketing edits listing details and owner details.
3. Co-listing Marketing is read-only (plus status toggle, D-021).
4. Co-listing assignment never transfers ownership.
5. Admin manages co-listing membership and can transfer primary ownership (audited, D-022). Admin never edits listing content.

## 3. Co-listing limits

1. Maximum three Marketing users per property including primary (two co-listing).
2. Only Admin adds/removes/replaces co-listing members.
3. Assigned user must be active with role `sales`. No duplicates.
4. Limits are enforced transactionally.

## 4. Owner privacy

1. Owner details are visible only to the primary Marketing owner.
2. Co-listing, other Marketing, Admin (ordinary workflows) and guests never see owner details.
3. The owner's phone is replaced by the representing Marketing user's contact everywhere.

## 5. Address privacy

1. General location is public.
2. Exact block, cluster detail, street/house number, floor and room are visible only to primary and co-listing Marketing of that listing.
3. Do not expose precise coordinates if they reveal the protected address.

## 6. Deal rules

1. Only Admin records an outcome for a property: **Deal** or **Cancelled**. Marketing reports to Admin outside the system.
2. Deal: Admin enters type (sale or rent), final price, date, buyer/tenant name and notes. A sale sets the property to `sold`; a rental sets it to `rented`. A commission record is created.
3. Cancelled: the owner is no longer selling or renting. The property becomes `archived` and no longer appears in any public or Marketing list by default. No commission. Admin enters an optional note.
4. Sold/Rented properties stay visible as history with a badge (D-024). Archived properties do not.
5. A property with a deal cannot be hard-deleted; archive instead.
6. Recording an outcome is one database transaction: lock property, validate it is not already sold/rented/archived, create the record, update property status, create commission (Deal only), write audit log.

## 7. Commission

Default percentage from `COMMISSION_DEFAULT_PERCENTAGE` (2.5), rounded to 2 decimals. The commission belongs to the property's primary Marketing at the time of the deal. Admin may override the percentage and set status (`pending` → `paid`/`cancelled`) only while `pending` (D-016). The default scheme is used until the business defines another one.

## 8. Status display

Sold/Rented listings stay visible with a small badge, are never shown as available, and are not offered as availability filters.

## 9. WhatsApp

Guests contact the representing Marketing user's own WhatsApp number with a prefilled message (title/code, price, general location). Owner phone is never used or returned.
