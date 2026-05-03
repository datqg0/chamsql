-- +goose Up
-- +goose StatementBegin
ALTER TABLE exam_participants ADD COLUMN updated_at TIMESTAMPTZ DEFAULT NOW();
ALTER TABLE exam_submissions ADD COLUMN updated_at TIMESTAMPTZ DEFAULT NOW();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE exam_participants DROP COLUMN IF EXISTS updated_at;
ALTER TABLE exam_submissions DROP COLUMN IF EXISTS updated_at;
-- +goose StatementEnd
