package pgdump_test

import (
	"testing"

	pgdump "github.com/PaddleHQ/go-pgdump"
)

func TestNewDumper(t *testing.T) {
	tests := []struct {
		name        string
		threads     int
		wantThreads int
	}{
		{name: "explicit thread count is preserved", threads: 8, wantThreads: 8},
		{name: "zero threads falls back to default", threads: 0, wantThreads: 50},
		{name: "negative threads falls back to default", threads: -1, wantThreads: 50},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			const connStr = "postgres://user:pass@localhost/db"
			d := pgdump.NewDumper(connStr, tt.threads)
			if d.ConnectionString != connStr {
				t.Errorf("ConnectionString = %q, want %q", d.ConnectionString, connStr)
			}
			if d.Parallels != tt.wantThreads {
				t.Errorf("Parallels = %d, want %d", d.Parallels, tt.wantThreads)
			}
			if d.DumpVersion == "" {
				t.Error("DumpVersion should not be empty")
			}
		})
	}
}
