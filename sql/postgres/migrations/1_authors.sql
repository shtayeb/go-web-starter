-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS authors (
  id  SERIAL PRIMARY KEY,
  name text    NOT NULL,
  bio  text
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS authors;
-- +goose StatementEnd