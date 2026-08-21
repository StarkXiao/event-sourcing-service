package domain

import "testing"

func TestBug002_AddMetadataOnNilMap(t *testing.T) {
	e := Event{}
	e.AddMetadata("request_id", "req-1")
	if e.Metadata["request_id"] != "req-1" {
		t.Fatalf("metadata = %#v", e.Metadata)
	}
}
