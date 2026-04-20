module github.com/PaddleHQ/go-pgdump

go 1.25.0

// Retract old unstable versions
retract (
	[v1.0.0, v1.0.9]
	[v0.2.0, v0.2.9]
	[v0.1.0, v0.1.9]
)

require (
	github.com/DATA-DOG/go-sqlmock v1.5.2
	github.com/lib/pq v1.10.9
	golang.org/x/sync v0.8.0
)
