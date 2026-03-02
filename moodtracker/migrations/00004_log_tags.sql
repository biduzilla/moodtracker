-- +goose Up
-- +goose StatementBegin
CREATE TABLE log_tags (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

  log_id UUID NOT NULL REFERENCES day_logs(id) ON DELETE CASCADE,
  tag_id UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE,

  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  UNIQUE (log_id, tag_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS log_tags;
-- +goose StatementEnd
