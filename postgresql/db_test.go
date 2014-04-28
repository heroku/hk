package postgresql

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"
)

type starterTest struct {
	plan      string
	isStarter bool
}

var starterTests = []starterTest{
	{"dev", true},
	{"basic", true},
	{"hobby-dev", true},
	{"hobby-basic", true},
	{"devcloud", true},
	{"crane", false},
	{"kappa", false},
	{"ronin", false},
	{"standard-yanari", false},
	{"premium-tengu", false},
	{"enterprise-ryu", false},
}

func TestIsStarterPlan(t *testing.T) {
	for _, st := range starterTests {
		db := DB{Plan: st.plan}
		if db.IsStarterPlan() != st.isStarter {
			t.Errorf("expected isStarter=%t for %s", st.isStarter, st.plan)
		}
	}
}

var pgInfoResponse = `
{
  "available_for_ingress": true,
  "created_at": "2014-01-14 23:35:22 +0000",
  "current_transaction": "1879",
  "database_name": "dbname",
  "database_password": "password",
  "database_user": "username",
  "info": [
    {
      "name": "Plan",
      "values": [ "Standard Tengu" ]
    },
    {
      "name": "Status",
      "values": [ "Available" ]
    },
    {
      "name": "Data Size",
      "values": [ "6.4 MB" ]
    },
    {
      "name": "Tables",
      "values": [ 0 ]
    },
    {
      "name": "PG Version",
      "values": [ "9.3.2" ]
    },
    {
      "name": "Connections",
      "values": [ 3 ]
    },
    {
      "name": "Fork/Follow",
      "values": [ "Available" ]
    },
    {
      "name": "Rollback",
      "values": [ "from 2014-01-15 23:45 UTC" ]
    },
    {
      "name": "Created",
      "values": [ "2014-01-14 23:35 UTC" ]
    },
    {
      "name": "Followers",
      "resolve_db_name": true,
      "values": []
    },
    {
      "name": "Forks",
      "resolve_db_name": true,
			"values": ["postgres://myfakedb.com/dbname"]
    },
    {
      "name": "Maintenance",
      "values": [ "not required" ]
    }
  ],
  "is_in_recovery?": true,
  "num_bytes": 6736056,
  "num_connections": 3,
  "num_connections_waiting": 2,
  "num_tables": 1,
  "plan": "standard-tengu",
  "postgresql_version": "9.3.2",
  "resource_url": "postgres://username:password@ec2-107-12-34-82.compute-1.amazonaws.com:5552/dbname",
  "service_port": 5552,
  "standalone?": false,
  "status_updated_at": "2014-01-16T00:43:52+00:00",
  "target_transaction": "5"
}`

