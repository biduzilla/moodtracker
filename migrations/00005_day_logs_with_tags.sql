-- +goose Up
-- +goose StatementBegin
CREATE VIEW day_logs_with_tags AS
SELECT
    dl.*,
    COALESCE(tg.tags, '{}') AS tags
FROM day_logs dl
LEFT JOIN (
    SELECT
        lt.log_id,
        ARRAY_AGG(t.name) AS tags
    FROM log_tags lt
    JOIN tags t ON t.id = lt.tag_id
    GROUP BY lt.log_id
) tg ON tg.log_id = dl.id
WHERE dl.deleted = false;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP VIEW IF EXISTS day_logs_with_tags;
-- +goose StatementEnd
