-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

-- packages table
CREATE TABLE `packages` (
    `id` integer PRIMARY KEY AUTOINCREMENT,
    `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `package_name` text NOT NULL,
    UNIQUE (`package_name`)
);

-- package_versions table
CREATE TABLE `package_versions` (
    `id` integer PRIMARY KEY AUTOINCREMENT,
    `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `package_id` integer NOT NULL,
    `version` integer NOT NULL,
    CONSTRAINT `fk_packages_versions` FOREIGN KEY (`package_id`) REFERENCES `packages`(`id`) ON DELETE CASCADE,
    UNIQUE (`package_id`, `version`)
);

-- Foreign key for latest_version_id
ALTER TABLE `packages`
    ADD COLUMN `latest_version_id` integer REFERENCES package_versions (id) ON DELETE SET NULL;

CREATE INDEX `idx_packages_latest_version_id` ON `packages`(`latest_version_id`);


CREATE TABLE `package_version_files` (
    `id` integer PRIMARY KEY AUTOINCREMENT,
    `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `package_version_id` integer NOT NULL,
    `file_name` text NOT NULL,
    `file_contents` text NOT NULL,
    CONSTRAINT `fk_package_versions_files` FOREIGN KEY (`package_version_id`) REFERENCES `package_versions`(`id`) ON DELETE CASCADE,
    UNIQUE (`package_version_id`, `file_name`)
);

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
DROP TABLE `package_version_files`;
DROP TABLE `package_versions`;
DROP TABLE `packages`;
