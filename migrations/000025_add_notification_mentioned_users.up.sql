ALTER TABLE notifications ADD COLUMN mentioned_user_ids bigint[] DEFAULT NULL;
