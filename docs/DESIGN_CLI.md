The terminal/tui tool that allows 
    - an admin user to interact with the server 
    - a worker to be assigned work from the server

### Identity

The terminal client may used by either the HUMAN, WORKER or ORCHESTRATOR from a shell.

SF_USERNAME/SF_PASSWORD must be provided, either environment variables or options in the commandline if the call requires authentication.  

# Authentication

All calls requiring authentication to the server MUST be Basic-Auth.

The user can provides this using environment variables:

```bash
export SF_USERNAME=XXX
export SF_PASSWORD=YYY
sf <command>
```

Note: the CLI never touches the database directly - it ALWAYS routes throught the SERVER.

## CLI Commands

Use single-hyphen for options, never double.

SF_URL or -url points to the SERVER

Option `-json` in any command in CLI prints to STDOUT the response in pretty-printed JSON, otherwse print human-readable

`$SF_URL` or `-url` specify the location of the SF_SERVER - which is normally https://localhost:8080

# Commands

```bash
# register a new user
sf register -username XXX -password YYYY (-email email@place.com) -type human|worker
```

```bash
# login (stores session token in ~/.config/sf/config.json)
$SF_PASSWORD=xxxx
sf login -username XXX (-password YYYY)
```


## Commands

```bash
# show usage for ALL commands
sf
```

Note: following calls require `SF_USERNAME` and `SF_PASSWORD` to be a valid authenticated user.

## Admin-Only Commands

These commands demand that the SF_USERNAME is the authenticated admin user.

Enable/Disable a user.  A user disabled will not be permitted by the server to carry out any actions - a 401 will be sent.   

```bash
sf enable-user -username XXXX
sf disable-user -username XXXX
```

The admin can set configuration values

```bash
# set a configuration value
sf config-set -key KEY -value VALUE

# delete a configuration value
sf config-rm -key KEY
```


```bash

# get all configuration
sf config-ls 

# get all configuration (as pretty-printed json)
sf config-ls -json

# list all projects
sf project list 

# activate project NAME as default working project (stores in ~/.config/sf/config.json)
sf project set $NAME

# deactivate project NAME as default working project (stores in ~/.config/sf/config.json) (revert to project "default")
sf project unset $NAME

# add a dependency so that B is blocking A
sf update -task_id A -blocked_by B

# change status of a task
sf update -task_id A -status xxxx

# update acceptance criteria
sf update -task_id A -acceptance_criteria "xxxx"

# update task title
sf update -task_id A -title "The title"

# Note that `sf update` requires the `-task_id` then any paramter should be modifiable by the corresponding input key.

# comment on a task
sf comment -task_id A -comment "The comment"


# delete a task
sf rm|delete -task_id N 

# list tasks that have hte specified attribute/value
sf list|ls (-project_id N) (-status N) (-type N) (-owner N)

# list all users
sf user list

# list all users of a given type
sf user list -type human|worker

# create a user
sf user create -user_id X -description desc -type human|worker

# claim a task (the current SF_USERNAME)
sf claim X -task_id Y

# request a task 
# normally the task chosed BY the orchestrator
sf request (-task_id Y)

# free a task 
# normally the task is whatever is being worked on so auto-identified
sf free X (-task_id Y)

# assign a task (to a specific user)
# normally this is not necessary as teh orchestrator will decide
sf assign -task_id Y -username XXXXX

# view user history
sf user history -user_id X

```

