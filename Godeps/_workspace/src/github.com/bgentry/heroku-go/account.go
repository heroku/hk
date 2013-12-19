// WARNING: This code is auto-generated from the Heroku Platform API JSON Schema
// by a Ruby script (gen/gen.rb). Changes should be made to the generation
// script rather than the generated files.

package heroku

import (
	"time"
)

// An account represents an individual signed up to use the Heroku platform.
type Account struct {
	// whether to allow third party web activity tracking
	AllowTracking bool `json:"allow_tracking"`

	// whether to utilize beta Heroku features
	Beta *bool `json:"beta"`

	// when account was created
	CreatedAt time.Time `json:"created_at"`

	// unique email address of account
	Email string `json:"email"`

	// unique identifier of an account
	Id string `json:"id"`

	// when account last authorized with Heroku
	LastLogin time.Time `json:"last_login"`

	// when account was updated
	UpdatedAt time.Time `json:"updated_at"`

	// whether account has been verified with billing information
	Verified bool `json:"verified"`
}

// Info for account.
//
// accountIdentity is the unique identifier of the Account.
func (c *Client) AccountInfo() (*Account, error) {
	var account Account
	return &account, c.Get(&account, "/account")
}

// Update account.
//
// accountIdentity is the unique identifier of the Account. password is the
// current password on the account. options is the struct of optional parameters
// for this action.
func (c *Client) AccountUpdate(password string, options *AccountUpdateOpts) (*Account, error) {
	params := struct {
		Password      string  `json:"password"`
		AllowTracking *bool   `json:"allow_tracking,omitempty"`
		Beta          *bool   `json:"beta,omitempty"`
		Name          *string `json:"name,omitempty"`
	}{
		Password:      password,
		AllowTracking: options.AllowTracking,
		Beta:          options.Beta,
		Name:          options.Name,
	}
	var accountRes Account
	return &accountRes, c.Patch(&accountRes, "/account", params)
}

// AccountUpdateOpts holds the optional parameters for AccountUpdate
type AccountUpdateOpts struct {
	// whether to allow third party web activity tracking
	AllowTracking *bool `json:"allow_tracking,omitempty"`
	// whether to utilize beta Heroku features
	Beta *bool `json:"beta,omitempty"`
	// full name of the account owner
	Name *string `json:"name,omitempty"`
}

// Change Email for account.
//
// accountIdentity is the unique identifier of the Account. password is the
// current password on the account. email is the unique email address of
// account.
func (c *Client) AccountChangeEmail(password string, email string) (*Account, error) {
	params := struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}{
		Password: password,
		Email:    email,
	}
	var accountRes Account
	return &accountRes, c.Patch(&accountRes, "/account", params)
}

// Change Password for account.
//
// accountIdentity is the unique identifier of the Account. newPassword is the
// the new password for the account when changing the password. password is the
// current password on the account.
func (c *Client) AccountChangePassword(newPassword string, password string) (*Account, error) {
	params := struct {
		NewPassword string `json:"new_password"`
		Password    string `json:"password"`
	}{
		NewPassword: newPassword,
		Password:    password,
	}
	var accountRes Account
	return &accountRes, c.Patch(&accountRes, "/account", params)
}
