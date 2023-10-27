<h2 align="center">A utility library for the GoShimmer and Hornet node software</h2>

<p align="center">
  <a href="https://discord.iota.org/" style="text-decoration:none;"><img src="https://img.shields.io/badge/Discord-9cf.svg?logo=discord" alt="Discord"></a>
    <a href="https://iota.stackexchange.com/" style="text-decoration:none;"><img src="https://img.shields.io/badge/StackExchange-9cf.svg?logo=stackexchange" alt="StackExchange"></a>
    <a href="https://github.com/iotaledger/hive.go/blob/master/LICENSE" style="text-decoration:none;"><img src="https://img.shields.io/github/license/iotaledger/hive.go.svg" alt="Apache 2.0 license"></a>
</p>

<p align="center">
  <a href="#about">About</a> ◈
  <a href="#prerequisites">Prerequisites</a> ◈
  <a href="#installation">Installation</a> ◈
  <a href="#getting-started">Getting started</a> ◈
  <a href="#supporting-the-project">Supporting the project</a> ◈
  <a href="#joining-the-discussion">Joining the discussion</a>
</p>

---

## About

Hive.go is a shared library that is used in the [GoShimmer](https://github.com/iotaledger/goshimmer), [Hornet](https://github.com/iotaledger/hornet) and [IOTA Core](https://github.com/iotaledger/iota-core) node software. This library contains shared:
* Data structures
* Utility methods
* Abstractions

This is beta software, so there may be performance and stability issues.
Please report any issues in our [issue tracker](https://github.com/iotaledger/hive.go/issues/new).

## Prerequisites

To use the library, you need to have at least [version 1.13 of Go](https://golang.org/doc/install) installed on your device.

To check if you have Go installed, run the following command:

```bash
go version
```

If Go is installed, you should see the version that's installed.

## Installation

To install Hive.go and its dependencies, you can use one of the following options:

* If you use Go modules, just import the packages that you want to use

    ```bash
    import (
    "github.com/iotaledger/hive.go/logger"
    "github.com/iotaledger/hive.go/node"
    )
    ```

* To download the library from GitHub, use the `go get` command

    ```bash
    go get github.com/iotaledger/hive.go
    ```

## Getting started

After you've [installed the library](#installation), you can use it in your project.

For example, to create a new `logger` instance:

```js
import "github.com/iotaledger/hive.go/logger"

log = logger.NewLogger('myNewLoggerName')
```

### Activating deadlock detection

To replace the mutexes in the `syncutils` package with [Go deadlock](https://github.com/sasha-s/go-deadlock), use the `deadlock` build flag when compiling your program.

## Supporting the project

If this library has been useful to you and you feel like contributing, consider submitting a [bug report](https://github.com/iotaledger/hive.go/issues/new), [feature request](https://github.com/iotaledger/hive.go/issues/new) or a [pull request](https://github.com/iotaledger/hive.go/pulls/).

See our [contributing guidelines](.github/CONTRIBUTING.md) for more information.

## Joining the discussion

If you want to get involved in the community, need help with getting set up, have any issues related to the library or just want to discuss IOTA, Distributed Registry Technology (DRT), and IoT with other people, feel free to join our [Discord](https://discord.iota.org/).
