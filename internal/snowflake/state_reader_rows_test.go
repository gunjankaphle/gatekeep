package snowflake

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestReadGrantsWithAdditionalColumns(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	role := `A"B`
	mock.ExpectQuery("SHOW ROLES").WillReturnRows(sqlmock.NewRows([]string{"name", "owner", "comment", "new_column"}).AddRow(role, "ADMIN", nil, "extra"))
	query := `SHOW GRANTS TO ROLE "A""B"`
	mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnRows(sqlmock.NewRows([]string{"created_on", "privilege", "granted_on", "name", "granted_to", "grantee_name", "grant_option", "granted_by", "granted_by_role_type"}).AddRow("2026-01-01", "SELECT", "TABLE", "DB.S.T", "ROLE", role, "false", "ADMIN", "ROLE"))
	reader := NewStateReader(&MockClient{QueryFunc: func(query string) (*sql.Rows, error) { return db.QueryContext(context.Background(), query) }})
	grants, err := reader.ReadGrants()
	if err != nil {
		t.Fatal(err)
	}
	if len(grants) != 1 || grants[0].Privilege != "SELECT" || grants[0].GranteeType != "ROLE" || grants[0].GranteeName != role {
		t.Fatalf("grants: %#v", grants)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestGrantQueryFailureIsFatal(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	mock.ExpectQuery("SHOW ROLES").WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("A"))
	mock.ExpectQuery(`SHOW GRANTS TO ROLE "A"`).WillReturnError(errors.New("permission denied"))
	reader := NewStateReader(&MockClient{QueryFunc: func(query string) (*sql.Rows, error) { return db.QueryContext(context.Background(), query) }})
	if _, err := reader.ReadGrants(); err == nil {
		t.Fatal("query failure was silently discarded")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
