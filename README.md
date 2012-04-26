## hk

Fast Heroku client.


### Installation

```bash
$ curl -s https://github.com/downloads/mmcgrana/hk/hk-release-darwin > /usr/local/bin/hk
$ chmod +x /usr/local/bin/hk
```


### Usage

```bash
$ hk help
Usage: hk <command> [-a <app>] [command-specific-options]

Supported hk commands are:
  addons          List add-ons
  addons-add      Add an add-on
  addons-open     Open an add-on page
  addons-remove   Remove an add-on 
  create          Create an app
  destroy         Destroy an app
  env             List config vars
  get             Get config var
  help            Show this help
  info            Show app info
  list            List apps
  login           Log in
  logout          Log out
  logs            Show logs
  open            Open app
  pg              List databases
  pg-info         Show database info
  pg-promote      Promote a database
  ps-psql         Open a psql database shell
  pg-wait         Await a database
  ps              List processes
  release         Show release info
  releases        List releases
  rename          Rename an app
  restart         Restart processes
  rollback        Rollback to a previous release
  run             Run a process
  set             Set config var
  scale           Scale processes
  run             Run a process
  stop            Stop a process
  token           Show auth token
  unset           Unset config vars
  version         Display version

See 'hk help <command>' for more information on a specific command.
```


### Development

```bash
$ cd hk
$ mate hk.go
$ go run hk.go list
$ go build
$ ./hk list
```
