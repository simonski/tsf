ENTITY.md

PROJECT = @docs/ENTITY_PROJECT.md
ROILE = @docs/ENTITY_ROLE.md
TASK = @docs/ENTITY_TASK.md
ORCHESTRATOR = @docs/DESIGN_ORCHESTRATOR.md
WORKER = @docs/DESIGN_WORKER.md

Introduction
----------

This document describes the responsibilities of the ORCHESTRATOR, WORKER and how they interact with TASK and ROLE

Motivation
----------

Task
----
A TASK is a unit of work.
A TASK has a ROLE assigned by the ORCHESTRATOR
A TASK status either Idle or Active
    Idle: the TASK is awaiting further work decided on by the ORCHESTRATOR.
    Active: the TASK is currently being worked on by an assigned WORKER
A TASK is either Open or Closed
    Open: The TASK is not yet finished - it is either in progress or pending future work.
    Closed: no more work is required on this TASK.  It is completely finished, the status will be Idle.

Assigning
---------

The ORCHESTRATOR monitors all TASKS and decides for each TASK on the OUTCOME by assessing the HISTORY of the TASK.

The decision it makes is 
    Assignment: which ROLE to assign to it to carry out work.
    Allocation: which WORKER to assign it to.

Working
-------

As a WORKER starts the work, it updates the TASK to be Active
As a WORKER completes its work, it updates the TASK to be Idle

One the WORKER completes, the TASK is then Idle and the ROLE is Unassigned.

The OUTCOME of the work is stored in the history of the TASK.  The ORCHESTRATOR is then able to look at the HISTORY (which is all of the OUTCOMEs the TASK has had) to make a decision as to what the next step is for this TASK.

