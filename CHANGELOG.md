# Changelog

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
