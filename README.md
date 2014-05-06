# hk

A fast Heroku CLI client.

## Overview

hk is a command line client to the Heroku runtime platform, designed to be as fast as possible.

### Disclaimer

This is beta software. It may still undergo substantial changes during the
coming months. You should expect that some functionality may break or be altered without warning.

## Motivation

### Fast as a feature

```bash
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

### Iterative Development

A release mechanism was created for hk in the beginning: the binary updates
itself. This gives us confidence in iterative development, which we value
highly, because it gives us the ability to release very often knowing users will
see the updates in a timely manner.

### The power of Go

hk demonstrates the power of the Go language, runtime, systems access, and
distribution story (a single, statically linked binary with no dependencies)
which are all very appealing to Heroku.

### Unix

Heroku loves Unix. This client should reflect that. Commands map to their unix
ancestors’ names and flags where applicable.

## Installation

> Please note that versions of hk installed from source are unsupported and
> should only be installed for development purposes.

### Mac OS X, Linux, BSD

Pre-built binaries are available for Mac OS X, Linux, and BSD. Once installed,
these binaries will automatically update themselves when new releases are
available.

To install a pre-built binary release, run the following one-liner:

```bash
$ L=/usr/local/bin/hk && curl -sL -A "`uname -sp`" https://hk.heroku.com/hk.gz | zcat >$L && chmod +x $L
```

The URL [https://hk.heroku.com/hk.gz](https://hk.heroku.com/hk.gz) will attempt
to detect your OS and CPU architecture based on the User-Agent, then redirect
you to the latest release for your platform.

### Windows

Currently, you need to have a [Go development environment][go-install] to
install hk on Windows. Compiled binaries with automatic updating are available
for Windows, but the installer is not ready yet.

	$ go get github.com/heroku/hk

Please note that this installation method is unsupported.

## Usage

The basic usage of hk is:

```
Usage: hk <command> [-a app] [options] [arguments]
```

For more details, and to learn about differences between hk and the Heroku Ruby
CLI, please see the [getting started guide](./doc/guide.md).

## Shell Completion

Shell completion scripts for hk have been written for zsh and bash. Both files
are located in [./contrib](./contrib/).

The zsh completion script completes all command names and help topics. It also
completes flags and other arguments for many commands:

![](http://cl.ly/image/3n1X0q3y2E17/Screen%20Shot%202014-03-07%20at%201.52.26%20PM.png)

![](http://cl.ly/image/0u3v0T2m352h/Image%202014-03-09%20at%2011.34.17%20AM.png)

![](http://f.cl.ly/items/2X200V0h2M0L1Q1w0x38/Image%202014-03-11%20at%203.12.23%20PM.png)

The bash completion script completes only command names at this time.

## Config

hk allows some configuration using git config.

### Use strict flag ordering / disable interspersed flags and non-flag arguments

Enable:

```
$ git config --global --bool hk.strict-flag-ordering true
```

Disable:

```
$ git config --global --unset hk.strict-flag-ordering
```

## Plugins

hk currently has a minimal plugin system. It may see substantial changes in the future, and those changes may break existing plugins or change the architecture at any time. Use this functionality at your own risk.

Plugins are executables located in HKPATH or, if HKPATH does not exist, in /usr/local/lib/hk/plugin. They are executed when hk does not know command X and an installed plugin X exists. The special case default plugin will be executed if hk has no command or installed plugin named X.

hk will set these environment variables for a plugin:

* HEROKU_API_URL - The url containing the username, password, and host to the api endpoint.
* HKAPP - The app as determined by the git heroku remote, if available.
* HKUSER - The username from either HEROKU_API_URL or .netrc
* HKPASS - The password from either HEROKU_API_URL or .netrc
* HKHOST - The hostname for the API endpoint

## Development

hk requires Go 1.2 or later and uses [Godep](https://github.com/kr/godep) to manage dependencies.

	$ cd hk
	$ vim main.go
	$ godep go build
	$ ./hk apps

Please follow the [contribution guidelines](./CONTRIBUTING.md) before submitting
a pull request.

[go-install]: http://golang.org/doc/install "Golang installation"
