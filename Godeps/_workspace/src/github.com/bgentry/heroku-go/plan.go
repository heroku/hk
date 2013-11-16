// WARNING: This code is auto-generated from the Heroku Platform API JSON Schema
// by a Ruby script (gen/gen.rb). Changes should be made to the generation
// script rather than the generated files.

package heroku

import (
	"time"
)

// Plans represent different configurations of add-ons that may be added to
// apps.
type Plan struct {
	// when plan was created
	CreatedAt time.Time `json:"created_at"`

	// description of plan
	Description string `json:"description"`

	// unique identifier of this plan
	Id string `json:"id"`

	// unique name of this plan
	Name string `json:"name"`

	// price
	Price struct {
		Cents int    `json:"cents"`
		Unit  string `json:"unit"`
	} `json:"price"`

	// release status for plan
	State string `json:"state"`

	// when plan was updated
	UpdatedAt time.Time `json:"updated_at"`
}

// Info for existing plan.
//
// planIdentity is the unique identifier of the Plan.
func (c *Client) PlanInfo(planIdentity string) (*Plan, error) {
	var plan Plan
	return &plan, c.Get(&plan, "/addon-services/{(%2Fschema%2Faddon-service%23%2Fdefinitions%2Fidentity)}/plans/"+planIdentity)
}

// List existing plans.
//
// lr is an optional ListRange that sets the Range options for the paginated
// list of results.
func (c *Client) PlanList(lr *ListRange) ([]Plan, error) {
	req, err := c.NewRequest("GET", "/addon-services/{(%2Fschema%2Faddon-service%23%2Fdefinitions%2Fidentity)}/plans", nil)
	if err != nil {
		return nil, err
	}

	if lr != nil {
		lr.SetHeader(req)
	}

	var plansRes []Plan
	return plansRes, c.DoReq(req, &plansRes)
}
