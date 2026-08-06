import { defineConfig } from "@playwright/test";

export default defineConfig({
  testDir: "./tests/e2e",
  timeout: 60_000,
  fullyParallel: false,
  workers: 1, // extension e2e is not safe to parallelize in one profile
  retries: 0,
  reporter: [["list"]],
  globalSetup: "./tests/e2e/global-setup.js",
  globalTeardown: "./tests/e2e/global-teardown.js",
  use: {
    trace: "retain-on-failure",
  },
});
