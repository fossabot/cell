ALTER TABLE users
    ALTER COLUMN password_hash TYPE BYTEA;
ALTER TABLE users
    ALTER COLUMN public_key TYPE BYTEA;