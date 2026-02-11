// Pollex settings page

const apiUrlInput = document.getElementById("api-url");
const btnTest = document.getElementById("btn-test");
const btnSave = document.getElementById("btn-save");
const statusEl = document.getElementById("status");

// Load saved URL on open
document.addEventListener("DOMContentLoaded", async () => {
  const result = await chrome.storage.local.get("apiUrl");
  apiUrlInput.value = result.apiUrl || "http://localhost:8090";
});

// Test connection
btnTest.addEventListener("click", async () => {
  statusEl.textContent = "Testing...";
  statusEl.className = "";

  const url = apiUrlInput.value.trim().replace(/\/+$/, "");
  try {
    const resp = await fetch(`${url}/api/health`);
    if (!resp.ok) throw new Error(`Status ${resp.status}`);
    const data = await resp.json();
    if (data.status === "ok") {
      statusEl.textContent = "Connected.";
      statusEl.className = "ok";
    } else {
      throw new Error("Unexpected response");
    }
  } catch (err) {
    statusEl.textContent = `Failed: ${err.message}`;
    statusEl.className = "err";
  }
});

// Save
btnSave.addEventListener("click", async () => {
  const url = apiUrlInput.value.trim().replace(/\/+$/, "");
  await chrome.storage.local.set({ apiUrl: url });
  statusEl.textContent = "Saved.";
  statusEl.className = "ok";
  setTimeout(() => { statusEl.textContent = ""; }, 2000);
});
