// SPDX-FileCopyrightText: 2024 NOI Techpark <digital@noi.bz.it>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package siri

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestWrapObjects(t *testing.T) {
	keys := map[string]bool{"L": true, "M": true}
	cases := []struct{ in, want string }{
		{`{"L":[{"a":1}]}`, `{"L":[{"a":1}]}`},
		{`{"L":{"a":1}}`, `{"L":[{"a":1}]}`},
		{`{"L" : { "M" : {"x":"}"} } , "z":"L"}`, `{"L" : [{ "M" : [{"x":"}"}] }] , "z":"L"}`},
		{`{"s":"\"L\":{","L":{}}`, `{"s":"\"L\":{","L":[{}]}`},
		{`{"o":{"L":{"M":[{"L":{"a":"\\"}}]}},"L":null}`, `{"o":{"L":[{"M":[{"L":[{"a":"\\"}]}]}]},"L":null}`},
		{`{"N":{"a":1}}`, `{"N":{"a":1}}`},
	}
	for _, c := range cases {
		got := string(wrapObjects([]byte(c.in), keys))
		if got != c.want {
			t.Errorf("wrapObjects(%s)\n got %s\nwant %s", c.in, got, c.want)
		}
		var a, b any
		if err := json.Unmarshal([]byte(got), &a); err != nil {
			t.Errorf("invalid JSON %s: %v", got, err)
		}
		json.Unmarshal([]byte(c.want), &b)
		if !reflect.DeepEqual(a, b) {
			t.Errorf("decoded mismatch for %s", c.in)
		}
	}
}

func TestWrapObjectsUnchangedInputNotCopied(t *testing.T) {
	in := []byte(`{"L":[{"a":{"b":1}}]}`)
	if out := wrapObjects(in, map[string]bool{"L": true}); &out[0] != &in[0] {
		t.Error("input without single objects should be returned as is")
	}
}
