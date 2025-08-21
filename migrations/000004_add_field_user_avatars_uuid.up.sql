ALTER TABLE user_avatars
    ADD COLUMN uuid UUID NOT NULL DEFAULT gen_random_uuid();