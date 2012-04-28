A sampling of `hk` terminal interactions, to give a sense of the tool's style and to help us achieve consistency across commands.

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
  stop            Stop a process
  token           Show auth token
  unset           Unset config vars
  version         Display version

See 'hk help <command>' for more information on a specific command.
```

```bash
$ hk help ps
Usage: hk ps -a <app>

List app processes.
```

```bash
$ hk version
0.0.1
```

```bash
$ hk list
morning-snow-4357
alpha-wolf-mark
sharp-mountain-8093
empty-autumn-6683
furious-autumn-6163
...
```

```bash
$ hk info -a morning-snow-4357
Name:     morning-snow-4357
Owner:    mark@heroku.com
Stack:    bamboo-mri-1.9.2
Git URL:  git@heroku.com:morning-snow-4357.git
Web URL:  http://morning-snow-4357.heroku.com/
```

```bash
$ hk env -a morning-snow-4357
DATABASE_URL=postgres://xmpvscckdw:xxx@ec2-107-20-254-132.compute-1.amazonaws.com/xmpvscckdw
URL=morning-snow-4357.heroku.com
RACK_ENV=production
STACK=bamboo-mri-1.9.2
...
```

```bash
$ hk get -a morning-snow-4357 RACK_ENV
production
```

```bash
$ hk ps -a morning-snow-4357
Process           State       Command
----------------  ----------  ------------------------
clock.1           up          bundle exec clockwork clock.rb
eventregistry.1   up          bundle exec rake resque:work QUEUE=event_registry VERBOSE=1 JOBS_PER_FORK=5
instworker.1      up          bundle exec rake resque:work QUEUE=instance VERBOSE=1 JOBS_PER_FORK=5
scheduler.1       up          bundle exec rake resque:scheduler
web.1             up          bundle exec thin start -p $PORT -e $RAILS_ENV
web.2             up          bundle exec thin start -p $PORT -e $RAILS_ENV
worker.1          up          bundle exec rake resque:work QUEUE=default,instance,reserved_instances VERBOSE=1 JOBS_PER_FORK=5
```

```bash
$ hk wat
Error: 'wat' is not an hk command. See 'hk help'.
```

```bash
$ hk list extra
Error: Unrecognized argument 'extra'. See 'hk help list'.
```

```bash
$ hk env -b morning-snow-4357
Error: Invalid usage. See 'hk help env'.
```

```bash
$ export HEROKU_API_KEY=x
$ hk list
Error: Unauthorized.
```
