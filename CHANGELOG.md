# Changelog

All notable changes to this project will be documented in this file.

## [0.6.1](https://github.com/inference-gateway/a2a-debugger/compare/v0.6.0...v0.6.1) (2025-10-03)

### ♻️ Improvements

* **tasks:** Show all Task attributes in list command ([#13](https://github.com/inference-gateway/a2a-debugger/issues/13)) ([632c930](https://github.com/inference-gateway/a2a-debugger/commit/632c930c26907129e32493f95830a83ad593c408)), closes [#12](https://github.com/inference-gateway/a2a-debugger/issues/12)

## [0.6.0](https://github.com/inference-gateway/a2a-debugger/compare/v0.5.2...v0.6.0) (2025-09-07)

### ✨ Features

* Add task resumption support to submit commands ([#9](https://github.com/inference-gateway/a2a-debugger/issues/9)) ([ccd9cec](https://github.com/inference-gateway/a2a-debugger/commit/ccd9cec204f28b86d30847ad4eff7ea243c7f41b)), closes [#8](https://github.com/inference-gateway/a2a-debugger/issues/8)
* **streaming:** Add comprehensive Task ID and summary display at end of streaming ([#7](https://github.com/inference-gateway/a2a-debugger/issues/7)) ([069cd58](https://github.com/inference-gateway/a2a-debugger/commit/069cd5844a353ae784955ea8fcde28615f6e7271)), closes [#6](https://github.com/inference-gateway/a2a-debugger/issues/6)

## [0.5.2](https://github.com/inference-gateway/a2a-debugger/compare/v0.5.1...v0.5.2) (2025-08-28)

### ♻️ Improvements

* Improve agent response display with message parts when streaming tasks ([41dfd28](https://github.com/inference-gateway/a2a-debugger/commit/41dfd2804198f9ac0d15284bb83f2a9005523905))

### 🐛 Bug Fixes

* Correct environment variable names in docker-compose.yml ([f9721a1](https://github.com/inference-gateway/a2a-debugger/commit/f9721a19af57ad356bee9316f3e6534ccd346b4e))

### 🔧 Miscellaneous

* Add initial configuration files for flox environment setup ([edede96](https://github.com/inference-gateway/a2a-debugger/commit/edede9677c48580f81dd4cb51f8cb7c4985c087b))

## [0.5.1](https://github.com/inference-gateway/a2a-debugger/compare/v0.5.0...v0.5.1) (2025-08-04)

### 🐛 Bug Fixes

* Display all message part types including tool calls ([#3](https://github.com/inference-gateway/a2a-debugger/issues/3)) ([54dc546](https://github.com/inference-gateway/a2a-debugger/commit/54dc546591846d51a09029671c26ffaeccf1d519)), closes [#2](https://github.com/inference-gateway/a2a-debugger/issues/2)

### 👷 CI

* Add Claude Code GitHub Workflow ([#1](https://github.com/inference-gateway/a2a-debugger/issues/1)) ([741230b](https://github.com/inference-gateway/a2a-debugger/commit/741230bd228efd486b1fc62dcabf5bb804dcc881))

### 📚 Documentation

* Add CLAUDE.md for project documentation and development guidelines ([caa0e90](https://github.com/inference-gateway/a2a-debugger/commit/caa0e90242ce3e784386ac3904770d7cb275bd60))

### 🔧 Miscellaneous

* Add issue templates for bug reports, feature requests, and refactor requests ([ae732b9](https://github.com/inference-gateway/a2a-debugger/commit/ae732b9a31aeca8dbaeab4c4187418ae7b2d6ad1))

### 🔨 Miscellaneous

* Install Claude code in Dockerfile ([13a2bbc](https://github.com/inference-gateway/a2a-debugger/commit/13a2bbc4c50695af878da018f950010ac275b8dd))

## [0.5.0](https://github.com/inference-gateway/a2a-debugger/compare/v0.4.0...v0.5.0) (2025-06-19)

### ✨ Features

* Add submit-streaming command to A2A CLI and update README with usage examples ([62e4614](https://github.com/inference-gateway/a2a-debugger/commit/62e461493e16f36e3a698cc9ecdaffc800d14825))

### 🔧 Miscellaneous

* Update README.md ([4700b7d](https://github.com/inference-gateway/a2a-debugger/commit/4700b7d12e03a81df5bc597eb23aefaa023b172a))

## [0.4.0](https://github.com/inference-gateway/a2a-debugger/compare/v0.3.1...v0.4.0) (2025-06-17)

### ✨ Features

* Add submit task command to A2A CLI and update README with usage examples ([f5ef0de](https://github.com/inference-gateway/a2a-debugger/commit/f5ef0dead022ff2f2d6400925f9a6598c8661ff9))

### ♻️ Improvements

* Update a2a-debugger service to use build context instead of image ([0c3732c](https://github.com/inference-gateway/a2a-debugger/commit/0c3732c650c5af27f417e3a01b72e8daa0cf128a))

### 🐛 Bug Fixes

* Correct build output path and entrypoint in Dockerfile ([bdb6320](https://github.com/inference-gateway/a2a-debugger/commit/bdb63203dce56a7879bffb00f4adc557b299285a))
* Improve error handling for MethodNotFoundError in A2A client commands ([cc16b6e](https://github.com/inference-gateway/a2a-debugger/commit/cc16b6efcf10716d25189512e7eebc9b82594305))

### 📚 Documentation

* Remove unnecessary config flag from README and update docker-compose entrypoint ([1099c8f](https://github.com/inference-gateway/a2a-debugger/commit/1099c8f678f437e2d179ade7e8c1d5c0668f5c2f))

## [0.3.1](https://github.com/inference-gateway/a2a-debugger/compare/v0.3.0...v0.3.1) (2025-06-17)

### ♻️ Improvements

* Change log level from Info to Debug for A2A client initialization and connection tests ([a15ad42](https://github.com/inference-gateway/a2a-debugger/commit/a15ad42c9aa01a53e3d3fedb7e307df98c79189c))

### 📚 Documentation

* Add example README, configuration file, and docker-compose for A2A Debugger setup ([9b535a7](https://github.com/inference-gateway/a2a-debugger/commit/9b535a7f888ed0b720dd66392a8f43278ba63150))
* Update example server URL in README to localhost ([f3ce233](https://github.com/inference-gateway/a2a-debugger/commit/f3ce23308d2de9ce16f56d04c41ea3d66a85ee75))

### 🔧 Miscellaneous

* Update TASK_VERSION to v3.44.0 ([813bd4e](https://github.com/inference-gateway/a2a-debugger/commit/813bd4e8c045af6cfa9b515cb0d63f3cc5313738))

## [0.3.0](https://github.com/inference-gateway/a2a-debugger/compare/v0.2.2...v0.3.0) (2025-06-17)

### ✨ Features

* Add version command and display version information in A2A Debugger ([6146892](https://github.com/inference-gateway/a2a-debugger/commit/614689232c9e8beef22067bb25461104d913521c))

## [0.2.2](https://github.com/inference-gateway/a2a-debugger/compare/v0.2.1...v0.2.2) (2025-06-17)

### ♻️ Improvements

* Use namespace based commands - all tasks related actions under tasks and all config related actions under a2a config ([afb3088](https://github.com/inference-gateway/a2a-debugger/commit/afb3088905177fda9fdcd591b42ec3f408e0a0f8))

### 📚 Documentation

* Add MIT license badge to README ([54e188e](https://github.com/inference-gateway/a2a-debugger/commit/54e188ee9c30d96ec7675526538c96e3f919ef65))

## [0.2.1](https://github.com/inference-gateway/a2a-debugger/compare/v0.2.0...v0.2.1) (2025-06-17)

### 🐛 Bug Fixes

* Update artifact upload script and improve formatting in workflow ([32f83d8](https://github.com/inference-gateway/a2a-debugger/commit/32f83d8159f96655534f46b312388c238a6308df))

### 🔧 Miscellaneous

* Update .gitattributes to mark generated and vendored files ([7ade7d3](https://github.com/inference-gateway/a2a-debugger/commit/7ade7d3707f4f5716fc39e31a671a0840335e7b9))

## [0.2.0](https://github.com/inference-gateway/a2a-debugger/compare/v0.1.2...v0.2.0) (2025-06-17)

### ✨ Features

* Add installation script for A2A Debugger with usage instructions ([cb360d1](https://github.com/inference-gateway/a2a-debugger/commit/cb360d1e53ea8c9733c9cc464d0f36567b32333b))

### ♻️ Improvements

* Rename a2a-debugger to a2a and update related configurations ([02b61f9](https://github.com/inference-gateway/a2a-debugger/commit/02b61f9c4ea76c66c29d9f956368bbfd8f8911c8))

### 🐛 Bug Fixes

* Ensure A2A client is initialized before executing commands ([3565b8d](https://github.com/inference-gateway/a2a-debugger/commit/3565b8dc73d553a263a1c66b75ec04f589897185))

### 📚 Documentation

* Add warning section to README.md about project status and breaking changes ([06681ef](https://github.com/inference-gateway/a2a-debugger/commit/06681ef089df0419f2fd6560ef3d0562f181a4c9))

## [0.1.2](https://github.com/inference-gateway/a2a-debugger/compare/v0.1.1...v0.1.2) (2025-06-17)

### 🐛 Bug Fixes

* Create LICENSE ([331e306](https://github.com/inference-gateway/a2a-debugger/commit/331e3060956da5008d67392dc24080849639280d))

## [0.1.1](https://github.com/inference-gateway/a2a-debugger/compare/v0.1.0...v0.1.1) (2025-06-17)

### 🐛 Bug Fixes

* Update .goreleaser.yaml for consistency in formatting and paths ([36a8fb3](https://github.com/inference-gateway/a2a-debugger/commit/36a8fb38acf3aff69b7c97b86eb229e6d5bfa3ec))

### 🔧 Miscellaneous

* Add install and uninstall tasks to Taskfile ([b6d6851](https://github.com/inference-gateway/a2a-debugger/commit/b6d6851e65ce2e82761ba567650324df0665a9e1))
