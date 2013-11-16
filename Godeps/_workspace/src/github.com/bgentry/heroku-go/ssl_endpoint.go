// WARNING: This code is auto-generated from the Heroku Platform API JSON Schema
// by a Ruby script (gen/gen.rb). Changes should be made to the generation
// script rather than the generated files.

package heroku

import (
	"time"
)

// SSL Endpoint is a public address serving custom SSL cert for HTTPS traffic to
// a Heroku app. Note that an app must have the ssl:endpoint addon installed
// before it can provision an SSL Endpoint using these APIs.
type SslEndpoint struct {
	// raw contents of the public certificate chain (eg: .crt or .pem file)
	CertificateChain string `json:"certificate_chain"`

	// canonical name record, the address to point a domain at
	Cname string `json:"cname"`

	// when endpoint was created
	CreatedAt time.Time `json:"created_at"`

	// unique identifier of this SSL endpoint
	Id string `json:"id"`

	// unique name for SSL endpoint
	Name string `json:"name"`

	// when endpoint was updated
	UpdatedAt time.Time `json:"updated_at"`
}

// Create a new SSL endpoint.
//
// appIdentity is the unique identifier of the ssl-endpoint's app.
// certificateChain is the raw contents of the public certificate chain (eg:
// .crt or .pem file). privateKey is the contents of the private key (eg .key
// file).
func (c *Client) SslEndpointCreate(appIdentity string, certificateChain string, privateKey string) (*SslEndpoint, error) {
	params := struct {
		CertificateChain string `json:"certificate_chain"`
		PrivateKey       string `json:"private_key"`
	}{
		CertificateChain: certificateChain,
		PrivateKey:       privateKey,
	}
	var sslEndpointRes SslEndpoint
	return &sslEndpointRes, c.Post(&sslEndpointRes, "/apps/"+appIdentity+"/ssl-endpoints", params)
}

// Delete existing SSL endpoint.
//
// appIdentity is the unique identifier of the ssl-endpoint's app.
// sslEndpointIdentity is the unique identifier of the SslEndpoint.
func (c *Client) SslEndpointDelete(appIdentity string, sslEndpointIdentity string) error {
	return c.Delete("/apps/" + appIdentity + "/ssl-endpoints/" + sslEndpointIdentity)
}

// Info for existing SSL endpoint.
//
// appIdentity is the unique identifier of the ssl-endpoint's app.
// sslEndpointIdentity is the unique identifier of the SslEndpoint.
func (c *Client) SslEndpointInfo(appIdentity string, sslEndpointIdentity string) (*SslEndpoint, error) {
	var sslEndpoint SslEndpoint
	return &sslEndpoint, c.Get(&sslEndpoint, "/apps/"+appIdentity+"/ssl-endpoints/"+sslEndpointIdentity)
}

// List existing SSL endpoints.
//
// appIdentity is the unique identifier of the ssl-endpoint's app. lr is an
// optional ListRange that sets the Range options for the paginated list of
// results.
func (c *Client) SslEndpointList(appIdentity string, lr *ListRange) ([]SslEndpoint, error) {
	req, err := c.NewRequest("GET", "/apps/"+appIdentity+"/ssl-endpoints", nil)
	if err != nil {
		return nil, err
	}

	if lr != nil {
		lr.SetHeader(req)
	}

	var sslEndpointsRes []SslEndpoint
	return sslEndpointsRes, c.DoReq(req, &sslEndpointsRes)
}

// Update an existing SSL endpoint.
//
// appIdentity is the unique identifier of the ssl-endpoint's app.
// sslEndpointIdentity is the unique identifier of the SslEndpoint. options is
// the struct of optional parameters for this call.
func (c *Client) SslEndpointUpdate(appIdentity string, sslEndpointIdentity string, options SslEndpointUpdateOpts) (*SslEndpoint, error) {
	var sslEndpointRes SslEndpoint
	return &sslEndpointRes, c.Patch(&sslEndpointRes, "/apps/"+appIdentity+"/ssl-endpoints/"+sslEndpointIdentity, options)
}

// SslEndpointUpdateOpts holds the optional parameters for SslEndpointUpdate
type SslEndpointUpdateOpts struct {
	// raw contents of the public certificate chain (eg: .crt or .pem file)
	CertificateChain *string `json:"certificate_chain,omitempty"`
	// contents of the private key (eg .key file)
	PrivateKey *string `json:"private_key,omitempty"`
	// indicates that a rollback should be performed
	Rollback *bool `json:"rollback,omitempty"`
}
