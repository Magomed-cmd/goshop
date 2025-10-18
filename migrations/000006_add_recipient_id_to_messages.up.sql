ALTER TABLE messages
ADD COLUMN recipient_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE;
CREATE INDEX idx_messages_recipient_id ON messages(recipient_id);