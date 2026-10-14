package config

import (
	"strings"
	"testing"
)

func TestManagedScopeValidation(t *testing.T) {
	valid := `version: "1.0"
managed:
  roles: [READER, RETIRED_ROLE]
  users: [service_user]
  objects:
    - type: SEMANTIC VIEW
      name: [DB, S, MODEL]
      privileges: [SELECT, REFERENCES]
roles:
  - name: READER
`
	if _, err := NewParser().Parse([]byte(valid)); err != nil {
		t.Fatal(err)
	}
	cases := []struct{ old, new string }{
		{"[READER, RETIRED_ROLE]", "[]"},
		{"[READER, RETIRED_ROLE]", "[SYSADMIN]"},
		{"[READER, RETIRED_ROLE]", "[READER, READER]"},
		{"[service_user]", "[service_user, service_user]"},
		{"[SELECT, REFERENCES]", "[OWNERSHIP]"},
		{"[SELECT, REFERENCES]", "[ALL PRIVILEGES]"},
		{"[SELECT, REFERENCES]", "[]"},
		{"SEMANTIC VIEW", "TABLE; DROP ROLE X"},
		{"SEMANTIC VIEW", "FUNCTION"},
		{"[DB, S, MODEL]", "[DB, S.MODEL]"},
		{"[DB, S, MODEL]", "[]"},
	}
	for _, tc := range cases {
		t.Run(tc.new, func(t *testing.T) {
			if _, err := NewParser().Parse([]byte(strings.Replace(valid, tc.old, tc.new, 1))); err == nil {
				t.Fatal("accepted unsafe scope")
			}
		})
	}
}
