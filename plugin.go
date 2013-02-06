package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

var (
	hkPath string
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

HKPLUGINMODE

  Either unset or it takes the value "info". If set to info, the
  plugin should print out a summary of itself in the following
  format:

    name version: short help

    long help

  Where name is the plugin's file name, version is the plugin's
  version string, short help is a one-line help message at most 50 chars,
  and long help is a complete help text including usage line, prose
  description, and list of options. Plugins are encouraged to follow the
  example set by built-in hk commands for the style of this documentation.
`,
}

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

type plugin string

func (p plugin) Name() string {
	return string(p)
}

func (p plugin) Short() string {
	_, short, _ := pluginInfo(string(p))
	return short
}

func pluginInfo(name string) (ver, short, long string) {
	if os.Getenv("HKPLUGINMODE") == "info" {
		return "", "[plugin exec loop]", "[plugin exec loop]\n"
	}
	var cmd exec.Cmd
	cmd.Args = []string{name}
	cmd.Path = lookupPlugin(name)
	cmd.Env = append([]string{"HKPLUGINMODE=info"}, os.Environ()...)
	buf, err := cmd.Output()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return "", "[unknown description]", "[unknown description]\n"
	}
	info := string(buf)
	if !strings.HasPrefix(info, name+" ") {
		return "", "[unknown description]", "[unknown description]\n"
	}
	info = info[len(name)+1:]
	i := strings.Index(info, ": ")
	if i < 0 {
		return "", "[unknown description]", "[unknown description]\n"
	}
	ver, info = info[:i], info[i+2:]
	i = strings.Index(info, "\n\n")
	if i < 0 {
		return "", "[unknown description]", "[unknown description]\n"
	}
	short, long = info[:i], info[i+2:]
	if len(short) > 50 || strings.Contains(short, "\n") {
		return "", "[unknown description]", "[unknown description]\n"
	}
	return ver, short, long
}
