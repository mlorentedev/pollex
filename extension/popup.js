// Pollex popup — wires UI to background service worker via messaging.

const MAX_CHARS = 1500;
const WARN_THRESHOLD = 0.9;
const MS_PER_CHAR = 36;
const SLOW_SECONDS = 45;
const DRAFT_DEBOUNCE_MS = 500;
const STALE_TIMEOUT_MS = 150000;

// --- DOM refs ---

const input = document.getElementById("input");
const charCount = document.getElementById("char-count");
const modelSelect = document.getElementById("model-select");
const modelLabel = document.getElementById("model-label");
const btnPolish = document.getElementById("btn-polish");
const btnCancel = document.getElementById("btn-cancel");
const btnCopy = document.getElementById("btn-copy");
const iconCopy = document.getElementById("icon-copy");
const iconCheck = document.getElementById("icon-check");
const copyLabel = document.getElementById("copy-label");
const btnSettings = document.getElementById("btn-settings");
const slowHint = document.getElementById("slow-hint");
const statusEl = document.getElementById("status");
const resultSection = document.getElementById("result-section");
const resultBox = document.getElementById("result");
const elapsedEl = document.getElementById("elapsed");
const progressSection = document.getElementById("progress-section");
const progressPct = document.getElementById("progress-pct");
const progressElapsed = document.getElementById("progress-elapsed");
const progressFill = document.getElementById("progress-fill");

// Settings panel
const btnBack = document.getElementById("btn-back");
const apiUrlInput = document.getElementById("api-url");
const apiKeyInput = document.getElementById("api-key");
const btnTest = document.getElementById("btn-test");
const btnSave = document.getElementById("btn-save");
const settingsStatus = document.getElementById("settings-status");
const serverVersion = document.getElementById("server-version");

// History
const historySection = document.getElementById("history-section");
const historyCount = document.getElementById("history-count");
const historyList = document.getElementById("history-list");
const historyDetail = document.getElementById("history-detail");
const btnHistoryBack = document.getElementById("btn-history-back");
const historyDetailMeta = document.getElementById("history-detail-meta");
const historyDetailInput = document.getElementById("history-detail-input");
const historyDetailOutput = document.getElementById("history-detail-output");
const btnHistoryCopy = document.getElementById("btn-history-copy");
const iconHistoryCopy = document.getElementById("icon-history-copy");
const iconHistoryCheck = document.getElementById("icon-history-check");
const historyCopyLabel = document.getElementById("history-copy-label");

let timerInterval = null;
let modelsLoaded = false;
let singleModelId = null;
let draftTimer = null;
let estimatedSeconds = 0;

// --- Init ---

document.addEventListener("DOMContentLoaded", async () => {
  await restoreDraft();
  await loadModels();
  await loadSettings();
  await recoverJobState();
  await loadHistory();
  updateCharCount();
  input.focus();
});

// --- Provider labels ---

const providerLabels = {
  ollama: "Local",
  mock: "Local",
  claude: "Cloud",
  "llama.cpp": "Local (GPU)",
};

// --- Models ---

async function loadModels() {
  try {
    const models = await fetchModels();

    // Single model: show static label instead of dropdown
    if (models.length === 1) {
      singleModelId = models[0].id;
      const provider = providerLabels[models[0].provider] || models[0].provider;
      modelLabel.textContent = `${models[0].name} · ${provider}`;
      modelLabel.classList.remove("hidden");
      modelSelect.classList.add("hidden");
    } else {
      singleModelId = null;
      modelLabel.classList.add("hidden");
      modelSelect.classList.remove("hidden");
      modelSelect.innerHTML = "";

      const groups = {};
      for (const m of models) {
        const label = providerLabels[m.provider] || m.provider;
        if (!groups[label]) groups[label] = [];
        groups[label].push(m);
      }

      const groupNames = Object.keys(groups);
      if (groupNames.length === 1) {
        for (const m of groups[groupNames[0]]) {
          modelSelect.appendChild(makeOption(m));
        }
      } else {
        for (const label of groupNames) {
          const optgroup = document.createElement("optgroup");
          optgroup.label = label;
          for (const m of groups[label]) {
            optgroup.appendChild(makeOption(m));
          }
          modelSelect.appendChild(optgroup);
        }
      }
    }

    btnPolish.disabled = false;
    modelsLoaded = true;
    fetchHealth().then(data => {
      if (data && data.version) serverVersion.textContent = `API ${data.version}`;
    }).catch(() => {});
  } catch {
    singleModelId = null;
    modelLabel.classList.add("hidden");
    modelSelect.classList.remove("hidden");
    modelSelect.innerHTML = '<option value="">Cannot connect</option>';
    btnPolish.disabled = true;
    modelsLoaded = false;
    showStatus("Cannot reach API — check Settings.", "error");
  }
}

