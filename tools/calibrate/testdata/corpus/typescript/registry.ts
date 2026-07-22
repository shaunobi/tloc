export interface Identified {
  readonly id: string;
  readonly updatedAt: Date;
}

type Listener<T> = (next: Readonly<T>, previous: Readonly<T> | undefined) => void;

export class Registry<T extends Identified> {
  private readonly records = new Map<string, T>();
  private readonly listeners = new Set<Listener<T>>();

  subscribe(listener: Listener<T>): () => void {
    this.listeners.add(listener);
    return () => this.listeners.delete(listener);
  }

  upsert(record: T): boolean {
    const previous = this.records.get(record.id);
    if (previous && previous.updatedAt >= record.updatedAt) return false;
    this.records.set(record.id, record);
    for (const listener of this.listeners) listener(record, previous);
    return true;
  }

  query(predicate: (record: Readonly<T>) => boolean): readonly T[] {
    return [...this.records.values()]
      .filter(predicate)
      .sort((left, right) => left.id.localeCompare(right.id));
  }
}
