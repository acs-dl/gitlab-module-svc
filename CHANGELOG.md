# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.4] - 2023-03-30

### Added

- Submodule to get user by id request
- Registrar routine

### Changed

- Registration in orchestrator every 10 minutes to be sure that modules is registered

## [1.0.3] - 2023-03-27

### Added

- Worker delete users and permissions that wasn't updated for some time
- `updated_at` column in users table 

## [1.0.2] - 2023-03-22

### Added

- Request to send roles by user status

## [1.0.1] - 2023-03-17

### Added

- Handling `Too Many requests` error from `API`

## [1.0.0] - 2023-03-15

### Added

- Sender.
- Receiver.
- Database.
- API handlers.

