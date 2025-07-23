-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS authors (
  id   INTEGER PRIMARY KEY,
  name text    NOT NULL,
  bio  text
);
-- +goose StatementEnd