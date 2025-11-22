-- Add soft delete support for hotels and rooms tables
-- Migration: 003_add_soft_delete.sql

-- Add deleted_at column to hotels table
ALTER TABLE hotels 
ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ DEFAULT NULL;

-- Add index for soft delete queries on hotels
CREATE INDEX IF NOT EXISTS idx_hotels_deleted_at ON hotels(deleted_at) WHERE deleted_at IS NULL;

-- Add deleted_at column to rooms table
ALTER TABLE rooms 
ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ DEFAULT NULL;

-- Add index for soft delete queries on rooms
CREATE INDEX IF NOT EXISTS idx_rooms_deleted_at ON rooms(deleted_at) WHERE deleted_at IS NULL;

-- Add created_at column to rooms table (for audit trail)
ALTER TABLE rooms 
ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ DEFAULT now();
