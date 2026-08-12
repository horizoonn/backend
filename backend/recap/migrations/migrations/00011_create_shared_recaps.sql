-- +goose Up
-- +goose StatementBegin
CREATE TABLE recap.shared_recaps
(
    token      VARCHAR(22) PRIMARY KEY,
    recap_id   UUID        NOT NULL
        REFERENCES recap.recaps (id) ON DELETE CASCADE,
    snapshot   JSONB       NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT unique_shared_recap_per_recap UNIQUE (recap_id),
    CONSTRAINT check_shared_recaps_token CHECK (
        char_length(token) = 22
        AND token ~ '^[A-Za-z0-9_-]{22}$'
    ),
    CONSTRAINT check_shared_recaps_snapshot CHECK (
        jsonb_typeof(snapshot) = 'object'
        AND snapshot <> '{}'::jsonb
    )
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS recap.shared_recaps;
-- +goose StatementEnd
