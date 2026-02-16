// Pollex service worker — persists polish requests across popup lifecycle.

importScripts("api.js");

const MAX_HISTORY = 7;
const STALE_TIMEOUT_MS = 150000;

let abortController = null;
let tickInterval = null;

chrome.runtime.onMessage.addListener((msg, _sender, sendResponse) => {
  if (msg.type === "POLISH_START") {
    handleStart(msg.payload).then(sendResponse);
    return true; // async response
  }
  if (msg.type === "POLISH_CANCEL") {
    handleCancel().then(sendResponse);
    return true;
  }
});

const MAX_TEXT_LENGTH = 1500;

async function handleStart({ text, modelId }) {
  if (!text || typeof text !== "string" || !text.trim()) {
    return { ok: false, error: "Text is required" };
  }
  if (!modelId || typeof modelId !== "string") {
    return { ok: false, error: "Model is required" };
  }
  if (text.length > MAX_TEXT_LENGTH) {
    return { ok: false, error: "Text too long" };
  }

  const { polishJob } = await chrome.storage.local.get("polishJob");
  if (polishJob && polishJob.status === "running") {
    return { ok: false, error: "Already running" };
  }

  await chrome.storage.local.set({
    polishJob: {
      status: "running",
      inputText: text,
      modelId,
      startedAt: Date.now(),
    },
  });

  startTick();
  doFetch(text, modelId);
  return { ok: true };
}

async function handleCancel() {
  if (abortController) {
    abortController.abort();
    abortController = null;
  }
  stopTick();
  await chrome.storage.local.set({
    polishJob: { status: "cancelled" },
  });
  return { ok: true };
}

function startTick() {
  let seconds = 0;
  stopTick();
  tickInterval = setInterval(() => {
    seconds++;
    chrome.runtime.sendMessage({ type: "POLISH_TICK", seconds }).catch(() => {});
  }, 1000);
}

function stopTick() {
  if (tickInterval) {
    clearInterval(tickInterval);
    tickInterval = null;
  }
}

async function doFetch(text, modelId) {
  abortController = new AbortController();
  try {
    const result = await fetchPolish(text, modelId, abortController.signal);
    stopTick();

    await chrome.storage.local.set({
      polishJob: {
        status: "completed",
        result: {
          polished: result.polished,
          model: result.model,
          elapsed_ms: result.elapsed_ms,
        },
      },
    });

    await appendHistory(text, result);
    await chrome.storage.local.remove("draftText");
  } catch (err) {
    stopTick();
    if (err.name === "AbortError") {
      await chrome.storage.local.set({
        polishJob: { status: "cancelled" },
      });
    } else {
      const safeMsg = (err.message || "Request failed").slice(0, 200);
      await chrome.storage.local.set({
        polishJob: { status: "failed", error: safeMsg },
      });
    }
  } finally {
    abortController = null;
  }
}

async function appendHistory(inputText, result) {
  const { history = [] } = await chrome.storage.local.get("history");
  const entry = {
    id: `h_${Date.now()}`,
    input: inputText,
    output: result.polished,
    model: result.model,
    elapsed_ms: result.elapsed_ms,
    timestamp: Date.now(),
  };
  history.unshift(entry);
  if (history.length > MAX_HISTORY) history.length = MAX_HISTORY;
  await chrome.storage.local.set({ history });
}
