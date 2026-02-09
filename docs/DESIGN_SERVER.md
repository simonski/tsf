The SERVER serves both the FRONTEND and the APIs.

The APIs are restful openAPI spec compatible and include a websocket for realtime access to current WORK state.

The server is really only a CRUD manager of state.  It will not make many "decisions" on what to do.

The SERVER is the ONLY component that read/writes the database.

## API

A CRUD api exists to manage tasks creation, it also exposes the openAPI specification.   

Versioning: use an openAPI /v1/ style to start with.

## Usage

```bash
# runs the server on the default server port
./sf server

# runs the server on the default server port, with explicit database
./sf server -f mydb.db
```

## OpenAPI

Each endpoint must have a fully documentated section in the OpenAPI specification.  

## Authentication

All calls to the server MUST be Basic-Auth OR using a session token once the user has logged in.

The credentials must be encrypted - argon2id.