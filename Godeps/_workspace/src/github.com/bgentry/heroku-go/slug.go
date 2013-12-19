// WARNING: This code is auto-generated from the Heroku Platform API JSON Schema
// by a Ruby script (gen/gen.rb). Changes should be made to the generation
// script rather than the generated files.

package heroku

import (
	"time"
)

// A slug is a snapshot of your application code that is ready to run on the
// platform.
type Slug struct {
	// HTTP verb and url where clients can fetch or store the release blob file
	Blob map[string]string `json:"blob"`

	// identification of the code with your version control system (eg: SHA of the git HEAD)
	Commit *string `json:"commit"`

	// when slug was created
	CreatedAt time.Time `json:"created_at"`

	// unique identifier of slug
	Id string `json:"id"`

	// hash mapping process type names to their respective command
	ProcessTypes map[string]string `json:"process_types"`

	// when slug was updated
	UpdatedAt time.Time `json:"updated_at"`
}

// Info for existing slug.
//
// appIdentity is the unique identifier of the slug's app. slugIdentity is the
// unique identifier of the Slug.
func (c *Client) SlugInfo(appIdentity string, slugIdentity string) (*Slug, error) {
	var slug Slug
	return &slug, c.Get(&slug, "/apps/"+appIdentity+"/slugs/"+slugIdentity)
}

// Create a new slug. For more information please refer to Deploying Slugs using
// the Platform API.
//
// appIdentity is the unique identifier of the slug's app. processTypes is the
// hash mapping process type names to their respective command. options is the
// struct of optional parameters for this action.
func (c *Client) SlugCreate(appIdentity string, processTypes map[string]string, options *SlugCreateOpts) (*Slug, error) {
	params := struct {
		ProcessTypes map[string]string `json:"process_types"`
		Commit       *string           `json:"commit,omitempty"`
	}{
		ProcessTypes: processTypes,
		Commit:       options.Commit,
	}
	var slugRes Slug
	return &slugRes, c.Post(&slugRes, "/apps/"+appIdentity+"/slugs", params)
}

// SlugCreateOpts holds the optional parameters for SlugCreate
type SlugCreateOpts struct {
	// identification of the code with your version control system (eg: SHA of the git HEAD)
	Commit *string `json:"commit,omitempty"`
}
