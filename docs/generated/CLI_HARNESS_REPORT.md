# CLI Harness Report

Source: `/Users/simon/code/ai/tsf/docs/CLI_HARNESS.md`

--------------------------------------------------------------------------------
name: summary
total: 30
passed: 30
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
capture: project_pos_id=id\s*:\s*([a-z0-9-]+)
actual_cmd: sf project create "my project"
actual_exit: 0
pass: true
actual_stdout: Project created: ⏎ created_at: 2026-02-20T09:52:17Z ⏎ created_by: d70c80e6-7c9e-4187-a8e4-4101f4d0e76e ⏎ description: my project ⏎ id: 0b917e00-cd29-47e1-bfe6-9d2a1efb23b7 ⏎ name: my project...
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: project list alias ls
cmd: sf project ls
exit: 0
stdout: Found
actual_cmd: sf project ls
actual_exit: 0
pass: true
actual_stdout: Found 2 projects: ⏎  ⏎ ID:   46eda86f-e9df-4a68-8be1-b626b47e2dc6 ⏎ Name: default ⏎ Desc: Default project for general tasks ⏎  ⏎ ID:   0b917e00-cd29-47e1-bfe6-9d2a1efb23b7 ⏎ Name: my pro...
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: project create explicit id
cmd: sf project create -project_id proj_cli_harness -name "CLI Harness Project" -description "project for cli harness"
exit: 0
stdout: Project created:
capture: project_id=id\s*:\s*([a-z0-9-]+)
actual_cmd: sf project create -project_id proj_cli_harness -name "CLI Harness Project" -description "project for cli harness"
actual_exit: 0
pass: true
actual_stdout: Project created: ⏎ created_at: 2026-02-20T09:52:17Z ⏎ created_by: d70c80e6-7c9e-4187-a8e4-4101f4d0e76e ⏎ description: project for cli harness ⏎ id: eeab37a7-72cb-46c7-8109-9555c1239a68 ⏎ nam...
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: project get by id
cmd: sf project get -project_id ${project_id}
exit: 0
stdout: ${project_id}
actual_cmd: sf project get -project_id eeab37a7-72cb-46c7-8109-9555c1239a68
actual_exit: 0
pass: true
actual_stdout: created_at: 2026-02-20T09:52:17Z ⏎ created_by: d70c80e6-7c9e-4187-a8e4-4101f4d0e76e ⏎ description: project for cli harness ⏎ id: eeab37a7-72cb-46c7-8109-9555c1239a68 ⏎ name: CLI Harness Projec...
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: project update by id
cmd: sf project update -project_id ${project_id} -name "CLI Harness Project Updated"
exit: 0
stdout: Project updated:
actual_cmd: sf project update -project_id eeab37a7-72cb-46c7-8109-9555c1239a68 -name "CLI Harness Project Updated"
actual_exit: 0
pass: true
actual_stdout: Project updated: ⏎ created_at: 2026-02-20T09:52:17Z ⏎ created_by: d70c80e6-7c9e-4187-a8e4-4101f4d0e76e ⏎ description: project for cli harness ⏎ id: eeab37a7-72cb-46c7-8109-9555c1239a68 ⏎ nam...
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: task create json and capture task id
cmd: sf task create -title "CLI Harness Task" -description "task for cli harness" -project_id ${project_id} -json
exit: 0
stdout: "id"
capture: task_id="id"\s*:\s*"([a-z0-9-]+)"
actual_cmd: sf task create -title "CLI Harness Task" -description "task for cli harness" -project_id eeab37a7-72cb-46c7-8109-9555c1239a68 -json
actual_exit: 0
pass: true
actual_stdout: Task created: ⏎ { ⏎   "created_at": "2026-02-20T09:52:17Z", ⏎   "created_by": "d70c80e6-7c9e-4187-a8e4-4101f4d0e76e", ⏎   "description": "task for cli harness", ⏎   "id": "edb097c9-a8c5-4a65...
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: task create positional title defaults
cmd: sf task create "bar" -project_id ${project_id} -json
exit: 0
stdout: "title";;"bar"
actual_cmd: sf task create "bar" -project_id eeab37a7-72cb-46c7-8109-9555c1239a68 -json
actual_exit: 0
pass: true
actual_stdout: Task created: ⏎ { ⏎   "created_at": "2026-02-20T09:52:17Z", ⏎   "created_by": "d70c80e6-7c9e-4187-a8e4-4101f4d0e76e", ⏎   "description": "", ⏎   "id": "52a19765-4342-4300-9294-87afd4554065",...
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: task list alias ls
cmd: sf task ls -project_id ${project_id}
exit: 0
stdout: Found
actual_cmd: sf task ls -project_id eeab37a7-72cb-46c7-8109-9555c1239a68
actual_exit: 0
pass: true
actual_stdout: Found 2 tasks: ⏎  ⏎ ID:     edb097c9-a8c5-4a65-bb93-1995a687a609 ⏎ Title:  CLI Harness Task ⏎ Status: idle ⏎ Priority: medium ⏎  ⏎ ID:     52a19765-4342-4300-9294-87afd4554065 ⏎ Title:...
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: task get by captured id
cmd: sf task get -task_id ${task_id}
exit: 0
stdout: ${task_id}
actual_cmd: sf task get -task_id edb097c9-a8c5-4a65-bb93-1995a687a609
actual_exit: 0
pass: true
actual_stdout: created_at: 2026-02-20T09:52:17Z ⏎ created_by: d70c80e6-7c9e-4187-a8e4-4101f4d0e76e ⏎ description: task for cli harness ⏎ id: edb097c9-a8c5-4a65-bb93-1995a687a609 ⏎ is_complete: false ⏎ is_d...
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: task update by captured id
cmd: sf task update -task_id ${task_id} -description "updated by cli harness"
exit: 0
stdout: Task updated:
actual_cmd: sf task update -task_id edb097c9-a8c5-4a65-bb93-1995a687a609 -description "updated by cli harness"
actual_exit: 0
pass: true
actual_stdout: Task updated: ⏎ created_at: 2026-02-20T09:52:17Z ⏎ created_by: d70c80e6-7c9e-4187-a8e4-4101f4d0e76e ⏎ description: updated by cli harness ⏎ id: edb097c9-a8c5-4a65-bb93-1995a687a609 ⏎ is_comp...
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: task delete alias rm by captured id
cmd: sf task rm -task_id ${task_id}
exit: 0
stdout: Task deleted
actual_cmd: sf task rm -task_id edb097c9-a8c5-4a65-bb93-1995a687a609
actual_exit: 0
pass: true
actual_stdout: Task deleted ⏎
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: user create
cmd: sf user create -username cli_harness_user -password pass123 -type human
exit: 0
stdout: User created:
capture: user_id=id\s*:\s*([a-z0-9-]+)
actual_cmd: sf user create -username cli_harness_user -password pass123 -type human
actual_exit: 0
pass: true
actual_stdout: User created: ⏎ created_at: 2026-02-20T09:52:18Z ⏎ id: 3786d9a1-d6b4-434d-8740-5ad8b7c20cc6 ⏎ is_active: true ⏎ type: human ⏎ updated_at: 2026-02-20T09:52:18Z ⏎ username: cli_harness_user ...
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
actual_cmd: sf user get -user_id 3786d9a1-d6b4-434d-8740-5ad8b7c20cc6
actual_exit: 0
pass: true
actual_stdout: created_at: 2026-02-20T09:52:18Z ⏎ id: 3786d9a1-d6b4-434d-8740-5ad8b7c20cc6 ⏎ is_active: true ⏎ type: human ⏎ updated_at: 2026-02-20T09:52:18Z ⏎ username: cli_harness_user ⏎
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
capture: role_id=id\s*:\s*([a-z0-9-]+)
actual_cmd: sf role create -title "CLI Harness Role" -description "role for harness" -goals "validate cli harness"
actual_exit: 0
pass: true
actual_stdout: Role created: ⏎ created_at: 2026-02-20T09:52:18Z ⏎ created_by: d70c80e6-7c9e-4187-a8e4-4101f4d0e76e ⏎ description: role for harness ⏎ goals: validate cli harness ⏎ id: 9f5c124b-3229-4214-9a8...
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: role list alias ls
cmd: sf role ls
exit: 0
stdout: Found
actual_cmd: sf role ls
actual_exit: 0
pass: true
actual_stdout: Found 5 roles: ⏎  ⏎ ID:   9f5c124b-3229-4214-9a86-3bfc315e8230 ⏎ Name: CLI Harness Role ⏎  ⏎ ID:   cddfb185-49f6-4855-a9c7-d8c94c8fbc2a ⏎ Name: analyst ⏎  ⏎ ID:   004a8e87-1476-4b84-bc...
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: role get by captured id
cmd: sf role get -role_id ${role_id}
exit: 0
stdout: ${role_id}
actual_cmd: sf role get -role_id 9f5c124b-3229-4214-9a86-3bfc315e8230
actual_exit: 0
pass: true
actual_stdout: created_at: 2026-02-20T09:52:18Z ⏎ created_by: d70c80e6-7c9e-4187-a8e4-4101f4d0e76e ⏎ description: role for harness ⏎ goals: validate cli harness ⏎ id: 9f5c124b-3229-4214-9a86-3bfc315e8230 ⏎...
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: role update by captured id
cmd: sf role update -role_id ${role_id} -description "updated description"
exit: 0
stdout: Role updated:
actual_cmd: sf role update -role_id 9f5c124b-3229-4214-9a86-3bfc315e8230 -description "updated description"
actual_exit: 0
pass: true
actual_stdout: Role updated: ⏎ created_at: 2026-02-20T09:52:18Z ⏎ created_by: d70c80e6-7c9e-4187-a8e4-4101f4d0e76e ⏎ description: updated description ⏎ goals:  ⏎ id: 9f5c124b-3229-4214-9a86-3bfc315e8230 �...
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: role delete alias rm
cmd: sf role rm -role_id ${role_id}
exit: 0
stdout: Role deleted
actual_cmd: sf role rm -role_id 9f5c124b-3229-4214-9a86-3bfc315e8230
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
actual_stdout: Config set: ⏎ created_at: 2026-02-20T09:52:19Z ⏎ key: cli.harness.enabled ⏎ updated_at: 2026-02-20T09:52:19Z ⏎ value: true ⏎
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: config get
cmd: sf config get -key cli.harness.enabled
exit: 0
stdout: cli.harness.enabled
actual_cmd: sf config get -key cli.harness.enabled
actual_exit: 0
pass: true
actual_stdout: created_at: 2026-02-20T09:52:19Z ⏎ key: cli.harness.enabled ⏎ updated_at: 2026-02-20T09:52:19Z ⏎ value: true ⏎
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
actual_cmd: sf project rm -project_id 0b917e00-cd29-47e1-bfe6-9d2a1efb23b7
actual_exit: 0
pass: true
actual_stdout: Project deleted ⏎
--------------------------------------------------------------------------------

