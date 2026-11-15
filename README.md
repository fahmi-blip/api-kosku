# KosKu API

Sistem backend manajemen kos-kosan berbasis Go, Fiber, PostgreSQL, dan GORM.

## Fitur Sistem

### Autentikasi dan Pengguna

- Registrasi akun penghuni dengan validasi email dan password.
- Login menggunakan access token JWT.
- Refresh token dengan rotasi token.
- Logout dengan pencabutan seluruh sesi refresh token pengguna.
- Melihat profil pengguna yang sedang login.
- Seed akun awal pemilik kos dari konfigurasi environment.

### Manajemen Kamar

- Melihat seluruh daftar kamar.
- Melihat kamar yang tersedia untuk disewa.
- Melihat detail kamar untuk pengguna terautentikasi.
- Menambah kamar baru.
- Memperbarui informasi dan status kamar.
- Menghapus kamar yang tidak sedang berstatus `occupied`.
- Mendukung status kamar `available`, `occupied`, dan `maintenance`.

### Manajemen Rental

- Penghuni dapat mengajukan sewa kamar.
- Validasi ketersediaan kamar, tanggal mulai, dan durasi sewa.
- Pencegahan pengajuan pending untuk kamar yang sama oleh penghuni yang sama.
- Pemilik dapat melihat seluruh pengajuan rental.
- Penghuni hanya dapat melihat pengajuan miliknya.
- Pemilik dapat menyetujui atau menolak pengajuan sewa.
- Penghuni dapat membatalkan pengajuan atau sewa miliknya.
- Status kamar otomatis berubah menjadi `occupied` ketika rental disetujui.
- Status kamar kembali menjadi `available` ketika rental yang disetujui dibatalkan.
- Riwayat perubahan status rental disimpan dalam `rental_status_logs`.

## Prasyarat
- Go 1.26.5+
- PostgreSQL

## Setup & Konfigurasi
1. Salin file environment:
   ```bash
   cp .env.example .env
   ```
2. Sesuaikan konfigurasi database dan kredensial seed pada `.env`.

## Cara Menjalankan
1. Unduh dependensi:
   ```bash
   go mod tidy
   ```
2. Jalankan aplikasi:
   ```bash
   go run main.go
   ```

Aplikasi akan otomatis menjalankan migrasi SQL, menyelaraskan model dengan GORM, dan membuat akun awal pemilik kos dari variabel env `SEED_OWNER_EMAIL` dan `SEED_OWNER_PASSWORD`. Data perubahan status rental disimpan pada tabel `rental_status_logs`.

### Pengujian Alur Sistem (curl)

Berikut langkah pengujian lengkap alur bisnis:

#### 1. Login Pemilik
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"pemilik@kosku.com","password":"pemilik12345"}'
```
*Simpan `access_token` dan `refresh_token` dari response.*

#### 1a. Cek Profil Sendiri (Me)
```bash
curl -X GET http://localhost:8080/api/v1/auth/me \
  -H "Authorization: Bearer <ACCESS_TOKEN>"
```

#### 1b. Refresh Token (Rotasi Token)
```bash
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{"refresh_token":"<REFRESH_TOKEN>"}'
```
*Token lama akan dicabut dan pasangan token baru dikembalikan.*

#### 1c. Logout (Tanpa Refresh Token)
```bash
curl -X POST http://localhost:8080/api/v1/auth/logout \
  -H "Authorization: Bearer <ACCESS_TOKEN>"
```
*Logout tidak memerlukan request body dan mencabut seluruh refresh token milik pengguna tersebut.*

#### 2. Tambah Kamar Baru (Pemilik)
```bash
curl -X POST http://localhost:8080/api/v1/rooms \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <ACCESS_TOKEN_PEMILIK>" \
  -d '{
    "code": "KMR-101",
    "name": "Kamar Deluxe 101",
    "price_per_month": 1500000,
    "facilities": "AC, Kamar Mandi Dalam, WiFi, Kasur Springbed"
  }'
```

#### 3. Cek Kamar yang Tersedia (Publik)
```bash
curl -X GET http://localhost:8080/api/v1/rooms/available
```

#### 4. Registrasi Penghuni Baru
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Budi Santoso",
    "email": "budi@example.com",
    "password": "password123",
    "phone": "081298765432"
  }'
```

#### 5. Login Penghuni
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"budi@example.com","password":"password123"}'
```
*Simpan `access_token` dan `refresh_token` penghuni dari response.*

#### 6. Pengajuan Sewa Kamar (Penghuni)
```bash
curl -X POST http://localhost:8080/api/v1/rentals \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <ACCESS_TOKEN_PENGHUNI>" \
  -d '{
    "room_id": 1,
    "start_date": "2026-12-01",
    "duration_months": 6
  }'
```

#### 7. Menyetujui Pengajuan Sewa (Pemilik)
```bash
curl -X PATCH http://localhost:8080/api/v1/rentals/1/approve \
  -H "Authorization: Bearer <ACCESS_TOKEN_PEMILIK>"
```
*Setelah disetujui dalam satu transaksi, status sewa berubah menjadi `approved` dan status kamar menjadi `occupied`.*

#### 8. Membatalkan Pengajuan atau Sewa (Penghuni)
```bash
curl -X PATCH http://localhost:8080/api/v1/rentals/1/cancel \
  -H "Authorization: Bearer <ACCESS_TOKEN_PENGHUNI>"
```
*Rental berstatus `pending` atau `approved` dapat dibatalkan oleh pemilik rental. Jika rental sudah `approved`, kamar dikembalikan menjadi `available`.*

#### 9. Melihat Riwayat Perubahan Status Rental
```bash
curl -X GET http://localhost:8080/api/v1/rentals/1/logs \
  -H "Authorization: Bearer <ACCESS_TOKEN>"
```

## Struktur Folder
```
kosku-backend/
├── app/
│   ├── model/        -> struct entitas dan bentuk request-response
│   ├── repository/   -> akses data: query ke database
│   └── service/      -> application service dan logika bisnis per domain
├── config/
│   ├── app.go        -> perakitan instance Fiber, middleware, dan error response
│   ├── env.go        -> variabel environment dengan godotenv
│   └── logger.go     -> log writer dan rotating log files
├── database/         -> koneksi database PostgreSQL, migrasi skema dan file SQL
│   ├── database.go   -> inisialisasi awal dan seeding akun
│   ├── migration.go  -> fungsi runner migrasi skema
│   └── postgres.go   -> koneksi pool PostgreSQL dengan GORM
├── helper/           -> fungsi bantu lintas package (jwt, authz, context, security, errors, response, request)
├── logs/             -> output log terotasi (tidak di-commit)
├── middleware/       -> auth, authz, dan middleware global (cors, helmet, recover, logger, requestid, requireJSON)
├── migrations/       -> berkas migrasi skema SQL (up)
├── route/
│   └── route.go      -> pendaftaran seluruh endpoint dalam satu file (Dependencies)
├── .env.example
└── main.go           -> titik masuk aplikasi dengan graceful shutdown
```
