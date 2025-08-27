-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS tokens (
	hash BLOB PRIMARY KEY,
	user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	expiry DATETIME NOT NULL,
	scope TEXT NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS tokens;
-- +goose StatementEnd