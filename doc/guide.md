# Getting started with hk

[hk][hk] is a new command line client to the Heroku platform, designed to be
fast and lightweight.

While hk is intended to replace the [Heroku Ruby CLI][ruby-cli], it's designed
as a completely new tool rather than a drop-in replacement. hk borrows from a
number of well-established Unix conventions, with commands mapping to their Unix
ancestors' names where applicable.

hk is distributed as an auto-updating executable. Once you've installed it, hk
will periodically perform a version check in the background. If a newer version
is found, hk will update itself automatically.

To install hk, please use the instructions on [hk.heroku.com][hk-install]. hk is
currently only being distributed for Mac OS X and Linux users. More
user-friendly installers, including one for Windows, are planned.

## Important differences from the Heroku Ruby CLI

The Heroku Ruby CLI organized its commands under nested namespaces, separated
with colons (i.e. `domains:add`). Most, but not all, of these namespaces were
pluralizezd.

hk, however, uses a simple, flat command space (i.e. `domains` and
`domain-add`). Commands use pluralized nouns where it's logical to do so, such
as lists of items (`apps`, `dynos`, `addons`, `releases`). The rest of the
commands are named with singular nouns because they deal with a single resource
(`addon-add`, `domain-remove`, `release-info`).

The other major difference is that hk uses strict POSIX ordering on command
options. That means that any option flags, such as `-a <app>` must come before
any non-flag arguments. Here are some examples:

```bash
$ hk restart -a myapp web      # valid
$ hk restart web -a myapp      # not valid

$ hk get -a myapp CONFIG_NAME  # valid
$ hk get CONFIG_NAME -a myapp  # not valid

$ hk run -a myapp ls           # valid, will run `ls` on myapp
$ hk run ls -a myapp           # will run the command `ls -a myapp` on whichever
                               # app is configured in git remotes, which may or
                               # may not be myapp
```

## Guide to hk commands for Heroku users

Many commands are similar in both the Heroku Ruby CLI and hk. However, some
commands have different names and take different arguments. This is a list of
frequently used commands, showing how to accomplish the same thing with either
CLI.

### Apps

#### Create an app

```bash
➜ heroku create myapp
```

```bash
➜ hk create myapp
```

With a region specified:

```bash
➜ heroku create --region eu myapp
```

```bash
➜ hk create -r eu myapp
```

#### List apps

```bash
➜ heroku list
```

```bash
➜ hk apps
```

#### Show an app's info

```bash
➜ heroku info
```

```bash
➜ hk info
```

#### Destroy an app

```bash
➜ heroku destroy -a myapp
```

```bash
➜ hk destroy myapp
```

The `heroku destroy` command can infer your app name from the current
directory's git remotes. For safety, however, `hk destroy` always requires you
to specify the name of the app you want to destroy.

This command can permanently destroy data, so it prompts for confirmation.

#### Rename an app

```bash
➜ heroku rename -a oldappname newappname
```

```bash
➜ hk rename oldappname newappname
```

#### View your application log

```bash
➜ heroku logs --tail
```

```bash
➜ hk log
```

The `hk log` command follows your application log stream by default (which
required a `--tail` flag in the Toolbelt).

### Dynos

#### Change dyno scale

```bash
➜ heroku ps:scale web=2 worker=4:PX
```

```bash
➜ hk scale web=2 worker=4:PX
```

This command is mostly identical in hk, except that it doesn't support scaling
by relative increments (i.e. `web+2`).

#### List dynos

```bash
➜ heroku ps
```

```bash
➜ hk dynos
```

### App configuration (environment)

hk aligns with Unix conventions for setting environment (config) variables. As
such, the hk commands to modify your app's "config vars" are the same as those
to modify the environment on a Unix machine.

#### Set an environment variable on an app

```bash
➜ heroku config:set KEY=value
```

```bash
➜ hk set KEY=value
```

#### List app's environment settings

```bash
➜ heroku config
```

```bash
➜ hk env
```

#### Show a single environment variable

```bash
➜ heroku config:get KEY
```

