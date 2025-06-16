CREATE TABLE IF NOT EXISTS users(
id uuid PRIMARY KEY DEFAULT gen_random_uuid(), 
username varchar(255) NOT NULL, 
role varchar(255) NOT NULL DEFAULT 'user',
email varchar(255) NOT NULl, password text NOT NULl);