# hk Design Research: Command Structure, Namespacing, and Naming

## Prior Art

### aws-cli

The new [aws cli](https://github.com/aws/aws-cli) has a straightforward command
structure. All commands are nested under the service they act on. A space
separates the service and the action.

#### Usage:

```
aws [options] <service> <action> [parameters]
```

#### Examples:

```bash
$ aws iam list-users
$ aws ec2 describe-instances
```

#### Command Listing

There are 31 top-level namespaces, one for each service. The most populous of
these is `ec2`, which has 149 actions, 27 of which are actions on instances:

```
bundle-instance                        cancel-reserved-instances-listing
cancel-spot-instance-requests          confirm-product-instance
create-instance-export-task            create-reserved-instances-listing
describe-instance-attribute            describe-instance-status
describe-instances                     describe-reserved-instances
describe-reserved-instances-listings   describe-reserved-instances-modifications
describe-reserved-instances-offerings  describe-spot-instance-requests
modify-instance-attribute              modify-reserved-instances
monitor-instances                      purchase-reserved-instances-offering
reboot-instances                       report-instance-status
request-spot-instances                 reset-instance-attribute
run-instances                          start-instances
stop-instances                         terminate-instances
unmonitor-instances
```

Heroku doesn't have multiple services to talk to, just a single primary API, so
this style of namespacing is not directly applicable.

### Force.com CLI

@ddollar's [Force.com CLI](https://github.com/heroku/force) began as a fork of
hk, but due to the number of commands, has expanded upon hk's command structure
by adding a level of nesting.

#### Usage:

```
force <noun> <action> <object-type> <object-id>
force <noun> <action> <object-type> [<fields>]
```

#### Examples:

```bash
$ force record get User 00Ei0000000000
$ force record create User Name:"David Dollar" Phone:0000000000
$ force record update User 00Ei0000000000 State:GA
$ force record delete User 00Ei0000000000
```

### Git

Git doesn't really have a command hierarchy. Subcommands (and plugins) each
represent a single action, often with a variety of flags and options. This model
doesn't necessarily directly map to an API client like hk, though.

#### Usage:

```
git [options] <command> [<args>]
```

#### Examples:

```bash
$ git push
$ git push -f origin master
```

#### Command Listing

While git has many options, its help only lists the most commonly-used commands:

```
add     bisect  branch  checkout  clone
commit  diff    fetch   grep      init
log     merge   mv      pull      push
rebase  reset   rm      show      status
tag
```

A full command listing reveals a plethora of hidden commands:

```
add                       diff-index         merge-one-file   request-pull
add--interactive          diff-tree          merge-ours       rerere
am                        difftool           merge-recursive  reset
annotate                  difftool--helper   merge-resolve    rev-list
apply                     fast-export        merge-subtree    rev-parse
archimport                fast-import        merge-tree       revert
archive                   fetch              mergetool        rm
bisect                    fetch-pack         mktag            send-email
bisect--helper            filter-branch      mktree           send-pack
blame                     fmt-merge-msg      mv               sh-i18n--envsubst
branch                    for-each-ref       name-rev         shell
bundle                    format-patch       notes            shortlog
cat-file                  fsck               p4               show
check-attr                fsck-objects       pack-objects     show-branch
check-ignore              gc                 pack-redundant   show-index
check-ref-format          get-tar-commit-id  pack-refs        show-ref
checkout                  grep               patch-id         stage
checkout-index            gui--askpass       peek-remote      stash
cherry                    hash-object        prune            status
cherry-pick               help               prune-packed     stripspace
citool                    http-backend       pull             submodule
clean                     http-fetch         push             subtree
clone                     http-push          quiltimport      svn
column                    imap-send          read-tree        symbolic-ref
commit                    index-pack         rebase           tag
commit-tree               init               receive-pack     tar-tree
config                    init-db            reflog           unpack-file
count-objects             instaweb           relink           unpack-objects
credential                log                remote           update-index
credential-cache          lost-found         remote-ext       update-ref
credential-cache--daemon  ls-files           remote-fd        update-server-info
credential-osxkeychain    ls-remote          remote-ftp       upload-archive
credential-store          ls-tree            remote-ftps      upload-pack
cvsexportcommit           mailinfo           remote-http      var
cvsimport                 mailsplit          remote-https     verify-pack
cvsserver                 merge              remote-testpy    verify-tag
daemon                    merge-base         remote-testsvn   web--browse
describe                  merge-file         repack           whatchanged
diff                      merge-index        replace          write-tree
diff-files                merge-octopus      repo-config
```

Certainly, many of these commands are infrequently used. But their
discoverability is also very poor.

### NPM

The Node.js npm CLI tool is similar in command structure to Git, except that it
adds an optional level of namespacing (similar to the Force.com CLI).

#### Usage

```
npm <command> [<subcommand>] [<options>]
```

#### Examples

```bash
$ npm install express
$ npm cache clean express
```

#### Command Listing

```
add-user    adduser      apihelp    author      bin
bugs        c            cache      completion  config
ddp         dedupe       deprecate  docs        edit
explore     faq          find       find-dupes  get
help        help-search  home       i           info
init        install      isntall    issues      la
link        list         ll         ln          login
ls          outdated     owner      pack        prefix
prune       publish      r          rb          rebuild
remove      repo         restart    rm          root
run-script  s            se         search      set
show        shrinkwrap   star       stars       start
stop        submodule    tag        test        tst
un          uninstall    unlink     unpublish   unstar
up          update       v          version     view
whoami
```

Individual commands, like [npm-cache](https://npmjs.org/doc/cli/npm-cache.html)
can have subcommands:

```
npm cache add <tarball file>
npm cache add <folder>
npm cache add <tarball url>
npm cache add <git url>
npm cache add <name>@<version>
npm cache ls [<path>]
npm cache clean [<pkg>[@<version>]]
```

The implementation of command is contained in a module specific to that command,
along with any subcommands it may contain.

### Heroku Gem

Uses colon-separated nesting, borrowed from the style of Rake (which borrowed
from Maven). However, the gem also makes heavy use of aliasing for common
actions. App actions, in particular, usually appear as a verb without any
namespace. This has worked ok as there has only ever been one top-level entity
(an app), but that will likely change in the near future.

#### Usage:

```
heroku <verb> [<options>]
heroku <noun>:<verb> [<options>]
```

#### Examples:

```bash
$ heroku create
$ heroku create myapp --region eu
$ heroku create --stack cedar myapp
$ heroku apps:create
$ heroku config:set -a myapp NAME=value
$ heroku config:unset NAME -a myapp
```
Notably, the heroku gem has very loose requirements for ordering of
optional arguments.

#### Command Listing

The `heroku` command contains many namespaces, most with multiple subcommands.
Namespaces are divided into the "primary help topics" and "additional topics":

```
addons  apps  auth      config  domains
logs    ps    releases  run     sharing

Additional topics:

account    certs    drains   fork         git
help       keys     labs     maintenance  pg
pgbackups  plugins  regions  stack        status
update     version
```

On average, each of these namespaces contains 3 or more subcommands. For
example, here is the `apps` namespace:

```bash
heroku apps:create [name]
heroku apps:destroy
heroku apps:info
heroku apps:open
heroku apps:rename <newname>
heroku apps:suspend
heroku apps:unsuspend
```

Also not shown are aliases, which further pollute the top-level namespace.

### hk (today)

hk currently has a flat command structure with no nesting. This has been
clean and simple thus far as there are few commands.

#### Usage:

```
hk [-a app] <command> [<options>] [arguments]
```

#### Examples:

```bash
$ hk create
$ hk create -r eu myapp
$ hk set NAME=value
$ hk -a myapp log -n 10
```

Note that the `-a <appname>` flag is currently a global option.

While the lack of namespacing has felt clean and simple thus far, we anticipate
that it will not be practical once the full set of functionality has been
implemented. The sheer number of options will be overwhelming if they're all
presented at once.

#### Command Listing

```
create apps dynos releases addons scale restart set unset env run log info open
rename destroy sshauth version

Additional commands:

api app get creds url
```

## Issues to Consider

### Namespacing

hk currently has a flat command structure with no namespacing. The heroku gem
uses a rake-style colon-separated namespacing. Some recent examples use spaces
to separate the namespace and the subcommand.

#### Plugins

While the current plugin design is far from final, we should consider it when
thinking about namespacing. Today, hk plugins today can only add new commands to
the toplevel namespace. If we add namespacing and nesting, plugins may want to
add new commands underneath a top-level namespace. Here are some pg extras
commands, for example:

```bash
$ hk pg cache_hit
$ hk pg outliers
$ hk pg mandelbrot
```

The other option is to allow each plugin to be a single command and/or
namespace. That would mean that you couldn't add stuff to the `app` or `pg`
namespaces, but you could do the following:

```bash
$ hk pg psql          # built-in
$ hk pgx cache_hit    # plugin pgx
$ hk pgx mandelbrot   # plugin pgx
```

### Noun Plurality

Plural nouns, as in `heroku apps`, only make sense when you're trying to list
all apps. Otherwise, the singular form, i.e. `heroku app`, is more appropriate
(for creating, updating, deleting). Maybe we should stick with singular nouns
everywhere? The Force.com CLI has taken this approach. For the app namespace,
this might look like:

```bash
$ hk app list
$ hk app info
$ hk app create
$ hk app destroy
$ hk app rename
$ hk app fork
```

And for dynos:

```bash
$ hk dyno run bash
$ hk dyno restart
$ hk dyno restart web
$ hk dyno scale web=20
$ hk dyno resize web=2x
```

I'll admit that run and restart both feel a bit strange to me being nested under
"dyno".

### Placement of App Flag

hk recently moved the `-a <appname>` flag to be a global option, appearing
immediately after the `hk`:

```bash
$ hk -a myapp set KEY=val
$ hk -a myapp log -n 100
```

This was done because we wanted to parse the app argument out of any command
rather than leaving it up to subcommands or plugins to parse that option.

However, even after several months of using hk regularly, that change still
feels unnatural to me. Others have reported the same. This is partially because
many users have a habit from the old CLI of appending `-a <appname>` to the end
of any command, as in `heroku config:set KEY=val -a myapp`. I think another part
of it is that the actual command gets buried, especially with long app names:

```bash
$ hk -a my-really-long-app-name log -n 20
```

Perhaps we can special-case this flag so it can still be placed at the end of
the command, yet be stripped from the commands that are passed to plugins (and
provided as `HKAPPNAME=foo` or whatever).

```bash
$ hk restart -a myapp
$ hk set KEY=val -a myapp
$ hk log -n 100 -a myapp
```
