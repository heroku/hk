# Building and Releasing hk

## Setup Go

Make sure your `$GOPATH` is setup correctly.

## Install an up-to-date Godep

```bash
$ go get -u github.com/tools/godep
```

## Clone hk

Check out hk to `$GOPATH/src/github.com/heroku/hk`.

## Godep restore

Change into your hk directory and run `godep restore`. This will restore the
current versions of hk's dependencies into your `$GOPATH`, which will make
`godep go build` the same as `go build`.

## Install hkdist

Change into `hk/hkdist` and run `godep go install`. This will build and install
the hkdist executable, hk's build tool and update distribution server.

## hkdist Overview

hkdist is hk's build tool and distribution server. It's a subpackage within the
hk repo. You should read about it in the [hkdist readme](./../hkdist/Readme.md).

## Install a cross-compilation Go environment

Since we're building hk for multiple environments from a single machine, we
need an environemnt for cross-compiling. The following instructions should do
the trick:

http://dave.cheney.net/2012/09/08/an-introduction-to-cross-compilation-with-go

## Branch workflow

hk has 3 main branches:

1. `master`, which is supposed to be ready-to-ship at any time
2. `staging`, the branch corresponding to hkstaging builds
3. `release`, the main release branch

The workflow for releasing is therefore `master -> staging -> release`.

## Release tags

Release versions are identified with git tags using a date convention:

```
➜  git tag | tail -n 8
v20140429
v20140429.1
v20140502
v20140503
v20140506
v20140507
v20140509
v20140509.1
```

If there's more than one release on a day, just increment with a `.1`, `.2`,
etc.

## Creating a release

### Preparing a staging build

#### Update commit for staging branch

```
git co staging
git rebase master
```

#### Add a git tag

GPG signing is not strictly required yet, because there is no verification of
these signatures, but it's still best practice.

```
git tag -sam "fix pg-info display of name when defaulting" v20140509.1
git push && git push —tags
```

#### Make staging build

You'll need to export the [required environment variables](./staging.md) for hkdist to run
locally. Once you've got those, just `cd /tmp` and run:

```bash
$ hkdist build
```

If you see any error messages during this time, something has probably gone
wrong. The build process will continue in that case, but it won't actually set
the new release to be the current release if there were errors. The exception is
errors during diff generation, because users can still update without those.

This function is also idempotent, so you can just re-run it if anything goes
wrong, and it will figure out which steps haven't been completed.

#### Optional: verify staging build

If you've made changes that have any chance of affecting the auto-updating
process, you should definitely verify that you can update and run the staging
build of hk.

Regardless, it's a good idea to have a copy of the staging build installed so
you can test with it:

```bash
L=/usr/local/bin/hks && curl -sL -A "`uname -sp`" https://hkdist-staging.herokuapp.com/hkstaging.gz | zcat >$L && chmod +x $L
```

### Preparing a staging build

#### Update commit for release branch

```
git co release
git rebase staging
git push
```

#### Make production build

Just as before, you'll need to export the [required environment variables](./production.md) (this
time for production) and run the following:

```bash
$ hkdist build
```

#### Verify production release

If you run `hk update`, you should now see it update to the latest hk version
that you just released.
