# Changelog

## [1.6.6](https://github.com/mlorentedev/pollex/compare/v1.6.5...v1.6.6) (2026-02-18)


### Bug Fixes

* parametrize multi-Jetson deploy scripts and fix q4_0 model bug ([a603aa0](https://github.com/mlorentedev/pollex/commit/a603aa03654113859ec7fee0b1063274b63c6b2b))

## [1.6.5](https://github.com/mlorentedev/pollex/compare/v1.6.4...v1.6.5) (2026-02-18)


### Bug Fixes

* simplify Makefile by removing broken deploy-prod target ([096b368](https://github.com/mlorentedev/pollex/commit/096b368c54604a89149f674d69393dc9bfde5946))

## [1.6.4](https://github.com/mlorentedev/pollex/compare/v1.6.3...v1.6.4) (2026-02-18)


### Bug Fixes

* add curl progress bar to deploy-prod to prevent tunnel timeout ([4b3bf44](https://github.com/mlorentedev/pollex/commit/4b3bf443b425dce4f343ce66f86a8d2a79031e86))

## [1.6.3](https://github.com/mlorentedev/pollex/compare/v1.6.2...v1.6.3) (2026-02-18)


### Bug Fixes

* deploy-prod downloads release binary from GitHub instead of SCP ([0d9bbfd](https://github.com/mlorentedev/pollex/commit/0d9bbfd220a7d7d7afa52f991aa2cd7748585998))

## [1.6.2](https://github.com/mlorentedev/pollex/compare/v1.6.1...v1.6.2) (2026-02-18)


### Bug Fixes

* correct goreleaser archive path from config.yaml.example to config.yaml ([ad4ff21](https://github.com/mlorentedev/pollex/commit/ad4ff2165c5d335aba3773310cbbae42fe4c6ae8))

## [1.6.1](https://github.com/mlorentedev/pollex/compare/v1.6.0...v1.6.1) (2026-02-18)


### Bug Fixes

* unify release-please and goreleaser into single workflow ([f8fbc8d](https://github.com/mlorentedev/pollex/commit/f8fbc8dcbbdf6433dea9a648067d4a908b1dfc25))

## [1.6.0](https://github.com/mlorentedev/pollex/compare/v1.5.0...v1.6.0) (2026-02-18)


### Features

* add deploy-prod target with pre-flight guardrails ([f6fae81](https://github.com/mlorentedev/pollex/commit/f6fae8126dd4c56083946a990f8a6c74ccaf6e75))

## [1.5.0](https://github.com/mlorentedev/pollex/compare/v1.4.0...v1.5.0) (2026-02-18)


### Features

* multi-node blue-green deployment ([fecd95f](https://github.com/mlorentedev/pollex/commit/fecd95f937136307922c693c88f2f082f4804c76))


### Bug Fixes

* add .gitattributes to enforce LF line endings for deploy scripts ([b7177e9](https://github.com/mlorentedev/pollex/commit/b7177e9442840da8c152d9fe5fbb9d6b7485ff1d))
* wrong model in llama server startup ([a8563ca](https://github.com/mlorentedev/pollex/commit/a8563cad86b8e0d5d8c593f5d4ab68c1a3453a12))

## [1.4.0](https://github.com/mlorentedev/pollex/compare/v1.3.1...v1.4.0) (2026-02-16)


### Features

* add backend version to health endpoint and extension settings ([7ad8a94](https://github.com/mlorentedev/pollex/commit/7ad8a949e03d933b097206a00ee666b3d226950e))
* add service worker, rolling history, progress bar, and prompt injection defense to extension ([014b4b2](https://github.com/mlorentedev/pollex/commit/014b4b2aa93d3c10d4dcf9cad71f36f4b7362a20))

## [1.3.1](https://github.com/mlorentedev/pollex/compare/v1.3.0...v1.3.1) (2026-02-16)


### Bug Fixes

* sync extension manifest version and fix release-please config ([7d0966c](https://github.com/mlorentedev/pollex/commit/7d0966c9c53864e4660e7275ae0fe4f5dfdf759f))

## [1.3.0](https://github.com/mlorentedev/pollex/compare/v1.2.0...v1.3.0) (2026-02-16)


### Features

* containerization (Dockerfile + compose) + observability alerting stack ([62d544a](https://github.com/mlorentedev/pollex/commit/62d544a0dd3bf26661e20e5480012cc33534ac39))
* SLOs/SLIs definition (ADR-007) + Prometheus alerting + Grafana dashboard ([cbc7eba](https://github.com/mlorentedev/pollex/commit/cbc7eba7df344832fad7f51db24d71aade36f860))


### Bug Fixes

* k6 load test + remove Ansible + repo cleanup ([d88a365](https://github.com/mlorentedev/pollex/commit/d88a36507e2882ef4fac55aa457d4a5373d0289d))

## [1.2.0](https://github.com/mlorentedev/pollex/compare/v1.1.0...v1.2.0) (2026-02-15)


### Features

* Prometheus metrics endpoint + structured JSON logging (slog) ([f9f44cf](https://github.com/mlorentedev/pollex/commit/f9f44cfb58b1ab0cf920631040180028b2316a63))
* Q4_0 quantization + mlock (22% faster), extension draft persistence, quality benchmark mode ([48badf7](https://github.com/mlorentedev/pollex/commit/48badf71fc3459077656637cf17e15ce0f64fcc2))

## [1.1.0](https://github.com/mlorentedev/pollex/compare/v1.0.0...v1.1.0) (2026-02-14)


### Features

* benchmark fixes, extension slow-text warning, timeout 120s, llama-server tuning ([d29495a](https://github.com/mlorentedev/pollex/commit/d29495ae215b0cd2e60e97975697b7be6c2145f7))

## 1.0.0 (2026-02-14)


### Features

* add API key auth + Cloudflare Tunnel for remote access ([f0581d0](https://github.com/mlorentedev/pollex/commit/f0581d0bafb196779259b67d6ef0467307ffacaa))
* add benchmark CLI, CI/CD pipelines, and system prompt improvement ([a4b2082](https://github.com/mlorentedev/pollex/commit/a4b208254245bbd98a22b782ce90eb415ed21d98))
* full implementation — Go API, browser extension, deploy ([dc84016](https://github.com/mlorentedev/pollex/commit/dc84016f890e17d10b5d7ee19c14865221e64bfe))
