## hk

Fast Heroku client.

### Overview

hk is a command line client to the Heroku runtime platform, designed to be as fast as possible.

#### Disclaimer

This is pre-alpha software. It will undergo substantial changes during the
coming months. Don't be surprised if it is broken or if the interface is
completely altered without warning.

Issues and pull requests are still welcome, but please understand that they may
be rejected if it's an issue we just aren't ready to tackle this early in the
project.

### Motivation

#### Fast as a feature

```
## version

$ time heroku version >/dev/null
real	0m1.813s

$ time hk version >/dev/null
real	0m0.016s

## list

$ time heroku apps >/dev/null
real	0m3.830s

$ time hk apps >/dev/null
real	0m0.785s
```

#### Focus on API

We believe this is evidence that a first-class Heroku API trumps any particular client.

#### Iterative Development

A release mechanism was created for hk in the beginning: the binary updates
itself. This gives us confidence in iterative development, which we value
highly, because it gives us the ability to release very often knowing users will
see the updates in a timely manner.

#### The power of Go

hk demonstrates the power of the Go language, runtime, systems access, and
distribution story (a single, statically linked binary with no dependencies) which are all very appealing to Heroku.

#### Unix

Heroku loves Unix. This client should reflect that. Commands are single,
unhyphenated words that map to their unix ancestors’ names and flags where
applicable.

### Installation

Currently, you need to have a [Go development environment](http://golang.org/doc/install) to install hk.
We intend to provide compiled binary forms of hk for easy installation,
but this isn't ready yet.

	$ go get github.com/kr/hk

### netrc

You'll need a line like this in $HOME/.netrc

	machine api.heroku.com login <email> password <apitoken>

### Usage

```
$ hk help

Usage: hk [-a app] [command] [options] [arguments]


Commands:

    create    create an app
    apps      list apps
    dynos     list dynos
    releases  list releases
    addons    list addons
    scale     change dyno counts
    restart   restart dynos
    set       set config var
    unset     unset config var
    env       list config vars
    run       run a process in a dyno
    log       stream app log lines
    info      show app info
    open      open app in a web browser
    rename    rename an app
    destroy   destroy an app
    sshauth   authorize ssh public keys
    version   show hk version

Run 'hk help [command]' for details.


Additional help topics:

    environ   environment variables used by hk
    plugins   interface to plugin commands
    more      additional commands, less frequently used
    about     information about hk (e.g. copyright, license, etc.)
```

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
	$ godep go get
	$ mate main.go
	$ godep go build
	$ ./hk list

### Release

	$ cd hk
	$ vim main.go # edit Version
	$ godep go build
