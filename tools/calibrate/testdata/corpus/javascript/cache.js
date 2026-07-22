export class ExpiringCache {
  #values = new Map();

  constructor(clock = () => Date.now()) {
    this.clock = clock;
  }

  set(key, value, ttlMilliseconds) {
    if (!Number.isFinite(ttlMilliseconds) || ttlMilliseconds <= 0) {
      throw new RangeError("ttlMilliseconds must be positive");
    }
    this.#values.set(key, {
      value,
      expiresAt: this.clock() + ttlMilliseconds,
    });
  }

  get(key) {
    const entry = this.#values.get(key);
    if (!entry || entry.expiresAt <= this.clock()) {
      this.#values.delete(key);
      return undefined;
    }
    return entry.value;
  }
}
