// Pollex popup — wires UI to api.js

const input = document.getElementById("input");
const modelSelect = document.getElementById("model-select");
const btnPolish = document.getElementById("btn-polish");
const btnCopy = document.getElementById("btn-copy");
const btnSettings = document.getElementById("btn-settings");
const status = document.getElementById("status");
const resultSection = document.getElementById("result-section");
const resultBox = document.getElementById("result");
const elapsed = document.getElementById("elapsed");

let abortController = null;
let timerInterval = null;

// --- Init ---

document.addEventListener("DOMContentLoaded", async () => {
  await loadModels();
  input.focus();
});

const providerLabels = {
  ollama: "Local",
  mock: "Local",
  claude: "Cloud",
};

async function loadModels() {
  try {
    const models = await fetchModels();
    modelSelect.innerHTML = "";

    // Group by provider category
    const groups = {};
    for (const m of models) {
      const label = providerLabels[m.provider] || m.provider;
      if (!groups[label]) groups[label] = [];
      groups[label].push(m);
    }

    const groupNames = Object.keys(groups);
    if (groupNames.length === 1) {
      // Single group — no optgroup needed
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
  } catch (err) {
    modelSelect.innerHTML = '<option value="">Cannot connect</option>';
    showStatus("Cannot reach API. Check settings.", "error");
  }
}

function makeOption(model) {
  const opt = document.createElement("option");
  opt.value = model.id;
  opt.textContent = model.name;
  return opt;
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
  btnPolish.textContent = "Cancel";
  btnPolish.disabled = false;

  // Start timer
  let seconds = 0;
  showStatus(`Polishing... 0s`, "polishing");
  timerInterval = setInterval(() => {
    seconds++;
    showStatus(`Polishing... ${seconds}s`, "polishing");
  }, 1000);

  // Switch button to cancel mode
  abortController = new AbortController();
  const cancelHandler = () => {
    abortController.abort();
  };
  btnPolish.removeEventListener("click", doPolish);
  btnPolish.addEventListener("click", cancelHandler);

  try {
    const result = await fetchPolish(text, modelId, abortController.signal);

    // Show result
    resultBox.textContent = result.polished;
    elapsed.textContent = `${(result.elapsed_ms / 1000).toFixed(1)}s`;
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

    // Restore polish button
    btnPolish.removeEventListener("click", cancelHandler);
    btnPolish.addEventListener("click", doPolish);
    btnPolish.textContent = "Polish";
    btnPolish.disabled = false;
  }
}

// --- Copy ---

btnCopy.addEventListener("click", async () => {
  const text = resultBox.textContent;
  if (!text) return;
  await navigator.clipboard.writeText(text);
  btnCopy.textContent = "Copied!";
  setTimeout(() => {
    btnCopy.textContent = "Copy";
  }, 1500);
});

// --- Settings ---

btnSettings.addEventListener("click", () => {
  chrome.runtime.openOptionsPage();
});

// --- Helpers ---

function showStatus(msg, type) {
  status.textContent = msg;
  status.className = `status ${type}`;
}

function hideStatus() {
  status.className = "status hidden";
}
