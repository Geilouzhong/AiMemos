-- Add enable_activity_tracking column to user table
ALTER TABLE user ADD COLUMN enable_activity_tracking INTEGER NOT NULL CHECK (enable_activity_tracking IN (0, 1)) DEFAULT 0;
