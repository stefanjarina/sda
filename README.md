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

- [MS SQL](https://www.microsoft.com/en-us/sql-server/sql-server-2019) - [Docker HUB page](https://hub.docker.com/_/microsoft-mssql-server)
- [Postgres](https://www.postgresql.org/) - [Docker HUB page](https://hub.docker.com/_/postgres)
- [Mariadb](https://mariadb.org/) - [Docker HUB page](https://hub.docker.com/_/mariadb)
- [Mongodb](https://www.mongodb.com/) - [Docker HUB page](https://hub.docker.com/_/mongo)
- [Redis](https://redis.io/) - [Docker HUB page](https://hub.docker.com/_/redis)
- [Cassandra](http://cassandra.apache.org/) - [Docker HUB page](https://hub.docker.com/_/cassandra)
- [Ravendb](https://ravendb.net/) - [Docker HUB page](https://hub.docker.com/r/ravendb/ravendb)
- [Clickhouse](https://clickhouse.yandex/) - [Docker HUB page](https://hub.docker.com/r/yandex/clickhouse-server)
- [Dremio](https://www.dremio.com/) - [Docker HUB page](https://hub.docker.com/r/dremio/dremio-oss)
- [Dynamodb](https://aws.amazon.com/dynamodb/) - [Docker HUB page](https://hub.docker.com/r/amazon/dynamodb-local/)
- [Elasticsearch](https://www.elastic.co/) - [Docker HUB page](https://hub.docker.com/_/elasticsearch)
- [Solr](https://lucene.apache.org/solr/) - [Docker HUB page](https://hub.docker.com/_/solr)
- [Neo4j](https://neo4j.com/) - [Docker HUB page](https://hub.docker.com/_/neo4j)
- [OrientDB](https://orientdb.com/) - [Docker HUB page](https://hub.docker.com/_/orientdb)
- [ArangoDB](https://www.arangodb.com/) - [Docker HUB page](https://hub.docker.com/_/arangodb)
- [RethinkDB](https://rethinkdb.com/) - [Docker HUB page](https://hub.docker.com/_/rethinkdb)
- [Presto](https://prestodb.io/) - [Docker HUB page](https://hub.docker.com/r/starburstdata/presto)
- [ScyllaDB](https://www.scylladb.com/) - [Docker HUB page](https://hub.docker.com/r/scylladb/scylla)
- [Firebird](https://firebirdsql.org/) - [Docker HUB page](https://hub.docker.com/r/jacobalberty/firebird)
- [Vertica](https://www.vertica.com/) - [Docker HUB page](https://hub.docker.com/r/jbfavre/vertica)
- [Crate](https://crate.io/) - [Docker HUB page](https://hub.docker.com/_/crate)
- [Portainer](https://www.portainer.io/) - [Docker HUB page](https://hub.docker.com/r/portainer/portainer)

## TODO

- [ ] better output for list command
- [ ] add logs command
- [ ] add --recreate flag to create command
- [ ] add tests
- [ ] support bulk actions (stop/start all, etc.)
- [ ] Generate documentation
- [ ] More general polish (e.g. typos, common messages to be similar, naming to be similar, etc.)
- [ ] Add GitHub Actions for CI/CD
    - [ ] Create GitHub release
- [ ] Improve installation instructions
    - [ ] create windows installer
    - [ ] investigate scoop, chocolatey, winget
    - [ ] create deb + rpm packages
- [ ] Add more customization options with sane defaults (e.g. custom ports, custom network, ...)
- [ ] Add versioning autoincrement (via tags?)
- [ ] Support calling docker compose maybe? (e.g. for more complex setups)

### Ultimate TODO for services

- [ ] elasticsearch - fix cli connect command
- [ ] SurrealDB - add support