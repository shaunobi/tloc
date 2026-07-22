# Contributing

Create a focused branch and keep each commit buildable. Before opening a change,
run the formatter, static checks, unit tests, and the small integration fixture.

Pull requests should explain the user-visible behavior, the alternatives that
were considered, and how the change was verified. Add regression coverage for
bug fixes. Generated files belong in the same commit as their source, with the
generation command noted in the description.

## Review checklist

- Public names and error messages are clear and stable.
- New I/O has cancellation, size limits, and useful context on failure.
- Concurrent code has deterministic results and no shared mutable state.
- Documentation describes defaults and important safety constraints.

Please avoid drive-by refactors in functional changes. Small preparatory cleanup
is welcome as a separate commit when it makes the main diff easier to review.
