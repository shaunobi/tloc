export class Histogram {
  #buckets;
  #counts;
  #total = 0;

  constructor(boundaries = [5, 10, 25, 50, 100]) {
    if (!boundaries.every((value, index) => value > (boundaries[index - 1] ?? 0))) {
      throw new TypeError("boundaries must be positive and increasing");
    }
    this.#buckets = [...boundaries];
    this.#counts = new Array(boundaries.length + 1).fill(0);
  }

  observe(value) {
    if (!Number.isFinite(value) || value < 0) {
      throw new RangeError(`invalid observation: ${value}`);
    }
    const index = this.#buckets.findIndex((limit) => value <= limit);
    this.#counts[index < 0 ? this.#buckets.length : index] += 1;
    this.#total += value;
  }

  snapshot() {
    const labels = [...this.#buckets.map(String), "+Inf"];
    return Object.freeze({
      buckets: Object.fromEntries(labels.map((label, i) => [label, this.#counts[i]])),
      count: this.#counts.reduce((sum, count) => sum + count, 0),
      total: this.#total,
    });
  }
}
