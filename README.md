## hk

Fast Heroku client.


### Overview

hk is a command line client to the Heroku runtime platform, designed to be as fast as possible.


### Motivation

```bash
$ time heroku version
2.27.3
real	0m1.813s

$ time hk version
0.4
real	0m0.016s

$ time heroku list > /dev/null
real	0m8.826s

$ time hk list > /dev/null
real  0m3.658s
```


### Installation

```bash
$ wget -qO- https://hk.heroku.com/hk.gz | zcat >/usr/local/bin/hk
$ chmod +x /usr/local/bin/hk
```


### Usage

```bash
$ hk help
Usage: hk <command> [options] [arguments]

Supported commands are:

  create     create an app
  destroy    destroy an app
  creds      show auth creds
  env        list config vars
  get        get config var
  set        set config var
  info       show app info
  list       list apps
  open       open app
  ps         list processes
  scale      change dyno counts
  tail       tail log files
  run        run a process
  version    show hk version
  help       show help

See 'hk help [command]' for more information about a command.
```


### Development

```bash
$ cd hk
$ go get
$ mate main.go
$ go build
$ ./hk list
```

### Release

```bash
$ cd hk
$ vim main.go # edit Version
$ go build
$ ./mkpatch oldver1 oldver2...
# put OS-ARCH-hk.gz and OS-ARCH-VER-next.hkdiff online
```

### Design

We maintain a collections of `hk` terminal interactions in [DESIGN.md](hk/blob/master/DESIGN.md), from which you can get a sense of the tool's style.
