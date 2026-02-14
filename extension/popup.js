// Pollex popup — wires UI to api.js

const MAX_CHARS = 10000;
const WARN_THRESHOLD = 0.9;

// --- DOM refs ---

const input = document.getElementById("input");
const charCount = document.getElementById("char-count");
const modelSelect = document.getElementById("model-select");
const btnPolish = document.getElementById("btn-polish");
const btnCancel = document.getElementById("btn-cancel");
const btnCopy = document.getElementById("btn-copy");
const iconCopy = document.getElementById("icon-copy");
const iconCheck = document.getElementById("icon-check");
const copyLabel = document.getElementById("copy-label");
const btnSettings = document.getElementById("btn-settings");
const statusEl = document.getElementById("status");
const resultSection = document.getElementById("result-section");
const resultBox = document.getElementById("result");
const elapsedEl = document.getElementById("elapsed");

// Settings panel
const btnBack = document.getElementById("btn-back");
const apiUrlInput = document.getElementById("api-url");
const apiKeyInput = document.getElementById("api-key");
const btnTest = document.getElementById("btn-test");
const btnSave = document.getElementById("btn-save");
const settingsStatus = document.getElementById("settings-status");

let abortController = null;
let timerInterval = null;
let modelsLoaded = false;

// --- Init ---

document.addEventListener("DOMContentLoaded", async () => {
  await loadModels();
  await loadSettings();
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

    btnPolish.disabled = false;
    modelsLoaded = true;
  } catch {
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

// --- Character count ---

input.addEventListener("input", () => {
  updateCharCount();
  clearTransientStatus();
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
}

function clearTransientStatus() {
  // Dismiss error/cancelled messages when user starts editing
  if (statusEl.classList.contains("error") || statusEl.classList.contains("cancelled")) {
    hideStatus();
  }
}

// --- Polish ---

btnPolish.addEventListener("click", doPolish);

input.addEventListener("keydown", (e) => {
  if ((e.ctrlKey || e.metaKey) && e.key === "Enter") {
    e.preventDefault();
    if (!btnPolish.disabled) doPolish();
  }
});

async function doPolish() {
  const text = input.value.trim();
  if (!text) return;

  const modelId = modelSelect.value;
  if (!modelId) return;

  // Reset UI
  resultSection.classList.add("hidden");
  btnPolish.disabled = true;
  btnCancel.classList.remove("hidden");

  // Start timer
  let seconds = 0;
  showStatus("Polishing... 0s", "polishing");
  timerInterval = setInterval(() => {
    seconds++;
    showStatus(`Polishing... ${seconds}s`, "polishing");
  }, 1000);

  abortController = new AbortController();

  try {
    const result = await fetchPolish(text, modelId, abortController.signal);

    resultBox.textContent = result.polished;
    elapsedEl.textContent = `${(result.elapsed_ms / 1000).toFixed(1)}s`;
    resultSection.classList.remove("hidden");
    hideStatus();
  } catch (err) {
    if (err.name === "AbortError") {
      showStatus("Cancelled.", "cancelled");
    } else {
      showStatus(err.message, "error");
    }
  } finally {
    clearInterval(timerInterval);
    abortController = null;
    btnCancel.classList.add("hidden");
    btnPolish.disabled = false;
  }
}

// --- Cancel ---

btnCancel.addEventListener("click", () => {
  if (abortController) abortController.abort();
});

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

function showStatus(msg, type) {
  statusEl.textContent = msg;
  statusEl.className = `status ${type}`;
}

function hideStatus() {
  statusEl.className = "status hidden";
}
