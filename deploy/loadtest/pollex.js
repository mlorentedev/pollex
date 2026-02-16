// Pollex — k6 Load Test
//
// Usage:
//   k6 run deploy/loadtest/pollex.js                                    # local mock
//   k6 run -e BASE_URL=https://pollex.mlorente.dev -e API_KEY=xxx deploy/loadtest/pollex.js
//   k6 run -e SCENARIO=burst deploy/loadtest/pollex.js                  # single scenario
//   k6 run -e SCENARIO=jetson deploy/loadtest/pollex.js                 # single-user Jetson
//   k6 run -e SCENARIO=soak deploy/loadtest/pollex.js                   # 30 min soak

import http from "k6/http";
import { check, sleep } from "k6";
import { Rate, Trend } from "k6/metrics";

// Custom metrics
const polishDuration = new Trend("polish_duration_ms", true);
const errorRate = new Rate("errors");

// Configuration
const BASE_URL = __ENV.BASE_URL || "http://localhost:8090";
const API_KEY = __ENV.API_KEY || "";
const SCENARIO = __ENV.SCENARIO || "all";

// Test data — realistic text samples ordered by size
const samples = [
  "i goes to store yesterday and buyed some foods",
  "The meeting was really good and we discussed about many things that is important for the project moving forward",
  "When I first started working on this project, I was not sure how to approach the problem. After doing some research and talking to my colleagues, I realized that the best approach would be to break it down into smaller pieces and tackle each one individually. This strategy has proven to be very effective.",
];

// Scenario definitions
const scenarios = {
  normal: {
    executor: "constant-arrival-rate",
    rate: 12,
    timeUnit: "1m",
    duration: "2m",
    preAllocatedVUs: 2,
    maxVUs: 5,
    exec: "polishFlow",
  },
  burst: {
    executor: "shared-iterations",
    vus: 5,
    iterations: 25,
    maxDuration: "2m",
    exec: "polishFlow",
  },
  soak: {
    executor: "constant-vus",
    vus: 1,
    duration: "30m",
    exec: "polishFlow",
  },
  // Single-user scenarios for Jetson Nano (1 GPU, sequential inference)
  jetson: {
    executor: "constant-vus",
    vus: 1,
    duration: "5m",
    exec: "polishFlow",
  },
};

// Select scenarios based on SCENARIO env var
function getScenarios() {
  if (SCENARIO === "all") {
    return { normal: scenarios.normal, burst: scenarios.burst };
  }
  if (scenarios[SCENARIO]) {
    return { [SCENARIO]: scenarios[SCENARIO] };
  }
  return { normal: scenarios.normal };
}

// Thresholds adapt to target: Jetson has Cloudflare Tunnel latency on health
const isJetson = SCENARIO === "jetson" || SCENARIO === "soak";
const healthP95 = isJetson ? "p(95)<2000" : "p(95)<500";

export const options = {
  scenarios: getScenarios(),
  thresholds: {
    http_req_duration: ["p(50)<20000", "p(95)<60000"], // SLO: p50 < 20s, p95 < 60s
    errors: ["rate<0.01"],                              // SLO: error rate < 1%
    "http_req_duration{name:health}": [healthP95],
  },
};

function headers() {
  const h = { "Content-Type": "application/json" };
  if (API_KEY) {
    h["X-API-Key"] = API_KEY;
  }
  return h;
}

// Health check (no auth required)
export function healthCheck() {
  const res = http.get(`${BASE_URL}/api/health`, {
    tags: { name: "health" },
  });
  check(res, {
    "health status 200": (r) => r.status === 200,
    "health has adapters": (r) => {
      try {
        return JSON.parse(r.body).adapters !== undefined;
      } catch {
        return false;
      }
    },
  });
}

// Polish flow — the main workload
export function polishFlow() {
  // 1. Health check (send API key to bypass rate limiter)
  const healthRes = http.get(`${BASE_URL}/api/health`, {
    headers: headers(),
    tags: { name: "health" },
  });
  check(healthRes, { "health ok": (r) => r.status === 200 });

  // 2. Get models
  const modelsRes = http.get(`${BASE_URL}/api/models`, {
    headers: headers(),
    tags: { name: "models" },
  });
  const modelsOk = check(modelsRes, { "models ok": (r) => r.status === 200 });

  if (!modelsOk) {
    errorRate.add(1);
    return;
  }

  // Pick first available model (API returns a flat array of ModelInfo)
  let modelId = "";
  try {
    const models = JSON.parse(modelsRes.body);
    if (Array.isArray(models) && models.length > 0) {
      modelId = models[0].id;
    }
  } catch {
    errorRate.add(1);
    return;
  }

  // 3. Polish a random sample
  const text = samples[Math.floor(Math.random() * samples.length)];
  const payload = JSON.stringify({ text: text, model_id: modelId });

  const polishRes = http.post(`${BASE_URL}/api/polish`, payload, {
    headers: headers(),
    tags: { name: "polish" },
    timeout: "120s",
  });

  const polishOk = check(polishRes, {
    "polish status 200": (r) => r.status === 200,
    "polish has result": (r) => {
      try {
        return JSON.parse(r.body).polished !== undefined;
      } catch {
        return false;
      }
    },
  });

  if (polishOk) {
    try {
      const body = JSON.parse(polishRes.body);
      polishDuration.add(body.elapsed_ms || polishRes.timings.duration);
    } catch {
      polishDuration.add(polishRes.timings.duration);
    }
    errorRate.add(0);
  } else {
    errorRate.add(1);
  }

  sleep(1);
}

export default polishFlow;
