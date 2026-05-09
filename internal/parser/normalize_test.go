package parser

import "testing"

func TestParseOneLiner(t *testing.T) {
	cases := []struct {
		in       string
		method   string
		resource string
		status   string
		ok       bool
	}{
		{"GET / -> 200", "GET", "/", "200", true},
		{"POST /users -> 201", "POST", "/users", "201", true},
		{"DELETE /users/1 -> 204", "DELETE", "/users/1", "204", true},
		{"INVALID", "", "", "", false},
		{"get / -> 200", "", "", "", false}, // lowercase method not accepted
		{"GET ", "", "", "", false},
	}
	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			tc, err := parseOneLiner(c.in)
			if (err == nil) != c.ok {
				t.Fatalf("ok = %v, want %v (err=%v)", err == nil, c.ok, err)
			}
			if !c.ok {
				return
			}
			if tc.Method != c.method || tc.Resource != c.resource || tc.Expect.Status != c.status {
				t.Errorf("got method=%q resource=%q status=%v",
					tc.Method, tc.Resource, tc.Expect.Status)
			}
		})
	}
}

func TestParseFullForm_KnownAndUnknownFields(t *testing.T) {
	in := map[string]any{
		"name":    "test",
		"method":  "POST",
		"timeout": "5s",
		// unknown / plugin-specific:
		"customField": "x",
		"body":        map[string]any{"a": 1},
	}
	tc, err := parseFullForm(in, "", "")
	if err != nil {
		t.Fatal(err)
	}
	if tc.Name != "test" || tc.Method != "POST" || tc.Timeout != "5s" {
		t.Errorf("known fields wrong: %+v", tc)
	}
	if tc.Extra["customField"] != "x" || tc.Extra["body"] == nil {
		t.Errorf("extra wrong: %v", tc.Extra)
	}
}

func TestParseFullForm_CompactMethodResource(t *testing.T) {
	in := map[string]any{
		"timeout": "1s",
	}
	tc, err := parseFullForm(in, "PATCH", "/items/1")
	if err != nil {
		t.Fatal(err)
	}
	if tc.Method != "PATCH" || tc.Resource != "/items/1" {
		t.Errorf("compact override wrong: %+v", tc)
	}
}

func TestNormalizeOne_BadType(t *testing.T) {
	if _, err := normalizeOne(123); err == nil {
		t.Error("int test entry should error")
	}
}
