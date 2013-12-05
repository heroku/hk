package main

import (
	"reflect"
	"testing"
)

var testAddonConfigs = []struct {
	in  []string
	out map[string]string
	err error
}{
	{
		[]string{"key=val", "really_long-conf=crazy-value.1"},
		map[string]string{"key": "val", "really_long-conf": "crazy-value.1"},
		nil,
	},
	{
		[]string{"k='single-quoted value'", "k2=\"double-quoted value\"", "k3='\"'"},
		map[string]string{"k": "single-quoted value", "k2": "double-quoted value", "k3": "\""},
		nil,
	},
}

func TestParseAddonAddConfig(t *testing.T) {
	for i, c := range testAddonConfigs {
		res, err := parseAddonAddConfig(c.in)
		if err != c.err {
			t.Errorf("%d. parseAddonAddConfig(%q).err => %q, want %q", i, c.in, err, c.err)
		}
		if !reflect.DeepEqual(*res, c.out) {
			t.Errorf("%d. parseAddonAddConfig(%q) => %v, want %v", i, c.in, *res, c.out)
		}
	}
}
