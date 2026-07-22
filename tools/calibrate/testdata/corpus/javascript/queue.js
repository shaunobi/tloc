export class TaskQueue {
  #pending = [];
  #running = 0;

  constructor(concurrency = 2) {
    if (!Number.isInteger(concurrency) || concurrency < 1) {
      throw new RangeError("concurrency must be a positive integer");
    }
    this.concurrency = concurrency;
  }

  add(task) {
    return new Promise((resolve, reject) => {
      this.#pending.push({ task, resolve, reject });
      queueMicrotask(() => this.#drain());
    });
  }

  #drain() {
    while (this.#running < this.concurrency && this.#pending.length > 0) {
      const item = this.#pending.shift();
      this.#running += 1;
      Promise.resolve()
        .then(item.task)
        .then(item.resolve, item.reject)
        .finally(() => {
          this.#running -= 1;
          this.#drain();
        });
    }
  }
}

// Preserve input order even when jobs finish out of order.
export async function mapConcurrent(values, mapper, concurrency = 2) {
  const queue = new TaskQueue(concurrency);
  return Promise.all(values.map((value, index) => queue.add(() => mapper(value, index))));
}
