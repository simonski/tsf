Use a sqlite backend that is accessed only via the server.

## Database Location

The default database location is `~/.config/task/task.db`, following the XDG Base Directory specification. This provides a standard location for user data that:
- Persists across code updates
- Separates user data from application code
- Is easily backed up
- Can be overridden with the `-f` flag

## Initialize database

```bash
# writes to the location specified with -f
./task initdb -f /path/to/database.db

# defaults to ~/.config/task/task.db
./task initdb
```

Note: `task initdb` is the ONLY command which touches the database directly as it is initialising the database.

Initialisation will:
- Create the database directory if it doesn't exist
- Initialize schema with all required tables
- Create an admin user with a generated password (printed to stdout)
- Set up system configuration
- Create the ORCHESTRATOR username/password in the USERS table.

Initialisation creates default config values in the config table:

orchestrator.heartbeat=1000
orchestrator.idle=10000
worker.heartbeat=1000
worker.idle=10000

creates default project `default`, public.
