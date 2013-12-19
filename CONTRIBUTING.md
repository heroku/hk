# Contribution Guidelines

hk's design is very much unfinished, and we're not yet ready to accept
contributions or requests related to design. At this time, please only submit
bug reports or pull requests that fix bugs.

Once we've finished some full iterations on the design, we'll be more open to
feature requests and discussions around design.

It will also take some time to come up with a written description of the design
philosophy and code quality standards for things like idiomatic Go. In the mean
time, we'll review and discuss on a case-by-case basis.

## Philosophy

Check out the [readme](https://github.com/heroku/hk/blob/master/README.md#motivation).

## Code Standards

We strive to write idiomatic Go code, and welcome any suggestions on how to
better leverage the expressive nature of Go code. Here are a few guidelines:

* Please `go fmt` everything before submitting.
* Package imports from 3rd parties should be listed separately, in the
  [goimports][goimports] style:

  ```go
  import (
    "fmt"
    "io"
    "strings"

    "github.com/bgentry/heroku-go"
    "github.com/heroku/hk/term"
  )
  ```

* Use `fmt` to print to stdout, and `log` to print to stderr. Stdout should be
  reserved for output that is intended to be piped into other processes, while
  stderr should be used for informational messaging or errors:

  ```go
  fmt.Println("dodging-samurai-42")
  fmt.Println("web.1  up  3h  bin/web")
  
  log.Println("Created dodging-samurai-42.")
  log.Println("This is an error :(")
  ```

## Style Guide

The style guide is available at [hk.heroku.com/styleguide.html][styleguide]. The
style guide can be automatically generated using hk's help output:

```bash
$ ./hk help styleguide > hkdist/public/styleguide.html
```

[goimports]: https://github.com/bradfitz/goimports
[styleguide]: https://hk.heroku.com/styleguide.html
