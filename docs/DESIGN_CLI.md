The terminal tool that allows 
    - an admin user to interact with the server 
    - a worker to be assigned work from the server

If the command is successful, exit 0.  If it fails, exit 1.

### Identity and Authentication

The terminal client may used by either the ADMIN, USER, WORKER or ORCHESTRATOR from a shell.

ADMIN: a human administrator user.
USER: a human non-administrator user.
WORKER: a nonhuman agent
ORCHESTRATOR: a nonhuman process that reviews all tasks and progress.

The caller must first "login" then their session token will be stored to $SF_HOME or ~/.config/sf/session.json at which point the session token will identify the caller.   At this point -userame and -password are no longer necessary until the session expires.

ALL calls requiring authentication to the server MUST be Basic-Auth.

The USER, ADMIN, WORKER and ORCHESTRATOR provide this using environment variables:

```bash
export SF_USERNAME=XXX
export SF_PASSWORD=YYY
sf <command>
```

Note: the CLI never touches the database directly - it ALWAYS routes throught the SERVER.

## CLI Commands

Use single-hyphen for options, never double (except for --force).

SF_URL or -url points to the SERVER

Option `-json` in any command in CLI prints to STDOUT the response in pretty-printed JSON, otherwse print human-readable

`$SF_URL` or `-url` specify the location of the SF_SERVER - which is normally https://localhost:8080

# Commands

## Login

Creates a session token valid for the `session.duration` configuration value (or permanently if `session.duration` is `0`).

```bash
# login (stores session token in ~/.config/sf/config.json)
# will request username and password if not supplied and not available via env vars
SF_USERNAME=xxx
SF_PASSWORD=yyy
sf login (-username XXX -password YYYY)
>username:
>password:
```

```bash
# register a new user
# this API is enabled if the server config `registration.enabled` = `true`
sf register -username XXX -password YYYY (-email email@place.com) 
```

## Commands

```bash
# show usage for ALL commands
# does not require authentication or any network call
sf
```

## User command

All `sf user` commands are usable only by an admin and will return forbidden to any other caller.

```bash
# An enabled user can login to the website and use the client
sf user enable -username XXXX

# A disabled user cannot login to the website or use the client
# A user disabled will not be permitted by the server to carry out any actions - a 401 will be sent.   
sf user disable -username XXXX

# list all users
sf user list

# reset a user passsord
sf user reset-password -username XXX -password YYYY
```

## Worker Commands

All `sf worker` commands are usable only by an admin and will return forbidden to any other caller.

```bash
## creates a new worker type user
sf worker create -worker_id XXX (-password YYYY)
> password is returned to STDOUT 

## list all workers
sf worker list

## enable/disable a worker
sf worker enable -worker_id XXX

## enable/disable a worker
sf worker disable -worker_id XXX

## resets a worker password
sf worker reset-password -worker_id XXX -password YYYY

```

## Config Admin Commands

All `sf config` commands are usable only by an admin and will return forbidden to any other caller.

```bash
# set a configuration value
sf config set -key KEY -value VALUE

# delete a configuration value
sf config rm -key KEY

# get all configuration
sf config list|ls

# get all configuration (as pretty-printed json)
sf config list|ls -json
```

### Projects

`sf project` commands are available to ADMIN, USER, WORKER and ORCHESTRATOR

```bash
# activate project NAME as default working project (stores in ~/.config/sf/config.json)
# this avoids the need for -project XXX

sf project create -project_id XXX -title XXX -description XXX -prefix ABC
sf project list -project_id XXX -title XXX -description XXX -prefix ABC
sf project udpate -project_id XXX -title XXX -description XXX -prefix ABC
sf project delete -project_id XXX -title XXX -description XXX -prefix ABC

# this avoids the need for -project XXX
# this sets a local (via the ~/.config/ts/config.json) setting to remember the current project
sf project set-default -project_id $NAME

# prints the current project, or 'default' if none set
# (via the ~/.config/ts/config.json)current project
sf project get-default 

# deactivate project NAME as default working project (stores in ~/.config/sf/config.json) (revert to project "default")
sf project unset-default -project_id $NAME

# list all projects
sf project list 

# update feature(s) of a project
sf project update -project_id XXXX (-title XXX -description YYYY -prefix ABC)

```

### Role

`sf role` commands are available to ADMIN, USER, WORKER and ORCHESTRATOR

```bash
# list all users
sf role list

# create a role
sf role create -title XXX -description XXX -goals XXXX

# create a role
sf role update -user_id X -description desc -type human|worker

# view role history
sf role history -user_id X
```

### Tasks

`sf task` commands are available to ADMIN, USER, WORKER and ORCHESTRATOR

```bash
# create a task
sf task create -title XXX -description YYY -acceptance_criteria ZZZ -type feature|bug|epic|chore
> returns task_id
> task_id will be the project prefix then a uuid e.g PREFIX_UUID

# returns the full task description including history
sf task get -task_id A 

# add a dependency so that B is blocking A
sf task update -task_id A -blocked_by B

# change status of a task
sf task update -task_id A -status xxxx

# update acceptance criteria
sf task update -task_id A -acceptance_criteria "xxxx"

# update task title
sf task update -task_id A -title "The title"

# update task title
sf task update -task_id A -description "The description"

# comment on a task
# The comment will store the text, the caller name, the date/time
sf task comment -task_id A -comment "The comment"

# delete a task
# soft deletes a task (marks it as is_deleted true)
sf task rm|delete -task_id N 

# list tasks that have hte specified attribute/value
# does not return the full json for each task (no history)
sf task list|ls (-project_id N) (-status N) (-type N) (-owner N)

```

### Worker-Only Task calls

Only WORKER can make the following calls or they will fail.
```bash
# request a task
# returns the task currently assigned to this worker_id
sf task request

# returns a task to the ORCHESTRATOR 
# the task will be marked as IDLE and the current_worker will be NULL
# the task history will show the task has been "freed"
sf task return (-task_id Y)
```

### Orchestrator/Admin only Task Calls

Only ADMIN and ORCHESTRATOR users can make these calls

```bash
# assign a task (to a specific user)
# normally this is not necessary as the orchestrator will decide
# when a task is assigned 3 things are related
# 1. the task
# 2. a worker to perform the work
# 3. a role the worker should perform the task with
sf task assign -task_id Y -worker_id XXXXX -role role_id

# un-assign a task (from a specific user)
# this removes the task from a given worker.  
sf task unassign -task_id Y -worker_id XXXXX
```

