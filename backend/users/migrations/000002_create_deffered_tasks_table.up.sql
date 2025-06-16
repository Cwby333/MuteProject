CREATE TABLE IF NOT EXISTS deffered_tasks(
id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
topic varchar(255) NOT NULL, data JSON NOT NULL, 
created_at timestamp DEFAULT now());