-- +migrate Up
CREATE TABLE `database_config`
(
  `id`         INTEGER PRIMARY KEY,
  `name`       TEXT NOT NULL UNIQUE,
  `dsn`        TEXT NOT NULL UNIQUE,
  `created_at` INTEGER,
  `updated_at` INTEGER,
  `deleted_at` INTEGER DEFAULT NULL
);

CREATE TABLE `doc`
(
  `id`          INTEGER PRIMARY KEY,
  `name`        TEXT NOT NULL,
  `path`        TEXT NOT NULL,
  `content`     TEXT    DEFAULT '',
  `description` TEXT    DEFAULT NULL,
  `db_name`     TEXT NOT NULL,
  `created_at`  INTEGER,
  `updated_at`  INTEGER,
  `deleted_at`  INTEGER DEFAULT NULL
);
-- +migrate Down
DROP TABLE `database_config`;
DROP TABLE `doc`;