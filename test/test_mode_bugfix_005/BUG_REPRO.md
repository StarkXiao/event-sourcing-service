# Bug Reproduction

Baseline: remote `origin/base_bug_005`.

Category: context. The projection wait ignores cancellation and returns success after an unconditional sleep.

Verify:

`仅执行命令，不修改文件： go test ./internal/application -count=1 -run '^TestBug005_WaitForProjectionHonorsCancellation$'`
