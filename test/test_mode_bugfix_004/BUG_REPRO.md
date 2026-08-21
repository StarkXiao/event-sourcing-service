# Bug Reproduction

Baseline: remote `origin/base_bug_004`.

Category: error. The projection lag error loses its sentinel identity because it is not wrapped.

Verify:

`仅执行命令，不修改文件： go test ./internal/application -count=1 -run '^TestBug004_ProjectionLagIsIdentifiable$'`
