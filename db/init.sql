\c docker;

CREATE TYPE currency AS ENUM ('USD', 'EUR');

CREATE TABLE account (
  user_id serial PRIMARY KEY,
  identifier VARCHAR(36) NOT NULL,
  currency currency NOT NULL,
  amount DECIMAL DEFAULT 0,
  created_on TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE payment (
  payment_id serial PRIMARY KEY,
  from_id INTEGER NOT NULL,
  to_id INTEGER NOT NULL,
  amount DECIMAL NOT NULL,
  transaction_time_utc TIMESTAMP NOT NULL,
  currency currency NOT NULL,
  CONSTRAINT payment_from_id_fk FOREIGN KEY (from_id)
      REFERENCES account (user_id) MATCH SIMPLE
      ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT payment_to_id_fk FOREIGN KEY (to_id)
      REFERENCES account (user_id) MATCH SIMPLE
      ON UPDATE NO ACTION ON DELETE NO ACTION
);

INSERT INTO account (identifier, currency, amount) VALUES
('first', 'USD', 1000),
('second', 'USD', 0),
('third', 'EUR', 10);

GRANT ALL PRIVILEGES on TABLE account TO docker;
