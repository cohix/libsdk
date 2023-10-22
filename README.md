# 💄 Lipstick

> libsdk (lipstick) is a Go library/sdk for resilient distributed backend systems. It pairs well with [Makeup](https://github.com/cohix/makeup) for local development.

Lipstick is in early development. It is being designed for applications composed of small nimble services that may (or may not) run in geographically distributed regions such as edge clouds.

The key concepts are a `fabric` to connect the services of an application easily and resiliently, and a `store` that automatically replicates data amongst replicas of running services (which may be geographically distant).

The goal is to interoperate fully with Go's stdlib (and therefore the entire ecosystem of `net/http`) seamlessly, which will make it easy to adopt incrementally.

More to come.

Copyright Connor Hicks and external contributors, 2023. Apache-2.0 licensed, see LICENSE file.
