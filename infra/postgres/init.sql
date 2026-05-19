-- Buat semua database untuk tiap service
-- Script ini dijalankan otomatis saat container PostgreSQL pertama kali start

CREATE DATABASE db_users;
CREATE DATABASE db_orders;
CREATE DATABASE db_payments;
CREATE DATABASE db_notifications;

-- konfirmasi
\echo '✓ All databases created: db_users, db_orders, db_payments, db_notifications'