```bash
➜ hk get KEY
```

#### Unset an environment variable on an app

```bash
➜ heroku config:unset KEY1 KEY2
```

```bash
➜ hk unset KEY1 KEY2
```

### Domain Names

#### List domain names

```bash
➜ heroku domains
```

```bash
➜ hk domains
```

#### Add a domain name

```bash
➜ heroku domains:add www.test.com
```

```bash
➜ hk domain-add www.test.com
```

#### Remove a domain name

```bash
➜ heroku domains:remove www.test.com
```

```bash
➜ hk domain-remove www.test.com
```

### Add-ons

#### List add-ons on an app

```bash
➜ heroku addons
```

```bash
➜ hk addons
```

#### Add an add-on

```bash
➜ heroku addons:add heroku-postgresql
```

```bash
➜ hk addon-add heroku-postgresql
```

With additional provisioning options:

```bash
➜ heroku addons:add heroku-postgresql --fork RED
```

```bash
➜ hk addon-add heroku-postgresql fork=red
```

Additional add-on config is provided as `key=value` pairs rather than
`--key value` flags.

#### Destroy an add-on

```bash
➜ heroku addons:remove redistogo
```

```bash
➜ hk addon-remove redistogo
```

Next, a Heroku Postgres addon:

```bash
➜ heroku addons:remove heroku-postgresql:dev
```

```bash
➜ hk addon-remove heroku-postgresql-blue
```

Add-ons in hk are referenced by their `name`. Usually this is just the addon
provider's name, but for Heroku Postgres, it's of the form:
`heroku-postgresql-color`. In either case, the name matches what's displayed in
`hk addons`.

This command can permanently destroy data, so it prompts for confirmation.

### Access Control (sharing with collaborators)

#### Add access for a user

```bash
➜ heroku sharing:add user@test.com
```

```bash
➜ hk access-add user@test.com
```

#### Remove access for a user

```bash
➜ heroku sharing:remove user@test.com
```

```bash
➜ hk access-remove user@test.com
```

### Releases

#### View an app's releases

```bash
➜ heroku releases
```

```bash
➜ hk releases
```

#### View release info

```bash
➜ heroku releases:info v123
```

```bash
➜ hk release-info v123
```

#### Rollback to a previous release

```bash
➜ heroku rollback
```

```bash
➜ hk rollback v122
```

The Toolbelt attempts to rollback by one version, but hk requires you to specify
the version you want to rollback to.

## Getting help with hk

hk has a simple help system. The most common commands are listed in the basic
help output, which is also available via `hk help`:

```
$ hk help
Usage: hk <command> [-a app] [options] [arguments]


Commands:

    create          create an app
    apps            list apps
    dynos           list dynos
    releases        list releases
    release-info    show release info
    rollback        roll back to a previous release
    addons          list addons
    addon-add       add an addon
    addon-destroy   destroy an addon
    scale           change dyno quantities and sizes
    restart         restart dynos
    set             set env var
    unset           unset env var
    env             list env vars
    run             run a process in a dyno
    log             stream app log lines
    info            show app info
    rename          rename an app
    destroy         destroy an app
    domains         list domains
    domain-add      add a domain
    domain-remove   remove a domain
    version         show hk version

Run 'hk help [command]' for details.


Additional help topics:

    commands  list all commands with usage
    environ   environment variables used by hk
    plugins   interface to plugin commands
    more      additional commands, less frequently used
    about     information about hk (e.g. copyright, license, etc.)
```

Commands that are used less frequently are listed under `hk help more`. For any
specific command, you can run `hk help <command>` to get the detailed help and
usage info for that command.

## Full Command List

```bash
$ hk help commands
```

