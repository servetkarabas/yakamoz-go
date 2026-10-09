ALTER TABLE authors ADD COLUMN IF NOT EXISTS role TEXT NOT NULL DEFAULT 'author';

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'authors_role_check'
    ) THEN
        ALTER TABLE authors ADD CONSTRAINT authors_role_check CHECK (role IN ('author', 'reviewer', 'admin'));
    END IF;
END $$;

CREATE UNIQUE INDEX IF NOT EXISTS authors_single_reviewer_idx ON authors (role) WHERE role = 'reviewer';
CREATE UNIQUE INDEX IF NOT EXISTS authors_single_admin_idx ON authors (role) WHERE role = 'admin';
