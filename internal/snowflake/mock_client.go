package snowflake

import (
	"database/sql"
	"fmt"
)

// MockClient is a mock implementation of Client for testing
type MockClient struct {
	QueryFunc func(query string) (*sql.Rows, error)
	ExecFunc  func(query string) (sql.Result, error)
	CloseFunc func() error
	Closed    bool
}

// Query calls the mock QueryFunc
func (m *MockClient) Query(query string) (*sql.Rows, error) {
	if m.QueryFunc != nil {
		return m.QueryFunc(query)
	}
	return nil, fmt.Errorf("QueryFunc not implemented")
}

// Exec calls the mock ExecFunc
func (m *MockClient) Exec(query string) (sql.Result, error) {
	if m.ExecFunc != nil {
		return m.ExecFunc(query)
	}
	return nil, fmt.Errorf("ExecFunc not implemented")
}

// Close calls the mock CloseFunc
func (m *MockClient) Close() error {
	m.Closed = true
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}