```
hk access [-a <app>]                                              # list access permissions
hk access-add [-a <app>] [-s] <email>                             # give a user access to an app
hk access-remove [-a <app>] <email>                               # remove a user's access to an app
hk account-feature-disable <feature>                              # disable an account feature
hk account-feature-enable <feature>                               # enable an account feature
hk account-feature-info <feature>                                 # show info for an account feature
hk account-features                                               # list account features
hk addon-add [-a <app>] <service>[:<plan>] [<config>=<value>...]  # add an addon
hk addon-destroy [-a <app>] <name>                                # destroy an addon
hk addon-open [-a <app>] <name>                                   # open an addon
hk addons [-a <app>] [<service>:<plan>...]                        # list addons
hk api <method> <path>                                            # make a single API request
hk apps [<name>...]                                               # list apps
hk create [-r <region>] [<name>]                                  # create an app
hk creds                                                          # show credentials
hk destroy <name>                                                 # destroy an app
hk domain-add [-a <app>] <domain>                                 # add a domain
hk domain-remove [-a <app>] <domain>                              # remove a domain
hk domains [-a <app>]                                             # list domains
hk drain-add [-a <app>] <url>                                     # add a log drain
hk drain-info [-a <app>] <id or url>                              # show info for a log drain
hk drain-remove [-a <app>] <id or url>                            # remove a log drain
hk drains [-a <app>]                                              # list log drains
hk dynos [-a <app>] [<name>...]                                   # list dynos
hk env [-a <app>]                                                 # list env vars
hk feature-disable [-a <app>] <feature>                           # disable an app feature
hk feature-enable [-a <app>] <feature>                            # enable an app feature
hk feature-info [-a <app>] <feature>                              # show info for an app feature
hk features [-a <app>]                                            # list app features
hk get [-a <app>] <name>                                          # get env var
hk help [<topic>]                                                 # 
hk info [-a <app>]                                                # show app info
hk key-add [<public-key-file>]                                    # add ssh public key
hk key-remove <fingerprint>                                       # remove an ssh public key
hk keys                                                           # list ssh public keys
hk log [-a <app>] [-n <lines>] [-s <source>] [-d <dyno>]          # stream app log lines
hk login <email>                                                  # log in to your Heroku account
hk logout                                                         # log out of your Heroku account
hk maintenance [-a <app>]                                         # show app maintenance mode
hk maintenance-disable [-a <app>]                                 # disable maintenance mode
hk maintenance-enable [-a <app>]                                  # enable maintenance mode
hk open [-a <app>]                                                # open app in a web browser
hk pg-info [-a <app>] <dbname>                                    # show Heroku Postgres database info
hk pg-list [-a <app>]                                             # list Heroku Postgres databases
hk pg-unfollow [-a <app>] <dbname>                                # stop a replica postgres database from following
hk psql [-a <app>] [-c <command>] [<dbname>]                      # open a psql shell to a Heroku Postgres database
hk regions                                                        # list regions
hk release-info [-a <app>] <version>                              # show release info
hk releases [-a <app>] [<version>...]                             # list releases
hk rename <oldname> <newname>                                     # rename an app
hk restart [-a <app>] [<type or name>]                            # restart dynos
hk rollback [-a <app>] <version>                                  # roll back to a previous release
hk run [-a <app>] [-s <size>] [-d] <command> [<argument>...]      # run a process in a dyno
hk scale [-a <app>] <type>=[<qty>]:[<size>]...                    # change dyno quantities and sizes
hk set [-a <app>] <name>=<value>...                               # set env var
hk status                                                         # display heroku platform status
hk transfer [-a <app>] <email>                                    # transfer app ownership to a collaborator
hk transfer-accept [-a <app>]                                     # accept an inbound app transfer
hk transfer-cancel [-a <app>]                                     # cancel an outbound app transfer
hk transfer-decline [-a <app>]                                    # decline an inbound app transfer
hk transfers [-a <app>]                                           # list existing app transfers
hk unset [-a <app>] <name>...                                     # unset env var
hk update                                                         # 
hk url [-a <app>]                                                 # show app url
hk version                                                        # show hk version
hk which-app [-a <app>]                                           # show which app is selected, if any
```

[hk]: https://github.com/heroku/hk "hk on Github"
[hk-install]: https://hk.heroku.com/ "hk: a fast Heroku CLI"
[ruby-cli]: https://github.com/heroku/heroku "Heroku Ruby CLI"
