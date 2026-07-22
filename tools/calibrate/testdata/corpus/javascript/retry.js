const sleep = (milliseconds, signal) =>
  new Promise((resolve, reject) => {
    const timer = setTimeout(resolve, milliseconds);
    signal?.addEventListener(
      "abort",
      () => {
        clearTimeout(timer);
        reject(signal.reason ?? new Error("operation aborted"));
      },
      { once: true },
    );
  });

export async function retry(operation, options = {}) {
  const {
    attempts = 3,
    baseDelay = 50,
    signal,
    retryable = () => true,
  } = options;

  let lastError;
  for (let attempt = 1; attempt <= attempts; attempt += 1) {
    signal?.throwIfAborted();
    try {
      return await operation({ attempt, signal });
    } catch (error) {
      lastError = error;
      if (attempt === attempts || !retryable(error)) break;
      // Capped exponential delay with deterministic behavior for tests.
      await sleep(Math.min(baseDelay * 2 ** (attempt - 1), 2_000), signal);
    }
  }

  throw new AggregateError([lastError], `operation failed after ${attempts} attempts`);
}
