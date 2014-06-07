# Production configuration

To release a production version of hk, export the follow variables to the local environment in which you'll be running the `hkdist` command.

```
export BUILDBRANCH=release
export BUILDNAME=hk
export DISTURL=https://hk.heroku.com/
export S3PATCHURL=https://hkpatch.s3.amazonaws.com/
export HKGENAPPNAME=hkgen
export S3DISTURL=xxxxxx
export S3_ACCESS_KEY=xxxxxx
export S3_SECRET_KEY=xxxxxx
```

The `S3DISTURL`, `S3_ACCESS_KEY` and `S3_SECRET_KEY` values can be found on the relevant instance of `hkdist` on Heroku.