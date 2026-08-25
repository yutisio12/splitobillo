CREATE TABLE IF NOT EXISTS receipts (
    id                BIGSERIAL PRIMARY KEY,
    session_id        VARCHAR(64)  NOT NULL,
    receipt_number    VARCHAR(128) NOT NULL DEFAULT '',
    merchant_name     VARCHAR(255) NOT NULL DEFAULT '',
    transaction_date  TIMESTAMPTZ,
    subtotal          BIGINT       NOT NULL DEFAULT 0,
    tax               BIGINT       NOT NULL DEFAULT 0,
    service_charge    BIGINT       NOT NULL DEFAULT 0,
    discount          BIGINT       NOT NULL DEFAULT 0,
    total             BIGINT       NOT NULL DEFAULT 0,
    image_url         VARCHAR(512) NOT NULL DEFAULT '',
    ocr_raw_text      TEXT         NOT NULL DEFAULT '',
    status            VARCHAR(32)  NOT NULL DEFAULT 'UPLOADED',
    created_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_receipts_session_id ON receipts (session_id);
CREATE INDEX IF NOT EXISTS idx_receipts_status ON receipts (status);

CREATE TABLE IF NOT EXISTS receipt_items (
    id            BIGSERIAL PRIMARY KEY,
    receipt_id    BIGINT       NOT NULL REFERENCES receipts (id) ON DELETE CASCADE,
    name          VARCHAR(255) NOT NULL DEFAULT '',
    quantity      INT          NOT NULL DEFAULT 1,
    unit_price    BIGINT       NOT NULL DEFAULT 0,
    total_price   BIGINT       NOT NULL DEFAULT 0,
    confidence    DOUBLE PRECISION NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_receipt_items_receipt_id ON receipt_items (receipt_id);

CREATE TABLE IF NOT EXISTS participants (
    id            BIGSERIAL PRIMARY KEY,
    receipt_id    BIGINT       NOT NULL REFERENCES receipts (id) ON DELETE CASCADE,
    name          VARCHAR(255) NOT NULL,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_participants_receipt_id ON participants (receipt_id);

CREATE TABLE IF NOT EXISTS item_assignments (
    id               BIGSERIAL PRIMARY KEY,
    receipt_item_id  BIGINT NOT NULL REFERENCES receipt_items (id) ON DELETE CASCADE,
    participant_id   BIGINT NOT NULL REFERENCES participants (id) ON DELETE CASCADE,
    quantity         INT    NOT NULL DEFAULT 1,
    amount           BIGINT NOT NULL DEFAULT 0,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_item_assignments_receipt_item_id ON item_assignments (receipt_item_id);
CREATE INDEX IF NOT EXISTS idx_item_assignments_participant_id ON item_assignments (participant_id);
