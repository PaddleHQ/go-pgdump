package pgdump

import (
	"bytes"
	"strings"
	"testing"
)

func TestEscapeReservedName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "reserved keyword uppercase", input: "USER", want: `"USER"`},
		{name: "reserved keyword lowercase", input: "user", want: `"user"`},
		{name: "reserved keyword mixed case", input: "User", want: `"User"`},
		{name: "non-reserved name", input: "customers", want: "customers"},
		{name: "non-reserved mixed case", input: "CustomerOrders", want: "CustomerOrders"},
		{name: "empty string", input: "", want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := escapeReservedName(tt.input)
			if got != tt.want {
				t.Errorf("escapeReservedName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestWriteHeader(t *testing.T) {
	info := DumpInfo{
		DumpVersion:   "1.2.3",
		ServerVersion: "PostgreSQL 16.0",
		ThreadsNumber: 8,
	}

	var buf bytes.Buffer
	if err := writeHeader(&buf, info); err != nil {
		t.Fatalf("writeHeader returned error: %v", err)
	}

	output := buf.String()
	wantContains := []string{
		"Go PostgreSQL Dump v1.2.3",
		"PostgreSQL 16.0",
		"8",
		"SET statement_timeout = 0;",
		"SET client_encoding = 'UTF8';",
	}
	for _, want := range wantContains {
		if !strings.Contains(output, want) {
			t.Errorf("header output missing %q\ngot:\n%s", want, output)
		}
	}
}

func TestWriteFooter(t *testing.T) {
	info := DumpInfo{
		CompleteTime: "2026-04-20 14:30:00 +0000 UTC",
	}

	var buf bytes.Buffer
	if err := writeFooter(&buf, info); err != nil {
		t.Fatalf("writeFooter returned error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Dump completed on 2026-04-20 14:30:00 +0000 UTC") {
		t.Errorf("footer output missing completion line\ngot:\n%s", output)
	}
}
