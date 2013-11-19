// WARNING: This code is auto-generated from the Heroku Platform API JSON Schema
// by a Ruby script (gen/gen.rb). Changes should be made to the generation
// script rather than the generated files.

package heroku

import (
	"time"
)

// The formation of processes that should be maintained for an app. Update the
// formation to scale processes or change dyno sizes. Commands and types are
// defined by the Procfile uploaded with an app.
type Formation struct {
	// command to use to launch this process
	Command string `json:"command"`

	// when process type was created
	CreatedAt time.Time `json:"created_at"`

	// unique identifier of this process type
	Id string `json:"id"`

	// number of processes to maintain
	Quantity int `json:"quantity"`

	// dyno size (default: "1")
	Size string `json:"size"`

	// type of process to maintain
	Type string `json:"type"`

	// when dyno type was updated
	UpdatedAt time.Time `json:"updated_at"`
}

// Info for a process type
//
// appIdentity is the unique identifier of the formation's app.
// formationIdentity is the unique identifier of the Formation.
func (c *Client) FormationInfo(appIdentity string, formationIdentity string) (*Formation, error) {
	var formation Formation
	return &formation, c.Get(&formation, "/apps/"+appIdentity+"/formation/"+formationIdentity)
}

// List process type formation
//
// appIdentity is the unique identifier of the formation's app. lr is an
// optional ListRange that sets the Range options for the paginated list of
// results.
func (c *Client) FormationList(appIdentity string, lr *ListRange) ([]Formation, error) {
	req, err := c.NewRequest("GET", "/apps/"+appIdentity+"/formation", nil)
	if err != nil {
		return nil, err
	}

	if lr != nil {
		lr.SetHeader(req)
	}

	var formationsRes []Formation
	return formationsRes, c.DoReq(req, &formationsRes)
}

// Update process type
//
// appIdentity is the unique identifier of the formation's app.
// formationIdentity is the unique identifier of the Formation. options is the
// struct of optional parameters for this call.
func (c *Client) FormationUpdate(appIdentity string, formationIdentity string, options FormationUpdateOpts) (*Formation, error) {
	var formationRes Formation
	return &formationRes, c.Patch(&formationRes, "/apps/"+appIdentity+"/formation/"+formationIdentity, options)
}

// FormationUpdateOpts holds the optional parameters for FormationUpdate
type FormationUpdateOpts struct {
	// number of processes to maintain
	Quantity *int `json:"quantity,omitempty"`
	// dyno size (default: "1")
	Size *string `json:"size,omitempty"`
}
