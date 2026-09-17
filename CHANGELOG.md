# Changelog

## [1.1.2](https://github.com/twangodev/lfm-api/compare/v1.1.1...v1.1.2) (2026-09-16)


### Bug Fixes

* handle Fastly proof-of-work challenges when fetching scrobbles ([5f58c8f](https://github.com/twangodev/lfm-api/commit/5f58c8f86b7c415e684e077623c36b1d01366d9d))

## [1.1.1](https://github.com/twangodev/lfm-api/releases/tag/v1.1.1) (2026-07-15)

### What's Changed

#### Bug Fixes

* `GetActiveScrobble` now returns an error on non-200 responses instead of silently returning an empty scrobble. Last.fm's Varnish layer intermittently serves HTTP 600 "Temporarily Unavailable", which was previously indistinguishable from "user is not scrobbling" — causing consumers like [lfm-cli](https://github.com/twangodev/lfm-cli) to tear down Discord presence mid-track. Callers can now detect transient failures and retain state.

#### Dependencies & Docs

* Bump golang.org/x/net from 0.29.0 to 0.32.0
* Bump mkdocs-open-in-new-tab from 1.0.3 to 1.0.8
* Bump mkdocs-git-committers-plugin-2 from 2.3.0 to 2.4.1
* Bump mkdocs-git-revision-date-localized-plugin from 1.2.5 to 1.2.8
* Bump mkdocs-rss-plugin from 1.12.2 to 1.15.0
* Bump urllib3 from 2.2.1 to 2.2.2
* Docs site analytics updates

**Full Changelog**: https://github.com/twangodev/lfm-api/compare/v1.1.0...v1.1.1

## [1.1.0](https://github.com/twangodev/lfm-api/releases/tag/v1.1.0) (2024-07-23)

### What's Changed

* Fixes go-http-client compatibility issue
* Add CI/CD workflows

**Full Changelog**: https://github.com/twangodev/lfm-api/compare/v1.0.4...v1.1.0

## [1.0.4](https://github.com/twangodev/lfm-api/releases/tag/v1.0.4) (2024-05-22)

### What's Changed

* Bump golang.org/x/net from 0.1.0 to 0.7.0 by @dependabot in https://github.com/twangodev/lfm-api/pull/2
* Create dependabot.yml by @twangodev in https://github.com/twangodev/lfm-api/pull/3
* Bump golang.org/x/net from 0.7.0 to 0.14.0 by @dependabot in https://github.com/twangodev/lfm-api/pull/4
* Bump github.com/bozd4g/go-http-client from 0.1.4 to 1.0.2 by @dependabot in https://github.com/twangodev/lfm-api/pull/5
* Bump golang.org/x/net from 0.14.0 to 0.15.0 by @dependabot in https://github.com/twangodev/lfm-api/pull/6
* Bump golang.org/x/net from 0.15.0 to 0.16.0 by @dependabot in https://github.com/twangodev/lfm-api/pull/7
* Bump golang.org/x/net from 0.16.0 to 0.17.0 by @dependabot in https://github.com/twangodev/lfm-api/pull/8
* Bump golang.org/x/net from 0.17.0 to 0.18.0 by @dependabot in https://github.com/twangodev/lfm-api/pull/9
* Bump golang.org/x/net from 0.18.0 to 0.19.0 by @dependabot in https://github.com/twangodev/lfm-api/pull/10
* Bump golang.org/x/net from 0.19.0 to 0.24.0 by @dependabot in https://github.com/twangodev/lfm-api/pull/14
* Bump jinja2 from 3.1.3 to 3.1.4 by @dependabot in https://github.com/twangodev/lfm-api/pull/16
* Update README.md by @twangodev in https://github.com/twangodev/lfm-api/pull/21
* Bump requests from 2.31.0 to 2.32.0 by @dependabot in https://github.com/twangodev/lfm-api/pull/20
* Bump mkdocs-material from 9.5.20 to 9.5.23 by @dependabot in https://github.com/twangodev/lfm-api/pull/19
* Bump golang.org/x/net from 0.24.0 to 0.25.0 by @dependabot in https://github.com/twangodev/lfm-api/pull/18

### New Contributors

* @dependabot made their first contribution in https://github.com/twangodev/lfm-api/pull/2

**Full Changelog**: https://github.com/twangodev/lfm-api/compare/v1.0.3...v1.0.4

## [1.0.3](https://github.com/twangodev/lfm-api/releases/tag/v1.0.3) (2022-11-07)

**Full Changelog**: https://github.com/twangodev/lfm-api/compare/v1.0.2...v1.0.3

## [1.0.2](https://github.com/twangodev/lfm-api/releases/tag/v1.0.2) (2022-11-06)

**Full Changelog**: https://github.com/twangodev/lfm-api/compare/v1.0.1...v1.0.2

Moves repository from lastfm-discordrpc to twangodev

## [1.0.1](https://github.com/twangodev/lfm-api/releases/tag/v1.0.1) (2022-11-04)

### What's Changed

* Merge v1.0.0 to main by @twangodev in https://github.com/lastfm-discordrpc/lfm-api/pull/1

### New Contributors

* @twangodev made their first contribution in https://github.com/lastfm-discordrpc/lfm-api/pull/1

**Full Changelog**: https://github.com/lastfm-discordrpc/lfm-api/compare/v1.0.0...v1.0.1

## [1.0.0](https://github.com/twangodev/lfm-api/releases/tag/v1.0.0) (2022-10-26)

**Full Changelog**: https://github.com/lastfm-discordrpc/lfm-api/commits/v1.0.0
