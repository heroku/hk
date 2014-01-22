package postgresql

import (
	"strings"
	"time"
)

type DB struct {
	Id     string
	Plan   string
	client *Client
}

type DBInfo struct {
	AvailableForIngress   bool   `json:"available_for_ingress"`
	CreatedAt             string `json:"created_at"`
	CurrentTransaction    string `json:"current_transaction"`
	DatabaseName          string `json:"database_name"`
	DatabasePassword      string `json:"database_password"`
	DatabaseUser          string `json:"database_user"`
	Info                  []InfoEntry
	IsInRecovery          bool `json:"is_in_recovery?"`
	NumBytes              int  `json:"num_bytes"`
	NumConnections        int  `json:"num_connections"`
	NumConnectionsWaiting int  `json:"num_connections_waiting"`
	NumTables             int  `json:"num_tables"`
	Plan                  string
	PostgresqlVersion     string    `json:"postgresql_version"`
	ResourceURL           string    `json:"resource_url"`
	ServicePort           string    `json:"service_port"`
	StatusUpdatedAt       time.Time `json:"status_updated_at"`
	Standalone            string    `json:"standalone?"`
	TargetTransaction     string    `json:"target_transaction"`
}

type InfoEntry struct {
	Name          string
	ResolveDBName bool `json:"resolve_db_name"`
	Values        []interface{}
}

func (d *DB) Info() (dbi DBInfo, err error) {
	err = d.client.Get(d.IsStarterPlan(), "/"+d.Id, &dbi)
	return
}

func (d *DB) Ingress() error {
	return d.client.Put(d.IsStarterPlan(), "/"+d.Id+"/ingress", nil)
}

// Whether the DB is a starter plan and should communicate with the starter API.
// Plan names ending in "dev" or "basic" are currently handled by the starter
// API while all others are handled by the production API.
func (d *DB) IsStarterPlan() bool {
	return strings.HasSuffix(d.Plan, "dev") || strings.HasSuffix(d.Plan, "basic")
}

func (d *DB) Reset() error {
	return d.client.Put(d.IsStarterPlan(), "/"+d.Id+"/reset", nil)
}

func (d *DB) RotateCredentials() error {
	return d.client.Post(d.IsStarterPlan(), "/"+d.Id+"/credentials_rotation", nil)
}

func (d *DB) Unfollow() error {
	return d.client.Put(d.IsStarterPlan(), "/"+d.Id+"/unfollow", nil)
}

type WaitStatus struct {
}

func (d *DB) WaitStatus() (ws WaitStatus, err error) {
	err = d.client.Get(d.IsStarterPlan(), "/"+d.Id+"/wait_status", &ws)
	return
}
