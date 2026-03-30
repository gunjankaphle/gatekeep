package snowflake

import (
	"database/sql"
	"fmt"
	"testing"
)

func TestNewStateReader(t *testing.T) {
	mock := &MockClient{}
	sr := NewStateReader(mock)

	if sr == nil {
		t.Fatal("Expected non-nil StateReader")
	}

	if sr.client != mock {
		t.Error("Expected StateReader to use provided client")
	}
}

func TestStateReader_ReadState_MockError(t *testing.T) {
	mock := &MockClient{
		QueryFunc: func(query string) (*sql.Rows, error) {
			return nil, fmt.Errorf("mock error")
		},
	}

	sr := NewStateReader(mock)
	_, err := sr.ReadState()

	if err == nil {
		t.Fatal("Expected error from ReadState with failing mock")
	}

	if err.Error() == "" {
		t.Error("Expected non-empty error message")
	}
}

func TestStateReader_Methods(t *testing.T) {
	// Test that all reader methods exist and can be called
	// (they will fail with mock, but we verify the API)
	mock := &MockClient{
		QueryFunc: func(query string) (*sql.Rows, error) {
			return nil, fmt.Errorf("expected mock error")
		},
	}

	sr := NewStateReader(mock)

	t.Run("ReadRoles", func(t *testing.T) {
		_, err := sr.ReadRoles()
		if err == nil {
			t.Error("Expected error from mock")
		}
	})

	t.Run("ReadUsers", func(t *testing.T) {
		_, err := sr.ReadUsers()
		if err == nil {
			t.Error("Expected error from mock")
		}
	})

	t.Run("ReadGrants", func(t *testing.T) {
		grants, err := sr.ReadGrants()
		// ReadGrants currently returns empty, no error
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if grants == nil {
			t.Error("Expected non-nil grants slice")
		}
	})

	t.Run("ReadDatabases", func(t *testing.T) {
		_, err := sr.ReadDatabases()
		if err == nil {
			t.Error("Expected error from mock")
		}
	})

	t.Run("ReadWarehouses", func(t *testing.T) {
		_, err := sr.ReadWarehouses()
		if err == nil {
			t.Error("Expected error from mock")
		}
	})
}

func TestTypes(t *testing.T) {
	// Test that types can be instantiated
	t.Run("State", func(t *testing.T) {
		state := State{
			Roles:      []Role{{Name: "TEST"}},
			Users:      []User{{Name: "test@example.com"}},
			Grants:     []Grant{},
			Databases:  []Database{{Name: "DB"}},
			Warehouses: []Warehouse{{Name: "WH"}},
		}

		if len(state.Roles) != 1 {
			t.Errorf("Expected 1 role, got %d", len(state.Roles))
		}
		if state.Roles[0].Name != "TEST" {
			t.Errorf("Expected role name TEST, got %s", state.Roles[0].Name)
		}
	})

	t.Run("Role", func(t *testing.T) {
		role := Role{
			Name:    "TEST_ROLE",
			Comment: "Test comment",
			Owner:   "ACCOUNTADMIN",
		}

		if role.Name != "TEST_ROLE" {
			t.Errorf("Expected name TEST_ROLE, got %s", role.Name)
		}
	})

	t.Run("User", func(t *testing.T) {
		user := User{
			Name:  "user@example.com",
			Roles: []string{"ROLE1", "ROLE2"},
		}

		if len(user.Roles) != 2 {
			t.Errorf("Expected 2 roles, got %d", len(user.Roles))
		}
	})

	t.Run("Grant", func(t *testing.T) {
		grant := Grant{
			GrantedOn:   "TABLE",
			GrantedTo:   "ROLE",
			Name:        "MYTABLE",
			Privilege:   "SELECT",
			GranteeType: "ROLE",
			GranteeName: "ANALYST",
		}

		if grant.Privilege != "SELECT" {
			t.Errorf("Expected privilege SELECT, got %s", grant.Privilege)
		}
	})
}
