-- +goose Up
-- +goose StatementBegin
CREATE TABLE tags (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    name TEXT NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id),

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ,
    deleted BOOLEAN NOT NULL DEFAULT false,

    created_by UUID REFERENCES users(id),
    updated_by UUID REFERENCES users(id),

    version INT NOT NULL DEFAULT 1
);

-- nome da tag é único por usuário (case-insensitive, não deletado)
CREATE UNIQUE INDEX uniq_tags_name_user_not_deleted
ON tags (LOWER(name), user_id)
WHERE deleted = false;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS tags;
-- +goose StatementEnd
