A sampling of `hk` terminal interactions, to give a sense of the tool's style and to help us achieve consistency across commands.

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
