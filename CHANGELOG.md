# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project adheres
to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]


## [0.2.1] - 2024-07-22

### Added
- Search for votes

## [0.2.0] - 2024-07-05

### Added
- Feed settings details


## [0.1.0] - 2024-07-01

### Added
- Push settings details

## [0.0.85] - 2024-06-13

### Changed
- Add device uuid for push tokens requests

## [0.0.84] - 2024-04-19

### Added
- Sorting recommendations by popularity index

## [0.0.83] - 2024-04-11

### Added
- Achievements handlers

## [0.0.82] - 2024-04-10

### Added
- Participated daos

## [0.0.81] - 2024-04-08

### Added
- Added tool endpoint for getting voting power for specific users

## [0.0.80] - 2024-03-31

### Fixed
- Skip votes if the proposal is canceled or spam

## [0.0.79] - 2024-03-29

### Fixed
- Proposal list request with huge limit

## [0.0.78] - 2024-03-28

### Fixed
- Remove duplicate dao ids

## [0.0.77] - 2024-03-27

### Added
- Monthly filter for analytics endpoints

## [0.0.76] - 2024-03-27

### Fixed
- Public profile endpoint to get the information from core
- Mutual dao. Skip incorrect dao id.

## [0.0.75] - 2024-03-25

### Added
- Public profile endpoints

## [0.0.74] - 2024-03-22

### Added
- Stats endpoint

## [0.0.73] - 2024-03-22

### Fixed
- Fixed empty result for featured proposals

## [0.0.72] - 2024-03-21

### Added
- Recommended DAO endpoint

## [0.0.71] - 2024-03-15

### Added
- Featured proposals endpoint

## [0.0.70] - 2024-03-15

### Added
- Total Vp for proposal votes

## [0.0.69] - 2024-03-13

### Added
- User field for top voter endpoint

### Changed
- core-web-sdk dependency name

## [0.0.68] - 2024-03-06

### Added
- Dao feed caching
- Nats metrics

## [0.0.67] - 2024-02-22

### Changed
- Use proposals from core storage instead of feed snapshot

## [0.0.66] - 2024-02-19

### Added
- Spam count for monthly new proposals

## [0.0.65] - 2024-02-19

### Fixed
- Fixed last activity at for me endpoint

## [0.0.64] - 2024-02-18

### Added
- Totals for Vp Avg

## [0.0.63] - 2024-02-16

### Fixed
- Filter only active proposals for user can vote

## [0.0.62] - 2024-02-15

### Fixed
- Proposal score calculation

## [0.0.61] - 2024-02-14

### Added
- Return empty votes list for guest

## [0.0.60] - 2024-02-13

### Added
- Added total and unread counters for mark-as... endpoints

### Fixed
- Fixed batch unread endoint
- Fixed linter warnings

## [0.0.59] - 2024-02-11

### Fixed
- Fixed empty proposals in me can vote endpoint 

## [0.0.58] - 2024-02-11

### Fixed
- Fixed inbox-api dependency

### Added
- Added mark-as-unread and mark-as-unread bath endpoints

## [0.0.57] - 2024-02-08

### Changed
- Generate avatars based on ens name instead of address

## [0.0.56] - 2024-02-
## Added
- New fields for dao

## [0.0.55] - 2024-02-06

## Added
- User votes to feed
- User vote first
- Empty response if the user hasn't votes

## [0.0.54] - 2024-02-01

### Added
- User votes
- dao.terms to proposal

## [0.0.53] - 2024-01-31

### Added
- Added user activity
- Added siwe replay protection 

## [0.0.52] - 2024-01-29

### Added
- Send custom pushes
- Endpoint for mark pushes as clicked

## [0.0.51] - 2024-01-17

### Added
- Voter buckets V2

## [0.0.50] - 2023-12-29

### Added
- Added ens to user profile

## [0.0.49] - 2023-12-20

### Added
- Added siwe endpoint and user/sessions control endpoinds 
- Added user sessions, roles
- Support new user api 

## [0.0.48] - 2023-12-19

### Fixed
- DAO avatars

## [0.0.47] - 2023-12-19

### Added
- DAO avatars

## [0.0.46] - 2023-12-14

### Added
- Author ens name field for votes
- User avatars for proposals and voters

## [0.0.45] - 2023-12-06

### Added
- Author ens name field for proposals

## [0.0.44] - 2023-12-04

### Added
- Added voting methods

## [0.0.43] - 2023-11-12

### Added
- Ecosystem charts

## [0.0.42] - 2023-11-08

### Changed
- Update ipfs resolve url

## [0.0.41] - 2023-11-07

### Changed
- Recently viewed dao from 10 to 30

## [0.0.40] - 2023-11-07

### Changed
- Return empty subscriptions instead of null
- Increased images size of ipfs avatars 180px instead of 90px

### Added 
- Subscription info for mutual daos

## [0.0.39] - 2023-11-03

### Changed
- Return dao object vs dao id for mutual daos

## [0.0.38] - 2023-11-02

### Changed
- Exclusive voters response

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
