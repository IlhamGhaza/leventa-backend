# Claude Code Instructions

Baca `AGENTS.md` terlebih dahulu. Setelah itu baca dokumen dalam `docs/` yang relevan dengan task.

## Cara bekerja

1. Inspect repository sebelum melakukan perubahan.
2. Identifikasi modul, pola, dependency, migration, dan test yang sudah ada.
3. Jangan mengganti stack atau arsitektur yang telah disepakati.
4. Implementasikan satu task secara end-to-end: migration (jika perlu), entity/model, repository, use case, handler, route, validation, dokumentasi API, dan tests.
5. Jalankan `gofmt`, `go vet`, dan test yang relevan.
6. Laporkan file yang diubah, keputusan penting, test yang dijalankan, serta hal yang belum selesai.

## Larangan

- Jangan menambahkan Redis atau Docker.
- Jangan mengubah aplikasi menjadi microservices.
- Jangan membuat role selector pada login.
- Jangan mempercayai `user_id`, role, owner ID, atau marketing ID dari request body untuk menentukan hak akses. Identitas aktor harus berasal dari token yang sudah diverifikasi.
- Jangan hanya menyembunyikan data sensitif di Flutter. Backend harus menghilangkannya dari response dan menolak akses yang tidak sah.
- Jangan mengizinkan Admin mengedit detail listing atau data pemilik melalui fitur co-listing management.
- Jangan memberikan hak edit properti kepada co-listing marketing.
- Jangan mengekspos nomor atau identitas pemilik kepada co-listing, marketing lain, customer, atau publik.
- Jangan membuat endpoint atau business rule yang bertentangan dengan dokumen spesifikasi.

## Saat requirement tidak jelas

Jangan menebak perilaku bisnis yang berdampak pada akses data, uang, kepemilikan listing, atau status transaksi. Tunjukkan ambiguitas dan ajukan satu pertanyaan terarah. Untuk detail implementasi non-bisnis, pilih solusi Go yang sederhana dan idiomatis lalu dokumentasikan.

## Definition of done

Task belum selesai hanya karena kode berhasil dikompilasi. Pastikan authorization, validasi, error handling, migration, tests, dan dokumentasi ikut diperbarui sesuai dampak perubahan.
