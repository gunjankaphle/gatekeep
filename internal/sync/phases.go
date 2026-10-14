package sync

import (
	"context"
	"fmt"

	"github.com/yourusername/gatekeep/internal/diff"
)

// executePhases preserves planner order while retaining parallelism within phases.
// Failed or skipped work stops subsequent phases, especially revocations.
func (o *Orchestrator) executePhases(ctx context.Context, operations []diff.SQLOperation, cfg Config) []OperationResult {
	results := make([]OperationResult, 0, len(operations))
	for start := 0; start < len(operations); {
		end := start + 1
		phase := operationPhase(operations[start].Type)
		for end < len(operations) && operationPhase(operations[end].Type) == phase {
			end++
		}
		batch := NewParallelExecutor(o.executor, cfg).Execute(ctx, operations[start:end])
		results = append(results, batch...)
		failed := false
		for _, result := range batch {
			if result.Status != OpStatusSuccess {
				failed = true
			}
		}
		if failed {
			for _, op := range operations[end:] {
				results = append(results, OperationResult{Operation: op, Status: OpStatusSkipped, Error: fmt.Sprintf("earlier execution phase %d did not complete successfully", phase)})
			}
			break
		}
		start = end
	}
	return results
}

func operationPhase(kind diff.OperationType) diff.ExecutionPhase {
	switch kind {
	case diff.OpCreateDatabase, diff.OpCreateWarehouse:
		return diff.PhaseCreateResources
	case diff.OpCreateRole:
		return diff.PhaseCreateRoles
	case diff.OpGrantRole:
		return diff.PhaseGrantRoleHierarchy
	case diff.OpGrantObject:
		return diff.PhaseGrantObjectPermissions
	case diff.OpCreateUser:
		return diff.PhaseCreateUsers
	case diff.OpGrantUserRole:
		return diff.PhaseGrantUserRoles
	case diff.OpRevokeObject, diff.OpRevokeRole, diff.OpRevokeUserRole:
		return diff.PhaseRevokePermissions
	default:
		return diff.PhaseDeleteRoles
	}
}
