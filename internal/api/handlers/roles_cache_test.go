package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yourusername/gatekeep/internal/repository"
)

func TestCompareGrantsAcrossRoles(t *testing.T) {
	a := []repository.CachedGrant{{GrantedOn: "TABLE", ObjectName: "DB.S.T", Privilege: "SELECT", GranteeName: "A"}, {GrantedOn: "TABLE", ObjectName: "DB.S.T", Privilege: "DELETE", GranteeName: "A"}}
	b := []repository.CachedGrant{{GrantedOn: "TABLE", ObjectName: "DB.S.T", Privilege: "SELECT", GranteeName: "B"}, {GrantedOn: "TABLE", ObjectName: "DB.S.T", Privilege: "INSERT", GranteeName: "B"}}
	diff := computeGrantsDiff(a, b)
	for kind, privilege := range map[string]string{"added": "INSERT", "removed": "DELETE", "unchanged": "SELECT"} {
		entries := diff[kind].([]map[string]string)
		if len(entries) != 1 || entries[0]["privilege"] != privilege {
			t.Fatalf("%s: %v", kind, entries)
		}
	}
	// Object types must remain part of permission identity.
	b[0].GrantedOn = "VIEW"
	if len(computeGrantsDiff(a, b)["unchanged"].([]map[string]string)) != 0 {
		t.Fatal("different object types matched")
	}
	empty, err := json.Marshal(computeGrantsDiff(nil, nil))
	if err != nil || string(empty) != `{"added":[],"removed":[],"unchanged":[]}` {
		t.Fatalf("empty result: %s, %v", empty, err)
	}
}

func TestUnavailableDependencies(t *testing.T) {
	cases := []struct {
		name    string
		handler http.HandlerFunc
	}{
		{"history", NewHistoryHandler(nil).ListHistory},
		{"detail", NewHistoryHandler(nil).GetSyncRunDetail},
		{"refresh", NewRolesCacheHandler(nil, nil, 10).TriggerRefresh},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			tc.handler(response, httptest.NewRequest(http.MethodGet, "/", nil))
			if response.Code != http.StatusServiceUnavailable {
				t.Fatalf("status %d", response.Code)
			}
		})
	}
}
