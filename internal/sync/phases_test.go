package sync

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/yourusername/gatekeep/internal/audit"
	"github.com/yourusername/gatekeep/internal/diff"
	"github.com/yourusername/gatekeep/internal/snowflake"
)

type phaseSQLResult struct{}

func (phaseSQLResult) LastInsertId() (int64, error) { return 0, nil }
func (phaseSQLResult) RowsAffected() (int64, error) { return 0, nil }

func TestExecutePhasesOrderingAndFailureStopsRevocations(t *testing.T) {
	for _, fail := range []bool{false, true} {
		t.Run(map[bool]string{false: "ordered", true: "failed prerequisite"}[fail], func(t *testing.T) {
			var mu sync.Mutex
			calls := []string{}
			client := &snowflake.MockClient{ExecFunc: func(query string) (sql.Result, error) {
				mu.Lock()
				defer mu.Unlock()
				calls = append(calls, query)
				if fail && query == "grant object" {
					return nil, errors.New("grant denied")
				}
				return phaseSQLResult{}, nil
			}}
			o := NewOrchestrator(nil, nil, snowflake.NewExecutor(client), diff.SyncModeStrict)
			ops := []diff.SQLOperation{{Type: diff.OpCreateRole, SQL: "create role"}, {Type: diff.OpGrantRole, SQL: "grant hierarchy"}, {Type: diff.OpGrantObject, SQL: "grant object"}, {Type: diff.OpRevokeObject, SQL: "revoke object"}}
			result := o.execute(context.Background(), ops, DefaultConfig())
			expected := []string{"create role", "grant hierarchy", "grant object", "revoke object"}
			if fail {
				expected = expected[:3]
			}
			if !reflect.DeepEqual(calls, expected) {
				t.Fatalf("execution order: %v", calls)
			}
			if len(result) != len(ops) {
				t.Fatalf("results lost: %+v", result)
			}
			if fail && (result[2].Status != OpStatusFailed || result[3].Status != OpStatusSkipped) {
				t.Fatalf("revocation wasn't stopped: %+v", result)
			}
		})
	}
}

type phaseAudit struct {
	*audit.NoOpLogger
	nextID   int64
	recorded []string
}

func (a *phaseAudit) LogOperation(_ context.Context, _ int64, _, _, _ string) (int64, error) {
	a.nextID++
	return a.nextID, nil
}
func (a *phaseAudit) RecordOperationSuccess(_ context.Context, id int64, _ int) error {
	a.recorded = append(a.recorded, fmt.Sprintf("success:%d", id))
	return nil
}
func (a *phaseAudit) RecordOperationFailure(_ context.Context, id int64, _ string, _ int) error {
	a.recorded = append(a.recorded, fmt.Sprintf("failed:%d", id))
	return nil
}
func (a *phaseAudit) RecordOperationSkipped(_ context.Context, id int64, _ string) error {
	a.recorded = append(a.recorded, fmt.Sprintf("skipped:%d", id))
	return nil
}

func TestSkippedRevocationsAreAudited(t *testing.T) {
	logger := &phaseAudit{NoOpLogger: audit.NewNoOpLogger()}
	client := &snowflake.MockClient{ExecFunc: func(query string) (sql.Result, error) { return nil, errors.New("grant denied") }}
	o := NewOrchestrator(nil, nil, snowflake.NewExecutor(client), diff.SyncModeStrict).WithAuditLogger(logger)
	result := o.executeWithAudit(context.Background(), []diff.SQLOperation{{Type: diff.OpGrantObject, SQL: "grant"}, {Type: diff.OpRevokeObject, SQL: "revoke"}}, DefaultConfig(), 42)
	if len(result) != 2 || result[1].Status != OpStatusSkipped {
		t.Fatalf("result: %+v", result)
	}
	if !reflect.DeepEqual(logger.recorded, []string{"failed:1", "skipped:2"}) {
		t.Fatalf("audit left a pending operation: %v", logger.recorded)
	}
}

func TestTimedOutGrantPreventsRevocation(t *testing.T) {
	revoked := false
	timedOut := false
	client := &snowflake.MockClient{ExecContextFunc: func(ctx context.Context, query string) (sql.Result, error) {
		if query == "grant" {
			<-ctx.Done()
			timedOut = true
			return nil, ctx.Err()
		}
		revoked = true
		return phaseSQLResult{}, nil
	}}
	o := NewOrchestrator(nil, nil, snowflake.NewExecutor(client), diff.SyncModeStrict)
	cfg := DefaultConfig()
	cfg.Timeout = 10 * time.Millisecond
	results := o.execute(context.Background(), []diff.SQLOperation{{Type: diff.OpGrantObject, SQL: "grant"}, {Type: diff.OpRevokeObject, SQL: "revoke"}}, cfg)
	if !timedOut || revoked || len(results) != 2 || results[0].Status != OpStatusFailed || results[1].Status != OpStatusSkipped {
		t.Fatalf("deadline did not stop revocation: %+v", results)
	}
}
