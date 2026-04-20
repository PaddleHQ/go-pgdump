package pgdump

import (
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestGetEnumTypes(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	defer func() { _ = db.Close() }()

	rows := sqlmock.NewRows([]string{"enum_name", "enum_values"}).
		AddRow("status", "active,inactive").
		AddRow("user", "admin,member")

	mock.ExpectQuery(`SELECT t\.typname AS enum_name`).WillReturnRows(rows)

	got, err := getEnumTypes(db)
	if err != nil {
		t.Fatalf("getEnumTypes returned error: %v", err)
	}

	wantContains := []string{
		"CREATE TYPE status AS ENUM ('active', 'inactive');",
		`CREATE TYPE "user" AS ENUM ('admin', 'member');`,
	}
	for _, want := range wantContains {
		if !strings.Contains(got, want) {
			t.Errorf("output missing %q\ngot:\n%s", want, got)
		}
	}
}

func TestGetExtensions(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	defer func() { _ = db.Close() }()

	rows := sqlmock.NewRows([]string{"extname"}).
		AddRow("uuid-ossp").
		AddRow("pg_trgm")

	mock.ExpectQuery(`SELECT extname\s+FROM pg_extension`).WillReturnRows(rows)

	got, err := getExtensions(db)
	if err != nil {
		t.Fatalf("getExtensions returned error: %v", err)
	}

	wantContains := []string{
		"CREATE EXTENSION IF NOT EXISTS uuid-ossp;",
		"CREATE EXTENSION IF NOT EXISTS pg_trgm;",
	}
	for _, want := range wantContains {
		if !strings.Contains(got, want) {
			t.Errorf("output missing %q\ngot:\n%s", want, got)
		}
	}
}

func TestGetPartitionBound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	defer func() { _ = db.Close() }()

	mock.ExpectQuery(`SELECT pg_get_expr`).
		WithArgs("orders_2026").
		WillReturnRows(sqlmock.NewRows([]string{"bound"}).
			AddRow("FOR VALUES FROM ('2026-01-01') TO ('2027-01-01')"))

	got, err := getPartitionBound(db, "orders_2026")
	if err != nil {
		t.Fatalf("getPartitionBound returned error: %v", err)
	}

	want := "FOR VALUES FROM ('2026-01-01') TO ('2027-01-01')"
	if got != want {
		t.Errorf("getPartitionBound = %q, want %q", got, want)
	}
}

func TestGetPartitionStrategy(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	defer func() { _ = db.Close() }()

	mock.ExpectQuery(`FROM pg_partitioned_table`).
		WithArgs("orders").
		WillReturnRows(sqlmock.NewRows([]string{"strategy", "columns"}).
			AddRow("RANGE", "created_at"))

	got, err := getPartitionStrategy(db, "orders")
	if err != nil {
		t.Fatalf("getPartitionStrategy returned error: %v", err)
	}

	want := "PARTITION BY RANGE (created_at)"
	if got != want {
		t.Errorf("getPartitionStrategy = %q, want %q", got, want)
	}
}

func TestGetCreatePartitionStatement(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	defer func() { _ = db.Close() }()

	mock.ExpectQuery(`SELECT pg_get_expr`).
		WithArgs("orders_2026").
		WillReturnRows(sqlmock.NewRows([]string{"bound"}).
			AddRow("FOR VALUES FROM ('2026-01-01') TO ('2027-01-01')"))

	table := tableInfo{Name: "orders_2026", IsPartition: true, ParentName: "orders"}
	got, err := getCreatePartitionStatement(db, table)
	if err != nil {
		t.Fatalf("getCreatePartitionStatement returned error: %v", err)
	}

	want := "CREATE TABLE orders_2026 PARTITION OF orders FOR VALUES FROM ('2026-01-01') TO ('2027-01-01');"
	if got != want {
		t.Errorf("getCreatePartitionStatement = %q, want %q", got, want)
	}
}

func TestScriptCommentsNoComments(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	defer func() { _ = db.Close() }()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT obj_description(c.oid)")).
		WithArgs("customers").
		WillReturnRows(sqlmock.NewRows([]string{"comment"}).AddRow(nil))

	mock.ExpectQuery(regexp.QuoteMeta("SELECT a.attname, col_description")).
		WithArgs("customers").
		WillReturnRows(sqlmock.NewRows([]string{"attname", "comment"}))

	got, err := scriptComments(db, "customers")
	if err != nil {
		t.Fatalf("scriptComments returned error: %v", err)
	}

	if got != "" {
		t.Errorf("expected empty output, got %q", got)
	}
}

func TestScriptCommentsWithComments(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	defer func() { _ = db.Close() }()

	tableCommentRows := sqlmock.NewRows([]string{"comment"}).AddRow("Customer records")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT obj_description(c.oid)")).
		WithArgs("customers").
		WillReturnRows(tableCommentRows)

	columnRows := sqlmock.NewRows([]string{"attname", "comment"}).
		AddRow("id", "Primary key").
		AddRow("name", "Customer's name")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT a.attname, col_description")).
		WithArgs("customers").
		WillReturnRows(columnRows)

	got, err := scriptComments(db, "customers")
	if err != nil {
		t.Fatalf("scriptComments returned error: %v", err)
	}

	wantContains := []string{
		"COMMENT ON TABLE customers IS 'Customer records';",
		"COMMENT ON COLUMN customers.id IS 'Primary key';",
		"COMMENT ON COLUMN customers.name IS 'Customer''s name';",
	}
	for _, want := range wantContains {
		if !strings.Contains(got, want) {
			t.Errorf("output missing %q\ngot:\n%s", want, got)
		}
	}
}
