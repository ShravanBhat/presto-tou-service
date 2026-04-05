-- Enable btree_gist extension for exclusion constraints on TIME ranges
CREATE EXTENSION IF NOT EXISTS btree_gist;

CREATE TABLE chargers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    timezone VARCHAR(100) NOT NULL DEFAULT 'UTC',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- TOU schedules support overnight/midnight-crossing periods.
-- We store start_time and end_time as TIME, and use an is_overnight flag
-- to correctly handle periods that cross midnight (e.g., 22:00 - 06:00).
CREATE TABLE tou_schedules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    charger_id UUID NOT NULL REFERENCES chargers(id) ON DELETE CASCADE,
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    price_per_kwh DECIMAL(10, 4) NOT NULL CHECK (price_per_kwh >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    -- Validate time ranges:
    -- start_time must be logically before end_time (00:00 represents the end resolving to 24:00:00 natively or conditionally)
    CONSTRAINT valid_time_range CHECK (
        start_time != end_time AND
        (start_time < end_time OR end_time = '00:00:00')
    ),
    -- Prevent overlapping schedules for the same charger
    -- Linear mapping across a single date successfully catches overlaps natively.
    EXCLUDE USING gist (
        charger_id WITH =,
        tsrange(
            '1970-01-01'::date + start_time,
            '1970-01-01'::date + CASE WHEN end_time = '00:00:00' THEN '24:00:00'::TIME ELSE end_time END
        ) WITH &&
    )
);

CREATE INDEX idx_tou_schedules_charger_id ON tou_schedules(charger_id);
CREATE INDEX idx_tou_schedules_charger_time ON tou_schedules(charger_id, start_time, end_time);
