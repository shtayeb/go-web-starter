-- +goose Up
-- +goose StatementBegin
CREATE TABLE authors (
  id   INTEGER PRIMARY KEY,
  name text    NOT NULL,
  bio  text
);
-- +goose StatementEnd