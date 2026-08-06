// Playwright global setup — boots the Pollex API in mock mode on a dedicated
// port (MOCK_API_PORT) so the e2e suite never collides with a dev server on
// :8090, and waits until /api/health is green before tests start.

import { spawn, execFileSync } from "node:child_process";
import { fileURLToPath } from "node:url";
import path from "node:path";
import fs from "node:fs";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
export const MOCK_API_PORT = 8099;

async function waitForHealth(url, timeoutMs = 30_000) {
  const deadline = Date.now() + timeoutMs;
  while (Date.now() < deadline) {
    try {
      const resp = await fetch(url);
      if (resp.ok) {
        const body = await resp.json();
        if (body.status === "ok") return true;
      }
    } catch {
      // not up yet
    }
    await new Promise((r) => setTimeout(r, 500));
  }
  return false;
}

export default async function globalSetup() {
  const repoRoot = path.resolve(__dirname, "../../..");
  const url = `http://127.0.0.1:${MOCK_API_PORT}/api/health`;

  // Already up (e.g. a dev server)? Use it.
  if (await waitForHealth(url, 2000)) return;

  // Build once so we don't rely on `go run` wrapper or a shell PATH.
  const bin = path.join(repoRoot, "dist", "pollex-e2e");
  fs.mkdirSync(path.dirname(bin), { recursive: true });
  execFileSync("go", ["build", "-o", bin, "./cmd/pollex"], {
    cwd: repoRoot,
    stdio: "inherit",
  });

  const proc = spawn(bin, ["--mock", "--port", String(MOCK_API_PORT)], {
    cwd: repoRoot,
    stdio: "pipe",
    env: { ...process.env },
  });
  proc.stdout?.on("data", (d) => process.stdout.write(`[pollex-e2e] ${d}`));
  proc.stderr?.on("data", (d) => process.stderr.write(`[pollex-e2e] ${d}`));

  if (!(await waitForHealth(url))) {
    throw new Error(`Pollex mock API did not become healthy at ${url}`);
  }

  // Persist for globalTeardown.
  global.__POLLEX_E2E_PROC = proc;
  global.__POLLEX_E2E_URL = url;
}
