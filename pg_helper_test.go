package main

import (
	"fmt"
	"reflect"
	"testing"
)

func TestHpgOptNames(t *testing.T) {
	expected := []string{"fork", "follow", "rollback"}
	if !reflect.DeepEqual(hpgOptNames, expected) {
		t.Fatalf("expected hpgOptNames to be %v, got %v", expected, hpgOptNames)
	}
}

var hpgOptResolveTests = []struct {
	opts   *map[string]string
	env    map[string]string
	result map[string]string
	err    error
}{
	{
		opts:   &map[string]string{"fork": "yellow"},
		env:    map[string]string{"HEROKU_POSTGRESQL_YELLOW_URL": "postgres://test.com"},
		result: map[string]string{"fork": "postgres://test.com"},
		err:    nil,
	},
	{
		opts: &map[string]string{
			"fork":     "HEROKU_POSTGRESQL_ORANGE_URL",
			"follow":   "orange",
			"rollback": "heroku-postgresql-yellow",
			"ignore":   "yes",
		},
		env: map[string]string{
			"HEROKU_POSTGRESQL_YELLOW_URL": "postgres://test.com",
			"HEROKU_POSTGRESQL_ORANGE_URL": "postgres://test.com/orange",
		},
		result: map[string]string{
			"fork":     "postgres://test.com/orange",
			"follow":   "postgres://test.com/orange",
			"rollback": "postgres://test.com",
			"ignore":   "yes",
		},
		err: nil,
	},
	{
		opts:   &map[string]string{"fork": "nope"},
		env:    map[string]string{"HEROKU_POSTGRESQL_YELLOW_URL": "postgres://test.com"},
		result: nil,
		err:    fmt.Errorf("could not resolve fork option \"nope\" to a heroku-postgresql addon"),
	},
	{
		opts:   &map[string]string{"fork": "postgres://test.com/1", "ignore": "yes"},
		env:    map[string]string{"HEROKU_POSTGRESQL_YELLOW_URL": "postgres://test.com"},
		result: map[string]string{"fork": "postgres://test.com/1", "ignore": "yes"},
		err:    nil,
	},
	{
		opts:   nil,
		env:    nil,
		result: nil,
		err:    nil,
	},
}

func TestHpgAddonOptResolve(t *testing.T) {
	for i, unit := range hpgOptResolveTests {
		err := hpgAddonOptResolve(unit.opts, unit.env)
		if unit.err != nil {
			if !reflect.DeepEqual(unit.err, err) {
				t.Fatalf("test %d: expected error %v on, got %v", i, unit.err, err)
			}
		} else if unit.opts != nil {
			if !reflect.DeepEqual(*unit.opts, unit.result) {
				t.Errorf("test %d: expected %v, got %v", i, unit.result, *unit.opts)
			}
		}
	}
}

func TestPgEnvToDBName(t *testing.T) {
	tests := []struct {
		in  string
		out string
	}{
		{"HEROKU_POSTGRESQL_BLUE_URL", "heroku-postgresql-blue"},
		{"HEROKU_POSTGRESQL_CRIMSON_URL", "heroku-postgresql-crimson"},
	}
	for _, ex := range tests {
		result := pgEnvToDBName(ex.in)
		if result != ex.out {
			t.Errorf("expected %s, got %s", ex.out, result)
		}
	}
}

func TestDBNameToPgEnv(t *testing.T) {
	tests := []struct {
		in  string
		out string
	}{
		{"heroku-postgresql-blue", "HEROKU_POSTGRESQL_BLUE_URL"},
		{"heroku-postgresql-crimson", "HEROKU_POSTGRESQL_CRIMSON_URL"},
		{"rose", "HEROKU_POSTGRESQL_ROSE_URL"},
		{"HEROKU_POSTGRESQL_ROSE", "HEROKU_POSTGRESQL_ROSE_URL"},
		{"HEROKU_POSTGRESQL_CRIMSON_URL", "HEROKU_POSTGRESQL_CRIMSON_URL"},
	}
	for _, ex := range tests {
		result := dbNameToPgEnv(ex.in)
		if result != ex.out {
			t.Errorf("expected %s, got %s", ex.out, result)
		}
	}
}
