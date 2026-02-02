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
As a WORKER completes its work, it updates the TASK 
One the WORKER completes, the TASK is then Idle and the ROLE is Unassigned.


A TASK and ROLE is then employed by a WORKER, where the taks becomes Active.




A TASK progresses through different STAGES during its lifecycle.

A STAGE will contain STATE of activity:
    Stage: Design
    State: Idle,Active,Success,Failure

Stage: 
    Design, 
    Development, 
    Test, 
    Release
Stage: Idle, ACtive, Succes, Failure

An EPIC contains one or more STORIES.

TASKS are worked on by a WORKER in a ROLE.

A WORKER is a unit of COMPUTE with NO MOTIVATION

A ROLE is the MOTIVATION: "As a Business Analyst..."
A ROLE is the MOTIVATION: "As an Engineer"
A ROLE is the MOTIVATION: "As a Product Owner"
A ROLE is the MOTIVATION: "As a Release Manager"

Each ROLE has its own MOTIVATION which is a set of RULES that it must apply to the TASK that it is working on.

The WORKER is just the unit of compute dedicated to carrying out the TASK in a given ROLE.

Once a WORKER concludes the TASK is complete - success or failure, then the TASK should be returned to the POOL allowing hte ORCHESTRATOR to decide what to do with the TASK.  

It is NOT up to the WORKER to decide what to do next, only to explain the OUTCOME of the work carried out.  

The ORCHESTRATOR then decides what the appropriate next STAGE is.  this means the ORCHESTRATOR could "Send back" the TASK to a previous ROLE, or "send on" the TASK to a subsequent ROLE.   In this way the 
ORCHESTRATOR may descide "ok, engineering role is complete, send to qa".   But it may decide "ok, engineering failed complaining it is not clear, send back to the BA".

In that manner the ORCHESTRATOR would then set the TARGET to be a particular ROLE and the STATUS to be IDLE, meaning the TASK is then in an "idle" pool with waiting to be assigned or claimed by a WORKER who is in the role previously assigned.