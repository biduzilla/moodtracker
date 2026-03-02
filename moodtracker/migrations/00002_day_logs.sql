-- +goose Up
-- +goose StatementBegin
CREATE TABLE day_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    date DATE NOT NULL,
    description TEXT,
    mood_label SMALLINT NOT NULL,

    user_id UUID NOT NULL REFERENCES users(id),
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ,
    deleted BOOLEAN NOT NULL DEFAULT false,

    created_by UUID REFERENCES users(id),
    updated_by UUID REFERENCES users(id),

    version INT NOT NULL DEFAULT 1
);

CREATE UNIQUE INDEX uniq_day_logs_user_date
ON day_logs (user_id, date)
WHERE deleted = false;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS day_logs;
-- +goose StatementEnd
