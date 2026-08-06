// Unit tests for the background service worker job lifecycle.
// Runs in Node with mocked chrome.storage / chrome.runtime / fetch.

import { describe, it, expect, beforeEach } from "vitest";
import {
  clearStorage,
  seedStorage,
  dumpStorage,
  fetchMock,
} from "./setup.js";
import {
  handleStart,
  handleCancel,
  doFetch,
  appendHistory,
  MAX_HISTORY,
  MAX_TEXT_LENGTH,
} from "../../background.js";

function okResponse(payload) {
  return {
    ok: true,
    status: 200,
    json: async () => payload,
  };
}

beforeEach(() => {
  clearStorage();
  fetchMock.impl = () => Promise.resolve(
    okResponse({ polished: "Fixed text.", model: "mock", elapsed_ms: 42 }),
  );
});

describe("handleStart — input validation", () => {
  it("rejects missing text", async () => {
    const res = await handleStart({ text: "", modelId: "mock" });
    expect(res).toEqual({ ok: false, error: "Text is required" });
  });

  it("rejects non-string text", async () => {
    const res = await handleStart({ text: 123, modelId: "mock" });
    expect(res).toEqual({ ok: false, error: "Text is required" });
  });

  it("rejects whitespace-only text", async () => {
    const res = await handleStart({ text: "   ", modelId: "mock" });
    expect(res).toEqual({ ok: false, error: "Text is required" });
  });

  it("rejects missing modelId", async () => {
    const res = await handleStart({ text: "hello", modelId: "" });
    expect(res).toEqual({ ok: false, error: "Model is required" });
  });

  it("rejects text over MAX_TEXT_LENGTH", async () => {
    const res = await handleStart({
      text: "a".repeat(MAX_TEXT_LENGTH + 1),
      modelId: "mock",
    });
    expect(res).toEqual({ ok: false, error: "Text too long" });
  });
});

describe("handleStart — job lifecycle", () => {
  it("refuses to start while another job is running", async () => {
    seedStorage({ polishJob: { status: "running", startedAt: Date.now() } });
    const res = await handleStart({ text: "hello", modelId: "mock" });
    expect(res).toEqual({ ok: false, error: "Already running" });
  });

  it("persists a running job and returns ok", async () => {
    const res = await handleStart({ text: "hello", modelId: "mock" });
    expect(res).toEqual({ ok: true });
    const { polishJob } = dumpStorage();
    expect(polishJob.status).toBe("running");
    expect(polishJob.inputText).toBe("hello");
    expect(polishJob.modelId).toBe("mock");
    expect(typeof polishJob.startedAt).toBe("number");
  });
});

describe("handleCancel", () => {
  it("marks the job cancelled", async () => {
    const res = await handleCancel();
    expect(res).toEqual({ ok: true });
    expect(dumpStorage().polishJob).toEqual({ status: "cancelled" });
  });
});

describe("doFetch — completion paths", () => {
  it("stores a completed result and removes the draft", async () => {
    seedStorage({ draftText: "hello" });
    await doFetch("hello", "mock");

    const { polishJob, draftText } = dumpStorage();
    expect(draftText).toBeUndefined();
    expect(polishJob.status).toBe("completed");
    expect(polishJob.result).toEqual({
      polished: "Fixed text.",
      model: "mock",
      elapsed_ms: 42,
    });
  });

  it("stores failed status with a truncated error", async () => {
    fetchMock.impl = () =>
      Promise.reject(new Error("x".repeat(300)));

    await doFetch("hello", "mock");

    const { polishJob } = dumpStorage();
    expect(polishJob.status).toBe("failed");
    expect(polishJob.error.length).toBeLessThanOrEqual(200);
  });

  it("stores cancelled status on AbortError", async () => {
    const err = new Error("aborted");
    err.name = "AbortError";
    fetchMock.impl = () => Promise.reject(err);

    await doFetch("hello", "mock");

    expect(dumpStorage().polishJob.status).toBe("cancelled");
  });
});

describe("appendHistory", () => {
  it("prepends entries and caps at MAX_HISTORY", async () => {
    for (let i = 0; i < MAX_HISTORY + 3; i++) {
      await appendHistory(`input ${i}`, {
        polished: `out ${i}`,
        model: "mock",
        elapsed_ms: i,
      });
    }
    const { history } = dumpStorage();
    expect(history.length).toBe(MAX_HISTORY);
    // Most recent first; oldest surviving entry is the last one within cap
    expect(history[0].input).toBe(`input ${MAX_HISTORY + 2}`);
    expect(history[history.length - 1].input).toBe("input 3");
  });

  it("creates an entry with all expected fields", async () => {
    await appendHistory("in", { polished: "out", model: "m", elapsed_ms: 7 });
    const { history } = dumpStorage();
    expect(history).toHaveLength(1);
    const entry = history[0];
    expect(entry.input).toBe("in");
    expect(entry.output).toBe("out");
    expect(entry.model).toBe("m");
    expect(entry.elapsed_ms).toBe(7);
    expect(typeof entry.id).toBe("string");
    expect(typeof entry.timestamp).toBe("number");
  });
});
