-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd
CREATE TABLE `messages` (
    `id` integer PRIMARY KEY AUTOINCREMENT,
    `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `package_id` integer NOT NULL,
    `name` text NOT NULL,
    `proto_body` text NOT NULL,
    CONSTRAINT `fk_messages_package` FOREIGN KEY (`package_id`) REFERENCES `packages`(`id`) ON DELETE CASCADE,
    UNIQUE (`package_id`, `name`)
);
CREATE TABLE `message_versions` (
    `id` integer PRIMARY KEY AUTOINCREMENT,
    `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `message_id` integer NOT NULL,
    `package_version_id` integer NOT NULL,
    `version` integer NOT NULL,
    `proto_body` text NOT NULL,
    `serialized_schema` text NOT NULL,

     CONSTRAINT `fk_package_versions_message_versions` FOREIGN KEY (`package_version_id`) REFERENCES `package_versions`(`id`) ON DELETE CASCADE,
     UNIQUE (`message_id`, `version`)
);

-- Foreign key for latest_version_id
ALTER TABLE `messages`
ADD COLUMN `latest_version_id` integer REFERENCES message_versions (id) ON DELETE SET NULL;

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
DROP TABLE `message_versions`;
DROP TABLE `messages`;
