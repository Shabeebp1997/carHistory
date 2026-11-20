BEGIN;

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================================
-- USERS
-- ============================================================
CREATE TABLE IF NOT EXISTS users (
    id                          UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    name                        VARCHAR(255) NOT NULL,
    email                       VARCHAR(255) UNIQUE,
    password_hash               VARCHAR(255) NOT NULL,
    mobile_number               VARCHAR(20)  UNIQUE,
    timezone                    VARCHAR(100) DEFAULT 'UTC',
    email_notifications_enabled BOOLEAN      DEFAULT TRUE,
    sms_notifications_enabled   BOOLEAN      DEFAULT FALSE,
    is_active                   BOOLEAN      DEFAULT TRUE,
    created_at                  TIMESTAMPTZ  DEFAULT NOW(),
    updated_at                  TIMESTAMPTZ  DEFAULT NOW(),
    CONSTRAINT users_contact_required CHECK (email IS NOT NULL OR mobile_number IS NOT NULL)
);

CREATE INDEX IF NOT EXISTS idx_users_email         ON users(email) WHERE email IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_users_mobile_number ON users(mobile_number) WHERE mobile_number IS NOT NULL;

CREATE TABLE IF NOT EXISTS password_reset_tokens (
    id         UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token      VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at    TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_password_reset_token ON password_reset_tokens(token);

CREATE TABLE IF NOT EXISTS refresh_tokens (
    id         UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token      VARCHAR(512) UNIQUE NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token ON refresh_tokens(token);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens(user_id);

-- ============================================================
-- VEHICLE TYPES
-- ============================================================
CREATE TABLE IF NOT EXISTS vehicle_types (
    id         SERIAL       PRIMARY KEY,
    name       VARCHAR(100) UNIQUE NOT NULL,
    is_active  BOOLEAN      DEFAULT TRUE,
    created_at TIMESTAMPTZ  DEFAULT NOW()
);

INSERT INTO vehicle_types (name) VALUES
    ('Car'),
    ('Bike'),
    ('Scooter'),
    ('SUV'),
    ('Van'),
    ('Truck'),
    ('Bus'),
    ('Commercial Vehicle'),
    ('Electric Car'),
    ('Electric Bike'),
    ('Other')
ON CONFLICT (name) DO NOTHING;

-- ============================================================
-- VEHICLES
-- ============================================================
CREATE TABLE IF NOT EXISTS vehicles (
    id                  UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    vehicle_type_id     INTEGER      NOT NULL REFERENCES vehicle_types(id),
    brand               VARCHAR(100),
    model               VARCHAR(100),
    variant             VARCHAR(100),
    manufacturing_year  INTEGER,
    purchase_date       DATE,
    registration_number VARCHAR(50),
    chassis_number      VARCHAR(100),
    engine_number       VARCHAR(100),
    rc_owner            VARCHAR(255),
    owner_contact       VARCHAR(20),
    fuel_type           VARCHAR(50),
    current_odometer    DECIMAL(12,2) DEFAULT 0,
    colour              VARCHAR(50),
    nickname            VARCHAR(100),
    notes               TEXT,
    is_archived         BOOLEAN      DEFAULT FALSE,
    created_by          UUID         NOT NULL REFERENCES users(id),
    created_at          TIMESTAMPTZ  DEFAULT NOW(),
    updated_at          TIMESTAMPTZ  DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_vehicles_created_by ON vehicles(created_by);
CREATE INDEX IF NOT EXISTS idx_vehicles_registration ON vehicles(registration_number);

-- ============================================================
-- VEHICLE MEMBERS (SHARING)
-- ============================================================
CREATE TABLE IF NOT EXISTS vehicle_members (
    id         UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    vehicle_id UUID         NOT NULL REFERENCES vehicles(id) ON DELETE CASCADE,
    user_id    UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role       VARCHAR(20)  NOT NULL CHECK (role IN ('owner', 'editor', 'viewer')),
    joined_at  TIMESTAMPTZ  DEFAULT NOW(),
    created_at TIMESTAMPTZ  DEFAULT NOW(),
    UNIQUE (vehicle_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_vehicle_members_vehicle_id ON vehicle_members(vehicle_id);
CREATE INDEX IF NOT EXISTS idx_vehicle_members_user_id ON vehicle_members(user_id);

-- ============================================================
-- VEHICLE INVITATIONS
-- ============================================================
CREATE TABLE IF NOT EXISTS vehicle_invitations (
    id             UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    vehicle_id     UUID         NOT NULL REFERENCES vehicles(id) ON DELETE CASCADE,
    invited_by     UUID         NOT NULL REFERENCES users(id),
    invited_email  VARCHAR(255) NOT NULL,
    role           VARCHAR(20)  NOT NULL CHECK (role IN ('editor', 'viewer')),
    token          VARCHAR(255) UNIQUE NOT NULL,
    status         VARCHAR(20)  DEFAULT 'pending' CHECK (status IN ('pending', 'accepted', 'rejected', 'expired')),
    expires_at     TIMESTAMPTZ  NOT NULL,
    created_at     TIMESTAMPTZ  DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_invitations_token ON vehicle_invitations(token);
CREATE INDEX IF NOT EXISTS idx_invitations_email ON vehicle_invitations(invited_email);
CREATE INDEX IF NOT EXISTS idx_invitations_vehicle_id ON vehicle_invitations(vehicle_id);

-- ============================================================
-- MAINTENANCE TYPES
-- ============================================================
CREATE TABLE IF NOT EXISTS maintenance_types (
    id          SERIAL       PRIMARY KEY,
    name        VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    is_active   BOOLEAN      DEFAULT TRUE,
    created_at  TIMESTAMPTZ  DEFAULT NOW()
);

INSERT INTO maintenance_types (name) VALUES
    ('General Service'),
    ('Engine Oil'),
    ('Oil Filter'),
    ('Air Filter'),
    ('Brake Pads'),
    ('Brake Fluid'),
    ('Coolant'),
    ('Battery'),
    ('Tyres'),
    ('Wheel Alignment'),
    ('Wheel Balancing'),
    ('AC Service'),
    ('Spark Plug'),
    ('Chain Service'),
    ('Clutch'),
    ('Suspension'),
    ('Wiper'),
    ('Repair'),
    ('Custom Maintenance')
ON CONFLICT (name) DO NOTHING;

-- ============================================================
-- MAINTENANCE RECORDS
-- ============================================================
CREATE TABLE IF NOT EXISTS maintenance_records (
    id                      UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    vehicle_id              UUID         NOT NULL REFERENCES vehicles(id) ON DELETE CASCADE,
    maintenance_type_id     INTEGER      NOT NULL REFERENCES maintenance_types(id),
    title                   VARCHAR(255) NOT NULL,
    date_performed          DATE         NOT NULL,
    odometer_at_maintenance DECIMAL(12,2),
    next_due_date           DATE,
    next_due_odometer       DECIMAL(12,2),
    cost                    DECIMAL(12,2),
    service_provider        VARCHAR(255),
    description             TEXT,
    remarks                 TEXT,
    created_by              UUID         NOT NULL REFERENCES users(id),
    created_at              TIMESTAMPTZ  DEFAULT NOW(),
    updated_at              TIMESTAMPTZ  DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_maintenance_records_vehicle_id ON maintenance_records(vehicle_id);
CREATE INDEX IF NOT EXISTS idx_maintenance_records_date ON maintenance_records(date_performed);

-- ============================================================
-- SERVICE RECORDS
-- ============================================================
CREATE TABLE IF NOT EXISTS service_records (
    id               UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    vehicle_id       UUID         NOT NULL REFERENCES vehicles(id) ON DELETE CASCADE,
    service_date     DATE         NOT NULL,
    odometer         DECIMAL(12,2),
    service_type     VARCHAR(100),
    service_centre   VARCHAR(255),
    technician       VARCHAR(255),
    labour_cost      DECIMAL(12,2) DEFAULT 0,
    parts_cost       DECIMAL(12,2) DEFAULT 0,
    other_cost       DECIMAL(12,2) DEFAULT 0,
    total_cost       DECIMAL(12,2) DEFAULT 0,
    next_service_date DATE,
    next_service_km  DECIMAL(12,2),
    remarks          TEXT,
    created_by       UUID         NOT NULL REFERENCES users(id),
    created_at       TIMESTAMPTZ  DEFAULT NOW(),
    updated_at       TIMESTAMPTZ  DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_service_records_vehicle_id ON service_records(vehicle_id);
CREATE INDEX IF NOT EXISTS idx_service_records_date ON service_records(service_date);

CREATE TABLE IF NOT EXISTS service_items (
    id                UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    service_record_id UUID         NOT NULL REFERENCES service_records(id) ON DELETE CASCADE,
    name              VARCHAR(255) NOT NULL,
    description       TEXT,
    cost              DECIMAL(12,2) DEFAULT 0,
    created_at        TIMESTAMPTZ  DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_service_items_service_record_id ON service_items(service_record_id);

-- ============================================================
-- INSURANCE RECORDS
-- ============================================================
CREATE TABLE IF NOT EXISTS insurance_records (
    id                UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    vehicle_id        UUID         NOT NULL REFERENCES vehicles(id) ON DELETE CASCADE,
    insurance_company VARCHAR(255) NOT NULL,
    policy_number     VARCHAR(100),
    policy_type       VARCHAR(100),
    start_date        DATE         NOT NULL,
    expiry_date       DATE         NOT NULL,
    premium           DECIMAL(12,2),
    coverage_amount   DECIMAL(14,2),
    remarks           TEXT,
    created_by        UUID         NOT NULL REFERENCES users(id),
    created_at        TIMESTAMPTZ  DEFAULT NOW(),
    updated_at        TIMESTAMPTZ  DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_insurance_records_vehicle_id ON insurance_records(vehicle_id);
CREATE INDEX IF NOT EXISTS idx_insurance_records_expiry ON insurance_records(expiry_date);

-- ============================================================
-- PUC RECORDS
-- ============================================================
CREATE TABLE IF NOT EXISTS puc_records (
    id                 UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    vehicle_id         UUID         NOT NULL REFERENCES vehicles(id) ON DELETE CASCADE,
    certificate_number VARCHAR(100),
    test_date          DATE         NOT NULL,
    expiry_date        DATE         NOT NULL,
    testing_centre     VARCHAR(255),
    emission_info      TEXT,
    remarks            TEXT,
    created_by         UUID         NOT NULL REFERENCES users(id),
    created_at         TIMESTAMPTZ  DEFAULT NOW(),
    updated_at         TIMESTAMPTZ  DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_puc_records_vehicle_id ON puc_records(vehicle_id);
CREATE INDEX IF NOT EXISTS idx_puc_records_expiry ON puc_records(expiry_date);

-- ============================================================
-- TYRE RECORDS
-- ============================================================
CREATE TABLE IF NOT EXISTS tyre_records (
    id               UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    vehicle_id       UUID         NOT NULL REFERENCES vehicles(id) ON DELETE CASCADE,
    brand            VARCHAR(100),
    model            VARCHAR(100),
    size             VARCHAR(50),
    position         VARCHAR(50),
    purchase_date    DATE,
    installation_date DATE,
    installation_km  DECIMAL(12,2),
    removal_date     DATE,
    removal_km       DECIMAL(12,2),
    cost             DECIMAL(12,2),
    warranty_expiry  DATE,
    remarks          TEXT,
    created_by       UUID         NOT NULL REFERENCES users(id),
    created_at       TIMESTAMPTZ  DEFAULT NOW(),
    updated_at       TIMESTAMPTZ  DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_tyre_records_vehicle_id ON tyre_records(vehicle_id);

-- ============================================================
-- BATTERY RECORDS
-- ============================================================
CREATE TABLE IF NOT EXISTS battery_records (
    id                UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    vehicle_id        UUID         NOT NULL REFERENCES vehicles(id) ON DELETE CASCADE,
    brand             VARCHAR(100),
    model             VARCHAR(100),
    serial_number     VARCHAR(100),
    installation_date DATE,
    installation_km   DECIMAL(12,2),
    replacement_date  DATE,
    replacement_km    DECIMAL(12,2),
    warranty_expiry   DATE,
    cost              DECIMAL(12,2),
    remarks           TEXT,
    created_by        UUID         NOT NULL REFERENCES users(id),
    created_at        TIMESTAMPTZ  DEFAULT NOW(),
    updated_at        TIMESTAMPTZ  DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_battery_records_vehicle_id ON battery_records(vehicle_id);

-- ============================================================
-- ODOMETER RECORDS
-- ============================================================
CREATE TABLE IF NOT EXISTS odometer_records (
    id          UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    vehicle_id  UUID         NOT NULL REFERENCES vehicles(id) ON DELETE CASCADE,
    reading     DECIMAL(12,2) NOT NULL,
    recorded_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    source      VARCHAR(100),
    notes       TEXT,
    created_by  UUID         NOT NULL REFERENCES users(id),
    created_at  TIMESTAMPTZ  DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_odometer_records_vehicle_id ON odometer_records(vehicle_id);
CREATE INDEX IF NOT EXISTS idx_odometer_records_recorded_at ON odometer_records(recorded_at DESC);

-- ============================================================
-- REMINDERS
-- ============================================================
CREATE TABLE IF NOT EXISTS reminders (
    id                  UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    vehicle_id          UUID         NOT NULL REFERENCES vehicles(id) ON DELETE CASCADE,
    title               VARCHAR(255) NOT NULL,
    reminder_type       VARCHAR(30)  NOT NULL CHECK (reminder_type IN ('date', 'odometer', 'date_or_odometer')),
    due_date            DATE,
    due_odometer        DECIMAL(12,2),
    advance_days        INTEGER,
    advance_km          DECIMAL(12,2),
    related_entity_type VARCHAR(50),
    related_entity_id   UUID,
    is_active           BOOLEAN      DEFAULT TRUE,
    created_by          UUID         NOT NULL REFERENCES users(id),
    created_at          TIMESTAMPTZ  DEFAULT NOW(),
    updated_at          TIMESTAMPTZ  DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_reminders_vehicle_id ON reminders(vehicle_id);
CREATE INDEX IF NOT EXISTS idx_reminders_active ON reminders(is_active) WHERE is_active = TRUE;

-- ============================================================
-- NOTIFICATIONS
-- ============================================================
CREATE TABLE IF NOT EXISTS notifications (
    id            UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id       UUID         NOT NULL REFERENCES users(id),
    vehicle_id    UUID         REFERENCES vehicles(id),
    reminder_id   UUID         REFERENCES reminders(id),
    channel       VARCHAR(20)  NOT NULL CHECK (channel IN ('email', 'sms')),
    status        VARCHAR(20)  DEFAULT 'pending' CHECK (status IN ('pending', 'sent', 'failed')),
    subject       VARCHAR(255),
    message       TEXT         NOT NULL,
    sent_at       TIMESTAMPTZ,
    error_message TEXT,
    created_at    TIMESTAMPTZ  DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications(user_id);
CREATE INDEX IF NOT EXISTS idx_notifications_status ON notifications(status);
CREATE INDEX IF NOT EXISTS idx_notifications_reminder_id ON notifications(reminder_id);

-- ============================================================
-- DOCUMENTS
-- ============================================================
CREATE TABLE IF NOT EXISTS documents (
    id                  UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    vehicle_id          UUID         NOT NULL REFERENCES vehicles(id) ON DELETE CASCADE,
    related_entity_type VARCHAR(50),
    related_entity_id   UUID,
    document_type       VARCHAR(100),
    file_name           VARCHAR(255) NOT NULL,
    file_path           VARCHAR(500) NOT NULL,
    file_size           BIGINT,
    mime_type           VARCHAR(100),
    uploaded_by         UUID         NOT NULL REFERENCES users(id),
    created_at          TIMESTAMPTZ  DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_documents_vehicle_id ON documents(vehicle_id);
CREATE INDEX IF NOT EXISTS idx_documents_entity ON documents(related_entity_type, related_entity_id);

-- ============================================================
-- EXPENSES
-- ============================================================
CREATE TABLE IF NOT EXISTS expenses (
    id                  UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    vehicle_id          UUID         NOT NULL REFERENCES vehicles(id) ON DELETE CASCADE,
    category            VARCHAR(50)  NOT NULL CHECK (category IN ('service', 'maintenance', 'insurance', 'tyre', 'battery', 'fuel', 'repair', 'other')),
    related_entity_type VARCHAR(50),
    related_entity_id   UUID,
    expense_date        DATE         NOT NULL,
    description         VARCHAR(255),
    amount              DECIMAL(12,2) NOT NULL,
    vendor              VARCHAR(255),
    odometer            DECIMAL(12,2),
    notes               TEXT,
    created_by          UUID         NOT NULL REFERENCES users(id),
    created_at          TIMESTAMPTZ  DEFAULT NOW(),
    updated_at          TIMESTAMPTZ  DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_expenses_vehicle_id ON expenses(vehicle_id);
CREATE INDEX IF NOT EXISTS idx_expenses_date ON expenses(expense_date);
CREATE INDEX IF NOT EXISTS idx_expenses_category ON expenses(category);

-- ============================================================
-- ACTIVITY LOGS
-- ============================================================
CREATE TABLE IF NOT EXISTS activity_logs (
    id          UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    vehicle_id  UUID         REFERENCES vehicles(id),
    user_id     UUID         NOT NULL REFERENCES users(id),
    action      VARCHAR(100) NOT NULL,
    entity      VARCHAR(100),
    entity_id   UUID,
    description TEXT,
    created_at  TIMESTAMPTZ  DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_activity_logs_vehicle_id ON activity_logs(vehicle_id);
CREATE INDEX IF NOT EXISTS idx_activity_logs_user_id ON activity_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_activity_logs_created_at ON activity_logs(created_at DESC);

COMMIT;
