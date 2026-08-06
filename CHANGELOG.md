# Changelog

## [1.8.2](https://github.com/mlorentedev/pollex/compare/v1.8.1...v1.8.2) (2026-08-06)


### Bug Fixes

* **ci:** add PRs to bitácora board via gh CLI ([#39](https://github.com/mlorentedev/pollex/issues/39)) ([59dca56](https://github.com/mlorentedev/pollex/commit/59dca56e58ef7cf70e8e0a5b5f9aec0c283b31a7))
* **ci:** add PRs to bitácora via GraphQL, not gh project CLI ([#46](https://github.com/mlorentedev/pollex/issues/46)) ([0eab527](https://github.com/mlorentedev/pollex/commit/0eab52769483c75eb4855f3e6b169daab559b9af))
* **ci:** skip Dependabot PRs in add-to-project ([#51](https://github.com/mlorentedev/pollex/issues/51)) ([db8a1e4](https://github.com/mlorentedev/pollex/commit/db8a1e42d5b88675ee011a59e610bfd9317bc37c))
* force auth off in mock mode so the extension connects out of the box ([#40](https://github.com/mlorentedev/pollex/issues/40)) ([08940fd](https://github.com/mlorentedev/pollex/commit/08940fdb4c201fea47a6f71f7bd6bba5395b824c))
* make deploy-secrets resolve keys via dotf when not in shell ([#52](https://github.com/mlorentedev/pollex/issues/52)) ([ffce7bb](https://github.com/mlorentedev/pollex/commit/ffce7bb329c9dfd293a9f64c39aa2b171bb8decd))

## [1.8.1](https://github.com/mlorentedev/pollex/compare/v1.8.0...v1.8.1) (2026-06-06)


### Bug Fixes

* remove trailing ... from gremlins paths in CI ([#33](https://github.com/mlorentedev/pollex/issues/33)) ([2ce872a](https://github.com/mlorentedev/pollex/commit/2ce872a306b7c542ec18fb0a8dd48c5e041b88dc))

## [1.8.0](https://github.com/mlorentedev/pollex/compare/v1.7.0...v1.8.0) (2026-06-06)


### Features

* rebrand nous-cloud -&gt; nan-cloud, language-aware prompt, unify deploy ([319082e](https://github.com/mlorentedev/pollex/commit/319082e62347999dd623c2150a3288270aa6a88f))

## [1.7.0](https://github.com/mlorentedev/pollex/compare/v1.6.8...v1.7.0) (2026-06-06)


### Features

* add FallbackChain adapter with availability-aware retry ([054a293](https://github.com/mlorentedev/pollex/commit/054a293a05067112bac600653f9004b4096caf92))
* add NaN cloud config fields and POLLEX_NAN_* overrides ([0e59fd4](https://github.com/mlorentedev/pollex/commit/0e59fd4d53a4473f698dac00f5058a4f8077d0b4))
* add NousAdapter for NaN (nan.builders) cloud inference ([495beb5](https://github.com/mlorentedev/pollex/commit/495beb524e0184e0420bd331eac84a17afd2b662))
* bound concurrent NaN calls with a semaphore to avoid 429s ([de148ef](https://github.com/mlorentedev/pollex/commit/de148ef4381e7c26da47707b576e7dfcd152130e))
* register NaN cloud engine as a single auto-fallback model ([0897c22](https://github.com/mlorentedev/pollex/commit/0897c22ef7c8b196d920631a28dd220292a35250))
* surface NaN cloud engine in extension and deploy tooling ([484572f](https://github.com/mlorentedev/pollex/commit/484572f2faf45a3815aa7b01f70dd17594838980))


### Bug Fixes

* accept NaN base URL with or without trailing /v1 ([6fd9447](https://github.com/mlorentedev/pollex/commit/6fd9447d2098bb048e6b6701a9701b174fe75b98))
* cap NaN per-model timeout so the chain fits the request budget ([2cf8bda](https://github.com/mlorentedev/pollex/commit/2cf8bdab8241e018fa0c11f70395c89e63db214d))

## [1.6.8](https://github.com/mlorentedev/pollex/compare/v1.6.7...v1.6.8) (2026-03-05)


### Bug Fixes

* improve accent color contrast for light mode buttons and links ([65af826](https://github.com/mlorentedev/pollex/commit/65af826962c087c2107988977372041beee8606b))
* remove gray/white/black overrides that broke Starlight theme contrast ([c963ed7](https://github.com/mlorentedev/pollex/commit/c963ed74ea04b58e34d9b2d057b0a5426c29d80a))

## [1.6.7](https://github.com/mlorentedev/pollex/compare/v1.6.6...v1.6.7) (2026-02-22)


### Bug Fixes

* fix cloudflared.service JetPack 4.6 compatibility + prometheus host label ([ca5ce76](https://github.com/mlorentedev/pollex/commit/ca5ce76e9bf04507961b4f22073478df90bb11b4))

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
