[![PkgGoDev](https://img.shields.io/badge/go.dev-docs-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/github.com/crazy-max/gotestlist)
[![Test workflow](https://img.shields.io/github/workflow/status/crazy-max/gotestlist/test?label=test&logo=github&style=flat-square)](https://github.com/crazy-max/gotestlist/actions?workflow=test)
[![Go Report](https://goreportcard.com/badge/github.com/crazy-max/gotestlist?style=flat-square)](https://goreportcard.com/report/github.com/crazy-max/gotestlist)
[![Codecov](https://img.shields.io/codecov/c/github/crazy-max/gotestlist?logo=codecov&style=flat-square)](https://codecov.io/gh/crazy-max/gotestlist)

## About

List tests in the given Go packages.

## Installation

```console
$ go install github.com/crazy-max/gotestlist@latest
```

## Usage

```console
$ gotestlist .
gotestlist     TestTests     /home/crazy/src/github.com/crazy-max/gotestlist/gotestlist_test.go
```

```console
$ gotestlist ./...
gotestlist     TestTests     /home/crazy/src/github.com/crazy-max/gotestlist/gotestlist_test.go
main           TestDirs      /home/crazy/src/github.com/crazy-max/gotestlist/cmd/gotestlist/gotestlist_test.go
```

```console
$ gotestlist github.com/crazy-max/gotestlist
gotestlist     TestTests     /home/crazy/src/github.com/crazy-max/gotestlist/gotestlist_test.go
```

```console
$ gotestlist github.com/crazy-max/gotestlist/...
gotestlist     TestTests     /home/crazy/src/github.com/crazy-max/gotestlist/gotestlist_test.go
main           TestDirs      /home/crazy/src/github.com/crazy-max/gotestlist/cmd/gotestlist/gotestlist_test.go
```

```console
$ gotestlist github.com/crazy-max/gotestlist github.com/crazy-max/gotestlist/cmd/gotestlist
gotestlist     TestTests     /home/crazy/src/github.com/crazy-max/gotestlist/gotestlist_test.go
main           TestDirs      /home/crazy/src/github.com/crazy-max/gotestlist/cmd/gotestlist/gotestlist_test.go
```

```console
$ gotestlist -f json ./... | jq
[
  {
    "name": "TestTests",
    "benchmark": false,
    "fuzz": false,
    "suite": "",
    "file": "/home/crazy/src/github.com/crazy-max/gotestlist/gotestlist_test.go",
    "pkg": "gotestlist"
  },
  {
    "name": "TestDirs",
    "benchmark": false,
    "fuzz": false,
    "suite": "",
    "file": "/home/crazy/src/github.com/crazy-max/gotestlist/cmd/gotestlist/gotestlist_test.go",
    "pkg": "main"
  }
]
```

```console
$ gotestlist -f "Pkg: {{.Pkg}} | TestName: {{.Name}} | File: {{.File}}" ./...
Pkg:     gotestlist     |     TestName:     TestTests     |     File:     /home/crazy/src/github.com/crazy-max/gotestlist/gotestlist_test.go
Pkg:     main           |     TestName:     TestDirs      |     File:     /home/crazy/src/github.com/crazy-max/gotestlist/cmd/gotestlist/gotestlist_test.go
```

## Contributing

Want to contribute? Awesome! The most basic way to show your support is to star the project, or to raise issues. You
can also support this project by [**becoming a sponsor on GitHub**](https://github.com/sponsors/crazy-max) or by making
a [Paypal donation](https://www.paypal.me/crazyws) to ensure this journey continues indefinitely!

Thanks again for your support, it is much appreciated! :pray:

## License

MIT. See `LICENSE` for more details.
