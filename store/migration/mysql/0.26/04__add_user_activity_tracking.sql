-- Add enable_activity_tracking column to user table
ALTER TABLE `user` ADD COLUMN `enable_activity_tracking` BOOLEAN NOT NULL DEFAULT FALSE;
