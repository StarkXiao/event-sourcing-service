# Bug Reproduction

Baseline: remote `origin/base_bug_003`.

Category: slice. The event copy helper allocates one fewer element and drops the final event.

Verify:

`仅执行命令，不修改文件： go test ./internal/application -count=1 -run '^TestBug003_QueriesKeepAllEvents$'`
