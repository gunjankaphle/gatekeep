package config

import "testing"

func TestGenericObjectValidation(t *testing.T) {
	for _, tc := range []struct {
		objectType, privilege string
		valid                 bool
	}{
		{"DYNAMIC TABLE", "SELECT", true}, {"SEMANTIC VIEW", "REFERENCES", true}, {"WAREHOUSE", "OPERATE", true}, {"NEW OBJECT TYPE", "NEW PRIVILEGE", true},
		{"TABLE; DROP ROLE X", "SELECT", false}, {"TABLE", "SELECT; DROP ROLE X", false}, {"TABLE", "OWNERSHIP", false}, {"TABLE", "ALL PRIVILEGES", false},
	} {
		cfg := &Config{Version: "1.0", Roles: []Role{{Name: "R"}}, Objects: []Object{{Type: tc.objectType, Name: []string{"DB", "S", "T"}, Grants: []Grant{{ToRole: "R", Privileges: []string{tc.privilege}}}}}}
		err := NewValidator().Validate(cfg)
		if (err == nil) != tc.valid {
			t.Errorf("%s/%s: %v", tc.objectType, tc.privilege, err)
		}
	}
}
