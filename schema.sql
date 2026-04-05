
CREATE TABLE chargers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    timezone VARCHAR(100) NOT NULL DEFAULT 'UTC',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE tou_schedules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    charger_id UUID NOT NULL REFERENCES chargers(id) ON DELETE CASCADE,
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    price_per_kwh DECIMAL(10, 4) NOT NULL CHECK (price_per_kwh >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_tou_schedules_charger_id ON tou_schedules(charger_id);
CREATE INDEX idx_tou_schedules_charger_time ON tou_schedules(charger_id, start_time, end_time);
