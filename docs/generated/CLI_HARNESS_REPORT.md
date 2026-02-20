# CLI Harness Report

Source: `/Users/simon/code/ai/tsf/docs/CLI_HARNESS.md`

--------------------------------------------------------------------------------
name: summary
total: 29
passed: 29
failed: 0
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: no-arg project shows help
cmd: sf project
exit: 0
stdout: SF PROJECT - Project Management;;SUBCOMMANDS
actual_cmd: sf project
actual_exit: 0
pass: true
actual_stdout: ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ �...
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: no-arg task shows help
cmd: sf task
exit: 0
stdout: SF TASK - Task Management;;SUBCOMMANDS
actual_cmd: sf task
actual_exit: 0
pass: true
actual_stdout: ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ �...
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: no-arg user shows help
cmd: sf user
exit: 0
stdout: SF USER - User Management;;SUBCOMMANDS
actual_cmd: sf user
actual_exit: 0
pass: true
actual_stdout: ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ �...
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: no-arg role shows help
cmd: sf role
exit: 0
stdout: SF ROLE - Role Management;;SUBCOMMANDS
actual_cmd: sf role
actual_exit: 0
pass: true
actual_stdout: ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ �...
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: no-arg config shows help
cmd: sf config
exit: 0
stdout: SF CONFIG - Configuration Management;;SUBCOMMANDS
actual_cmd: sf config
actual_exit: 0
pass: true
actual_stdout: ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ �...
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: project create positional name
cmd: sf project create "my project"
exit: 0
stdout: Project created:
capture: project_pos_id="id"\s*:\s*"([^"]+)"
actual_cmd: sf project create "my project"
actual_exit: 0
pass: true
actual_stdout: Project created: ⏎ {"id":"cf08dd8b-cf81-476d-8f4a-44f41bd98fb7","name":"my project","description":"my project","status":"active","visibility":"public","created_at":"2026-02-20T09:07:16Z","updated_at...
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: project list alias ls
cmd: sf project ls
exit: 0
stdout: Found
actual_cmd: sf project ls
actual_exit: 0
pass: true
actual_stdout: Found 2 projects: ⏎  ⏎ ID:   34d334da-6452-495a-a084-b042cd238f88 ⏎ Name: default ⏎ Desc: Default project for general tasks ⏎  ⏎ ID:   cf08dd8b-cf81-476d-8f4a-44f41bd98fb7 ⏎ Name: my pro...
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: project create explicit id
cmd: sf project create -project_id proj_cli_harness -name "CLI Harness Project" -description "project for cli harness"
exit: 0
stdout: Project created:
capture: project_id="id"\s*:\s*"([^"]+)"
actual_cmd: sf project create -project_id proj_cli_harness -name "CLI Harness Project" -description "project for cli harness"
actual_exit: 0
pass: true
actual_stdout: Project created: ⏎ {"id":"b7346a18-01dc-43a4-98f4-707b378af54e","name":"CLI Harness Project","description":"project for cli harness","status":"active","visibility":"public","created_at":"2026-02-20T...
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: project get by id
cmd: sf project get -project_id ${project_id}
exit: 0
stdout: ${project_id}
actual_cmd: sf project get -project_id b7346a18-01dc-43a4-98f4-707b378af54e
actual_exit: 0
pass: true
actual_stdout: {"id":"b7346a18-01dc-43a4-98f4-707b378af54e","name":"CLI Harness Project","description":"project for cli harness","status":"active","visibility":"public","created_at":"2026-02-20T09:07:16Z","updated_a...
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: project update by id
cmd: sf project update -project_id ${project_id} -name "CLI Harness Project Updated"
exit: 0
stdout: Project updated:
actual_cmd: sf project update -project_id b7346a18-01dc-43a4-98f4-707b378af54e -name "CLI Harness Project Updated"
actual_exit: 0
pass: true
actual_stdout: Project updated: ⏎ {"id":"b7346a18-01dc-43a4-98f4-707b378af54e","name":"CLI Harness Project Updated","description":"project for cli harness","status":"active","visibility":"public","created_at":"202...
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: task create json and capture task id
cmd: sf task create -title "CLI Harness Task" -description "task for cli harness" -project_id ${project_id} -json
exit: 0
stdout: "id"
capture: task_id="id"\s*:\s*"([^"]+)"
actual_cmd: sf task create -title "CLI Harness Task" -description "task for cli harness" -project_id b7346a18-01dc-43a4-98f4-707b378af54e -json
actual_exit: 0
pass: true
actual_stdout: Task created: ⏎ {"id":"73a5eae8-242a-4d03-a76b-ba66c6d7afde","project_id":"b7346a18-01dc-43a4-98f4-707b378af54e","title":"CLI Harness Task","type":"task","description":"task for cli harness","status...
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: task list alias ls
cmd: sf task ls -project_id ${project_id}
exit: 0
stdout: Found
actual_cmd: sf task ls -project_id b7346a18-01dc-43a4-98f4-707b378af54e
actual_exit: 0
pass: true
actual_stdout: Found 1 tasks: ⏎  ⏎ ID:     73a5eae8-242a-4d03-a76b-ba66c6d7afde ⏎ Title:  CLI Harness Task ⏎ Status: idle ⏎ Priority: medium ⏎  ⏎
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: task get by captured id
cmd: sf task get -task_id ${task_id}
exit: 0
stdout: ${task_id}
actual_cmd: sf task get -task_id 73a5eae8-242a-4d03-a76b-ba66c6d7afde
actual_exit: 0
pass: true
actual_stdout: {"id":"73a5eae8-242a-4d03-a76b-ba66c6d7afde","project_id":"b7346a18-01dc-43a4-98f4-707b378af54e","title":"CLI Harness Task","type":"task","description":"task for cli harness","status":"idle","priority...
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: task update by captured id
cmd: sf task update -task_id ${task_id} -description "updated by cli harness"
exit: 0
stdout: Task updated:
actual_cmd: sf task update -task_id 73a5eae8-242a-4d03-a76b-ba66c6d7afde -description "updated by cli harness"
actual_exit: 0
pass: true
actual_stdout: Task updated: ⏎ {"id":"73a5eae8-242a-4d03-a76b-ba66c6d7afde","project_id":"b7346a18-01dc-43a4-98f4-707b378af54e","title":"CLI Harness Task","type":"task","description":"updated by cli harness","stat...
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: task delete alias rm by captured id
cmd: sf task rm -task_id ${task_id}
exit: 0
stdout: Task deleted
actual_cmd: sf task rm -task_id 73a5eae8-242a-4d03-a76b-ba66c6d7afde
actual_exit: 0
pass: true
actual_stdout: Task deleted ⏎
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: user create
cmd: sf user create -username cli_harness_user -password pass123 -type human
exit: 0
stdout: User created:
capture: user_id="id"\s*:\s*"([^"]+)"
actual_cmd: sf user create -username cli_harness_user -password pass123 -type human
actual_exit: 0
pass: true
actual_stdout: User created: ⏎ {"id":"32cbb3a2-895a-4e78-9eae-62070d84e9de","username":"cli_harness_user","type":"human","is_active":true,"created_at":"2026-02-20T09:07:17Z","updated_at":"2026-02-20T09:07:17Z"} �...
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: user list alias ls
cmd: sf user ls
exit: 0
stdout: Found
actual_cmd: sf user ls
actual_exit: 0
pass: true
actual_stdout: Found 5 users: ⏎  ⏎ Username: admin ⏎ Type:     human ⏎ Active:   true ⏎  ⏎ Username: cli_harness_user ⏎ Type:     human ⏎ Active:   true ⏎  ⏎ Username: orchestrator ⏎ Type:     ...
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: user get by captured id
cmd: sf user get -user_id ${user_id}
exit: 0
stdout: ${user_id}
actual_cmd: sf user get -user_id 32cbb3a2-895a-4e78-9eae-62070d84e9de
actual_exit: 0
pass: true
actual_stdout: {"id":"32cbb3a2-895a-4e78-9eae-62070d84e9de","username":"cli_harness_user","type":"human","is_active":true,"created_at":"2026-02-20T09:07:17Z","updated_at":"2026-02-20T09:07:17Z"} ⏎  ⏎
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: user delete alias rm maps to disable
cmd: sf user rm -username cli_harness_user
exit: 0
stdout: soft-deleted (disabled)
actual_cmd: sf user rm -username cli_harness_user
actual_exit: 0
pass: true
actual_stdout: User 'cli_harness_user' soft-deleted (disabled) ⏎
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: role create and capture role id
cmd: sf role create -title "CLI Harness Role" -description "role for harness" -goals "validate cli harness"
exit: 0
stdout: Role created:
capture: role_id="id"\s*:\s*"([^"]+)"
actual_cmd: sf role create -title "CLI Harness Role" -description "role for harness" -goals "validate cli harness"
actual_exit: 0
pass: true
actual_stdout: Role created: ⏎ {"id":"68086042-6972-4eec-b491-d5e78082f828","name":"CLI Harness Role","description":"role for harness","goals":"validate cli harness","scope":"global","is_active":true,"created_at":...
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: role list alias ls
cmd: sf role ls
exit: 0
stdout: Found
actual_cmd: sf role ls
actual_exit: 0
pass: true
actual_stdout: Found 5 roles: ⏎  ⏎ ID:   68086042-6972-4eec-b491-d5e78082f828 ⏎ Name: CLI Harness Role ⏎  ⏎ ID:   0d1f14f6-c451-4dc6-80b3-0e07c8faf0d2 ⏎ Name: analyst ⏎  ⏎ ID:   ad7af9f1-6791-480e-98...
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: role get by captured id
cmd: sf role get -role_id ${role_id}
exit: 0
stdout: ${role_id}
actual_cmd: sf role get -role_id 68086042-6972-4eec-b491-d5e78082f828
actual_exit: 0
pass: true
actual_stdout: {"id":"68086042-6972-4eec-b491-d5e78082f828","name":"CLI Harness Role","description":"role for harness","goals":"validate cli harness","scope":"global","is_active":true,"created_at":"2026-02-20T09:07:...
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: role update by captured id
cmd: sf role update -role_id ${role_id} -description "updated description"
exit: 0
stdout: Role updated:
actual_cmd: sf role update -role_id 68086042-6972-4eec-b491-d5e78082f828 -description "updated description"
actual_exit: 0
pass: true
actual_stdout: Role updated: ⏎ {"id":"68086042-6972-4eec-b491-d5e78082f828","name":"","description":"updated description","goals":"","scope":"global","is_active":true,"created_at":"2026-02-20T09:07:18Z","updated_a...
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: role delete alias rm
cmd: sf role rm -role_id ${role_id}
exit: 0
stdout: Role deleted
actual_cmd: sf role rm -role_id 68086042-6972-4eec-b491-d5e78082f828
actual_exit: 0
pass: true
actual_stdout: Role deleted ⏎
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: config set
cmd: sf config set -key cli.harness.enabled -value true
exit: 0
stdout: Config set:
actual_cmd: sf config set -key cli.harness.enabled -value true
actual_exit: 0
pass: true
actual_stdout: Config set: ⏎ {"key":"cli.harness.enabled","value":"true","created_at":"2026-02-20T09:07:18Z","updated_at":"2026-02-20T09:07:18Z"} ⏎  ⏎
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: config get
cmd: sf config get -key cli.harness.enabled
exit: 0
stdout: cli.harness.enabled
actual_cmd: sf config get -key cli.harness.enabled
actual_exit: 0
pass: true
actual_stdout: {"key":"cli.harness.enabled","value":"true","created_at":"2026-02-20T09:07:18Z","updated_at":"2026-02-20T09:07:18Z"} ⏎  ⏎
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: config list alias ls
cmd: sf config ls
exit: 0
stdout: Found
actual_cmd: sf config ls
actual_exit: 0
pass: true
actual_stdout: Found 5 config entries: ⏎  ⏎ cli.harness.enabled = true ⏎ orchestrator.heartbeat = 1000 ⏎ orchestrator.idle = 10000 ⏎ worker.heartbeat = 1000 ⏎ worker.idle = 10000 ⏎
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: config delete alias rm
cmd: sf config rm -key cli.harness.enabled
exit: 0
stdout: deleted
actual_cmd: sf config rm -key cli.harness.enabled
actual_exit: 0
pass: true
actual_stdout: Config key 'cli.harness.enabled' deleted ⏎
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: project delete alias rm
cmd: sf project rm -project_id ${project_pos_id}
exit: 0
stdout: Project deleted
actual_cmd: sf project rm -project_id cf08dd8b-cf81-476d-8f4a-44f41bd98fb7
actual_exit: 0
pass: true
actual_stdout: Project deleted ⏎
--------------------------------------------------------------------------------

