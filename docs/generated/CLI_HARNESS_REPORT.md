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
capture: project_pos_id=id\s*:\s*([a-z0-9-]+)
actual_cmd: sf project create "my project"
actual_exit: 0
pass: true
actual_stdout: Project created: ⏎ created_at: 2026-02-20T09:25:55Z ⏎ created_by: 5d41793b-9ac3-4e5e-b910-bce94c914dde ⏎ description: my project ⏎ id: d6218e2f-3e14-42c4-a1c9-2289981cd409 ⏎ name: my project...
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: project list alias ls
cmd: sf project ls
exit: 0
stdout: Found
actual_cmd: sf project ls
actual_exit: 0
pass: true
actual_stdout: Found 2 projects: ⏎  ⏎ ID:   67c973f4-52b6-4aa1-ac19-8b4b216f4de1 ⏎ Name: default ⏎ Desc: Default project for general tasks ⏎  ⏎ ID:   d6218e2f-3e14-42c4-a1c9-2289981cd409 ⏎ Name: my pro...
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
actual_stdout: Project created: ⏎ created_at: 2026-02-20T09:25:55Z ⏎ created_by: 5d41793b-9ac3-4e5e-b910-bce94c914dde ⏎ description: project for cli harness ⏎ id: 818ca1e6-3229-4aea-a6dc-3f77ce2e9635 ⏎ nam...
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: project get by id
cmd: sf project get -project_id ${project_id}
exit: 0
stdout: ${project_id}
actual_cmd: sf project get -project_id 818ca1e6-3229-4aea-a6dc-3f77ce2e9635
actual_exit: 0
pass: true
actual_stdout: created_at: 2026-02-20T09:25:55Z ⏎ created_by: 5d41793b-9ac3-4e5e-b910-bce94c914dde ⏎ description: project for cli harness ⏎ id: 818ca1e6-3229-4aea-a6dc-3f77ce2e9635 ⏎ name: CLI Harness Projec...
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: project update by id
cmd: sf project update -project_id ${project_id} -name "CLI Harness Project Updated"
exit: 0
stdout: Project updated:
actual_cmd: sf project update -project_id 818ca1e6-3229-4aea-a6dc-3f77ce2e9635 -name "CLI Harness Project Updated"
actual_exit: 0
pass: true
actual_stdout: Project updated: ⏎ created_at: 2026-02-20T09:25:55Z ⏎ created_by: 5d41793b-9ac3-4e5e-b910-bce94c914dde ⏎ description: project for cli harness ⏎ id: 818ca1e6-3229-4aea-a6dc-3f77ce2e9635 ⏎ nam...
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: task create json and capture task id
cmd: sf task create -title "CLI Harness Task" -description "task for cli harness" -project_id ${project_id} -json
exit: 0
stdout: "id"
capture: task_id="id"\s*:\s*"([a-z0-9-]+)"
actual_cmd: sf task create -title "CLI Harness Task" -description "task for cli harness" -project_id 818ca1e6-3229-4aea-a6dc-3f77ce2e9635 -json
actual_exit: 0
pass: true
actual_stdout: Task created: ⏎ { ⏎   "created_at": "2026-02-20T09:25:55Z", ⏎   "created_by": "5d41793b-9ac3-4e5e-b910-bce94c914dde", ⏎   "description": "task for cli harness", ⏎   "id": "a764337f-40ce-464f...
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: task list alias ls
cmd: sf task ls -project_id ${project_id}
exit: 0
stdout: Found
actual_cmd: sf task ls -project_id 818ca1e6-3229-4aea-a6dc-3f77ce2e9635
actual_exit: 0
pass: true
actual_stdout: Found 1 tasks: ⏎  ⏎ ID:     a764337f-40ce-464f-b5dd-edc95e1185ca ⏎ Title:  CLI Harness Task ⏎ Status: idle ⏎ Priority: medium ⏎  ⏎
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: task get by captured id
cmd: sf task get -task_id ${task_id}
exit: 0
stdout: ${task_id}
actual_cmd: sf task get -task_id a764337f-40ce-464f-b5dd-edc95e1185ca
actual_exit: 0
pass: true
actual_stdout: created_at: 2026-02-20T09:25:55Z ⏎ created_by: 5d41793b-9ac3-4e5e-b910-bce94c914dde ⏎ description: task for cli harness ⏎ id: a764337f-40ce-464f-b5dd-edc95e1185ca ⏎ is_complete: false ⏎ is_d...
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: task update by captured id
cmd: sf task update -task_id ${task_id} -description "updated by cli harness"
exit: 0
stdout: Task updated:
actual_cmd: sf task update -task_id a764337f-40ce-464f-b5dd-edc95e1185ca -description "updated by cli harness"
actual_exit: 0
pass: true
actual_stdout: Task updated: ⏎ created_at: 2026-02-20T09:25:55Z ⏎ created_by: 5d41793b-9ac3-4e5e-b910-bce94c914dde ⏎ description: updated by cli harness ⏎ id: a764337f-40ce-464f-b5dd-edc95e1185ca ⏎ is_comp...
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: task delete alias rm by captured id
cmd: sf task rm -task_id ${task_id}
exit: 0
stdout: Task deleted
actual_cmd: sf task rm -task_id a764337f-40ce-464f-b5dd-edc95e1185ca
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
actual_stdout: User created: ⏎ created_at: 2026-02-20T09:25:56Z ⏎ id: b1296497-7c77-45d5-b93e-a0f83f356a0a ⏎ is_active: true ⏎ type: human ⏎ updated_at: 2026-02-20T09:25:56Z ⏎ username: cli_harness_user ...
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
actual_cmd: sf user get -user_id b1296497-7c77-45d5-b93e-a0f83f356a0a
actual_exit: 0
pass: true
actual_stdout: created_at: 2026-02-20T09:25:56Z ⏎ id: b1296497-7c77-45d5-b93e-a0f83f356a0a ⏎ is_active: true ⏎ type: human ⏎ updated_at: 2026-02-20T09:25:56Z ⏎ username: cli_harness_user ⏎
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
actual_stdout: Role created: ⏎ created_at: 2026-02-20T09:25:56Z ⏎ created_by: 5d41793b-9ac3-4e5e-b910-bce94c914dde ⏎ description: role for harness ⏎ goals: validate cli harness ⏎ id: 3d59de12-536f-4adb-aae...
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: role list alias ls
cmd: sf role ls
exit: 0
stdout: Found
actual_cmd: sf role ls
actual_exit: 0
pass: true
actual_stdout: Found 5 roles: ⏎  ⏎ ID:   3d59de12-536f-4adb-aaeb-42cdd97f01f9 ⏎ Name: CLI Harness Role ⏎  ⏎ ID:   8b4a06cd-a390-4861-a6a9-71dc5385d451 ⏎ Name: analyst ⏎  ⏎ ID:   0f099840-cb2c-433f-b1...
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: role get by captured id
cmd: sf role get -role_id ${role_id}
exit: 0
stdout: ${role_id}
actual_cmd: sf role get -role_id 3d59de12-536f-4adb-aaeb-42cdd97f01f9
actual_exit: 0
pass: true
actual_stdout: created_at: 2026-02-20T09:25:56Z ⏎ created_by: 5d41793b-9ac3-4e5e-b910-bce94c914dde ⏎ description: role for harness ⏎ goals: validate cli harness ⏎ id: 3d59de12-536f-4adb-aaeb-42cdd97f01f9 ⏎...
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: role update by captured id
cmd: sf role update -role_id ${role_id} -description "updated description"
exit: 0
stdout: Role updated:
actual_cmd: sf role update -role_id 3d59de12-536f-4adb-aaeb-42cdd97f01f9 -description "updated description"
actual_exit: 0
pass: true
actual_stdout: Role updated: ⏎ created_at: 2026-02-20T09:25:56Z ⏎ created_by: 5d41793b-9ac3-4e5e-b910-bce94c914dde ⏎ description: updated description ⏎ goals:  ⏎ id: 3d59de12-536f-4adb-aaeb-42cdd97f01f9 �...
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: role delete alias rm
cmd: sf role rm -role_id ${role_id}
exit: 0
stdout: Role deleted
actual_cmd: sf role rm -role_id 3d59de12-536f-4adb-aaeb-42cdd97f01f9
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
actual_stdout: Config set: ⏎ created_at: 2026-02-20T09:25:57Z ⏎ key: cli.harness.enabled ⏎ updated_at: 2026-02-20T09:25:57Z ⏎ value: true ⏎
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: config get
cmd: sf config get -key cli.harness.enabled
exit: 0
stdout: cli.harness.enabled
actual_cmd: sf config get -key cli.harness.enabled
actual_exit: 0
pass: true
actual_stdout: created_at: 2026-02-20T09:25:57Z ⏎ key: cli.harness.enabled ⏎ updated_at: 2026-02-20T09:25:57Z ⏎ value: true ⏎
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
actual_cmd: sf project rm -project_id d6218e2f-3e14-42c4-a1c9-2289981cd409
actual_exit: 0
pass: true
actual_stdout: Project deleted ⏎
--------------------------------------------------------------------------------

