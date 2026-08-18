---
id: lesson-054-prompt-language-detection-polish-in-the-same-
type: lesson
status: active
created: "2026-06-05"
owner: manu
tags: [pollex, lesson, llm, prompt, i18n]
---

# Prompt language detection — polish in the same language as input

**Context:** The polishing prompt was hardcoded to English only.

**Problem:** Users writing in Spanish, Portuguese, etc. got English-polished text instead of polished text in their language.

**Solution:** Updated system prompt to include `Language: Detect the language of the input text and preserve it in the output. Polish in the same language the user wrote in.`

**Tags:** `#llm` `#prompt` `#i18n`

---
