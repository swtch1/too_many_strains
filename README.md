# too many strains!

Do you have so many strains of Mary Jane that you can't keep up with all of them?  Now that's my kind if problem...
but a problem none the less.  Use this server to track all of those strains and their associated traits so you
can get to that sticky info when you need it.

## Features
- fast, concurrent API for creation, reading, updating, and deleting strain traits
- automatic database setup/migration
- database seeding through JSON file ingestion
- well tested, of course...

## Database Initialization and Migration
Use the migration script to bootstrap your database.  If using a fresh database server
the migration script will create all of the necessary tables, or upgrade to the latest version of the schema
if the database is not on the correct version.

Run the migration wrapper script to run the migration with defaults.
```bash
./migrate_db.sh
```

Or use a custom configuration from the migration script's directory.
```bash
cd cmd/database-migration
go run . --help

# use strains file from project root
go run . --database-seed-file ../../strains.json
```

## API Server
Run the API server to interact with the strains database through RESTful API requests. Note that the server depends on
a populated and running database so make sure to connect to one or run the database migration first.
```bash
./build.sh
./bin/tms --version
./bin/tms --help
./bin/tms &

curl http://127.0.0.1:8888/api/strains/id/1 | jq .
```

## Testing
Unit tests should be run from the root directory in the normal way.
```bash
go test -v ./...
```

To run integration testing as well, enable the integration flag.
```bash
go test -v ./... -tags=integration
```

Integration tests will connect to a database to connect to.  This can be fully fledged database
server or a local test database in a container.  A testing database will be created on the server
and deleted after the tests are run.

When integration testing the database is assumed to have username `root` and password `password`.
