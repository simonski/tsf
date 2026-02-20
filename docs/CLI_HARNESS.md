# CLI Harness Scenarios

These scenarios are executed in order by the Go CLI harness.
Each block defines one terminal command and expected result.

--------------------------------------------------------------------------------
name: no-arg project shows help
cmd: sf project
exit: 0
stdout: SF PROJECT - Project Management;;SUBCOMMANDS
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: no-arg task shows help
cmd: sf task
exit: 0
stdout: SF TASK - Task Management;;SUBCOMMANDS
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: no-arg user shows help
cmd: sf user
exit: 0
stdout: SF USER - User Management;;SUBCOMMANDS
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: no-arg role shows help
cmd: sf role
exit: 0
stdout: SF ROLE - Role Management;;SUBCOMMANDS
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: no-arg config shows help
cmd: sf config
exit: 0
stdout: SF CONFIG - Configuration Management;;SUBCOMMANDS
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: project create positional name
cmd: sf project create "my project"
exit: 0
stdout: Project created:
capture: project_pos_id="id"\s*:\s*"([^"]+)"
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: project list alias ls
cmd: sf project ls
exit: 0
stdout: Found
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: project create explicit id
cmd: sf project create -project_id proj_cli_harness -name "CLI Harness Project" -description "project for cli harness"
exit: 0
stdout: Project created:
capture: project_id="id"\s*:\s*"([^"]+)"
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: project get by id
cmd: sf project get -project_id ${project_id}
exit: 0
stdout: ${project_id}
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: project update by id
cmd: sf project update -project_id ${project_id} -name "CLI Harness Project Updated"
exit: 0
stdout: Project updated:
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: task create json and capture task id
cmd: sf task create -title "CLI Harness Task" -description "task for cli harness" -project_id ${project_id} -json
exit: 0
stdout: "id"
capture: task_id="id"\s*:\s*"([^"]+)"
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: task list alias ls
cmd: sf task ls -project_id ${project_id}
exit: 0
stdout: Found
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: task get by captured id
cmd: sf task get -task_id ${task_id}
exit: 0
stdout: ${task_id}
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: task update by captured id
cmd: sf task update -task_id ${task_id} -description "updated by cli harness"
exit: 0
stdout: Task updated:
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: task delete alias rm by captured id
cmd: sf task rm -task_id ${task_id}
exit: 0
stdout: Task deleted
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: user create
cmd: sf user create -username cli_harness_user -password pass123 -type human
exit: 0
stdout: User created:
capture: user_id="id"\s*:\s*"([^"]+)"
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: user list alias ls
cmd: sf user ls
exit: 0
stdout: Found
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: user get by captured id
cmd: sf user get -user_id ${user_id}
exit: 0
stdout: ${user_id}
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: user delete alias rm maps to disable
cmd: sf user rm -username cli_harness_user
exit: 0
stdout: soft-deleted (disabled)
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: role create and capture role id
cmd: sf role create -title "CLI Harness Role" -description "role for harness" -goals "validate cli harness"
exit: 0
stdout: Role created:
capture: role_id="id"\s*:\s*"([^"]+)"
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: role list alias ls
cmd: sf role ls
exit: 0
stdout: Found
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: role get by captured id
cmd: sf role get -role_id ${role_id}
exit: 0
stdout: ${role_id}
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: role update by captured id
cmd: sf role update -role_id ${role_id} -description "updated description"
exit: 0
stdout: Role updated:
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: role delete alias rm
cmd: sf role rm -role_id ${role_id}
exit: 0
stdout: Role deleted
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: config set
cmd: sf config set -key cli.harness.enabled -value true
exit: 0
stdout: Config set:
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: config get
cmd: sf config get -key cli.harness.enabled
exit: 0
stdout: cli.harness.enabled
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: config list alias ls
cmd: sf config ls
exit: 0
stdout: Found
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: config delete alias rm
cmd: sf config rm -key cli.harness.enabled
exit: 0
stdout: deleted
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: project delete alias rm
cmd: sf project rm -project_id ${project_pos_id}
exit: 0
stdout: Project deleted
--------------------------------------------------------------------------------
