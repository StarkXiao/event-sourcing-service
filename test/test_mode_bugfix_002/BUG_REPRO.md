# Bug Reproduction

Baseline: remote `origin/base_bug_002`.

Category: nil. Metadata enrichment writes to a nil map when an event has no metadata map.

Verify:

`仅执行命令，不修改文件： go test ./internal/domain -count=1 -run '^TestBug002_AddMetadataOnNilMap$'`
