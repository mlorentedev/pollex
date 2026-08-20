---
id: lesson-059-nan-gateway-reasoning-lives-in-reasoning-cont
type: lesson
status: active
created: "2026-06-05"
owner: manu
tags: [pollex, lesson, nan, llm, api, performance]
---

# NaN gateway: reasoning lives in `reasoning_content`, suppress with `enable_thinking:false`

**Context:** Integrating nan.builders OpenAI-compatible gateway for cloud inference (mimo-v2.5, qwen3.6, gemma4).

**Problem:** Reasoning models return a non-standard `reasoning_content` field alongside `choices[0].message.content`. Leaving `enable_thinking` at default causes reasoning tokens to be generated silently, burning quota and adding 3–10× latency.

**Solution:** Send `chat_template_kwargs: {"enable_thinking": false}` in every request. Parse only `choices[0].message.content`; ignore `reasoning_content` entirely. Confirmed: reasoning_tokens=0 in smoke test.

**Why:** The gateway honors `enable_thinking:false` at the model level, not just in the response schema. Without it, reasoning runs even if you don't read the output.

**Tags:** `#nan` `#llm` `#api` `#performance`

---
