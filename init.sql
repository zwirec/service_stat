CREATE TYPE ACTION AS ENUM ('login', 'logout', 'like', 'commentary');

CREATE TYPE SEX AS ENUM ('M', 'F');

CREATE TABLE IF NOT EXISTS users
(
  id  SERIAL NOT NULL
    CONSTRAINT table_name_pkey
    PRIMARY KEY,
  age INTEGER,
  sex SEX
);

CREATE UNIQUE INDEX IF NOT EXISTS table_name_id_uindex
  ON users (id);

CREATE INDEX IF NOT EXISTS users_id_age_sex_idx
  ON users (id, age, sex);


CREATE UNIQUE INDEX IF NOT EXISTS table_name_id_uindex
  ON users (id);

CREATE INDEX IF NOT EXISTS users_id_age_sex_idx
  ON users (id, age, sex);

CREATE TABLE IF NOT EXISTS stats
(
  "user" INTEGER
    CONSTRAINT stats_users_id_fk
    REFERENCES users,
  action ACTION,
  date   DATE,
  cnt    INTEGER DEFAULT 1,
  CONSTRAINT user_time_uniq
  UNIQUE (user, date)
);

CREATE INDEX stats_time_idx
  ON stats (date);

CREATE INDEX IF NOT EXISTS stats_time_idx
  ON stats (time);