func TestDBInfo(t *testing.T) {
	var dbi DBInfo
	err := json.Unmarshal([]byte(pgInfoResponse), &dbi)
	if err != nil {
		t.Fatal(err)
	}
	if dbi.AvailableForIngress != true {
		t.Errorf("expected AvailableForIngress=true, got %t", dbi.AvailableForIngress)
	}
	if err != nil {
		t.Fatal(err)
	}
	if dbi.CreatedAt != "2014-01-14 23:35:22 +0000" {
		t.Errorf("expected CreatedAt=%s, got %s", "2014-01-14 23:35:22 +0000", dbi.CreatedAt)
	}
	if dbi.CurrentTransaction != "1879" {
		t.Errorf("expected CurrentTransaction=%s, got %s", "1879", dbi.CurrentTransaction)
	}
	if dbi.DatabaseName != "dbname" {
		t.Errorf("expected DatabaseName=%s, got %s", "dbname", dbi.DatabaseName)
	}
	if dbi.DatabasePassword != "password" {
		t.Errorf("expected DatabasePassword=%s, got %s", "password", dbi.DatabasePassword)
	}
	if dbi.DatabaseUser != "username" {
		t.Errorf("expected DatabaseUser=%s, got %s", "username", dbi.DatabaseUser)
	}

	// InfoEntries
	if len(dbi.Info) != 12 {
		t.Errorf("expected 12 info entries, got %d", len(dbi.Info))
	} else if !reflect.DeepEqual(dbi.Info[0], InfoEntry{
		Name:   "Plan",
		Values: []interface{}{"Standard Tengu"},
	}) {
		t.Errorf("unexpected values in dbi.Info[0]: %v", dbi.Info[0])
	} else if !reflect.DeepEqual(dbi.Info[9], InfoEntry{
		Name:          "Followers",
		ResolveDBName: true,
		Values:        []interface{}{},
	}) {
		t.Errorf("unexpected values in dbi.Info[9]: %v", dbi.Info[9])
	}

	if dbi.IsInRecovery != true {
		t.Errorf("expected IsInRecovery=true, got %t", dbi.IsInRecovery)
	}
	if dbi.NumBytes != 6736056 {
		t.Errorf("expected NumBytes=%d, got %d", 6736056, dbi.NumBytes)
	}
	if dbi.NumConnections != 3 {
		t.Errorf("expected NumConnections=%d, got %d", 3, dbi.NumConnections)
	}
	if dbi.NumConnectionsWaiting != 2 {
		t.Errorf("expected NumConnectionsWaiting=%d, got %d", 2, dbi.NumConnectionsWaiting)
	}
	if dbi.NumTables != 1 {
		t.Errorf("expected NumTables=%d, got %d", 1, dbi.NumTables)
	}
	if dbi.Plan != "standard-tengu" {
		t.Errorf("expected Plan=%s, got %s", "standard-tengu", dbi.Plan)
	}
	if dbi.PostgresqlVersion != "9.3.2" {
		t.Errorf("expected PostgresqlVersion=%s, got %s", "9.3.2", dbi.PostgresqlVersion)
	}
	if dbi.ResourceURL != "postgres://username:password@ec2-107-12-34-82.compute-1.amazonaws.com:5552/dbname" {
		t.Errorf("expected ResourceURL=%s, got %s", "postgres://username:password@ec2-107-12-34-82.compute-1.amazonaws.com:5552/dbname", dbi.ResourceURL)
	}
	if dbi.ServicePort != 5552 {
		t.Errorf("expected ServicePort=%d, got %d", 5552, dbi.ServicePort)
	}
	if dbi.Standalone != false {
		t.Errorf("expected Standalone=%t, got %t", false, dbi.Standalone)
	}
	tUpdated, err := time.Parse(time.RFC3339, "2014-01-16T00:43:52+00:00")
	if err != nil {
		t.Fatal(err)
	}
	if !tUpdated.Equal(dbi.StatusUpdatedAt) {
		t.Errorf("expected StatusUpdatedAt=%s, got %s", tUpdated, dbi.StatusUpdatedAt)
	}
	if dbi.TargetTransaction != "5" {
		t.Errorf("expected TargetTransaction=%s, got %s", "5", dbi.TargetTransaction)
	}
}

func TestInfoEntryListNamed(t *testing.T) {
	var dbi DBInfo
	err := json.Unmarshal([]byte(pgInfoResponse), &dbi)
	if err != nil {
		t.Fatal(err)
	}
	ie := dbi.Info.Named("Plan")
	if ie == nil {
		t.Fatal("expected to find Plan info")
	}
	expected := InfoEntry{
		Name:   "Plan",
		Values: []interface{}{"Standard Tengu"},
	}
	if !reflect.DeepEqual(*ie, expected) {
		t.Errorf("expected %+v, got %+v", expected, *ie)
	}
}

var getStringTests = []struct {
	Name            string
	ExpectedValue   string
	ExpectedResolve bool
}{
	{Name: "Plan", ExpectedValue: "Standard Tengu"},
	{Name: "Status", ExpectedValue: "Available"},
	{Name: "Tables", ExpectedValue: "0"},
	{Name: "Status", ExpectedValue: "Available"},
	{Name: "PG Version", ExpectedValue: "9.3.2"},
	{Name: "Forks", ExpectedValue: "postgres://myfakedb.com/dbname", ExpectedResolve: true},
	{Name: "Followers", ExpectedValue: "", ExpectedResolve: true},
	{Name: "Maintenance", ExpectedValue: "not required"},
}

func TestGetString(t *testing.T) {
	var dbi DBInfo
	err := json.Unmarshal([]byte(pgInfoResponse), &dbi)
	if err != nil {
		t.Fatal(err)
	}
	for _, st := range getStringTests {
		value, resolve := dbi.Info.GetString(st.Name)
		if value != st.ExpectedValue {
			t.Errorf("expected %s value of %q, got %q", st.Name, st.ExpectedValue, value)
		}
		if resolve != st.ExpectedResolve {
			t.Errorf("expected %s resolve of %t, got %t", st.Name, st.ExpectedResolve, resolve)
		}
	}
}
