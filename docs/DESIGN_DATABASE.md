Use a sqlite backend that is accessed only via the server.

## Database Location

The default database location is `~/.config/sf/sf.db`, following the XDG Base Directory specification. This provides a standard location for user data that:
- Persists across code updates
- Separates user data from application code
- Is easily backed up
- Can be overridden with the `-f` flag

## Initialize database

```bash
# writes to the location specified with -f
./sf initdb -f /path/to/database.db

# defaults to ~/.config/sf/sf.db
./sf initdb 

# defaults to ~/.config/sf/sf.db, --force overwrites
./sf initdb --force

```

Note: `sf initdb` is the ONLY command which touches the database directly as it is initialising the database.

Initialisation will:
- Create the database directory if it doesn't exist
- Initialize schema with all required tables
- Create an ADMIN user with a generated password (printed to stdout)
- Create a normal USER with a generated password (printed to stdout)
- Create 1 WORKER with a generated password (printed to stdout)
- Create a default project
- Set up system configuration values

Initialisation creates default config values in the config table:

```bash
orchestrator.heartbeat=1000
orchestrator.idle=10000
worker.heartbeat=1000
worker.idle=10000
```

creates default project `default`, public.
