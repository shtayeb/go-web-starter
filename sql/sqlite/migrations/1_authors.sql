-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS authors (
  id  INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL,
  bio  TEXT
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS authors;
-- +goose StatementEnd