# too_many_strains

Do you have so many strains of Mary Jane that you can't keep up with all of them?  Now that's my kind if problem...
but a problem none the less.  Use this server to track all of those strains and their associated traits so you
can get to that sticky info when you need it.

# Features
- fast, concurrent API for creation, reading, updating, and deleting strain traits
- automatic database setup/migration
- database seeding through JSON file ingestion
- well tested, of course

## Testing
Unit tests should be run from the main directory in the normal way.
```bash
go test -v ./...
```

To run integration testing as well, enable the integration flag.
```bash
go test -v ./... -args -Integration=true
```

Integration tests assumes you are connecting to a database.  This can be fully fledged database
server or a local test database in a container.  A testing database will be created on the server
and deleted after the tests are run.

When integration testing the database is assumed to have username `root` and password `password`.
