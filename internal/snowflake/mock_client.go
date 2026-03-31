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

// MockResult implements sql.Result for testing
type MockResult struct {
	AffectedRows int64
	LastID       int64
}

// LastInsertId returns the last insert ID
func (m *MockResult) LastInsertId() (int64, error) {
	return m.LastID, nil
}

// RowsAffected returns the number of rows affected
func (m *MockResult) RowsAffected() (int64, error) {
	return m.AffectedRows, nil
}

// MockError implements error for testing
type MockError struct {
	Msg string
}

// Error returns the error message
func (m *MockError) Error() string {
	return m.Msg
}
