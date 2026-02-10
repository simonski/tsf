DESIGN.md

## Introduction

`sf` is an Agentic Software Factory experiment.  A SPECIFICATION acts as INPUT, software is the OUTPUT.

The approach is to 
    - decompose work into smaller pieces
    - allocate that work to a series of agents with different objectives (analysis, architecture, engineering, testing, et cetera)
    - appraise the quality work via consensus
    - introduce a learning feedback loop

## Documents

@docs/ESSAY.md, 
@docs/DESIGN_COMMON.md, 
@docs/DESIGN_CLI.md, 
@docs/DESIGN_DATABASE.md, 
@docs/DESIGN_FRONTEND.md, 
@docs/DESIGN_ORCHESTRATOR.md, 
@docs/DESIGN_SERVER.md, 
<!-- // @docs/DESIGN_TUI.md,  -->
@docs/DESIGN_WORKER.md, 
@docs/ENTITY.md
@docs/ENTITY_PROJECT.md
@docs/ENTITY_TASK.md
@docs/ENTITY_ROLE.md


## Breakdown of Work

A unit of WORK is expected to proceed through a LIFECYCLE which goes from

    INPUT -> SYSTEM -> OUTPUT

That is, there is an 

    IDEA -> DESIGN -> IMPLEMENTATION -> TESTING -> OUTPUT

Where each discrete step occurs back-and-forth in a sequence of iteration.  

CONSENSUS is the vital ingredient where a VISION (a very high level statement with a desired outcome) is decomposed into GOALS or OBJECTIVES.  These GOALS or OBJECTIVES are further decomposed to OKRs, KPIs, EPICS, TASKs, subdividing as much as is necessary so that:

- the concept is understood
- the outcome can be agreed as SUCCESS or FAILURE (meets or does not meet the objective)

Each action is carried out by an AGENT (A software process run in an LLM) which takes the prior OUTPUT as the INPUT along with their ROLE.  

Examples of WORK could be:
    - writing a specification
    - fixing a bug
    - implementing a requireemnt
    - breaking down a requirement to EPICS 
    - breaking down EPICs to STORIES
    - clarifying a requirement with a a Product Owner
    - writing documentation
    - running User Acceptance Testing
    - ensuring Acceptance Criteria are applied across all stories
    - ensuring software engineering standards are maintained

The idea is that an iterative engineering chain of responsibility is created - a software factory - where the goal is to create software solutions based on english language descriptions.

CONSENSUS based ITERATION

In this way the iteration is NONLINEAR and this is VERY IMPORTANT to understand - you cannot predict which ROLE will next receive WORK once it is freed.   All roles in the workflow
will ultimately need to achieve consensus - which is to say until everyone agrees, the WORK is not complete.

## Technology Components

- FRONTEND browser frontend (DESIGN_FRONTEND.md) (lightweight css/js/html website that renders a kanban board of work).  embedded in go:embed single binary
- CLI go terminal frontend (DESIGN_CLI.md) (go client binary TUI application) permits users to manage tickets from terminal and agents to use the terminal to access the work
- SERVER backend (DESIGN_SERVER.md) (go server binary) - headless set of APIs to manage the task state and persist to the DATABASE
- ORCHESTRATOR backend (DESIGN_ORCHESTRATOR.md) (go server binary) - this process continually queries the server task status and decides on the next step of a given task - where it should go, what actions should be taken.  It is an automated process that will read the tasks and infer what the appropriate action should be.
- DATABASE (DESIGN_DATABASE) a sqlite database that retains all state.  accessible only via the SERVER
- WORKER a software process that requests/receives WORK from the server and performs the work (normally delegating to an LLM)

Note: the entire output is a single go binary, "sf" which contains all components - the FRONTEND, CLI and SERVER, the ORCHESTRATOR, it can then be used in any mode the user wants.

The SERVER serves both the FRONTEND and the APIs.

The APIs are restful openAPI spec compatible.

## Entities

TASK - @docs/ENTITY_TASK.md
PROJECT - @docs/ENTITY_PROJECT.md
ROLE - @docs/ENTITY_ROLE.md

The client can default to a project by setting PROJECT_ID as an envinronment variable, or supply `-project_id` as an option.  The default project id is `default`

## Users/Authentication

A USERS table should exist that contains
    username
    type (human, worker, orchestrator)
    password

All requests to the SERVER must be authenticated using Basic Auth.

A human can login using Passkeys if they wish.
An agent can authenticate using JWT or username/password.

