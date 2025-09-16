CREATE TABLE messages (
                          id BIGSERIAL PRIMARY KEY,
                          uuid UUID NOT NULL DEFAULT gen_random_uuid(),
                          user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                          content TEXT NOT NULL,
                          created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_messages_created_at ON messages(created_at);
CREATE INDEX idx_messages_user_id ON messages(user_id);