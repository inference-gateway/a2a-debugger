# Changelog

All notable changes to this project will be documented in this file.

## [0.8.2](https://github.com/inference-gateway/a2a-debugger/compare/v0.8.1...v0.8.2) (2026-05-07)

### ♻️ Improvements

* **deps:** Remove devcontainer ([d346703](https://github.com/inference-gateway/a2a-debugger/commit/d34670320509905fcc5dedeecdd677432b547255))
* **docs:** Remove redundant copilot-instructions ([7f97852](https://github.com/inference-gateway/a2a-debugger/commit/7f97852683c4931b01aecee44c458498ef517cab))
* **types:** Consume A2A types from adk/types, drop local generation ([b929114](https://github.com/inference-gateway/a2a-debugger/commit/b9291145586ee6b78ba3b31245b345f32aa474fe))
* Use the correct environment variables ([01ad8a9](https://github.com/inference-gateway/a2a-debugger/commit/01ad8a9ddb8a591b1e80ddd5e733ffe8f4290967))

### 🐛 Bug Fixes

* **streaming:** Discriminate events structurally; switch example to mock-agent ([2385de5](https://github.com/inference-gateway/a2a-debugger/commit/2385de5a9fdecd6f614497d534f5cdfbf6492fa8))

### 👷 CI

* Bump all actions to latest ([4b5ab22](https://github.com/inference-gateway/a2a-debugger/commit/4b5ab2248030635d0580edc1114240d9f727bdca))

### 📚 Documentation

* Add AGENTS.md ([ea16585](https://github.com/inference-gateway/a2a-debugger/commit/ea16585402a7c713196e6f43b86949cca72bfc8d))
* Update CLAUDE.md for clarity on CLI command structure and testing against mock A2A server ([c54dbf2](https://github.com/inference-gateway/a2a-debugger/commit/c54dbf26621cd283a70e24f211c1034c8a045384))

### 🔧 Miscellaneous

* Add .env to git ignore ([68eb604](https://github.com/inference-gateway/a2a-debugger/commit/68eb604fe473cb9d64b8e6c2304a576469a9fe92))
* **deps:** Update dependencies to latest versions (Go 1.26.2, go-task 3.48.0, claude-code 2.1.123) ([6088465](https://github.com/inference-gateway/a2a-debugger/commit/6088465db8c6457d7c6ad07991ed25c0d9741e35))
* Remove outdated issue templates for bug, feature, and refactor requests ([8b36c50](https://github.com/inference-gateway/a2a-debugger/commit/8b36c50bc24fa297de6f9f0e5d692a3b70f59ae3))
* Sort imports ([f4d2b72](https://github.com/inference-gateway/a2a-debugger/commit/f4d2b726f670f8fe4acb8a46923ecb8f48414b54))

## [0.8.1](https://github.com/inference-gateway/a2a-debugger/compare/v0.8.0...v0.8.1) (2026-05-07)

### 👷 CI

* Bump checkout and setup-go actions versions ([bd64aa5](https://github.com/inference-gateway/a2a-debugger/commit/bd64aa5039881c274d831b794edef099a3fada66))
* **deps:** Bump golangci-lint to latest ([b45dce8](https://github.com/inference-gateway/a2a-debugger/commit/b45dce8428f579d32a56c78647560a9560ce65ad))
* **deps:** Update golangci-lint installation script to use the latest version v2.12.2 ([135b290](https://github.com/inference-gateway/a2a-debugger/commit/135b2900b4c84e107f21e670a3344e8cd09cfa80))
* Update task installation version to v3.48.0 in CI workflows ([cd0f5f7](https://github.com/inference-gateway/a2a-debugger/commit/cd0f5f7d4cd6a6bf32256d73df80d18a75a7a306))

## [0.8.0](https://github.com/inference-gateway/a2a-debugger/compare/v0.7.0...v0.8.0) (2025-10-09)

### ✨ Features

* **tasks:** Add --include-artifacts flag to tasks list command ([#21](https://github.com/inference-gateway/a2a-debugger/issues/21)) ([1e91d73](https://github.com/inference-gateway/a2a-debugger/commit/1e91d73fb363d8e553ecaeb16fdd4adca52d5b6b)), closes [#18](https://github.com/inference-gateway/a2a-debugger/issues/18)
* **tasks:** Add --include-history flag to tasks list command ([#20](https://github.com/inference-gateway/a2a-debugger/issues/20)) ([6c9826f](https://github.com/inference-gateway/a2a-debugger/commit/6c9826f5d0ea93299b4713cd4057e6e8f071153c)), closes [#19](https://github.com/inference-gateway/a2a-debugger/issues/19)

### 🔧 Miscellaneous

* **deps:** Update create-github-app-token action to v2.1.4 ([f5afc25](https://github.com/inference-gateway/a2a-debugger/commit/f5afc2595652e7b1d6b30243dbf8bf7d48f3d2ac))

## [0.7.0](https://github.com/inference-gateway/a2a-debugger/compare/v0.6.1...v0.7.0) (2025-10-05)

### ✨ Features

* **cli:** Add YAML and JSON output format support ([#15](https://github.com/inference-gateway/a2a-debugger/issues/15)) ([a9a54a2](https://github.com/inference-gateway/a2a-debugger/commit/a9a54a28ba14cd6bc8129d888effc3ea12726f67)), closes [#14](https://github.com/inference-gateway/a2a-debugger/issues/14)

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
