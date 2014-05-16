# hkdist

A suite of tools for building hk, generating binary diffs, and distributing
them for updates.

## Prerequisites

install [godep](https://github.com/kr/godep) to manage dependencies:

```bash
$ go get github.com/kr/godep
```

## Functions of hkdist

### Web server

`hkdist web` is the web server for [hk.heroku.com](https://hk.heroku.com). It
serves intial downloads of hk. It also has an API that tracks hk's versions for
each OS.

It's this API that hk talks to when it needs to check for an update. If
there is a new version available, the server will return the new version number
and the SHA256 hash for the full binary on the given platform.

### Build tool

`hkdist build` runs the cross-compiled builds for hk. When you run it, it clones
a fresh copy of hk to the current directory (I usually run from `/tmp`). It then
finds the latest git tag on the repo to determine which version is the current.
It will only build from tags of the format `vYYYYMMDD` or `vYYYYMMDD.X`,
corresponding to the date of the tag and an optional point release for that day
(useful if you're releasing more than once in a day).

Once it knows which version to build, the tool will build for each platform,
upload the resulting binary, then generate binary diffs for the past 8 versions.
The diffs are actually generated in parallel using Heroku apps called `hkgen`
and `hkgen-staging`.

This command is idempotent, so you can just re-run it if anything goes wrong,
and it will figure out which steps haven't been completed.

### Diff generation

`hkdist gen` will generate and upload a binary diff (bsdiff) file for the given
command, on the specified platform, between the two versions specified. Example:

```bash
$ hkdist gen hkstaging darwin-amd64 v20140501 v20140529
```

## environment

### DISTURL (build)

url for the hk directory server (e.g. https://hk.heroku.com)

### S3DISTURL (build, gen, web)

url for the S3 bucket for distributing full hk binaries (e.g. https://hkdist.s3.amazonaws.com/)

### S3PATCHURL (gen)

url for the S3 bucket for distributing patches
(e.g. https://hkpatch.s3.amazonaws.com/)

### S3_ACCESS_KEY (build, gen), S3_SECRET_KEY (build, gen)

aws iam credentials

### BUILDNAME (build)

the name of the binary to build (e.g. "hk" or "hk-canary")

### BUILDBRANCH (build)

the name of the git branch to build from (e.g. "release" or "canary")

### DATABASE_URL (web)

postgres:// url

### PORT (web)

which tcp port to listen on

### HKGENAPPNAME (build)

the name of the heroku app to generate diffs with

## Creating a new release

To release a new version of hk:

1. make a gpg-signed git tag of the form vYYYYMMDD or vYYYYMMDD.N

    ```bash
    $ git tag -sam "added a new feature" vYYYYMMDD
    ```

2. push the release branch to that tag on github

    ```bash
    $ git co release
    $ git reset --hard vYYYYMMDD
    $ git push origin release
    ```

3. run hkdist build to build for all platforms and generate diffs

    ```bash
    ## Export environment variables for target build ##
    $ hkdist build
      ...
    ```
