# Fresh Container

## The goal of the tool

Given a version constraint and a container image name, `fresh-container`
determines whether the image is fresh or stale. For stale images `fresh-container`
gives the name of the latest tag that satisfies the constraint provided by the
user.

fresh-container brings [semantic versioning](https://semver.org/) checks to the
world of containers.

## Making an analogy with software development

Developers are used to express the dependencies of their programs using
semantic versioning constrains.

For example a Node.js application relying on `left-pad` could force only
certain versions of this library to be used by specifying a constraint like
`>= 1.1.0 < 1.2.0`. This would force `npm` to install the latest version
of the library that satisfies the constraint.

Going back to the container world, imagine the following scenario. A developer
needs to deploy an instance of the [nginx](https://hub.docker.com/_/nginx)
container.
The maintainers of the nginx image are providing tags compatible with semantic
versioning. That allows the developer to start by deploying version 1.9.0 and
then, using `fresh-container`, be aware of patch releases of this container
image.

## Usage

This is the simplest way to use `fresh-container`:

```bash
$ fresh-container check --constraint ">= 1.9.0 < 1.10.0" nginx:1.9.0

The 'docker.io/library/nginx' container image can be upgraded from the '1.9.0' tag to the '1.9.15' one and still satisfy the '>= 1.9.0 < 1.10.0' constraint.
```

Behind the scenes `fresh-container` will query the container registry hosting
the image to gather the list of all the available tags.
The tags that do not respect semantic versioning will be ignored and finally
the tool will evaluate the constraint provided by the user.

The tool can also provide output in json format by using the `-o json` flag.

### Expressing constraint

`fresh-container` relies on the [blang/semver](https://github.com/blang/semver)
library.

Constraints have to be expressed using [this](https://github.com/blang/semver#ranges)
syntax.

## Server mode

Querying the remote container registries to fetch all the available tags of a
container image is an expensive operation. That gets even worse when multiple
containers have to be inspected on a regular basis.

The `fresh-container` binary can operate in a server mode to alleviate this issue.

Consider the following command:

```bash
$ fresh-container server
```

This will start a simple web server exposing a REST interface that can be used
to query the stale status of a container image.

The server will query the remote container registries and cache the results
into an in-memory database (fresh-container relies on [dgraph-io/badger](https://github.com/dgraph-io/badger)
for that).

The entries of the cache expire after a customizable interval of time.

The `fresh-container` binary can perform a request against a remote server by
using the following command:

```bash
$ fresh-container check --server fresh-service.local.lan --constraint ">= 1.0.0 < 2.0.0" influxdb:1.2.3
```


## Configuration

`fresh-container` has a simple json configuration file that covers the following
options:

  * registry configuration: that allows the user to specify special connection
    options on a per registry basis. For example: authentication credentials,
    tls options,...
  * cache TTL (hours): this tunes the amount of time container image tags are
    stored inside of the in-memory database used by `fresh-container` server.

Registry configurations are stored inside of a map where the url of the
registry is used as key, while the value is a struct with the following
attributes:

  * `auth_domain`: alternate URL for registry authentication (ex. auth.docker.io) (default: none)
  * `insecure`: do not verify tls certificates (default: false)
  * `non_ssl`: do not use ssl secure connection (default: false)
  * `skip_ping`: skip pinging the registry while establishing connection (default: false)
  * `username`: username for the registry (default: none)
  * `password`: password for the registry (default: none)

You can find a simple configuration under the `examples` directory.

