---
id: pollex-lessons-index
type: index
status: active
created: "2026-05-10"
owner: manu
tags: [pollex, lessons, index]
---

# Lessons Learned Index

| # | Date | Title | File | Tags |
|---|---|---|---|---|
| 001 | 2026-03-04 | Starlight CSS overrides must be scoped to accent variables only | [lesson-001-starlight-css-overrides-must-be-scoped-to-acc.md](./lesson-001-starlight-css-overrides-must-be-scoped-to-acc.md) | `starlight`, `astro`, `css`, `accessibility`, `docs-site` |
| 002 | 2026-03-04 | Light-mode Starlight buttons need accent ≥ 6:1 contrast vs white | [lesson-002-light-mode-starlight-buttons-need-accent-6-1-.md](./lesson-002-light-mode-starlight-buttons-need-accent-6-1-.md) | `starlight`, `wcag`, `contrast`, `design-system`, `docs-site` |
| 003 | 2026-03-28 | Default prompt path is relative to working directory, not binary | [lesson-003-default-prompt-path-is-relative-to-working-di.md](./lesson-003-default-prompt-path-is-relative-to-working-di.md) | `go`, `paths`, `deployment` |
| 004 | 2026-03-28 | Extracting functions from `main()` enables testability | [lesson-004-extracting-functions-from-main-enables-testab.md](./lesson-004-extracting-functions-from-main-enables-testab.md) | `go`, `testing`, `architecture` |
| 005 | 2026-03-28 | `httptest.NewServer` > `httptest.NewRecorder` for E2E | [lesson-005-httptest-newserver-httptest-newrecorder-for-e.md](./lesson-005-httptest-newserver-httptest-newrecorder-for-e.md) | `go`, `testing` |
| 006 | 2026-03-28 | Middleware order matters | [lesson-006-middleware-order-matters.md](./lesson-006-middleware-order-matters.md) | `go`, `http`, `middleware` |
| 007 | 2026-03-28 | `http.MaxBytesError` requires `errors.As()` | [lesson-007-http-maxbyteserror-requires-errors-as.md](./lesson-007-http-maxbyteserror-requires-errors-as.md) | `go`, `error-handling` |
| 008 | 2026-03-28 | Rate limiter sliding window with `[]time.Time` | [lesson-008-rate-limiter-sliding-window-with-time-time.md](./lesson-008-rate-limiter-sliding-window-with-time-time.md) | `go`, `rate-limiting` |
| 009 | 2026-03-28 | `signal.Notify` needs buffer 1 | [lesson-009-signal-notify-needs-buffer-1.md](./lesson-009-signal-notify-needs-buffer-1.md) | `go`, `signals` |
| 010 | 2026-04-15 | JetPack 4.6.6 is the last supported version for Jetson Nano 4GB | [lesson-010-jetpack-4-6-6-is-the-last-supported-version-f.md](./lesson-010-jetpack-4-6-6-is-the-last-supported-version-f.md) | `jetson`, `jetpack`, `cuda` |
| 011 | 2026-04-15 | First boot takes ~45 min on SD card | [lesson-011-first-boot-takes-45-min-on-sd-card.md](./lesson-011-first-boot-takes-45-min-on-sd-card.md) | `jetson`, `sd-card` |
| 012 | 2026-04-15 | sshd doesn't start until OEM setup is complete | [lesson-012-sshd-doesn-t-start-until-oem-setup-is-complet.md](./lesson-012-sshd-doesn-t-start-until-oem-setup-is-complet.md) | `jetson`, `ssh` |
| 013 | 2026-04-15 | CUDA is not in PATH by default | [lesson-013-cuda-is-not-in-path-by-default.md](./lesson-013-cuda-is-not-in-path-by-default.md) | `jetson`, `cuda`, `path` |
| 014 | 2026-04-15 | Passwordless sudo required for remote deploy scripts | [lesson-014-passwordless-sudo-required-for-remote-deploy-.md](./lesson-014-passwordless-sudo-required-for-remote-deploy-.md) | `deployment`, `sudo`, `ssh` |
| 015 | 2026-04-15 | SSH to Jetson requires jump host | [lesson-015-ssh-to-jetson-requires-jump-host.md](./lesson-015-ssh-to-jetson-requires-jump-host.md) | `ssh`, `proxmox` |
| 016 | 2026-04-15 | SCP to protected paths fails due to permissions | [lesson-016-scp-to-protected-paths-fails-due-to-permissio.md](./lesson-016-scp-to-protected-paths-fails-due-to-permissio.md) | `deployment`, `scp`, `permissions` |
| 017 | 2026-04-15 | `zstd` required for Ollama | [lesson-017-zstd-required-for-ollama.md](./lesson-017-zstd-required-for-ollama.md) | `ollama`, `dependencies` |
| 018 | 2026-04-15 | llama.cpp repo migrated from `ggerganov` to `ggml-org` | [lesson-018-llama-cpp-repo-migrated-from-ggerganov-to-ggm.md](./lesson-018-llama-cpp-repo-migrated-from-ggerganov-to-ggm.md) | `llama`, `docker` |
| 019 | 2026-04-15 | Test with real Docker before deploying | [lesson-019-test-with-real-docker-before-deploying.md](./lesson-019-test-with-real-docker-before-deploying.md) | `testing`, `llama` |
| 020 | 2026-04-15 | CMake 3.14+ required for llama.cpp | [lesson-020-cmake-3-14-required-for-llama-cpp.md](./lesson-020-cmake-3-14-required-for-llama-cpp.md) | `cmake`, `jetson` |
| 021 | 2026-04-15 | `-DCMAKE_CUDA_STANDARD=14` is mandatory for Jetson Nano | [lesson-021-dcmake-cuda-standard-14-is-mandatory-for-jets.md](./lesson-021-dcmake-cuda-standard-14-is-mandatory-for-jets.md) | `cmake`, `cuda`, `jetson` |
| 022 | 2026-04-15 | Full cmake flags for Jetson Nano | [lesson-022-full-cmake-flags-for-jetson-nano.md](./lesson-022-full-cmake-flags-for-jetson-nano.md) | `cmake`, `jetson`, `arm64` |
| 023 | 2026-04-15 | NEON stubs go in `ggml-cpu-impl.h`, NOT in `ggml-cpu-quants.c` | [lesson-023-neon-stubs-go-in-ggml-cpu-impl-h-not-in-ggml-.md](./lesson-023-neon-stubs-go-in-ggml-cpu-impl-h-not-in-ggml-.md) | `arm`, `neon`, `llama` |
| 024 | 2026-04-15 | gcc-8 on aarch64 provides `vld1q_*_x2` but NOT `_x4` | [lesson-024-gcc-8-on-aarch64-provides-vld1q-x2-but-not-x4.md](./lesson-024-gcc-8-on-aarch64-provides-vld1q-x2-but-not-x4.md) | `arm`, `gcc`, `llama` |
| 025 | 2026-04-15 | WMMA (fattn-wmma-f16.cu) requires Volta+ (compute 7.0) | [lesson-025-wmma-fattn-wmma-f16-cu-requires-volta-compute.md](./lesson-025-wmma-fattn-wmma-f16-cu-requires-volta-compute.md) | `cuda`, `wmma`, `jetson` |
| 026 | 2026-04-15 | `cuda_bf16.h` stub must do `typedef half nv_bfloat16` | [lesson-026-cuda-bf16-h-stub-must-do-typedef-half-nv-bflo.md](./lesson-026-cuda-bf16-h-stub-must-do-typedef-half-nv-bflo.md) | `cuda`, `jetson` |
| 027 | 2026-04-15 | `<charconv>` is C++17, not available with nvcc C++14 | [lesson-027-charconv-is-c-17-not-available-with-nvcc-c-14.md](./lesson-027-charconv-is-c-17-not-available-with-nvcc-c-14.md) | `cuda`, `c`, `jetson` |
| 028 | 2026-04-15 | Don't blanket-replace `static constexpr` in functions | [lesson-028-don-t-blanket-replace-static-constexpr-in-fun.md](./lesson-028-don-t-blanket-replace-static-constexpr-in-fun.md) | `sed`, `llama` |
| 029 | 2026-04-15 | `crypto/subtle.ConstantTimeCompare` prevents timing attacks | [lesson-029-crypto-subtle-constanttimecompare-prevents-ti.md](./lesson-029-crypto-subtle-constanttimecompare-prevents-ti.md) | `security`, `go` |
| 030 | 2026-04-15 | `Cf-Connecting-Ip` header for Cloudflare Tunnel | [lesson-030-cf-connecting-ip-header-for-cloudflare-tunnel.md](./lesson-030-cf-connecting-ip-header-for-cloudflare-tunnel.md) | `cloudflare`, `tunnel`, `rate-limiting` |
| 031 | 2026-04-15 | `host_permissions: ["<all_urls>"]` required in Manifest V3 | [lesson-031-host-permissions-all-urls-required-in-manifes.md](./lesson-031-host-permissions-all-urls-required-in-manifes.md) | `chrome-extension`, `manifest-v3` |
| 032 | 2026-05-01 | Q4_0 vs Q4_K_M: Q4_0 is ~23% faster on Jetson Nano | [lesson-032-q4-0-vs-q4-k-m-q4-0-is-23-faster-on-jetson-na.md](./lesson-032-q4-0-vs-q4-k-m-q4-0-is-23-faster-on-jetson-na.md) | `quantization`, `jetson`, `performance` |
| 033 | 2026-05-01 | `--mlock` prevents model paging | [lesson-033-mlock-prevents-model-paging.md](./lesson-033-mlock-prevents-model-paging.md) | `jetson`, `llama`, `performance` |
| 034 | 2026-05-01 | 1500 char limit in extension | [lesson-034-1500-char-limit-in-extension.md](./lesson-034-1500-char-limit-in-extension.md) | `extension`, `limits` |
| 035 | 2026-05-01 | `promauto` registers metrics automatically | [lesson-035-promauto-registers-metrics-automatically.md](./lesson-035-promauto-registers-metrics-automatically.md) | `prometheus`, `metrics` |
| 036 | 2026-05-01 | Chrome popup lifecycle — destroyed on focus loss | [lesson-036-chrome-popup-lifecycle-destroyed-on-focus-los.md](./lesson-036-chrome-popup-lifecycle-destroyed-on-focus-los.md) | `chrome-extension`, `service-worker` |
| 037 | 2026-05-01 | `chrome.storage.onChanged` is the reactive bridge | [lesson-037-chrome-storage-onchanged-is-the-reactive-brid.md](./lesson-037-chrome-storage-onchanged-is-the-reactive-brid.md) | `chrome-extension`, `storage` |
| 038 | 2026-05-01 | Stale job detection | [lesson-038-stale-job-detection.md](./lesson-038-stale-job-detection.md) | `chrome-extension`, `stale-jobs` |
| 039 | 2026-05-01 | Timer ticks best-effort | [lesson-039-timer-ticks-best-effort.md](./lesson-039-timer-ticks-best-effort.md) | `chrome-extension`, `timer` |
| 040 | 2026-05-01 | Input validation in service worker, not just popup | [lesson-040-input-validation-in-service-worker-not-just-p.md](./lesson-040-input-validation-in-service-worker-not-just-p.md) | `chrome-extension`, `validation` |
| 041 | 2026-05-01 | Error truncation (200 chars) | [lesson-041-error-truncation-200-chars.md](./lesson-041-error-truncation-200-chars.md) | `chrome-extension`, `errors` |
| 042 | 2026-05-01 | Prompt injection defense | [lesson-042-prompt-injection-defense.md](./lesson-042-prompt-injection-defense.md) | `security`, `llm`, `prompt-injection` |
| 043 | 2026-05-01 | Progress bar ETA: pad +15% | [lesson-043-progress-bar-eta-pad-15.md](./lesson-043-progress-bar-eta-pad-15.md) | `chrome-extension`, `ux` |
| 044 | 2026-05-01 | `alpine:3.21` minimal base for Docker | [lesson-044-alpine-3-21-minimal-base-for-docker.md](./lesson-044-alpine-3-21-minimal-base-for-docker.md) | `docker`, `alpine` |
| 045 | 2026-05-01 | `--mount=type=cache` in Docker build | [lesson-045-mount-type-cache-in-docker-build.md](./lesson-045-mount-type-cache-in-docker-build.md) | `docker`, `performance` |
| 046 | 2026-05-01 | GitHub renders Mermaid natively | [lesson-046-github-renders-mermaid-natively.md](./lesson-046-github-renders-mermaid-natively.md) | `docs`, `mermaid` |
| 047 | 2026-05-01 | Assets in `docs/assets/` | [lesson-047-assets-in-docs-assets.md](./lesson-047-assets-in-docs-assets.md) | `docs`, `organization` |
| 048 | 2026-06-05 | `cloudflared tunnel route dns` doesn't overwrite | [lesson-048-cloudflared-tunnel-route-dns-doesn-t-overwrit.md](./lesson-048-cloudflared-tunnel-route-dns-doesn-t-overwrit.md) | `cloudflare`, `tunnel`, `dns` |
| 049 | 2026-06-05 | Don't stop the inactive node's tunnel | [lesson-049-don-t-stop-the-inactive-node-s-tunnel.md](./lesson-049-don-t-stop-the-inactive-node-s-tunnel.md) | `cloudflare`, `tunnel`, `multi-node` |
| 050 | 2026-06-05 | Restarting cloudflared kills your SSH | [lesson-050-restarting-cloudflared-kills-your-ssh.md](./lesson-050-restarting-cloudflared-kills-your-ssh.md) | `cloudflare`, `tunnel`, `ssh` |
| 051 | 2026-06-05 | SSH multiplexing (`ControlMaster`) critical for Cloudflare Tunnel | [lesson-051-ssh-multiplexing-controlmaster-critical-for-c.md](./lesson-051-ssh-multiplexing-controlmaster-critical-for-c.md) | `ssh`, `cloudflare`, `performance` |
| 052 | 2026-06-05 | `build-llamacpp.sh` downloaded wrong model | [lesson-052-build-llamacpp-sh-downloaded-wrong-model.md](./lesson-052-build-llamacpp-sh-downloaded-wrong-model.md) | `deployment`, `llama` |
| 053 | 2026-06-05 | `User=manu` in systemd fails only for cloudflared | [lesson-053-user-manu-in-systemd-fails-only-for-cloudflar.md](./lesson-053-user-manu-in-systemd-fails-only-for-cloudflar.md) | `systemd`, `jetson`, `cloudflare` |
| 054 | 2026-06-05 | Prompt language detection — polish in the same language as input | [lesson-054-prompt-language-detection-polish-in-the-same-.md](./lesson-054-prompt-language-detection-polish-in-the-same-.md) | `llm`, `prompt`, `i18n` |
| 055 | 2026-06-05 | NaN Cloud branding: "NaN Cloud (auto)" — not "Nous Cloud" | [lesson-055-nan-cloud-branding-nan-cloud-auto-not-nous-cl.md](./lesson-055-nan-cloud-branding-nan-cloud-auto-not-nous-cl.md) | `branding`, `nan`, `cloud` |
| 056 | 2026-06-05 | Cloudflare DNS route conflict requires --overwrite-dns | [lesson-056-cloudflare-dns-route-conflict-requires-overwr.md](./lesson-056-cloudflare-dns-route-conflict-requires-overwr.md) | `cloudflare`, `dns`, `tunnel` |
| 057 | 2026-06-05 | Makefile `deploy-tunnel-route` had argument order wrong | [lesson-057-makefile-deploy-tunnel-route-had-argument-ord.md](./lesson-057-makefile-deploy-tunnel-route-had-argument-ord.md) | `makefile`, `cloudflare` |
| 058 | 2026-06-05 | Kubelab (GPU) label for llama.cpp in extension | [lesson-058-kubelab-gpu-label-for-llama-cpp-in-extension.md](./lesson-058-kubelab-gpu-label-for-llama-cpp-in-extension.md) | `extension`, `ui` |
| 059 | 2026-06-05 | NaN gateway: reasoning lives in `reasoning_content`, suppress with `enable_thinking:false` | [lesson-059-nan-gateway-reasoning-lives-in-reasoning-cont.md](./lesson-059-nan-gateway-reasoning-lives-in-reasoning-cont.md) | `nan`, `llm`, `api`, `performance` |
| 060 | 2026-06-05 | Account-wide rate limit on nan.builders is shared with all tooling | [lesson-060-account-wide-rate-limit-on-nan-builders-is-sh.md](./lesson-060-account-wide-rate-limit-on-nan-builders-is-sh.md) | `nan`, `rate-limiting`, `concurrency`, `cloud` |
| 061 | 2026-06-05 | FallbackChain error policy: advance on availability, fail fast on client errors | [lesson-061-fallbackchain-error-policy-advance-on-availab.md](./lesson-061-fallbackchain-error-policy-advance-on-availab.md) | `go`, `fallback`, `error-handling`, `nan` |
| 062 | 2026-06-06 | gremlins v0.6.0 appends `...` internally — pass bare package paths in CI | [lesson-062-gremlins-v0-6-0-appends-internally-pass-bare-.md](./lesson-062-gremlins-v0-6-0-appends-internally-pass-bare-.md) | `gremlins`, `mutation-testing`, `ci`, `go` |
| 063 | 2026-08-05 | `--mock` must force auth off — dotfiles leaks POLLEX_API_KEY into every shell | [lesson-063-mock-must-force-auth-off-dotfiles-leaks-polle.md](./lesson-063-mock-must-force-auth-off-dotfiles-leaks-polle.md) | `go`, `extension`, `dev-loop`, `auth`, `dotfiles` |
| 064 | 2026-08-05 | `gh project item-add` fails with "unknown owner type" under a fine-grained PAT — use GraphQL in workflows | [lesson-064-gh-project-item-add-fails-with-unknown-owner-.md](./lesson-064-gh-project-item-add-fails-with-unknown-owner-.md) | `ci`, `github-projects`, `fine-grained-pat`, `graphql`, `bitacora` |
| 065 | 2026-08-05 | Astro build artifacts (`site/.astro/`) must not be committed | [lesson-065-astro-build-artifacts-site-astro-must-not-be-.md](./lesson-065-astro-build-artifacts-site-astro-must-not-be-.md) | `astro`, `docs-site`, `gitignore`, `generated-artifacts` |
| 066 | 2026-08-06 | GitHub does not expose secrets to Dependabot PRs — skip them in workflows that need secrets | [lesson-066-github-does-not-expose-secrets-to-dependabot-.md](./lesson-066-github-does-not-expose-secrets-to-dependabot-.md) | `ci`, `dependabot`, `secrets`, `github-actions`, `bitacora` |
| 067 | 2026-08-06 | `make deploy-secrets` failed with "POLLEX_API_KEY not set" in a plain shell — auto-resolve via `dotf secrets run` | [lesson-067-make-deploy-secrets-failed-with-pollex-api-ke.md](./lesson-067-make-deploy-secrets-failed-with-pollex-api-ke.md) | `makefile`, `deploy`, `secrets`, `dotf`, `jetson` |
| 068 | 2026-08-05 | Extension e2e in CI needs a headed Chromium under `xvfb-run` | [lesson-068-extension-e2e-in-ci-needs-a-headed-chromium-u.md](./lesson-068-extension-e2e-in-ci-needs-a-headed-chromium-u.md) | `ci`, `playwright`, `extension`, `e2e`, `github-actions` |
| 069 | 2026-08-05 | `release-please` needs a PAT, not `GITHUB_TOKEN`, once CI is a required check | [lesson-069-release-please-needs-a-pat-not-github-token-o.md](./lesson-069-release-please-needs-a-pat-not-github-token-o.md) | `ci`, `release-please`, `fine-grained-pat`, `branch-protection`, `github-actions` |
| 070 | 2026-08-08 | A job inserted mid-job in Actions YAML silently steals the trailing steps | [lesson-070-a-job-inserted-mid-job-in-actions-yaml-silent.md](./lesson-070-a-job-inserted-mid-job-in-actions-yaml-silent.md) | `ci`, `github-actions`, `yaml`, `false-green`, `branch-protection` |
