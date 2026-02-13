// Pollex API client — shared HTTP layer for the extension.
// All calls go through the Go backend; no API keys stored here.

const DEFAULT_API_URL = "http://localhost:8090";
const POLISH_TIMEOUT_MS = 70000;

async function getApiUrl() {
  const result = await chrome.storage.local.get("apiUrl");
  return result.apiUrl || DEFAULT_API_URL;
}

async function fetchHealth(signal) {
  const base = await getApiUrl();
  const opts = signal ? { signal } : {};
  const resp = await fetch(`${base}/api/health`, opts);
  if (!resp.ok) throw new Error(`Health check failed: ${resp.status}`);
  return resp.json();
}

async function fetchModels() {
  const base = await getApiUrl();
  const resp = await fetch(`${base}/api/models`);
  if (!resp.ok) throw new Error(`Failed to load models: ${resp.status}`);
  return resp.json();
}

async function fetchPolish(text, modelId, signal) {
  const base = await getApiUrl();

  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), POLISH_TIMEOUT_MS);

  // If caller provides a signal, abort our controller when theirs aborts
  if (signal) {
    signal.addEventListener("abort", () => controller.abort());
  }

  try {
    const resp = await fetch(`${base}/api/polish`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ text, model_id: modelId }),
      signal: controller.signal,
    });

    if (!resp.ok) {
      const body = await resp.json().catch(() => ({}));
      throw new Error(body.error || `Request failed: ${resp.status}`);
    }

    return resp.json();
  } finally {
    clearTimeout(timeout);
  }
}
