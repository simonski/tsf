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
actual_stdout: Project created: ⏎ created_at: 2026-02-20T10:11:22Z ⏎ created_by: 8e458f97-a918-49f9-ba41-8b01def6444f ⏎ description: my project ⏎ id: 8a6b5c06-d3f8-493a-b48a-f85f34adcd9e ⏎ name: my project...
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: project list alias ls
cmd: sf project ls
exit: 0
stdout: Found
actual_cmd: sf project ls
actual_exit: 0
pass: true
actual_stdout: Found 2 projects: ⏎  ⏎ ID:   b24cabd7-f345-4ac3-a189-e7d5abf30809 ⏎ Name: default ⏎ Desc: Default project for general tasks ⏎  ⏎ ID:   8a6b5c06-d3f8-493a-b48a-f85f34adcd9e ⏎ Name: my pro...
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
actual_stdout: Project created: ⏎ created_at: 2026-02-20T10:11:23Z ⏎ created_by: 8e458f97-a918-49f9-ba41-8b01def6444f ⏎ description: project for cli harness ⏎ id: 5eee4317-b269-4495-9c7b-7e512667edeb ⏎ nam...
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: project get by id
cmd: sf project get -project_id ${project_id}
exit: 0
stdout: ${project_id}
actual_cmd: sf project get -project_id 5eee4317-b269-4495-9c7b-7e512667edeb
actual_exit: 0
pass: true
actual_stdout: created_at: 2026-02-20T10:11:23Z ⏎ created_by: 8e458f97-a918-49f9-ba41-8b01def6444f ⏎ description: project for cli harness ⏎ id: 5eee4317-b269-4495-9c7b-7e512667edeb ⏎ name: CLI Harness Projec...
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: project update by id
cmd: sf project update -project_id ${project_id} -name "CLI Harness Project Updated"
exit: 0
stdout: Project updated:
actual_cmd: sf project update -project_id 5eee4317-b269-4495-9c7b-7e512667edeb -name "CLI Harness Project Updated"
actual_exit: 0
pass: true
actual_stdout: Project updated: ⏎ created_at: 2026-02-20T10:11:23Z ⏎ created_by: 8e458f97-a918-49f9-ba41-8b01def6444f ⏎ description: project for cli harness ⏎ id: 5eee4317-b269-4495-9c7b-7e512667edeb ⏎ nam...
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: task create json and capture task id
cmd: sf task create -title "CLI Harness Task" -description "task for cli harness" -project_id ${project_id} -json
exit: 0
stdout: "id"
capture: task_id="id"\s*:\s*"([a-z0-9-]+)"
actual_cmd: sf task create -title "CLI Harness Task" -description "task for cli harness" -project_id 5eee4317-b269-4495-9c7b-7e512667edeb -json
actual_exit: 0
pass: true
actual_stdout: Task created: ⏎ { ⏎   "created_at": "2026-02-20T10:11:23Z", ⏎   "created_by": "8e458f97-a918-49f9-ba41-8b01def6444f", ⏎   "description": "task for cli harness", ⏎   "id": "2166759d-a3cf-4d9d...
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: task create positional title defaults
cmd: sf task create "bar" -project_id ${project_id} -json
exit: 0
stdout: "title";;"bar"
actual_cmd: sf task create "bar" -project_id 5eee4317-b269-4495-9c7b-7e512667edeb -json
actual_exit: 0
pass: true
actual_stdout: Task created: ⏎ { ⏎   "created_at": "2026-02-20T10:11:23Z", ⏎   "created_by": "8e458f97-a918-49f9-ba41-8b01def6444f", ⏎   "description": "", ⏎   "id": "1c5be552-da44-49ff-ac26-8488a95d42cf",...
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: task list alias ls
cmd: sf task ls -project_id ${project_id}
exit: 0
stdout: Found
actual_cmd: sf task ls -project_id 5eee4317-b269-4495-9c7b-7e512667edeb
actual_exit: 0
pass: true
actual_stdout: Found 2 tasks: ⏎  ⏎ ID:     2166759d-a3cf-4d9d-a31a-21bc334c4ecb ⏎ Title:  CLI Harness Task ⏎ Status: idle ⏎ Priority: medium ⏎  ⏎ ID:     1c5be552-da44-49ff-ac26-8488a95d42cf ⏎ Title:...
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: task get by captured id
cmd: sf task get -task_id ${task_id}
exit: 0
stdout: ${task_id}
actual_cmd: sf task get -task_id 2166759d-a3cf-4d9d-a31a-21bc334c4ecb
actual_exit: 0
pass: true
actual_stdout: created_at: 2026-02-20T10:11:23Z ⏎ created_by: 8e458f97-a918-49f9-ba41-8b01def6444f ⏎ description: task for cli harness ⏎ id: 2166759d-a3cf-4d9d-a31a-21bc334c4ecb ⏎ is_complete: false ⏎ is_d...
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: task update by captured id
cmd: sf task update -task_id ${task_id} -description "updated by cli harness"
exit: 0
stdout: Task updated:
actual_cmd: sf task update -task_id 2166759d-a3cf-4d9d-a31a-21bc334c4ecb -description "updated by cli harness"
actual_exit: 0
pass: true
actual_stdout: Task updated: ⏎ created_at: 2026-02-20T10:11:23Z ⏎ created_by: 8e458f97-a918-49f9-ba41-8b01def6444f ⏎ description: updated by cli harness ⏎ id: 2166759d-a3cf-4d9d-a31a-21bc334c4ecb ⏎ is_comp...
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: task delete alias rm by captured id
cmd: sf task rm -task_id ${task_id}
exit: 0
stdout: Task deleted
actual_cmd: sf task rm -task_id 2166759d-a3cf-4d9d-a31a-21bc334c4ecb
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
actual_stdout: User created: ⏎ created_at: 2026-02-20T10:11:25Z ⏎ id: ff970826-b9d5-4600-a29f-4cb4f9963920 ⏎ is_active: true ⏎ type: human ⏎ updated_at: 2026-02-20T10:11:25Z ⏎ username: cli_harness_user ...
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
actual_cmd: sf user get -user_id ff970826-b9d5-4600-a29f-4cb4f9963920
actual_exit: 0
pass: true
actual_stdout: created_at: 2026-02-20T10:11:25Z ⏎ id: ff970826-b9d5-4600-a29f-4cb4f9963920 ⏎ is_active: true ⏎ type: human ⏎ updated_at: 2026-02-20T10:11:25Z ⏎ username: cli_harness_user ⏎
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
actual_stdout: Role created: ⏎ created_at: 2026-02-20T10:11:25Z ⏎ created_by: 8e458f97-a918-49f9-ba41-8b01def6444f ⏎ description: role for harness ⏎ goals: validate cli harness ⏎ id: 66dcbd4f-d049-4c6d-bc1...
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: role list alias ls
cmd: sf role ls
exit: 0
stdout: Found
actual_cmd: sf role ls
actual_exit: 0
pass: true
actual_stdout: Found 5 roles: ⏎  ⏎ ID:   66dcbd4f-d049-4c6d-bc1d-fc2dc63cd854 ⏎ Name: CLI Harness Role ⏎  ⏎ ID:   45d92add-d70e-42dd-8979-a3501fc3a653 ⏎ Name: analyst ⏎  ⏎ ID:   853fb145-cf82-4233-98...
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: role get by captured id
cmd: sf role get -role_id ${role_id}
exit: 0
stdout: ${role_id}
actual_cmd: sf role get -role_id 66dcbd4f-d049-4c6d-bc1d-fc2dc63cd854
actual_exit: 0
pass: true
actual_stdout: created_at: 2026-02-20T10:11:25Z ⏎ created_by: 8e458f97-a918-49f9-ba41-8b01def6444f ⏎ description: role for harness ⏎ goals: validate cli harness ⏎ id: 66dcbd4f-d049-4c6d-bc1d-fc2dc63cd854 ⏎...
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: role update by captured id
cmd: sf role update -role_id ${role_id} -description "updated description"
exit: 0
stdout: Role updated:
actual_cmd: sf role update -role_id 66dcbd4f-d049-4c6d-bc1d-fc2dc63cd854 -description "updated description"
actual_exit: 0
pass: true
actual_stdout: Role updated: ⏎ created_at: 2026-02-20T10:11:25Z ⏎ created_by: 8e458f97-a918-49f9-ba41-8b01def6444f ⏎ description: updated description ⏎ goals:  ⏎ id: 66dcbd4f-d049-4c6d-bc1d-fc2dc63cd854 �...
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: role delete alias rm
cmd: sf role rm -role_id ${role_id}
exit: 0
stdout: Role deleted
actual_cmd: sf role rm -role_id 66dcbd4f-d049-4c6d-bc1d-fc2dc63cd854
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
actual_stdout: Config set: ⏎ created_at: 2026-02-20T10:11:26Z ⏎ key: cli.harness.enabled ⏎ updated_at: 2026-02-20T10:11:26Z ⏎ value: true ⏎
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
name: config get
cmd: sf config get -key cli.harness.enabled
exit: 0
stdout: cli.harness.enabled
actual_cmd: sf config get -key cli.harness.enabled
actual_exit: 0
pass: true
actual_stdout: created_at: 2026-02-20T10:11:26Z ⏎ key: cli.harness.enabled ⏎ updated_at: 2026-02-20T10:11:26Z ⏎ value: true ⏎
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
actual_cmd: sf project rm -project_id 8a6b5c06-d3f8-493a-b48a-f85f34adcd9e
actual_exit: 0
pass: true
actual_stdout: Project deleted ⏎
--------------------------------------------------------------------------------

