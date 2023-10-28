# 💄 Lipstick

> libsdk (lipstick) is a Go library/sdk for resilient distributed backend systems. It pairs well with [Makeup](https://github.com/cohix/makeup) for local development.

Lipstick is in early development. It is being designed for applications composed of many small services that may (or may not) run in geographically distributed regions such as edge clouds.

libsdk is built around 2 main components:
* `fabric`: a durable messaging system to connect services of an application easily and resiliently.
* `store`: a database that automatically replicates amongst instances of running services (which may be geographically distant) using the fabric as its communication layer.

The goal is for `fabric` and `store` to be nearly invisible, and to interoperate fully with Go's stdlib (and therefore the entire ecosystem of `net/http`) seamlessly, which will make it easy to adopt incrementally.

The current status of the project is early development focused on hardening the data replication method using NATS as a fabric and SQLite as a data store. More to come.

Copyright Connor Hicks and external contributors, 2023. Apache-2.0 licensed, see LICENSE file.
