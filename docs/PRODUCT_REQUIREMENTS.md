# Product Requirements

## 1. Product overview

Internal property listing system for one company. Not an e-commerce marketplace, no customer accounts, no chat (WhatsApp and phone are used outside the system).

Roles (each has an account, created by Admin only):
1. Marketing (backend role value: `sales`)
2. Admin

The `staff` role is removed (D-027). Customers are **guests** without an account; they only view the public catalog.

## 2. Main workflow

`Property owner calls a Marketing user or the office → the Marketing user concerned creates the listing (house, apartment, factory, shop, ...) → listing is published with privacy-filtered data → Guest sees the listing and contacts the Marketing user via WhatsApp → Marketing tells Admin about a deal or a cancellation (outside the system) → Admin records Deal (property becomes sold/rented, commission created) or Cancelled (owner no longer sells; property is hidden from lists)`

- The Marketing user who creates the listing becomes `primary_marketing_id`.
- Admin may assign up to two co-listing Marketing users.
- Admin may change the primary Marketing (ownership transfer, D-022).

## 3. Guest capabilities (no account)

- Browse, search, filter and view public property details.
- See the representing Marketing user's contact (WhatsApp) in place of the owner's phone.

Guests cannot see owner identity/contact or the exact private address, and cannot do anything else.

## 4. Marketing capabilities

- Create listings (property types: house, apartment, shop_house/ruko, factory).
- Edit only their own primary-owned listings, including images and owner details.
- View owner details and exact address only for listings they own as primary.
- Co-listing Marketing: read-only view of the listing including exact address, never owner data; may toggle availability status (D-021).
- Use their own WhatsApp number as the public contact.
- View their own commission records.

Marketing must not edit another Marketing user's listing, see owner data of other listings, change their own role, or record deals.

## 5. Admin capabilities

- Create and manage Marketing and Admin accounts.
- Manage co-listing membership (max three Marketing per property including primary).
- Change the primary Marketing of a listing (audited ownership transfer).
- Record a **Deal** (sale or rent, with buyer/tenant name) or **Cancelled** for a property after Marketing reports it.
- Manage commission: default percentage, override and status (D-016).
- View audit logs.
- Admin cannot edit listing content or owner data, and has no owner PII access.

## 6. Property search (public)

Filters: price range, general location (province, city/regency, district, area), land area range, building area range, property type, transaction type (sale/rent). Sold/Rented are not offered as availability filter options; historical listings keep a SOLD/RENTED badge.

## 7. Property description

Reference: https://syahira.co.id/wdk-listing/cluster-mawar/ (screenshot in `docs/syahira_prop_desc.png`). A listing page has the sections below. Fields are optional unless marked; fields not relevant to the property type (e.g. bedrooms for a factory) are left empty.

**Header:** title, price, transaction type (Jual/Sewa), short description, area label (e.g. "Kota Wisata Cibubur"), image gallery with thumbnails.

**Data & Keterangan:** transaction purpose, city/regency (Daerah), sub-district (Kelurahan), postal code, land area, building area, house orientation (Rumah Menghadap), sale price, monthly maintenance fee (Biaya Perawatan/bulan). The reference page also shows the owner's phone here; this system must **not**. It is replaced by the Marketing contact (D-028).

**Dokumen Kepemilikan:** checkboxes per property type: house (SHM, HGB, IMB, PBB, Girik, Lainnya), apartment (SHMSRS, HGB, IMB, PBB, SLF, Lainnya), shop_house (SHM, HGB, IMB, PBB, SLF, Lainnya), factory (SHM, HGB, HGU, IMB, PBB, SLF, Lainnya). When a box is checked, the primary marketing can upload that document (PDF/JPEG/PNG); primary and co-listing marketing can download it to send to a customer outside the system. Documents are never public (D-029).

**Spesifikasi Ruangan:** floors, living room, dining room, master bedrooms, children bedrooms, other bedrooms, master bathrooms, children bathrooms, other bathrooms, family room, study room, kitchen.

**Fasilitas:** electricity capacity (VA/KVA), air conditioners (count), TV antenna/subscription provider, water source (e.g. PAM), furnishing (e.g. furnished).

**Detail Outdoor:** checkboxes swimming pool, garage, front yard, back yard, balcony.

**Jarak ke Fasilitas Umum:** distance in meters to school, bus terminal, MRT/LRT, toll gate, hospital.

**Peta:** map pin at an approximate location. Never an exact pin that reveals the restricted address (D-011/D-021).

**Contact card:** the representing Marketing user's name, photo and WhatsApp. No email and no owner data.

**Not included from the reference page:** the "Kirim Pertanyaan" form (Name, Phone, Email, Pertanyaan). Questions go through WhatsApp only (D-026). Featured/similar property blocks and the property-type and location counters are optional public list features, not required.

Property types (`type`): `house` (Rumah), `apartment` (Apartemen), `shop_house` (Ruko), `factory` (Pabrik).

Keep a flexible free-text description in addition to the structured fields (paragraphs, bullets). Do not force all narrative into rigid columns.

## 8. Contact

WhatsApp only, using the representing Marketing user's number with a URL-encoded prefilled message (title/code, price, general location). No email button. The owner's number is never returned.
