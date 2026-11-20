-- ============================================================
-- 002: Profile pictures, admin flag, purchase price,
--      media attachments, fuel/charging logs
-- All statements are idempotent (safe to re-run)
-- ============================================================

-- Users: profile picture + admin flag
ALTER TABLE users ADD COLUMN IF NOT EXISTS profile_picture  VARCHAR(500);
ALTER TABLE users ADD COLUMN IF NOT EXISTS is_admin         BOOLEAN NOT NULL DEFAULT FALSE;

-- Vehicles: purchase price
ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS purchase_price DECIMAL(12,2);

-- ── Media attachments ────────────────────────────────────────
CREATE TABLE IF NOT EXISTS media (
    id          UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    entity_type VARCHAR(50)  NOT NULL,
    entity_id   UUID         NOT NULL,
    file_path   VARCHAR(500) NOT NULL,
    file_name   VARCHAR(255) NOT NULL,
    media_type  VARCHAR(20)  NOT NULL DEFAULT 'image',
    mime_type   VARCHAR(100),
    file_size   BIGINT       NOT NULL DEFAULT 0,
    uploaded_by UUID         REFERENCES users(id) ON DELETE SET NULL,
    created_at  TIMESTAMPTZ  DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_media_entity ON media(entity_type, entity_id);

-- ── Fuel / charging logs ─────────────────────────────────────
CREATE TABLE IF NOT EXISTS fuel_logs (
    id               UUID          PRIMARY KEY DEFAULT uuid_generate_v4(),
    vehicle_id       UUID          NOT NULL REFERENCES vehicles(id) ON DELETE CASCADE,
    logged_by        UUID          NOT NULL REFERENCES users(id),
    fuel_type        VARCHAR(20)   NOT NULL DEFAULT 'petrol',
    quantity         DECIMAL(10,3),
    price_per_unit   DECIMAL(10,2),
    total_cost       DECIMAL(12,2) NOT NULL,
    odometer_reading DECIMAL(10,2),
    fuel_station     VARCHAR(255),
    date             DATE          NOT NULL DEFAULT CURRENT_DATE,
    notes            TEXT,
    created_at       TIMESTAMPTZ   DEFAULT NOW(),
    updated_at       TIMESTAMPTZ   DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_fuel_logs_vehicle ON fuel_logs(vehicle_id);
CREATE INDEX IF NOT EXISTS idx_fuel_logs_date    ON fuel_logs(vehicle_id, date);
