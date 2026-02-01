The SERVER serves both the FRONTEND and the APIs.

The APIs are restful openAPI spec compatible.

The server is really only a CRUD manager of state.  It will not make many "decisions" on what to do.

The SERVER is the ONLY component that read/writes the database.

## API

A CRUD api exists to manage tasks creation, it also exposes the openAPI specification.   

Versioning: use an openAPI /v1/ style to start with.

## Usage

```bash
# runs the server on the default server port
./task server

# runs the server on the default server port, with explicit database
./task server -f mydb.db
```

## OpenAPI

Each endpoint must have a fully documentated section in the OpenAPI specification.  

## Authentication

All calls to the server MUST be Basic-Auth.   

The credentials must be encrypted - argon2id.   

The server will maintain a USERS table which will contain all details of each type of user.

Password reset is available to the admin-only via `./task reset-password -username X -password Y`