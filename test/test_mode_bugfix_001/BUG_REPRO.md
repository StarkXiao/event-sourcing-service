# Bug Reproduction

Baseline: remote `origin/base_bug_001`.

Category: concurrency. The projection processing counter is updated concurrently without synchronization.

Verify:

`仅执行命令，不修改文件： go test ./internal/projection -race -count=1 -run '^TestBug001_ProjectionCountersRace$'`
