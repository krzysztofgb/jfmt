# Changelog

## [0.2.2](https://github.com/krzysztofgb/jfmt/compare/v0.2.1...v0.2.2) (2026-04-24)


### Bug Fixes

* **brew:** pre-generate completions to avoid executing quarantined binary ([f74c135](https://github.com/krzysztofgb/jfmt/commit/f74c1356852889e502a92c1564e3c6e3a1649226))

## [0.2.1](https://github.com/krzysztofgb/jfmt/compare/v0.2.0...v0.2.1) (2026-04-24)


### Bug Fixes

* **brew:** remove quarantine before completion generation ([966c680](https://github.com/krzysztofgb/jfmt/commit/966c680aa3ee755a873c82b55feefce6eab4d5a3))

## [0.2.0](https://github.com/krzysztofgb/jfmt/compare/v0.1.0...v0.2.0) (2026-04-24)


### Features

* add shell completions support ([560df3d](https://github.com/krzysztofgb/jfmt/commit/560df3da80cca0a1aaab15809f79b6802608c1cf))
* **brew:** add post-install caveats for shell completions ([2fa073a](https://github.com/krzysztofgb/jfmt/commit/2fa073aa35e50df367898c4012f69bd3e9e40dc5))


### Bug Fixes

* scope CI badge to main branch ([6c52123](https://github.com/krzysztofgb/jfmt/commit/6c521231569ec054a3f137b997164386695cc34b))

## [0.1.0](https://github.com/krzysztofgb/jfmt/compare/v0.0.6...v0.1.0) (2026-04-24)


### Features

* add --check and --diff flags ([07499d6](https://github.com/krzysztofgb/jfmt/commit/07499d6a96d78b36c0d8f37f4a0b71a3722f9886))
* add --config flag for explicit config file path ([ac13a5c](https://github.com/krzysztofgb/jfmt/commit/ac13a5cb05c38e47677d4b936eaa6a2727698c66))
* add --quiet/-q flag to suppress non-error output ([9ddc357](https://github.com/krzysztofgb/jfmt/commit/9ddc357e0baee6fea351ac44950421186ff30f9e))
* add --recursive flag for directory traversal ([f397aa1](https://github.com/krzysztofgb/jfmt/commit/f397aa1658c0429e5a91747dd59e035fd5ebf5bf))
* add --stdin-filename flag for error message context ([466ca0e](https://github.com/krzysztofgb/jfmt/commit/466ca0ed383f65b09a9c5e23e9c2e91e85bc4226))
* add FormatReader/FixReader, --print-config, and exit code 2 for --check ([8c350e2](https://github.com/krzysztofgb/jfmt/commit/8c350e244f83fa9e54edc857ae6b1509a9aebd30))
* add FormatString, ValidateString, FixString convenience functions ([68a78bf](https://github.com/krzysztofgb/jfmt/commit/68a78bf4dd7e25ef62639c1fd43d5dbfa2b74d5c))
* add TOML config file support with --no-config flag ([64197cb](https://github.com/krzysztofgb/jfmt/commit/64197cbaa079c0210493688613f38d4bafb45834))
* group flags by category in help output ([0436ea2](https://github.com/krzysztofgb/jfmt/commit/0436ea2628062fad6456a7b2218268f74d7bcaab))
* **lint:** add gocritic linter ([42f4230](https://github.com/krzysztofgb/jfmt/commit/42f4230ce2774c61ad0e6b37ac3e6041b2e6dc74))
* migrate CLI to cobra for shell completion support ([83ca430](https://github.com/krzysztofgb/jfmt/commit/83ca4303f0ec6fce98c6d75b862267cf4774f3bd))
* show help when invoked interactively with no arguments ([76086db](https://github.com/krzysztofgb/jfmt/commit/76086db24a5cd4d51d784bb1755ab7983bf5878c))


### Bug Fixes

* make Fix idempotent for quoted keys and consecutive trailing commas ([1cc59c8](https://github.com/krzysztofgb/jfmt/commit/1cc59c8d083903a050c7ce5bdc3e0d56ce873ed8))
* **test:** use filepath.Join for XDG config path assertion on Windows ([4bf49c9](https://github.com/krzysztofgb/jfmt/commit/4bf49c9a53745ca63f23fce4f1ac4507780fa976))