function makeOption(model) {
  const opt = document.createElement("option");
  opt.value = model.id;
  opt.textContent = model.name;
  return opt;
}

// --- Draft persistence ---

async function restoreDraft() {
  const result = await chrome.storage.local.get("draftText");
  if (result.draftText) input.value = result.draftText;
}

function saveDraft() {
  clearTimeout(draftTimer);
  draftTimer = setTimeout(() => {
    chrome.storage.local.set({ draftText: input.value });
  }, DRAFT_DEBOUNCE_MS);
}

// --- Character count ---

input.addEventListener("input", () => {
  updateCharCount();
  clearTransientStatus();
  clearStaleJob();
  saveDraft();
});

function updateCharCount() {
  const len = input.value.length;
  charCount.textContent = `${len.toLocaleString()} / ${MAX_CHARS.toLocaleString()}`;

  charCount.classList.remove("warning", "error");
  if (len > MAX_CHARS) {
    charCount.classList.add("error");
  } else if (len > MAX_CHARS * WARN_THRESHOLD) {
    charCount.classList.add("warning");
  }

  const estimatedSec = Math.round((len * MS_PER_CHAR) / 1000);
  if (estimatedSec > SLOW_SECONDS) {
    const min = Math.floor(estimatedSec / 60);
    const sec = estimatedSec % 60;
    const timeStr = min > 0 ? `~${min}m ${sec}s` : `~${estimatedSec}s`;
    slowHint.textContent = `Estimated processing time: ${timeStr}`;
    slowHint.classList.remove("hidden");
  } else {
    slowHint.classList.add("hidden");
  }
}

function clearTransientStatus() {
  if (statusEl.classList.contains("error") || statusEl.classList.contains("cancelled")) {
    hideStatus();
  }
}

function clearStaleJob() {
  chrome.storage.local.get("polishJob", ({ polishJob }) => {
    if (polishJob && polishJob.status !== "running") {
      chrome.storage.local.remove("polishJob");
    }
  });
}

// --- Polish (via service worker) ---

btnPolish.addEventListener("click", doPolish);

input.addEventListener("keydown", (e) => {
  if ((e.ctrlKey || e.metaKey) && e.key === "Enter") {
    e.preventDefault();
    if (!btnPolish.disabled) doPolish();
  }
});

async function doPolish() {
  const text = input.value.trim();
  if (!text) {
    input.focus();
    input.classList.add("shake");
    input.addEventListener("animationend", () => input.classList.remove("shake"), { once: true });
    return;
  }

  const modelId = singleModelId || modelSelect.value;
  if (!modelId) return;

  // Reset UI
  resultSection.classList.add("hidden");
  btnPolish.disabled = true;
  btnCancel.classList.remove("hidden");
  estimatedSeconds = Math.max(1, Math.round((text.length * MS_PER_CHAR * 1.15) / 1000));
  showPolishingStatus(0);
  startLocalTimer(0);

  const resp = await chrome.runtime.sendMessage({
    type: "POLISH_START",
    payload: { text, modelId },
  });

  if (!resp || !resp.ok) {
    stopLocalTimer();
    hideProgress();
    btnCancel.classList.add("hidden");
    btnPolish.disabled = false;
    showStatus(resp?.error || "Failed to start.", "error");
  }
  // On success, UI updates arrive via storage.onChanged
}

// --- Cancel (via service worker) ---

btnCancel.addEventListener("click", async () => {
  await chrome.runtime.sendMessage({ type: "POLISH_CANCEL" });
});

// --- Timer ---

function startLocalTimer(startSeconds) {
  let seconds = startSeconds;
  stopLocalTimer();
  timerInterval = setInterval(() => {
    seconds++;
    showPolishingStatus(seconds);
  }, 1000);
}

function stopLocalTimer() {
  if (timerInterval) {
    clearInterval(timerInterval);
    timerInterval = null;
  }
}

// --- Storage change listener (reacts to background writes) ---

