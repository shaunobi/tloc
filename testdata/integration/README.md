# Integration fixture

`scc-files.txt` is the accepted-file oracle for `project/`. It was verified
against the pinned `github.com/boyter/scc/v3` v3.7.0 release with its default
`.gitignore`, `.ignore`, and `.sccignore` behavior. The integration test compares
tloc's selected paths to this oracle before asserting language and folder
rollups.

The ignored files deliberately exercise all three ignore sources, including a
directory excluded through `.gitignore`.
