# Changelog

## [dev]

- Complete rewrite to fix corrupted CSV data
- Use `sync.Mutex` to prevent race conditions with interface syncing
- Breaking change: `-sslcert` & `-sslkey` is now `--sslcert` & `--sslkey`
- Add `version` and `update` sub-commands


## [0.2.5]

- Switch to go mods
- Remove duplicate help section
- Switch to axllent/ghru for updating
- Update modules & Makefile


## [0.2.4]

- Save stats on SIGTERM
- Save stats so they are current on refresh


## [0.2.3]

- Switch to packr


## [0.2.2]

- HTTPS support


## [0.2.1]

- Auto-detect websocket protocol


## [0.2.0]

- Gzip static http response
- Add basic auth


## [0.1.0]

- Switch to CSV database format for portability
- Remove Sqlite
- `-d` now implies database directory


## [0.0.3]

- Update help, no default network interface


## [0.0.2]

- Add updater


## [0.0.1]

- First release