chrome.storage.onChanged.addListener((changes) => {
  if (changes.polishJob) {
    const job = changes.polishJob.newValue;
    if (!job) return;
    handleJobUpdate(job);
  }
  if (changes.history) {
    renderHistory(changes.history.newValue || []);
  }
});

function handleJobUpdate(job) {
  if (job.status === "completed") {
    stopLocalTimer();
    hideProgress();
    resultBox.textContent = job.result.polished;
    elapsedEl.textContent = `${(job.result.elapsed_ms / 1000).toFixed(1)}s`;
    resultSection.classList.remove("hidden");
    hideStatus();
    btnCancel.classList.add("hidden");
    btnPolish.disabled = false;
  } else if (job.status === "failed") {
    stopLocalTimer();
    hideProgress();
    showStatus(job.error || "Request failed.", "error");
    btnCancel.classList.add("hidden");
    btnPolish.disabled = false;
  } else if (job.status === "cancelled") {
    stopLocalTimer();
    hideProgress();
    showStatus("Cancelled.", "cancelled");
    btnCancel.classList.add("hidden");
    btnPolish.disabled = false;
  }
}

// --- Tick listener (best-effort timer from background) ---

chrome.runtime.onMessage.addListener((msg) => {
  if (msg.type === "POLISH_TICK") {
    showPolishingStatus(msg.seconds);
  }
});

// --- Job recovery on popup open ---

async function recoverJobState() {
  const { polishJob } = await chrome.storage.local.get("polishJob");
  if (!polishJob) return;

  if (polishJob.status === "running") {
    const elapsed = Date.now() - polishJob.startedAt;
    if (elapsed > STALE_TIMEOUT_MS) {
      // Stale job — mark failed, but don't clutter the UI
      await chrome.storage.local.set({
        polishJob: { status: "failed", error: "Request timed out." },
      });
    } else {
      // Resume timer UI — only case where we show state on reopen
      const seconds = Math.floor(elapsed / 1000);
      estimatedSeconds = Math.max(1, Math.round(((polishJob.inputText || "").length * MS_PER_CHAR * 1.15) / 1000));
      showPolishingStatus(seconds);
      startLocalTimer(seconds);
      btnPolish.disabled = true;
      btnCancel.classList.remove("hidden");
    }
  } else {
    // Completed/failed/cancelled — clean interface, result lives in history
    await chrome.storage.local.remove("polishJob");
  }
}

// --- History ---

async function loadHistory() {
  const { history = [] } = await chrome.storage.local.get("history");
  renderHistory(history);
}

function renderHistory(history) {
  if (history.length === 0) {
    historySection.classList.add("hidden");
    return;
  }

  historySection.classList.remove("hidden");
  historyCount.textContent = `(${history.length})`;
  historyList.innerHTML = "";

  for (const entry of history) {
    const item = document.createElement("div");
    item.className = "history-item";
    item.addEventListener("click", () => showHistoryDetail(entry));

    const text = document.createElement("div");
    text.className = "history-item-text";
    text.textContent = entry.output;

    const meta = document.createElement("div");
    meta.className = "history-item-meta";
    meta.textContent = `${entry.model} · ${formatRelativeTime(entry.timestamp)}`;

    item.appendChild(text);
    item.appendChild(meta);
    historyList.appendChild(item);
  }
}

function showHistoryDetail(entry) {
  historyDetailMeta.textContent = `${entry.model} · ${formatRelativeTime(entry.timestamp)} · ${(entry.elapsed_ms / 1000).toFixed(1)}s`;
  historyDetailInput.textContent = entry.input;
  historyDetailOutput.textContent = entry.output;
  historyDetail.classList.remove("hidden");

  // Hide main content sections
  document.querySelector(".input-group").classList.add("hidden");
  document.querySelector(".controls").classList.add("hidden");
  statusEl.classList.add("hidden");
  progressSection.classList.add("hidden");
  resultSection.classList.add("hidden");
  historySection.classList.add("hidden");
}

btnHistoryBack.addEventListener("click", () => {
  historyDetail.classList.add("hidden");

  // Restore main content sections
  document.querySelector(".input-group").classList.remove("hidden");
  document.querySelector(".controls").classList.remove("hidden");
  loadHistory();
});

