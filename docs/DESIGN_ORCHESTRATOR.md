This describes the orchestrator.

## Building

The orchestrator is built in the same binary as the server and client.  

The orchestrator NEVER touches the database - it calls the SERVER which is the only component that touches the database.

# run the orchestrator

```bash
# Runs and REGISTERs the orchestrator on the TASK_URL server
# implicit TASK_URL= default server url 
./task orchestrator
```

```bash
# REGISTERs the orchestrator on the specified -url
./task orchestrator -url https://localhost:8080
```

## Hearbeat

The orchestrator periodically issues a heartbeat to the SERVER to indicate it is running.  If the orchestrator does not heartbeat, the server will move the orchestrator to an IDLE state. If it heartbeats, it will persist the last active satae and designate the orchestrator as ACTIVE.  

These values are configurable via a `./task config key value` admin-only call.

Default config key/values in milliseconds:
`orchestrator.heartbeat`: 1000 
`orchestrator.idle`: 10000

The orchestrator can also be run "within" the server process rather than standalone if it is passed during the start of the server command:

```bash
./task server orchestrator
```

This is just so it is easier to run fewer components.

## Usage

The goal of the orchestrator is to ensure all tasks are eithe rsuccessfully completed, or that they definitively cannot complete and are sent to a HUMAN to clarify/refine/fix.

It queries all non-closed tasks and decides what it should do (meaning which WORKER to send the work on to).

