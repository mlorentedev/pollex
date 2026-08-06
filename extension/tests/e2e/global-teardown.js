// Playwright global teardown — stops the mock API started in global-setup.js.

export default async function globalTeardown() {
  const proc = global.__POLLEX_E2E_PROC;
  if (proc) {
    proc.kill("SIGTERM");
    await new Promise((r) => proc.once("exit", r));
    console.log("[pollex-e2e] mock API stopped");
  }
}
