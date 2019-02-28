-- CREATE USER docker;
-- CREATE DATABASE docker;
-- GRANT ALL PRIVILEGES ON DATABASE docker TO docker;

\c docker;

CREATE TABLE account (
  user_id serial PRIMARY KEY,
  identifier VARCHAR(36) NOT NULL,
  created_on TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO account (identifier) VALUES
('first'),
('second'),
('third');

GRANT ALL PRIVILEGES on TABLE account TO docker;