btnHistoryCopy.addEventListener("click", async () => {
  const text = historyDetailOutput.textContent;
  if (!text) return;

  await navigator.clipboard.writeText(text);

  iconHistoryCopy.classList.add("hidden");
  iconHistoryCheck.classList.remove("hidden");
  historyCopyLabel.textContent = "Copied";

  setTimeout(() => {
    iconHistoryCheck.classList.add("hidden");
    iconHistoryCopy.classList.remove("hidden");
    historyCopyLabel.textContent = "Copy";
  }, 1500);
});

function formatRelativeTime(timestamp) {
  const diff = Date.now() - timestamp;
  const seconds = Math.floor(diff / 1000);
  if (seconds < 60) return "just now";
  const minutes = Math.floor(seconds / 60);
  if (minutes < 60) return `${minutes}m ago`;
  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `${hours}h ago`;
  const days = Math.floor(hours / 24);
  return `${days}d ago`;
}

// --- Copy ---

btnCopy.addEventListener("click", async () => {
  const text = resultBox.textContent;
  if (!text) return;

  await navigator.clipboard.writeText(text);

  // Swap to check icon
  iconCopy.classList.add("hidden");
  iconCheck.classList.remove("hidden");
  copyLabel.textContent = "Copied";

  setTimeout(() => {
    iconCheck.classList.add("hidden");
    iconCopy.classList.remove("hidden");
    copyLabel.textContent = "Copy";
  }, 1500);
});

// --- Settings panel ---

btnSettings.addEventListener("click", () => {
  document.body.classList.add("settings-open");
});

btnBack.addEventListener("click", () => {
  document.body.classList.remove("settings-open");
  settingsStatus.textContent = "";
  settingsStatus.className = "settings-status";
});

async function loadSettings() {
  const result = await chrome.storage.local.get(["apiUrl", "apiKey"]);
  apiUrlInput.value = result.apiUrl || "http://localhost:8090";
  apiKeyInput.value = result.apiKey || "";
}

btnTest.addEventListener("click", async () => {
  settingsStatus.textContent = "Testing...";
  settingsStatus.className = "settings-status";
  btnTest.disabled = true;

  try {
    const url = apiUrlInput.value.trim().replace(/\/+$/, "");
    const controller = new AbortController();
    const timeout = setTimeout(() => controller.abort(), 5000);

    try {
      const resp = await fetch(`${url}/api/health`, { signal: controller.signal });
      clearTimeout(timeout);
      if (!resp.ok) throw new Error(`Status ${resp.status}`);
      const data = await resp.json();
      if (data.status === "ok") {
        settingsStatus.textContent = "Connected.";
        settingsStatus.className = "settings-status ok";
      } else {
        throw new Error("Unexpected response");
      }
    } catch (err) {
      clearTimeout(timeout);
      if (err.name === "AbortError") {
        settingsStatus.textContent = "Timed out after 5s.";
      } else {
        settingsStatus.textContent = `Failed: ${err.message}`;
      }
      settingsStatus.className = "settings-status err";
    }
  } finally {
    btnTest.disabled = false;
  }
});

btnSave.addEventListener("click", async () => {
  const url = apiUrlInput.value.trim().replace(/\/+$/, "");
  const key = apiKeyInput.value.trim();
  await chrome.storage.local.set({ apiUrl: url, apiKey: key });

  settingsStatus.textContent = "Saved. Reloading models...";
  settingsStatus.className = "settings-status ok";

  // Reload models with new URL
  await loadModels();

  if (modelsLoaded) {
    settingsStatus.textContent = "Saved.";
    settingsStatus.className = "settings-status ok";
  } else {
    settingsStatus.textContent = "Saved, but cannot reach API.";
    settingsStatus.className = "settings-status err";
  }

  setTimeout(() => {
    settingsStatus.textContent = "";
    settingsStatus.className = "settings-status";
  }, 3000);
});

// --- Helpers ---

function showPolishingStatus(seconds) {
  const pct = estimatedSeconds > 0
    ? Math.min(99, Math.round((seconds / estimatedSeconds) * 100))
    : 0;
  progressPct.textContent = `${pct}%`;
  progressElapsed.textContent = `${seconds}s`;
  progressFill.style.width = `${pct}%`;
  progressSection.classList.remove("hidden");
  hideStatus();
}

function hideProgress() {
  progressSection.classList.add("hidden");
  progressFill.style.width = "0%";
}

function showStatus(msg, type) {
  statusEl.textContent = msg;
  statusEl.className = `status ${type}`;
}

function hideStatus() {
  statusEl.className = "status hidden";
}
