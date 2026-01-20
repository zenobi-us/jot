# Changelog

## [0.0.3](https://github.com/zenobi-us/opennotes/compare/0.1.0-next.1...v0.0.3) (2026-01-20)


### ⚠ BREAKING CHANGES

* Remove Node.js dependencies and package.json

### Bug Fixes

* adjust for prerelease tag offset in version computation ([14aa55a](https://github.com/zenobi-us/opennotes/commit/14aa55af7b1e394c7985461a05ba3d1217bf4f60))
* **ci:** correct release-please outputs printing in workflow ([35b45a6](https://github.com/zenobi-us/opennotes/commit/35b45a68533bc9326ad7e82ae6d694b83d604676))
* force tag creation for release workflow ([933add0](https://github.com/zenobi-us/opennotes/commit/933add04b2300e1c1315dc3b6a6fd8ec474112e8))
* ignore coverage files ([5a9a27e](https://github.com/zenobi-us/opennotes/commit/5a9a27e91246d8bbe5f03d000e8d1159b809c0a8))
* resolve all bats test failures and security issues ([#6](https://github.com/zenobi-us/opennotes/issues/6)) ([9353f1c](https://github.com/zenobi-us/opennotes/commit/9353f1c70fe38cd8cb9759dc0b0f53be76c448f4))


### Code Refactoring

* migrate from Node.js to Go-native version management ([51846b0](https://github.com/zenobi-us/opennotes/commit/51846b0b167a00295605761b880f2f3c694b9873))

## [0.0.2](https://github.com/zenobi-us/opennotes/compare/v0.0.1...v0.0.2) (2026-01-17)


### Bug Fixes

* **publish:** fetch git tags in checkout action ([5d6af78](https://github.com/zenobi-us/opennotes/commit/5d6af785086ebf729603f28baf38badd3cb24adb))

## 0.0.1 (2026-01-17)


### Features

* **cli:** add --sql flag to notes search command ([780acdd](https://github.com/zenobi-us/opennotes/commit/780acdd9dcbc321d2fba805f0c633e54fc6abe56))
* **core:** add wiki notebook and notes management system ([c3ae87f](https://github.com/zenobi-us/opennotes/commit/c3ae87fe787ba792b1acfb44aef57101dca362bc))
* **db:** add GetReadOnlyDB() method for safe query execution ([bffdf90](https://github.com/zenobi-us/opennotes/commit/bffdf901ec691aaafa76cc1e281eaca2e4141f6a))
* **display:** add RenderSQLResults() for table formatting ([a4dcc91](https://github.com/zenobi-us/opennotes/commit/a4dcc91b02ef03855ed5144947cb786f5c6db36d))
* Go rewrite with comprehensive testing and CI/CD ([#1](https://github.com/zenobi-us/opennotes/issues/1)) ([62d21ab](https://github.com/zenobi-us/opennotes/commit/62d21abb6c746ad8b609ac6755ce0145e741ff11))
* **init:** add init command and refactor ConfigService ([e34ed99](https://github.com/zenobi-us/opennotes/commit/e34ed99275f5d39f089183847764e297ef23519d))
* **sql:** add NoteService.ExecuteSQLSafe() for query orchestration ([5b9d9e2](https://github.com/zenobi-us/opennotes/commit/5b9d9e259e58df4f86654b633335c0719423b6f7))
* **sql:** add ValidateSQL() for safe query execution ([74ce4af](https://github.com/zenobi-us/opennotes/commit/74ce4afeffb987056412fe3dcfd15e274f2accff))


### Bug Fixes

* correct GoReleaser configuration for opennotes build ([744dc50](https://github.com/zenobi-us/opennotes/commit/744dc505011a09667949e47e9c66d2395552dacb))
* **publish:** add GITHUB_TOKEN and clarify release target ([4079f6b](https://github.com/zenobi-us/opennotes/commit/4079f6be8e5460ab9ebe4a7b4419e8f460a38824))
* **publish:** create git tag before GoReleaser runs ([bf8fdbf](https://github.com/zenobi-us/opennotes/commit/bf8fdbfe4b82e9f92bb8e570f1386c0d9f9f3500))
* **publish:** use git tags instead of calculating prerelease versions ([a144cb7](https://github.com/zenobi-us/opennotes/commit/a144cb764b8ea3d5c453372130e80767aea67821))
* **types:** resolve all TypeScript type errors ([bf69925](https://github.com/zenobi-us/opennotes/commit/bf69925146ca05743ac600378643fd4c2d05ed5f))

## Changelog

All notable changes to this project will be documented here by Release Please.
