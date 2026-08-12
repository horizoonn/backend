-- +goose Up
-- +goose StatementBegin
CREATE TABLE recap.recaps
(
    id                    UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id               UUID        NOT NULL REFERENCES recap.users (id) ON DELETE CASCADE,
    year                  SMALLINT    NOT NULL,
    archetype             VARCHAR(32) NOT NULL,
    archetype_title       TEXT        NOT NULL,
    archetype_description TEXT        NOT NULL,
    archetype_reasons     JSONB       NOT NULL,
    slides                JSONB       NOT NULL,
    generated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT unique_recap_per_user_year UNIQUE (user_id, year),
    CONSTRAINT check_recaps_year CHECK (year BETWEEN 2015 AND 2100),
    CONSTRAINT check_recaps_slides_is_non_empty_array
        CHECK (
            CASE
                WHEN jsonb_typeof(slides) = 'array'
                    THEN jsonb_array_length(slides) > 0
                ELSE FALSE
            END
        ),
    CONSTRAINT check_recaps_archetype_title_not_blank
        CHECK (btrim(archetype_title) <> ''),
    CONSTRAINT check_recaps_archetype_description_not_blank
        CHECK (btrim(archetype_description) <> ''),
    CONSTRAINT check_recaps_archetype_reasons_is_non_empty_array
        CHECK (
            CASE
                WHEN jsonb_typeof(archetype_reasons) = 'array'
                    THEN jsonb_array_length(archetype_reasons) > 0
                ELSE FALSE
            END
        ),
    CONSTRAINT check_recaps_archetype CHECK (archetype IN
        ('collector', 'dealmaker', 'negotiator', 'explorer'))
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS recap.recaps;
-- +goose StatementEnd
