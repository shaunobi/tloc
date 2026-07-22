namespace Example.Calibration;

public sealed class RouteMatcher
{
    private readonly Dictionary<string, Func<string, ValueTask<string>>> _handlers =
        new(StringComparer.OrdinalIgnoreCase);

    public void Register(string method, string path, Func<string, ValueTask<string>> handler)
    {
        ArgumentException.ThrowIfNullOrWhiteSpace(method);
        ArgumentException.ThrowIfNullOrWhiteSpace(path);
        ArgumentNullException.ThrowIfNull(handler);
        _handlers[Key(method, path)] = handler;
    }

    public async ValueTask<string> DispatchAsync(string method, Uri request, string body)
    {
        ArgumentNullException.ThrowIfNull(request);
        if (!_handlers.TryGetValue(Key(method, request.AbsolutePath), out var handler))
        {
            return $"404: no route for {method.ToUpperInvariant()} {request.AbsolutePath}";
        }

        return await handler(body).ConfigureAwait(false);
    }

    private static string Key(string method, string path) =>
        $"{method.Trim()}:{path.TrimEnd('/')}";
}
