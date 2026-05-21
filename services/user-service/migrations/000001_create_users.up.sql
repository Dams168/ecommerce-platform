-- buat tabel users
CREATE TABLE IF NOT EXISTS users (
    -- id pakai UUID bukan angka (1,2,3)
    -- alasan: kalau pakai angka, orang bisa tebak ada berapa user
    -- UUID: "a3f8c2d1-..." tidak bisa ditebak
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    name       VARCHAR(100) NOT NULL,
    email      VARCHAR(255) NOT NULL,

    -- password disimpan dalam bentuk hash, bukan teks asli
    -- kalau database bocor, password tetap aman
    password   VARCHAR(255) NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- buat index di kolom email supaya pencarian cepat
-- tanpa index, database harus scan semua baris untuk cari satu email
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email ON users(email);