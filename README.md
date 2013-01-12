## hk

Fast Heroku client.

### Overview

hk is a command line client to the Heroku runtime platform, designed to be as fast as possible.

### Motivation

**Fast as a feature**

	## version

	$ time heroku version >/dev/null
	real	0m1.813s

	$ time hk version >/dev/null
	real	0m0.016s

	## list

	$ time heroku list >/dev/null
	real	0m8.826s

	$ time hk list >/dev/null
	real  0m3.658s

**Focus on API**

We believe this is evidence that a first-class Heroku API trumps any particular client.

**Iterative Development**

A release mechanism was created for hk in the beginning: the binary updates
itself. This gives us confidence in iterative development, which we value
highly, because it gives us the ability to release very often knowing users will
see the updates in a timely manner.

**The power of Go**

hk demonstrates the power of the Go language, runtime, systems access, and
distribution story (a single, statically linked binary with no dependencies) which are all very appealing to Heroku.

**Unix**

Heroku loves Unix. This client should reflect that. Commands are single,
unhyphenated words that map to their unix ancestors’ names and flags where
applicable.

### Installation

	$ wget -qO- https://hk.heroku.com/hk.gz | zcat >/usr/local/bin/hk
	$ chmod +x /usr/local/bin/hk

### netrc

You'll need a line like this in $HOME/.netrc

	machine api.heroku.com login <email> password <apitoken>

### Usage

	$ hk help
	Usage: hk <command> [options] [arguments]

	Supported commands are:

  	create     create an app
    rename     rename an app
  	destroy    destroy an app
  	creds      show auth creds
  	env        list config vars
  	get        get config var
  	set        set config var
  	info       show app info
    rels       show releases and info
  	list       list apps
  	open       open app
  	ps         list processes
  	scale      change dyno counts
  	tail       tail log files
  	run        run a process
  	version    show hk version
  	help       show help

	See 'hk help [command]' for more information about a command.

## Plugins

Plugins are executables located in HKPATH or, if HKPATH does not exist, in /usr/local/lib/hk/plugin. They are executed when hk does not know command X and an installed plugin X exists. The special case default plugin will be executed iif hk has no command or installed plugin named X.

hk will set these environment variables for a plugin:

* HEROKU_API_URL - The url containing the username, password, and host to the api endpoint.
* HKAPP - The app as determined by the git heroku remote, if available.
* HKUSER - The username from either HEROKU_API_URL or .netrc
* HKPASS - The password from either HEROKU_API_URL or .netrc
* HKHOST - The hostname for the API endpoint

### Development

	$ cd hk
	$ go get
	$ mate main.go
	$ go build
	$ ./hk list

### Release

	$ cd hk
	$ vim main.go # edit Version
	$ go build
	$ ./pub # publish ./hk
	$ ./def # make `./hk version` the default
