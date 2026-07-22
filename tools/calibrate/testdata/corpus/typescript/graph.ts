export type NodeId = string & { readonly __brand: "NodeId" };

export class DirectedGraph<T> {
  private readonly values = new Map<NodeId, T>();
  private readonly edges = new Map<NodeId, Set<NodeId>>();

  add(id: NodeId, value: T): void {
    this.values.set(id, value);
    this.edges.set(id, this.edges.get(id) ?? new Set());
  }

  connect(from: NodeId, to: NodeId): void {
    if (!this.values.has(from) || !this.values.has(to)) {
      throw new Error(`cannot connect missing node ${from} -> ${to}`);
    }
    this.edges.get(from)!.add(to);
  }

  *walk(start: NodeId): IterableIterator<readonly [NodeId, T]> {
    const pending = [start];
    const visited = new Set<NodeId>();
    while (pending.length > 0) {
      const id = pending.shift()!;
      if (visited.has(id)) continue;
      visited.add(id);
      const value = this.values.get(id);
      if (value === undefined) continue;
      yield [id, value] as const;
      pending.push(...(this.edges.get(id) ?? []));
    }
  }
}
