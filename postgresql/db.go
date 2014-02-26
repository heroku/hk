package postgresql

import (
	"fmt"
	"strings"
	"time"
)

type DB struct {
	Id     string
	Plan   string
	client *Client
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
	return strings.HasSuffix(d.Plan, "dev") ||
		strings.HasSuffix(d.Plan, "basic") ||
		// special exception for devcloud plans:
		strings.HasSuffix(d.Plan, "devcloud")
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

func (d *DB) WaitStatus() (ws WaitStatus, err error) {
	err = d.client.Get(d.IsStarterPlan(), "/"+d.Id+"/wait_status", &ws)
	return
}

type WaitStatus struct {
}

type DBInfo struct {
	AvailableForIngress   bool          `json:"available_for_ingress"`
	CreatedAt             string        `json:"created_at"`
	CurrentTransaction    string        `json:"current_transaction"`
	DatabaseName          string        `json:"database_name"`
	DatabasePassword      string        `json:"database_password"`
	DatabaseUser          string        `json:"database_user"`
	Following             string        `json:"following"`
	Info                  InfoEntryList `json:"info"`
	IsInRecovery          bool          `json:"is_in_recovery?"`
	NumBytes              int           `json:"num_bytes"`
	NumConnections        int           `json:"num_connections"`
	NumConnectionsWaiting int           `json:"num_connections_waiting"`
	NumTables             int           `json:"num_tables"`
	Plan                  string        `json:"plan"`
	PostgresqlVersion     string        `json:"postgresql_version"`
	ResourceURL           string        `json:"resource_url"`
	ServicePort           string        `json:"service_port"`
	StatusUpdatedAt       time.Time     `json:"status_updated_at"`
	Standalone            string        `json:"standalone?"`
	TargetTransaction     string        `json:"target_transaction"`
}

func (dbi *DBInfo) IsFollower() bool {
	return dbi.Following != ""
}

type InfoEntryList []InfoEntry

func (iel *InfoEntryList) Named(name string) *InfoEntry {
	if iel == nil {
		return nil
	}
	for i := range *iel {
		if (*iel)[i].Name == name {
			return &(*iel)[i]
		}
	}
	return nil
}

func (iel *InfoEntryList) GetString(key string) (valstr string, resolve bool) {
	ie := iel.Named(key)
	if ie == nil {
		return
	}
	resolve = ie.ResolveDBName
	if len(ie.Values) > 0 {
		valstr = fmt.Sprintf("%v", ie.Values[0])
	}
	return
}

type InfoEntry struct {
	Name          string
	ResolveDBName bool `json:"resolve_db_name"`
	Values        []interface{}
}
