package main

import (
	"bytes"
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/yourusername/gatekeep/internal/snowflake"
	gatesync "github.com/yourusername/gatekeep/internal/sync"
)

func TestCLIOrchestratorDryRunAndExecution(t *testing.T) {
	for _, dry := range []bool{true, false} {
		t.Run(map[bool]string{true: "dry-run", false: "execute"}[dry], func(t *testing.T) {
			path := testConfig(t)
			// Dry-runs must not even try to initialize an audit database.
			if dry {
				t.Setenv("POSTGRES_DSN", "invalid DSN")
			} else {
				t.Setenv("POSTGRES_DSN", "")
			}
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = db.Close() }() //nolint:errcheck // test cleanup
			mock.ExpectQuery("SHOW ROLES").WillReturnRows(sqlmock.NewRows([]string{"name"}))
			mock.ExpectQuery("SHOW USERS").WillReturnRows(sqlmock.NewRows([]string{"name"}))
			mock.ExpectQuery("SHOW ROLES").WillReturnRows(sqlmock.NewRows([]string{"name"}))
			mock.ExpectQuery("SHOW DATABASES").WillReturnRows(sqlmock.NewRows([]string{"name"}))
			mock.ExpectQuery("SHOW WAREHOUSES").WillReturnRows(sqlmock.NewRows([]string{"name"}))
			if !dry {
				mock.ExpectExec(`CREATE ROLE IF NOT EXISTS "TEST_ROLE"`).WillReturnResult(sqlmock.NewResult(0, 1))
			}
			client := &snowflake.MockClient{QueryFunc: func(query string) (*sql.Rows, error) { return db.QueryContext(context.Background(), query) }, ExecFunc: func(query string) (sql.Result, error) { return db.ExecContext(context.Background(), query) }}
			args := []string{"sync", "--config", path, "--format", "json"}
			if dry {
				args = append(args, "--dry-run")
			}
			var out, errOut bytes.Buffer
			code := run(context.Background(), args, &out, &errOut, func(ctx context.Context, options syncOptions) (*gatesync.Result, error) {
				return executeWithClient(ctx, options, client)
			})
			if code != 0 {
				t.Fatalf("exit %d: %s %s", code, out.String(), errOut.String())
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}
