-- +goose Up
-- +goose StatementBegin
CREATE TABLE teams (
    name TEXT PRIMARY KEY
);

CREATE TABLE users (
    id        TEXT PRIMARY KEY,
    username  TEXT NOT NULL,
    team_name TEXT NOT NULL REFERENCES teams(name) ON DELETE RESTRICT,
    is_active BOOLEAN NOT NULL DEFAULT true
);

CREATE TABLE pull_requests (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    author_id   TEXT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    status      TEXT NOT NULL CHECK (status IN ('OPEN', 'MERGED')),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    merged_at   TIMESTAMPTZ
);

CREATE TABLE pull_request_reviewers (
    pr_id   TEXT NOT NULL REFERENCES pull_requests(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL REFERENCES users(id)        ON DELETE RESTRICT,
    PRIMARY KEY (pr_id, user_id)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS pull_request_reviewers;
DROP TABLE IF EXISTS pull_requests;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS teams;
-- +goose StatementEnd
