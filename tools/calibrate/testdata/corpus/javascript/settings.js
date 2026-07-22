const DEFAULTS = Object.freeze({ theme: "system", pageSize: 25, alerts: true });

export function parseSettings(text) {
  let candidate;
  try {
    candidate = JSON.parse(text);
  } catch (cause) {
    throw new SyntaxError("settings must contain valid JSON", { cause });
  }
  if (candidate === null || Array.isArray(candidate) || typeof candidate !== "object") {
    throw new TypeError("settings must be a JSON object");
  }

  const settings = { ...DEFAULTS, ...candidate };
  if (!["light", "dark", "system"].includes(settings.theme)) {
    throw new RangeError(`unsupported theme: ${settings.theme}`);
  }
  settings.pageSize = Number(settings.pageSize);
  if (!Number.isInteger(settings.pageSize) || settings.pageSize < 1) {
    throw new RangeError("pageSize must be a positive integer");
  }

  // Unknown keys are retained so newer producers remain forward-compatible.
  return Object.freeze(settings);
}

export function changedKeys(before, after) {
  return [...new Set([...Object.keys(before), ...Object.keys(after)])]
    .filter((key) => !Object.is(before[key], after[key]))
    .sort();
}
