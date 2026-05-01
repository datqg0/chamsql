-- +goose Up
-- +goose StatementBegin
ALTER TABLE problems
    ADD COLUMN IF NOT EXISTS source_pdf_url VARCHAR(1000);

COMMENT ON COLUMN problems.source_pdf_url IS 'MinIO URL của file PDF gốc mà bài toán được extract từ đó';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE problems DROP COLUMN IF EXISTS source_pdf_url;
-- +goose StatementEnd
