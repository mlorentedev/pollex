// Vitest setup — in-memory mocks of the chrome.* APIs used by the extension,
// plus a fetch stub. Exported helpers let individual tests seed/assert storage.

const store = new Map();

export const storageMock = {
  local: {
    async get(key) {
      if (key === undefined) return Object.fromEntries(store);
      if (typeof key === "string") return { [key]: store.get(key) };
      if (Array.isArray(key)) {
        const out = {};
        for (const k of key) if (store.has(k)) out[k] = store.get(k);
        return out;
      }
      if (typeof key === "object" && key !== null) {
        const out = {};
        for (const [k, v] of Object.entries(key)) {
          out[k] = store.has(k) ? store.get(k) : v;
        }
        return out;
      }
      return {};
    },
    async set(items) {
      for (const [k, v] of Object.entries(items)) store.set(k, structuredClone(v));
    },
    async remove(key) {
      if (Array.isArray(key)) for (const k of key) store.delete(k);
      else store.delete(key);
    },
    async clear() {
      store.clear();
    },
  },
};

export const runtimeMock = {
  onMessage: {
    addListener: () => {},
  },
  sendMessage: async () => {},
};

export const fetchMock = {
  impl: () => {
    throw new Error("fetchMock not configured for this test");
  },
};

export function seedStorage(items) {
  for (const [k, v] of Object.entries(items)) store.set(k, structuredClone(v));
}

export function dumpStorage() {
  return Object.fromEntries(store);
}

export function clearStorage() {
  store.clear();
}

// Install globals before modules under test import them.
globalThis.chrome = {
  storage: storageMock,
  runtime: runtimeMock,
};

globalThis.fetch = (...args) => fetchMock.impl(...args);
