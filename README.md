# sda (Simple Docker Apps)

A cli application for simply creating various servers or apps in a docker containers

This tool is not really meant to be robust, but rather a quick way to get a server up and running in a docker with
already set up volumes and ports as well as sitting in the same network as other services.

It is meant only for development purposes as it does not provide any security or performance optimizations needed for use in production.

It replaces for me a workflow of going to docker hub, finding an image, reading the docs, and running it via `docker run` every time I need it.
Instead I add it to config file and it is ready to go wherever I need it.

I provide a hefty list of already supported servers, but you can easily add your own by editing the `sda.yaml` file
in `$HOME/.config/sda` directory.

## Disclaimers

**This is PoC, a rewrite of a [powershell module](https://github.com/stefanjarina/SimpleDockerApps) I've used for a long time**
**The API might still be changing**

**This is also project written while learning language and because I use it often
when I need some dirty server in docker**

## Installation

- Download linux or windows binary from [Releases](https://github.com/stefanjarina/sda/releases/latest)

- Using Golang

```powershell
go install github.com/stefanjarina/sda@latest
```

## Usage

```bash
# list all running services
sda list   # -or-  sda list -r
# list all available services
sda list -a
# list all stopped services
sda list -s
# list all created services (running + stopped)
sda list -c

# Create a new service with defaults
sda create mssql
# Create a new service with custom password
sda create mssql -p mypassword
# Create a new service with custom password and do not start it immediately
sda create mssql -p mypassword --no-start

# Start a service
sda start mssql

# Stop a service
sda stop mssql

# Remove a service
sda remove mssql
# Remove a service and all volumes
sda remove mssql --volumes

# Show service info
sda show mssql
# Show service info in json
sda show mssql --json

# Connect to a service (if supported)
sda connect mssql
# Connect to a service with custom password
sda connect mssql -p mypassword
# Open in browser (if supported)
sda connect ravendb --web
```

## Supported Services

| name            | Website                                                              | Docker HUB                                                      |
|-----------------|----------------------------------------------------------------------|-----------------------------------------------------------------|
| mssql           | [MS SQL](https://www.microsoft.com/en-us/sql-server/sql-server-2019) | [Docker HUB](https://hub.docker.com/_/microsoft-mssql-server)   |
| postgres        | [Postgres](https://www.postgresql.org/)                              | [Docker HUB](https://hub.docker.com/_/postgres)                 |
| mariadb         | [Mariadb](https://mariadb.org/)                                      | [Docker HUB](https://hub.docker.com/_/mariadb)                  |
| mysql           | [Mariadb](https://mariadb.org/)                                      | [Docker HUB](https://hub.docker.com/_/mysql)                    |
| mongodb         | [Mongodb](https://www.mongodb.com/)                                  | [Docker HUB](https://hub.docker.com/_/mongo)                    |
| redis           | [Redis](https://redis.io/)                                           | [Docker HUB](https://hub.docker.com/_/redis)                    |
| redispersistent | [Redis](https://redis.io/)                                           | [Docker HUB](https://hub.docker.com/_/redis)                    |
| cassandra       | [Cassandra](http://cassandra.apache.org/)                            | [Docker HUB](https://hub.docker.com/_/cassandra)                |
| ravendb         | [Ravendb](https://ravendb.net/)                                      | [Docker HUB](https://hub.docker.com/r/ravendb/ravendb)          |
| clickhouse      | [Clickhouse](https://clickhouse.yandex/)                             | [Docker HUB](https://hub.docker.com/r/yandex/clickhouse-server) |
| dremio          | [Dremio](https://www.dremio.com/)                                    | [Docker HUB](https://hub.docker.com/r/dremio/dremio-oss)        |
| dynamodb        | [Dynamodb](https://aws.amazon.com/dynamodb/)                         | [Docker HUB](https://hub.docker.com/r/amazon/dynamodb-local/)   |
| elasticsearch   | [Elasticsearch](https://www.elastic.co/)                             | [Docker HUB](https://hub.docker.com/_/elasticsearch)            |
| solr            | [Solr](https://lucene.apache.org/solr/)                              | [Docker HUB](https://hub.docker.com/_/solr)                     |
| neo4j           | [Neo4j](https://neo4j.com/)                                          | [Docker HUB](https://hub.docker.com/_/neo4j)                    |
| orientdb        | [OrientDB](https://orientdb.com/)                                    | [Docker HUB](https://hub.docker.com/_/orientdb)                 |
| arangodb        | [ArangoDB](https://www.arangodb.com/)                                | [Docker HUB](https://hub.docker.com/_/arangodb)                 |
| rethinkdb       | [RethinkDB](https://rethinkdb.com/)                                  | [Docker HUB](https://hub.docker.com/_/rethinkdb)                |
| presto          | [Presto](https://prestodb.io/)                                       | [Docker HUB](https://hub.docker.com/r/starburstdata/presto)     |
| scylladb        | [ScyllaDB](https://www.scylladb.com/)                                | [Docker HUB](https://hub.docker.com/r/scylladb/scylla)          |
| firebird        | [Firebird](https://firebirdsql.org/)                                 | [Docker HUB](https://hub.docker.com/r/jacobalberty/firebird)    |
| vertica         | [Vertica](https://www.vertica.com/)                                  | [Docker HUB](https://hub.docker.com/r/jbfavre/vertica)          |
| crate           | [Crate](https://crate.io/)                                           | [Docker HUB](https://hub.docker.com/_/crate)                    |
| couchbase       | [Couchbase](https://www.couchbase.com/)                              | [Docker HUB](https://hub.docker.com/_/couchbase)                |
| marklogic       | [MarkLogic Server](https://www.progress.com/marklogic)               | [Docker HUB](https://hub.docker.com/r/marklogicdb/marklogic-db) |
| surrealdb       | [SurrealDB](https://surrealdb.com/)                                  | [Docker HUB](https://hub.docker.com/r/surrealdb/surrealdb)      |
| aerospike       | [Aerospike](https://aerospike.com/)                                  | [Docker HUB](https://hub.docker.com/_/aerospike)                |
| portainer       | [Portainer](https://www.portainer.io/)                               | [Docker HUB](https://hub.docker.com/r/portainer/portainer)      |

## TODO

- [ ] better output for list command
- [ ] add logs command
- [ ] add --recreate flag to create command
- [ ] add tests
- [ ] support bulk actions (stop/start all, etc.)
- [ ] Generate documentation
- [ ] More general polish (e.g. typos, common messages to be similar, naming to be similar, etc.)
- [X] Add GitHub Actions for CI/CD
  - [X] Create GitHub release
- [ ] Improve installation instructions
  - [ ] create windows installer
  - [ ] investigate scoop, chocolatey, winget
  - [ ] create deb + rpm packages
  - [ ] investigate snap
  - [ ] investigate AUR
- [ ] Add more customization options with sane defaults (e.g. custom ports, custom network, ...)
- [ ] Add versioning autoincrement (via tags?)
- [ ] Support calling docker compose maybe? (e.g. for more complex setups)

### Ultimate TODO for services

- [ ] elasticsearch - fix cli connect command
- [x] SurrealDB - add support
