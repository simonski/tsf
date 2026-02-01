The terminal/tui tool that allows you to interact with the server

### Identity

The terminal client may used by either the HUMAN, WORKER or ORCHESTRATOR from a shell.

TASK_USERNAME/TASK_PASSWORD must be provided, either environment variables or options in the commandline if the call requires authentication.  

# Authentication

All calls requiring authentication to the server MUST be Basic-Auth.

The user can provides this using environment variables:

```bash
export TASK_USERNAME=XXX
export TASK_PASSWORD=YYY
task <command>
```

Note: the CLI never touches the database directly - it ALWAYS routes throught the SERVER.

## CLI Commands

Use single-hyphen for options, never double.

TASK_URL or -url points to the SERVER

Option `-json` in any command in CLI prints to STDOUT the response in pretty-printed JSON, otherwse print human-readable

`$TASK_URL` or `-url` specify the location of the TASK_SERVER - which is normally https://localhost:8080

# Commands

```bash
task register -username XXX -password YYYY -type human|worker
```

## Admin-Only Commands

These commands demand that the TASK_USERNAME is the authenticated admin user.

Enable/Disable a user.  A user disabled will not be permitted by the server to carry out any actions - a 401 will be sent.   

```bash
task enable-user -username XXXX
task disable-user -username XXXX
```

### Configuration Commands

The admin can set configuration values

```bash
# set a configuration value
task config-set -key KEY -value VALUE

# delete a configuration value
task config-rm -key KEY

```


## Commands

```bash
# show usage for ALL commands
task
```

Note: following calls require `TASK_USERNAME` and `TASK_PASSWORD` to be a valid authenticated user.

```bash

# get all configuration
task config-ls 

# get all configuration (as pretty-printed json)
task config-ls -json

# list all projects
task project list 

# activate project NAME as default working project (stores in ~/.config/task/config.json)
task project set $NAME

# deactivate project NAME as default working project (stores in ~/.config/task/config.json) (revert to project "default")
task project unset $NAME

# add a dependency so that B is blocking A
task update -task_id A -blocked_by B

# change status of a task
task update -task_id A -status xxxx

# update acceptance criteria
task update -task_id A -acceptance_criteria "xxxx"

# update task title
task update -task_id A -title "The title"

Note that `task update` requires the `-task_id` then any paramter shoudl be modifiable by the corresponding input key.

# delete a task
task rm|delete -task_id N 

# list tasks that have hte specified attribute/value
task list|ls (-project_id N) (-status N) (-type N) (-owner N)

# list all users
task user list

# list all users of a given type
task user list -type human|worker

# create a user
task user create -user_id X -description desc -type human|worker

# claim a task (the current TASK_USERNAME)
task claim X -task_id Y

# request a task
task request 

# free a task 
task free X -task_id Y

# assign a task (to a specific user)
task assign X -task_id Y -username XXXXX

# view user history
task user history -user_id X

```

