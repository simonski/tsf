# NOTES

LLMs can:
  abductive reasoning: no
  inductive reasoning: yes
  deductive reasoning: yes

- ROLE&RESPONSIBILITY

- COUNCIL/CONSENSUS

- NO IMPLICIT MEMORY

- ROLE, WORKER, SKILL, STORY, TASK

 Vision → Goal → Initiative → Epic → Story → Task → Subtask

Terms of reference for the work.
WORK is the catchall term for
  REQUIREMENT/SPECIFICATION
  VISION
  GOAL: 
  EPIC
  STORY
  TASK (can have SUBTASK)
  DEFECT/BUG 

Status
  Idle
  Active

Outcomes
  Distance from the acceptance criteria
    meets acceptance criteria
    
  Success? 
  Failure?





The ORCHESTRATOR can be designed MANUALLY!
Just write out the intent and the role and execute it.
Ensure the steps are in sequence so that you can score it.
Then automate that.


Consensus
  All roles in the project should score a project based on their own criteria.
  0 - 1
  Once we have achieved a project config consensus - stddeviation and/or average.

GIT and history
  A pre-state (commit hash)
  A post-state (commit hash)
  can be used independently IF those values are stored
  Then an observer can verify if the work was carried out (VERIFY/TRUST)
  Can then score the difference between versions.

  Allows for building.
RISK
  not all is git
  not all is buildable
  most is a hodge-podge of design and contradictions.

  A role can describe the back/forth options for advice.

  Consensus could be objective or subjective.   Human supplied as an agent as well as virtual.

  A role editor is required.
  A worker editor is required.

Tenets:
  the software is 1% of the effort
  tokens will be burned
  validation is where the whole game is
  decompose work is the whole thing
  we will reduce human labour and achieve a better result
    faster
    more reliable
    repeatable

  the literal token cost is not the concern right now.  the outcome will inform us if it is practical

  the decomposing of work si as challenging as the risk reduction doing of the work
  because decomposing/analysis IS the work






  simplicity / run on laptop
    (call it sf because t is too far for my fingers)
    - building: single binary outcome `make clean build test`
    - running: solves problems quickly
      sf server
      sf worker
      sf <api-call>
    - distributed by design.  can run 1 worker or 100
    - simple to account for thins
      sf explain <work>
        should explain "why" story got to a given place, who said what, what the conversations were and the outcome achieved.

