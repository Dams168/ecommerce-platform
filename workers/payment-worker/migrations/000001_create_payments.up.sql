CREATE TABLE IF NOT EXISTS payments (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- order_id dari order-service
    -- tidak pakai foreign key karena beda database
    order_id   UUID NOT NULL UNIQUE,

    user_id    UUID NOT NULL,
    amount     DECIMAL(12,2) NOT NULL,

    -- status: pending, completed, failed
    status     VARCHAR(50) NOT NULL DEFAULT 'pending',

    -- untuk idempotency — catat kapan diproses
    processed_at TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- index untuk cek idempotency dengan cepat
CREATE UNIQUE INDEX IF NOT EXISTS idx_payments_order_id ON payments(order_id);
