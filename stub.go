package main

import (
	"fmt"
)

func runStub(cmd *Command, args []string) {
	fmt.Println(cmd.Name, "is a stub")
}

var cmdAddonAdd = &Command{
	Run:      runStub,
	Usage:    "addon-add <name>",
	Short:    "add an addon",
	NeedsApp: true,
}

var cmdAddonRemove = &Command{
	Run:      runStub,
	Usage:    "addon-remove <name>",
	Short:    "remove an addon",
	NeedsApp: true,
}

var cmdLogin = &Command{
	Run:   runStub,
	Usage: "login <user>",
	Short: "login to your heroku account",
}

var cmdLogout = &Command{
	Run:   runStub,
	Usage: "logout",
	Short: "log out of your heroku account",
}

var cmdDomains = &Command{
	Run:      runStub,
	Usage:    "domains",
	Short:    "list domains",
	NeedsApp: true,
}

var cmdDomainAdd = &Command{
	Run:      runStub,
	Usage:    "domain-add <domain>",
	Short:    "add a domain",
	NeedsApp: true,
}

var cmdDomainRemove = &Command{
	Run:      runStub,
	Usage:    "domain-remove <domain>",
	Short:    "remove a domain",
	NeedsApp: true,
}

var cmdResize = &Command{
	Run:      runStub,
	Usage:    "resize <type>=<size> ...",
	Short:    "resize dynos for process type",
	NeedsApp: true,
}

var cmdReleaseInfo = &Command{
	Run:      runStub,
	Usage:    "release-info <release>",
	Short:    "show info for a release",
	NeedsApp: true,
}

var cmdReleaseCreate = &Command{
	Run:      runStub,
	Usage:    "release-create [<args>]",
	Short:    "create new release" + extra,
	NeedsApp: true,
}

var cmdRollback = &Command{
	Run:      runStub,
	Usage:    "rollback <release>",
	Short:    "rollback to an old release",
	NeedsApp: true,
}

var cmdCollaborators = &Command{
	Run:      runStub,
	Usage:    "collaborators <user>",
	Short:    "list collaborators" + extra,
	NeedsApp: true,
}

var cmdCollaboratorAdd = &Command{
	Run:      runStub,
	Usage:    "collaborator-add <user>",
	Short:    "add a collaborator" + extra,
	NeedsApp: true,
}

var cmdCollaboratorRemove = &Command{
	Run:      runStub,
	Usage:    "collaborator-remove <user>",
	Short:    "remove a collaborator" + extra,
	NeedsApp: true,
}

var cmdTransfer = &Command{
	Run:      runStub,
	Usage:    "transfer <user>",
	Short:    "transfer an app to another user" + extra,
	NeedsApp: true,
}

var cmdSslEndpoints = &Command{
	Run:      runStub,
	Usage:    "ssl-endpoints",
	Short:    "list ssl endpoints" + extra,
	NeedsApp: true,
}

var cmdSslEndpointAdd = &Command{
	Run:      runStub,
	Usage:    "ssl-endpoint-add <certfile> <keyfile>",
	Short:    "add an ssl endpoint" + extra,
	NeedsApp: true,
}

var cmdSslEndpointRemove = &Command{
	Run:      runStub,
	Usage:    "ssl-endpoint-remove <endpoint>",
	Short:    "remove an ssl endpoint" + extra,
	NeedsApp: true,
}

var cmdSslEndpointUpdate = &Command{
	Run:      runStub,
	Usage:    "ssl-endpoint-update <certfile> <keyfile>",
	Short:    "update cert on an ssl endpoint" + extra,
	NeedsApp: true,
}

var cmdSslEndpointRollback = &Command{
	Run:      runStub,
	Usage:    "ssl-endpoint-rollback",
	Short:    "rollback cert on an ssl endpoint" + extra,
	NeedsApp: true,
}

var cmdDrains = &Command{
	Run:      runStub,
	Usage:    "drains [-l]",
	Short:    "list log drains" + extra,
	NeedsApp: true,
}

var cmdDrainAdd = &Command{
	Run:      runStub,
	Usage:    "drain-add <url>",
	Short:    "add a log drain" + extra,
	NeedsApp: true,
}

var cmdDrainRemove = &Command{
	Run:      runStub,
	Usage:    "drain-remove <url>",
	Short:    "remove a log drain" + extra,
	NeedsApp: true,
}

var cmdFork = &Command{
	Run:      runStub,
	Usage:    "fork [-r <region>] [<newname>]",
	Short:    "fork an app" + extra,
	NeedsApp: true,
}

var cmdKeys = &Command{
	Run:   runStub,
	Usage: "keys",
	Short: "list ssh keys for account" + extra,
}

var cmdKeyRemove = &Command{
	Run:   runStub,
	Usage: "key-remove <key>",
	Short: "remove an ssh key from account" + extra,
}

var cmdAccountFeatures = &Command{
	Run:   runStub,
	Usage: "account-features",
	Short: "list account features" + extra,
}

var cmdAccountFeatureInfo = &Command{
	Run:   runStub,
	Usage: "account-feature-info",
	Short: "show account feature info" + extra,
}

var cmdAccountFeatureEnable = &Command{
	Run:   runStub,
	Usage: "account-feature-enable <feature>",
	Short: "enable account feature" + extra,
}

var cmdAccountFeatureDisable = &Command{
	Run:   runStub,
	Usage: "account-feature-disable <feature>",
	Short: "disable account feature" + extra,
}

var cmdAppFeatures = &Command{
	Run:      runStub,
	Usage:    "app-features",
	Short:    "list app features" + extra,
	NeedsApp: true,
}

var cmdAppFeatureInfo = &Command{
	Run:      runStub,
	Usage:    "app-feature-info",
	Short:    "show app feature info" + extra,
	NeedsApp: true,
}

var cmdAppFeatureEnable = &Command{
	Run:      runStub,
	Usage:    "app-feature-enable <feature>",
	Short:    "enable app feature" + extra,
	NeedsApp: true,
}

var cmdAppFeatureDisable = &Command{
	Run:      runStub,
	Usage:    "app-feature-disable <feature>",
	Short:    "disable app feature" + extra,
	NeedsApp: true,
}

var cmdMaintenance = &Command{
	Run:      runStub,
	Usage:    "maintenance",
	Short:    "show maintenance mode status" + extra,
	NeedsApp: true,
}

var cmdMaintenanceEnable = &Command{
	Run:      runStub,
	Usage:    "maintenance-enable",
	Short:    "enable app maintenance mode" + extra,
	NeedsApp: true,
}

var cmdMaintenanceDisable = &Command{
	Run:      runStub,
	Usage:    "maintenance-disable",
	Short:    "disable app maintenance mode" + extra,
	NeedsApp: true,
}

var cmds = `
addon-add
addon-remove
login
logout
domains
domain-add
domain-remove
dyno-resize
release-info
release-create
release-rollback
share
unshare
transfer
ssl-endpoints
ssl-endpoint-add
ssl-endpoint-remove
ssl-endpoint-update
ssl-endpoint-rollback
drains
drain-add
drain-remove
fork
keys
key-remove
account-features
account-feature-info
account-feature-enable
account-feature-disable
app-features
app-feature-info
app-feature-enable
app-feature-disable
maintenance
maintenance-on
maintenance-off
pg
pg-info
pg-promote
pg-psql
pg-reset
pg-unfollow
pg-wait
pg-credentials
pgbackups
pgbackup-create
pgbackup-destroy
pgbackup-restore
pgbackup-transfer
pgbackup-url
plugin-add
plugin-remove
regions
status
`
