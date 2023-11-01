# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project adheres
to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.0.37] - 2023-11-01

### Added
- Recent dao endpoint
- Add viewed dao to inbox-storage

## [0.0.36] - 2023-10-31

### Added
- Top voters by avg vp
- Daos where voters also participate

### Changed
- Percent succeeded proposals response

## [0.0.35] - 2023-10-23

### Added
- Percent succeeded proposals

## [0.0.34] - 2023-10-20

### Added
- Added replace ipfs links in proposal body

## [0.0.33] - 2023-10-18

### Added
- Exclusive voters, proposals by month analytics endpoints

## [0.0.32] - 2023-10-16

### Fixed
- Fixed ipfs wrapper for DAO avatars

## [0.0.31] - 2023-10-12

### Changed
- Protect user subscriptions storage

## [0.0.30] - 2023-10-12

### Added
- Added option for fetching archived items only
- Added endpoint for unarchive item

## [0.0.29] - 2023-10-09

### Fixed
- Filling proposals count from dao info

### Added
- Voters count field for dao models

## [0.0.27] - 2023-10-06

### Changed
- Add proposal ends soon mapping

## [0.0.25] - 2023-09-12

### Changed
- Mark votes choice field as json.RawMessage due to multiple values

## [0.0.24] - 2023-09-08

### Added
- Add analytics endpoint

## [0.0.23] - 2023-09-04

### Added
- Proposal timeline field

## [0.0.22] - 2023-08-26

### Changed
- Add basic cache implementation for dao objects

## [0.0.21] - 2023-07-18

### Changed
- Extend vote model

## [0.0.20] - 2023-07-17

### Fixed
- Fixed fetching feed for unread elements by default

## [0.0.19] - 2023-07-17

### Fixed
- Fixed proposal.quorum field for proposals without defined quorum

## [0.0.18] - 2023-07-17

### Fixed
- Fixed proposal.quorum field calculation (temporary solution)
- Fixed getting dao info in the dao feed

### Removed
- Removed unnecessary mock data

## [0.0.17] - 2023-07-15

### Added
- Added inbox feed endpoints (getting feed, mark as read, mark as archived etc)

### Fixed
- Fixed feed timeline structure

## [0.0.16] - 2023-07-15

### Fixed
- Fixed ipfs links in dao feed

## [0.0.15] - 2023-07-15

### Fixed
- Fixed again structure for dao flat feed (the same as mocks)

## [0.0.14] - 2023-07-14

### Fixed
- Fixed structure for dao flat feed (the same as mocks)

## [0.0.13] - 2023-07-14

### Fixed
- Supported feed.timeline field

## [0.0.12] - 2023-07-14

### Fixed
- Fixed activity since format

## [0.0.11] - 2023-07-14

### Fixed
- Updated core-web-sdk to v0.0.9

## [0.0.10] - 2023-07-14

### Added
- DAO activity since field

## [0.0.9] - 2023-07-13

### Fixed
- Fixed notifications settings endpoint as in mock version

## [0.0.8] - 2023-07-13

### Added
- Subscriptions count for authenticated dao top request

### Fixed
- Fixed avatars links in dao/top endpoint

## [0.0.7] - 2023-07-12

### Added
- Getting user session in middleware for any authed request

## [0.0.6] - 2023-07-12

### Changed
- Changed title parameter for filtering proposals to query

### Fixed
- Fixed getting subscriptions list
- Fixed marshaling common.Time if real time element is nil

## [0.0.5] - 2023-07-11

### Fixed
- Fixed missed fields
- Updated core-web-sdk dependency to v0.0.5

## [0.0.4] - 2023-07-11

### Added
- Proposal top endpoint

### Fixed
- Fixed Dockerfile

### Added
- Add settings routes

## [0.0.3] - 2023-07-07

### Added
- Filtering proposals by title

## [0.0.2] - 2023-07-07

### Added
- Flat feed by DAO

## [0.0.1] - 2023-07-03

### Added
- Added skeleton app
- Added daos handlers
- Added proposals handlers
