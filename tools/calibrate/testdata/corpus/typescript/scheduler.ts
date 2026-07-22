export interface ScheduledJob<T> {
  readonly name: string;
  readonly run: (signal: AbortSignal) => Promise<T>;
}

export async function runWithDeadline<T>(
  job: ScheduledJob<T>,
  timeoutMs: number,
  parentSignal?: AbortSignal,
): Promise<T> {
  if (!Number.isFinite(timeoutMs) || timeoutMs <= 0) {
    throw new RangeError("timeoutMs must be positive");
  }

  const controller = new AbortController();
  const timeout = setTimeout(
    () => controller.abort(new Error(`job ${job.name} exceeded ${timeoutMs}ms`)),
    timeoutMs,
  );
  const cancel = () => controller.abort(parentSignal?.reason);
  parentSignal?.addEventListener("abort", cancel, { once: true });

  try {
    return await job.run(controller.signal);
  } catch (cause) {
    throw new Error(`scheduled job ${job.name} failed`, { cause });
  } finally {
    clearTimeout(timeout);
    parentSignal?.removeEventListener("abort", cancel);
  }
}
