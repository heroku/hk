package main

import (
	"log"
	"net/url"
	"os"
	"os/exec"
	"syscall"
)

var helpPlugins = &Command{
	Usage: "plugins",
	Short: "interface to plugin commands",
	Long: `
Plugin commands extend hk's functionality.

Plugins are located in one of the directories in hk's search path,
HKPATH, and executed for unrecognized hk commands. If a plugin named
"default" exists, it will be run when no suitably-named plugin can
be found. (See 'hk help environ' for details on HKPATH.)

The arguments to the plugin are the arguments to hk, not including
"hk" itself.

Several environment variables will also be set:

HEROKU_API_URL

  This follows the same format as the variable read by hk. For a
  plugin, this variable is always set and it always includes a
  username and password.

  (See 'hk help environ' for details of the format.)

HKUSER

  The username from HEROKU_API_URL, for convenience.

HKPASS

  The password from HEROKU_API_URL, for convenience.

HKHOST

  The hostname (and port, if any) from HEROKU_API_URL, for
  convenience.

HKAPP

  The name of the heroku app in the current directory, if there is a
  git remote named "heroku" with the proper URL format.

HKVERSION

  The version string of hk that executed the plugin.
`,
}

var hkPath string

func init() {
	const defaultPluginPath = "/usr/local/lib/hk/plugin"
	hkPath = os.Getenv("HKPATH")
	if hkPath == "" {
		hkPath = defaultPluginPath
	}

}

func execPlugin(path string, args []string) error {
	u, err := url.Parse(apiURL)
	if err != nil {
		log.Fatal(err)
	}

	hkuser, hkpass := getCreds(u)
	u.User = url.UserPassword(hkuser, hkpass)
	hkapp, _ := app()
	env := []string{
		"HEROKU_API_URL=" + u.String(),
		"HKAPP=" + hkapp,
		"HKUSER=" + hkuser,
		"HKPASS=" + hkpass,
		"HKHOST=" + u.Host,
		"HKVERSION=" + Version,
	}

	return syscall.Exec(path, args, append(env, os.Environ()...))
}

func findPlugin(name string) (path string) {
	path = lookupPlugin(name)
	if path == "" {
		path = lookupPlugin("default")
	}
	return path
}

// NOTE: lookupPlugin is not threadsafe for anything needing the PATH env var.
func lookupPlugin(name string) string {
	opath := os.Getenv("PATH")
	defer os.Setenv("PATH", opath)
	os.Setenv("PATH", hkPath)

	path, err := exec.LookPath(name)
	if err != nil {
		if e, ok := err.(*exec.Error); ok && e.Err == exec.ErrNotFound {
			return ""
		}
		log.Fatal(err)
	}
	return path
}
