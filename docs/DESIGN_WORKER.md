A CLI invocation of the `task` binary that that runs as a daemon seeking WORK from the server.

Effectively a for-loop that 
    - requests work from the server (where it is calling the serfver with the WORKER_ID)
    - performs the work (delegating to an LLM)
    - commits the work once complete, updating the server
    - waits for more work from the server (using -wait N seconds) or the wait period is set by the server once registered as a server config value.
    - periodically heartbeats its status back to the server, where the server responds with OK and any config necessary)

## Registration

Once a WORKER status, it issues a REGISTER call which is returned with the configuration appropriate to this WORKER, or a REJECTED which causes the WORKER to terminate exit code 1.

On success, a WORKER issues a REQUEST_WORK which the server will either return work or return nothing.  IF work is provided, the WORKER performs it.  If it is not, the worker sleeps for the configured period, then requests work again.

## Roles

A WORKER will only receive work and has no real IDENTITY until the SERVER assigns an IDENTITY.  An IDENTITY is like a role or motivation - e.g. Business Analyst, Programmer, Tester, DevOps, Product Manager, User Acceptance Tester.  The ROLE will provide MOTIVATION that is teh context of "how is the worker supposed to try to solve the Task given to it"

This ROLE is provided to the WORKER by the SERVER whenever work is assigned.

## Usage

# implicit SF_URL= default server url 
# uses the same SF_USERNAME/SF_PASSWORD variables as a human user.
./sf worker -worker_id USERNAME

# explicit SF_URL=https://localhost
export SF_URL=https://localhost
./sf worker -username USERNAME

# override url with passed value
./sf worker -username USERNAME -url https://server:4333

## Heartbeat

The WORKER will register itself as "alive" and issue a HEARTBEAT back to the server periodically to indicate it exists along with its status "e.g. IDLE, WORKING on TASK 1"

Default config key/values in milliseconds:
`worker.heartbeat`: 1000 
`worker.idle`: 10000

When a worker starts it is saying "I am ready for work, my name is `SF_USERNAME`." 

If a worker starts with the same name as a currently active worker, the server should reject the worker and the worker should exit error code 1.

