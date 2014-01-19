package main

import "testing"

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
	}
	for _, ex := range tests {
		result := dbNameToPgEnv(ex.in)
		if result != ex.out {
			t.Errorf("expected %s, got %s", ex.out, result)
		}
	}
}
