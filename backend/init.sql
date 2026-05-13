CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TYPE truck_status AS ENUM ('ENROUTE', 'SHOP', 'STOP', 'UNAVAILABLE');
CREATE TYPE ownership_type AS ENUM ('MS Exp', 'Flinn', 'Owner O');
CREATE TYPE expense_category AS ENUM (
  'PM Service',
  'Oil change',
  'Tire issue',
  'Engine issue',
  'Towing',
  'Road Service',
  'Body work',
  'Leakage',
  'Kris Shop',
  'Truck Wash/Detailing',
  'Electrical issue',
  'Fluids/Truck Parts',
  'Brakes/Drums/Rotors',
  'Scale',
  'Other'
);
CREATE TYPE transponder_status AS ENUM ('Active', 'Inactive');
CREATE TYPE trailer_availability AS ENUM ('Ready', 'N/A', 'Returned', 'SALE');

-- ─────────────────────────────────────────
--  TRUCKS
-- ─────────────────────────────────────────
CREATE TABLE trucks (
  id                          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  unit_number                 VARCHAR(20)      NOT NULL UNIQUE,
  vin                         VARCHAR(17),
  year                        SMALLINT,
  make                        VARCHAR(20),
  company                     VARCHAR(50),
  ownership                   ownership_type,
  plate_number                VARCHAR(20),
  plate_state                 VARCHAR(10),
  status                      truck_status     NOT NULL DEFAULT 'ENROUTE',
  status_changed_at           DATE,
  status_note                 TEXT,
  samsara_id                  VARCHAR(30),
  dot_inspection_expires_at   DATE,
  dot_inspection_form_url     TEXT,
  next_pm_odometer            INT,
  next_oil_change_odometer    INT,
  notes                       TEXT,
  active                      BOOLEAN          NOT NULL DEFAULT TRUE,
  created_at                  TIMESTAMPTZ      NOT NULL DEFAULT NOW(),
  updated_at                  TIMESTAMPTZ      NOT NULL DEFAULT NOW()
);

-- ─────────────────────────────────────────
--  TRAILERS
-- ─────────────────────────────────────────
CREATE TABLE trailers (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  unit_number   VARCHAR(20)          NOT NULL UNIQUE,
  vin           VARCHAR(17),
  plate_number  VARCHAR(20),
  year          SMALLINT,
  make          VARCHAR(30),
  usage_type    VARCHAR(20),
  location      VARCHAR(50),
  availability  trailer_availability,
  notes         TEXT,
  created_at    TIMESTAMPTZ          NOT NULL DEFAULT NOW(),
  updated_at    TIMESTAMPTZ          NOT NULL DEFAULT NOW()
);

-- ─────────────────────────────────────────
--  TRANSPONDERS
-- ─────────────────────────────────────────
CREATE TABLE transponders (
  id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  truck_id                UUID REFERENCES trucks(id) ON DELETE SET NULL,
  transponder_number      VARCHAR(30),
  old_transponder_number  VARCHAR(30),
  mc_company              VARCHAR(20),
  status                  transponder_status NOT NULL DEFAULT 'Active',
  notes                   TEXT,
  created_at              TIMESTAMPTZ        NOT NULL DEFAULT NOW(),
  updated_at              TIMESTAMPTZ        NOT NULL DEFAULT NOW()
);

-- ─────────────────────────────────────────
--  TABLETS
-- ─────────────────────────────────────────
CREATE TABLE tablets (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  truck_id        UUID REFERENCES trucks(id) ON DELETE SET NULL,
  imei            VARCHAR(20),
  phone_number    VARCHAR(20),
  device_make     VARCHAR(30),
  device_model    VARCHAR(50),
  contract_type   VARCHAR(30),
  contract_start  DATE,
  contract_end    DATE,
  status          VARCHAR(20),
  notes           TEXT,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ─────────────────────────────────────────
--  MAINTENANCE LOGS
-- ─────────────────────────────────────────
CREATE TABLE maintenance_logs (
  id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  truck_id             UUID REFERENCES trucks(id) ON DELETE SET NULL,
  trailer_id           UUID REFERENCES trailers(id) ON DELETE SET NULL,
  expense_date         DATE             NOT NULL,
  week_label           VARCHAR(20),
  driver_name          VARCHAR(60),
  amount               NUMERIC(10, 2)   NOT NULL DEFAULT 0,
  category             expense_category NOT NULL,
  payment_type         VARCHAR(20),
  description          TEXT,
  reference_number     VARCHAR(100),
  who_covers           VARCHAR(30),
  paid_by              VARCHAR(30),
  telegram_message     TEXT,
  manager_verified     BOOLEAN          NOT NULL DEFAULT FALSE,
  accounting_verified  BOOLEAN          NOT NULL DEFAULT FALSE,
  invoice_file_url     TEXT,
  created_at           TIMESTAMPTZ      NOT NULL DEFAULT NOW(),
  updated_at           TIMESTAMPTZ      NOT NULL DEFAULT NOW(),
  CONSTRAINT chk_entity CHECK (truck_id IS NOT NULL OR trailer_id IS NOT NULL)
);

-- ─────────────────────────────────────────
--  INDEXES
-- ─────────────────────────────────────────
CREATE INDEX idx_trucks_status        ON trucks(status);
CREATE INDEX idx_trucks_unit          ON trucks(unit_number);
CREATE INDEX idx_trucks_dot_expiry    ON trucks(dot_inspection_expires_at);
CREATE INDEX idx_transponders_truck   ON transponders(truck_id);
CREATE INDEX idx_tablets_truck        ON tablets(truck_id);
CREATE INDEX idx_mlog_truck           ON maintenance_logs(truck_id);
CREATE INDEX idx_mlog_trailer         ON maintenance_logs(trailer_id);
CREATE INDEX idx_mlog_date            ON maintenance_logs(expense_date);
CREATE INDEX idx_mlog_category        ON maintenance_logs(category);