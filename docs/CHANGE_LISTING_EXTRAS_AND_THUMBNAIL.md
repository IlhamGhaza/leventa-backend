# Change: area luar & fasilitas umum custom, pilih thumbnail foto

Tanggal: 2026-11-16 · Diminta oleh: tim FE (Leventa Mobile) · Status: **sudah diimplementasi di repo ini**, perlu review & migrasi di tiap environment.

## Kenapa

1. **Area luar** hanya 5 switch tetap (`has_pool`, `has_garage`, `has_front_yard`, `has_back_yard`, `has_balcony`). Sales tidak bisa mencatat area lain (rooftop, gazebo, carport, dll.).
2. **Jarak ke fasilitas umum** hanya 5 kolom tetap (`dist_school`, `dist_bus_terminal`, `dist_mrt_lrt`, `dist_toll`, `dist_hospital`). Fasilitas lain (pasar, masjid, mall, dll.) tidak bisa dicatat.
3. **Foto**: tabel `property_images` punya `is_thumbnail` dan repo punya `SetThumbnail`, tapi tidak ada endpoint, dan `properties.thumbnail` tidak pernah diisi. Akibatnya kartu di endpoint list (yang tidak memuat `images`) tampil tanpa foto, dan Sales tidak bisa memilih sampul.

Kolom tetap **dipertahankan** (tidak ada breaking change). Item custom ditambahkan di sebelahnya.

## Ringkasan perubahan API

| | Sebelum | Sesudah |
|---|---|---|
| Body create/update listing | — | `extra_outdoor`, `extra_distances` (opsional) |
| Response detail & list | — | `extra_outdoor`, `extra_distances` (selalu array, tidak pernah `null`) |
| Endpoint baru | — | `PUT /properties/{id}/images/{imageId}/thumbnail` |
| Upload foto | `is_thumbnail` selalu `false` | foto pertama otomatis jadi sampul |
| Hapus foto | sampul bisa hilang | foto sampul dihapus → foto berikutnya dipromosikan; foto terakhir → `thumbnail` dikosongkan |
| `thumbnail` di list | selalu kosong | URL foto sampul |

### Field baru

```json
{
  "extra_outdoor": ["Rooftop", "Gazebo"],
  "extra_distances": [
    { "name": "Pasar", "meters": 800 },
    { "name": "Masjid", "meters": 150 }
  ]
}
```

Validasi (422 bila dilanggar):

| Field | Aturan |
|---|---|
| `extra_outdoor` | maks 10 item; tiap item string 1–50 karakter |
| `extra_distances` | maks 10 item; `name` wajib, 1–50 karakter; `meters` integer 0–1.000.000 |

Normalisasi di server (bukan error): nama di-trim, item kosong dibuang, duplikat (tidak peka huruf besar/kecil) dibuang — item pertama yang dipakai.

Semantik PATCH/PUT `/properties/{id}`:
- field **tidak dikirim** → daftar tidak berubah;
- `[]` → daftar dikosongkan;
- array berisi → **mengganti seluruh** daftar (bukan append).

Privasi: field ini publik (sama seperti `has_*` / `dist_*`), tidak berisi data pemilik. Hak edit tetap hanya Primary Marketing (aturan `requirePrimary` yang sama).

Filter pencarian: item custom **tidak** bisa dipakai sebagai filter (teks bebas). Bila nanti perlu filter, item yang sering muncul dipindah jadi master data / kolom tetap.

### Endpoint baru: pilih sampul

```
PUT /api/v1/properties/{id}/images/{imageId}/thumbnail
Authorization: Bearer <token Sales>
```

- 200 `Thumbnail updated successfully`
- 403 bukan Primary Marketing listing tersebut (co-listing juga ditolak)
- 404 listing/foto tidak ada, atau foto milik listing lain

Efek: hanya satu foto `is_thumbnail = true` per listing, dan URL-nya disalin ke `properties.thumbnail`.

## Database — migrasi `000026_listing_extras_and_thumbnail`

Up:
- `properties.extra_outdoor JSONB NOT NULL DEFAULT '[]'` + CHECK harus array.
- `properties.extra_distances JSONB NOT NULL DEFAULT '[]'` + CHECK harus array.
- Backfill thumbnail: per listing, bila ada >1 `is_thumbnail` disisakan satu; bila tidak ada, foto pertama (`sort_order`, `created_at`) dijadikan sampul; `properties.thumbnail` diisi URL-nya.
- Partial unique index `uq_property_images_one_thumbnail ON property_images(property_id) WHERE is_thumbnail`.

Down: hapus index, constraint, dan dua kolom JSONB. Backfill thumbnail tidak di-revert (data tetap valid untuk skema lama).

Jalankan per environment:

```bash
DATABASE_URL=... go run ./cmd/migrate -up
```

## File yang berubah

- `database/migrations/000026_listing_extras_and_thumbnail.{up,down}.sql`
- `internal/domain/entity/property.go` — `ExtraOutdoor`, `ExtraDistances`, tipe `NearbyPlace`
- `internal/usecase/property_details.go` — field + validasi di create/patch/view, normalisasi
- `internal/domain/repository/property_repository.go`, `internal/repository/postgres/other_repositories.go` — `SetThumbnail` ikut transaksi, sinkron `properties.thumbnail`, `uuid.Nil` = kosongkan
- `internal/usecase/property_usecase.go` — `UploadImage` (sampul otomatis), `SetThumbnail`, `DeleteImage` (promosi sampul), semuanya dalam transaksi
- `internal/delivery/http/handler/property_handler.go`, `internal/routes/routes.go` — endpoint baru
- `docs/openapi.yaml`, `docs/DATABASE.md`
- Test: `internal/usecase/property_details_test.go` (validasi), `internal/usecase/listing_extras_integration_test.go` (simpan/normalisasi/patch, sampul + otorisasi + promosi)

## Hasil test

`go build`, `go vet`, `gofmt` bersih; `go test ./...` lulus dengan `TEST_DATABASE_URL` (migrasi 26 juga diuji down → up).

## Catatan lain (belum diubah)

- `toDetailsView` tidak mengisi `cert_shmsrs`, `cert_hgu`, `cert_slf`, jadi ketiganya selalu `false` di response walau tersimpan `true`. Perlu diperbaiki terpisah.
- Upload foto tidak mengisi `sort_order` (semua 0); urutan foto jadi tidak pasti. Perlu endpoint urutkan foto bila dibutuhkan.
