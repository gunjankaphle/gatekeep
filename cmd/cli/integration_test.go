package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"os"
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

func TestCLIScopedStrictReconciliation(t *testing.T) {
	for _, tc := range []struct {
		name                 string
		dry, fail, converged bool
		code, operations     int
	}{
		{name: "preview", dry: true, operations: 2},
		{name: "execute", operations: 2},
		{name: "failed grant prevents revoke", fail: true, code: 1, operations: 2},
		{name: "repeat has no changes", converged: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := testConfig(t)
			content := `version: "1.0"
managed:
  roles: [TEST_ROLE]
  objects:
    - type: WAREHOUSE
      name: [TEST_WH]
      privileges: [USAGE, MONITOR]
roles:
  - name: TEST_ROLE
objects:
  - type: WAREHOUSE
    name: [TEST_WH]
    grants:
      - to_role: TEST_ROLE
        privileges: [USAGE]
`
			if err := os.WriteFile(path, []byte(content), 0600); err != nil {
				t.Fatal(err)
			}
			t.Setenv("POSTGRES_DSN", "")
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = db.Close() }() //nolint:errcheck // test cleanup
			mock.ExpectQuery("SHOW ROLES").WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("TEST_ROLE"))
			mock.ExpectQuery("SHOW USERS").WillReturnRows(sqlmock.NewRows([]string{"name"}))
			mock.ExpectQuery("SHOW ROLES").WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("TEST_ROLE"))
			privilege := "MONITOR"
			if tc.converged {
				privilege = "USAGE"
			}
			mock.ExpectQuery(`SHOW GRANTS TO ROLE "TEST_ROLE"`).WillReturnRows(sqlmock.NewRows([]string{"privilege", "granted_on", "name", "granted_to"}).AddRow(privilege, "WAREHOUSE", "TEST_WH", "ROLE").AddRow("OPERATE", "WAREHOUSE", "TEST_WH", "ROLE").AddRow("USAGE", "WAREHOUSE", "OTHER_WH", "ROLE"))
			mock.ExpectQuery("SHOW DATABASES").WillReturnRows(sqlmock.NewRows([]string{"name"}))
			mock.ExpectQuery("SHOW WAREHOUSES").WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("TEST_WH"))
			if !tc.dry && !tc.converged {
				grant := mock.ExpectExec(`GRANT USAGE ON WAREHOUSE "TEST_WH" TO ROLE "TEST_ROLE"`)
				if tc.fail {
					grant.WillReturnError(errors.New("grant denied"))
				} else {
					grant.WillReturnResult(sqlmock.NewResult(0, 1))
					mock.ExpectExec(`REVOKE MONITOR ON WAREHOUSE "TEST_WH" FROM ROLE "TEST_ROLE"`).WillReturnResult(sqlmock.NewResult(0, 1))
				}
			}
			client := &snowflake.MockClient{QueryFunc: func(query string) (*sql.Rows, error) { return db.QueryContext(context.Background(), query) }, ExecFunc: func(query string) (sql.Result, error) { return db.ExecContext(context.Background(), query) }}
			args := []string{"sync", "--config", path, "--mode", "strict", "--format", "json"}
			if tc.dry {
				args = append(args, "--dry-run")
			}
			var out, errOut bytes.Buffer
			code := run(context.Background(), args, &out, &errOut, func(ctx context.Context, options syncOptions) (*gatesync.Result, error) {
				return executeWithClient(ctx, options, client)
			})
			if code != tc.code {
				t.Fatalf("exit %d: %s %s", code, out.String(), errOut.String())
			}
			var output cliOutput
			if err := json.Unmarshal(out.Bytes(), &output); err != nil {
				t.Fatal(err)
			}
			if len(output.Operations) != tc.operations {
				t.Fatalf("unexpected operations: %+v", output)
			}
			if tc.fail && output.Operations[1].Status != "skipped" {
				t.Fatalf("revoke wasn't skipped: %+v", output)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}
