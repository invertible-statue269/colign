-- Drop old comments tables (no data, schema change)
DROP TABLE IF EXISTS comment_replies;
DROP TABLE IF EXISTS comments;

CREATE TABLE comments (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    change_id BIGINT NOT NULL REFERENCES changes(id) ON DELETE CASCADE,
    document_type TEXT NOT NULL,
    quoted_text TEXT NOT NULL DEFAULT '',
    body TEXT NOT NULL,
    user_id BIGINT NOT NULL REFERENCES users(id),
    resolved BOOLEAN NOT NULL DEFAULT FALSE,
    resolved_by BIGINT REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_comments_change_doc ON comments(change_id, document_type);
CREATE INDEX idx_comments_resolved_by ON comments(resolved_by) WHERE resolved_by IS NOT NULL;

CREATE TABLE comment_replies (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    comment_id BIGINT NOT NULL REFERENCES comments(id) ON DELETE CASCADE,
    body TEXT NOT NULL,
    user_id BIGINT NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_comment_replies_comment ON comment_replies(comment_id);
CREATE INDEX idx_comment_replies_user_id ON comment_replies(user_id);
