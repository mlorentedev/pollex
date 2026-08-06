// E2E test for the Pollex Chrome extension (MV3, load-unpacked).
// Launches Chromium (full build — required for extensions) with the extension
// directory, opens the popup page, and verifies the real polish flow against
// the mock API started in global-setup.js.

import { test, expect, chromium } from "@playwright/test";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { MOCK_API_PORT } from "./global-setup.js";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const EXTENSION_PATH = path.resolve(__dirname, "../..");
const API_URL = `http://127.0.0.1:${MOCK_API_PORT}`;

// Chromium of Playwright (not headless shell) + the extension as load-unpacked.
async function launchExtensionContext() {
  const context = await chromium.launchPersistentContext("", {
    headless: false, // extensions require the full (headed) Chromium build
    args: [
      `--disable-extensions-except=${EXTENSION_PATH}`,
      `--load-extension=${EXTENSION_PATH}`,
    ],
  });

  // Grab the service worker to resolve the extension ID.
  let [sw] = context.serviceWorkers();
  if (!sw) sw = await context.waitForEvent("serviceworker", { timeout: 15_000 });
  const extensionId = new URL(sw.url()).host;
  return { context, extensionId };
}

// Seed chrome.storage so the popup points at the mock API.
async function seedApiUrl(sw, url) {
  await sw.evaluate((apiUrl) => {
    chrome.storage.local.set({ apiUrl });
  }, url);
}

test.describe("Pollex extension popup", () => {
  test("loads models and polishes text end to end", async () => {
    const { context, extensionId } = await launchExtensionContext();
    const [sw] = context.serviceWorkers();
    await seedApiUrl(sw, API_URL);

    const page = await context.newPage();
    await page.goto(`chrome-extension://${extensionId}/popup.html`);

    // Models load: single-model mode shows the static label with the mock name.
    await expect(page.locator("#model-label")).toBeVisible({ timeout: 10_000 });
    await expect(page.locator("#model-label")).toContainText("Mock (dev)");
    await expect(page.locator("#btn-polish")).toBeEnabled();

    // Server version banner from /api/health.
    await expect(page.locator("#server-version")).toContainText("API dev", {
      timeout: 10_000,
    });

    // Type text and polish.
    await page.fill("#input", "me and my team was working on this project for two months.");
    await page.click("#btn-polish");

    // Progress appears, then the result section.
    await expect(page.locator("#result-section")).toBeVisible({ timeout: 15_000 });
    const result = await page.locator("#result").textContent();
    expect(result.trim().length).toBeGreaterThan(0);

    // History records the entry.
    await expect(page.locator("#history-section")).toBeVisible();
    await expect(page.locator(".history-item").first()).toBeVisible();

    await context.close();
  });

  test("shows a clear error when the API is unreachable", async () => {
    const { context, extensionId } = await launchExtensionContext();
    const [sw] = context.serviceWorkers();
    // Point at a port with nothing listening.
    await seedApiUrl(sw, "http://127.0.0.1:1");

    const page = await context.newPage();
    await page.goto(`chrome-extension://${extensionId}/popup.html`);

    await expect(page.locator("#status")).toContainText("Cannot reach API", {
      timeout: 10_000,
    });
    await expect(page.locator("#btn-polish")).toBeDisabled();

    await context.close();
  });
});
