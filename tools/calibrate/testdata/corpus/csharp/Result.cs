namespace Example.Calibration;

public readonly record struct Result<T>(T? Value, string? Error)
{
    public bool IsSuccess => Error is null;

    public static Result<T> Success(T value) => new(value, null);

    public static Result<T> Failure(string message) =>
        new(default, string.IsNullOrWhiteSpace(message)
            ? throw new ArgumentException("An error message is required.", nameof(message))
            : message);

    public Result<TResult> Map<TResult>(Func<T, TResult> transform)
    {
        ArgumentNullException.ThrowIfNull(transform);
        return IsSuccess
            ? Result<TResult>.Success(transform(Value!))
            : Result<TResult>.Failure(Error!);
    }

    public T Unwrap() => IsSuccess
        ? Value!
        : throw new InvalidOperationException($"Cannot unwrap failed result: {Error}");
}
