-- +migrate Up
CREATE TABLE `database_config`
(
  `id`         int(11)     NOT NULL AUTO_INCREMENT,
  `uuid`       varchar(50) NOT NULL,
  `name`       varchar(60)  DEFAULT NULL,
  `dsn`        varchar(200) DEFAULT NULL,
  `created_at` datetime     DEFAULT NULL,
  `updated_at` datetime     DEFAULT NULL,
  `deleted_at` datetime     DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `dnname` (`name`) USING BTREE
) ENGINE = InnoDB
  AUTO_INCREMENT = 6
  DEFAULT CHARSET = utf8mb4;

CREATE TABLE `doc`
(
  `id`          int(11)     NOT NULL AUTO_INCREMENT,
  `uuid`        varchar(50) NOT NULL,
  `name`        varchar(50)  DEFAULT NULL,
  `path`        varchar(100) DEFAULT NULL,
  `content`     text,
  `description` varchar(500) DEFAULT NULL,
  `db_name`     varchar(60)  DEFAULT NULL,
  `created_at`  datetime     DEFAULT NULL,
  `updated_at`  datetime     DEFAULT NULL,
  `deleted_at`  datetime     DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `path` (`path`) USING BTREE
) ENGINE = InnoDB
  AUTO_INCREMENT = 29
  DEFAULT CHARSET = utf8mb4;
-- +migrate Down
DROP TABLE IF EXISTS `doc`;
DROP TABLE IF EXISTS `database_config`;