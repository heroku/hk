// WARNING: This code is auto-generated from the Heroku Platform API JSON Schema
// by a Ruby script (gen/gen.rb). Changes should be made to the generation
// script rather than the generated files.

package heroku

import (
	"time"
)

// Add-ons represent add-ons that have been provisioned for an app.
type Addon struct {
	// when add-on was updated
	CreatedAt time.Time `json:"created_at"`

	// unique identifier of add-on
	Id string `json:"id"`

	// identity of add-on plan
	Plan struct {
		Id   string `json:"id"`
		Name string `json:"name"`
	} `json:"plan"`

	// when add-on was updated
	UpdatedAt time.Time `json:"updated_at"`
}

// Create a new add-on.
//
// appIdentity is the unique identifier of the addon's app. plan is the unique
// identifier of this plan or unique name of this plan. options is the struct of
// optional parameters for this call.
func (c *Client) AddonCreate(appIdentity string, plan string, options AddonCreateOpts) (*Addon, error) {
	params := struct {
		Plan   string             `json:"plan"`
		Config *map[string]string `json:"config,omitempty"`
	}{
		Plan:   plan,
		Config: options.Config,
	}
	var addonRes Addon
	return &addonRes, c.Post(&addonRes, "/apps/"+appIdentity+"/addons", params)
}

// AddonCreateOpts holds the optional parameters for AddonCreate
type AddonCreateOpts struct {
	// custom add-on provisioning options
	Config *map[string]string `json:"config,omitempty"`
}

// Delete an existing add-on.
//
// appIdentity is the unique identifier of the addon's app. addonIdentity is the
// unique identifier of the Addon.
func (c *Client) AddonDelete(appIdentity string, addonIdentity string) error {
	return c.Delete("/apps/" + appIdentity + "/addons/" + addonIdentity)
}

// Info for an existing add-on.
//
// appIdentity is the unique identifier of the addon's app. addonIdentity is the
// unique identifier of the Addon.
func (c *Client) AddonInfo(appIdentity string, addonIdentity string) (*Addon, error) {
	var addon Addon
	return &addon, c.Get(&addon, "/apps/"+appIdentity+"/addons/"+addonIdentity)
}

// List existing add-ons.
//
// appIdentity is the unique identifier of the addon's app. lr is an optional
// ListRange that sets the Range options for the paginated list of results.
func (c *Client) AddonList(appIdentity string, lr *ListRange) ([]Addon, error) {
	req, err := c.NewRequest("GET", "/apps/"+appIdentity+"/addons", nil)
	if err != nil {
		return nil, err
	}

	if lr != nil {
		lr.SetHeader(req)
	}

	var addonsRes []Addon
	return addonsRes, c.DoReq(req, &addonsRes)
}

// Update an existing add-on.
//
// appIdentity is the unique identifier of the addon's app. addonIdentity is the
// unique identifier of the Addon. plan is the unique identifier of this plan or
// unique name of this plan.
func (c *Client) AddonUpdate(appIdentity string, addonIdentity string, plan string) (*Addon, error) {
	params := struct {
		Plan string `json:"plan"`
	}{
		Plan: plan,
	}
	var addonRes Addon
	return &addonRes, c.Patch(&addonRes, "/apps/"+appIdentity+"/addons/"+addonIdentity, params)
}
