# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed
- Updated minimum Go version to 1.22
- Removed local replace directive for independent module usage

### Added
- MIT License
- Comprehensive package documentation (doc.go)

## [0.1.0] - 2024-12-24

### Added
- Initial release
- PostgreSQL adapter implementation for toutago-datamapper
- Full CRUD operations (Create, Read, Update, Delete)
- Bulk insert support with multi-row VALUES
- RETURNING clause for auto-generated IDs (SERIAL, IDENTITY)
- Named parameter substitution ({param_name})
- PostgreSQL-specific optimizations
- Optimistic locking support
- Connection pooling configuration
- Custom SQL execution and stored procedures
- CQRS pattern support via source configuration

[Unreleased]: https://github.com/toutaio/toutago-datamapper-postgres/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/toutaio/toutago-datamapper-postgres/releases/tag/v0.1.0
