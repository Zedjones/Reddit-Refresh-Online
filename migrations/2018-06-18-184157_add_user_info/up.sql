CREATE TABLE user_info (
  email VARCHAR PRIMARY KEY,
  interval FLOAT NOT NULL
);

CREATE TABLE devices (
  id SERIAL PRIMARY KEY,
  email VARCHAR NOT NULL,
  device_id VARCHAR NOT NULL
)