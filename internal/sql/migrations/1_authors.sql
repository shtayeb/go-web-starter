-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS authors (
  id  SERIAL PRIMARY KEY,
  name text    NOT NULL,
  bio  text
);
-- +goose StatementEnd