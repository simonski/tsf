## Introduction

2026-02-15/16

Can a system of agents replace the traditional SDLC? should it? - The Software Factory (tsf) is an approach to answer this question.

The GOAL is to automate and augment the current SDLC process where we expect to improve productivity. The desired outcomes are:

  - programmers will operate at a higher level of abstraction
  - the number of programmers in the loop can be reduced for the same or greater amounts of work
  - the number of agents in the loop is increased
  - the quality of the work either is maintained or improves
  - further understanding of agentic engineering practices
  - data to understand represent cost-value in software engineering

## Key challenges

- Trust/Hallucinations

How do we establish trust if the human is taken out of the loop? How do we trust a computer will both create the code and tell us that the code is correct? What do we do about hallucinations?

- Repeatability

How can we demonstrate a given requirement will result in an expected outcome, repeatably?

- Provability

How can we "show our workings" to demonstrate WHY a given outcome occurred?

TSF approach is to model the roles and processes we have now and automate them.  Applying this process across the software engineering stack is the experiment this approach is attempting to prove.

Insights
--------

- Software Development is Nonlinear
- Teams are fallible and inconsistent
- Consensus in teams is what drives progression
- Work is carried out by a Role
- There is no implicit memory
- Work is a living document
- A council is required to move a piece of work
- LLMs are the critical dependency
- What is success?
- Evolution

## Software Development is Nonlinear

I *thought* that software went from IDEA through PROCESS to REALITY.   Zooming in we applied different roles, lifecycles and methods of describing for example with roles:

   IDEA -> Product Manager -> 
   Product Owner -> 
   Business Analyst -> 
   Architect -> 
   Tech Lead -> 
   Engineer -> 
   QA -> 
   Release Manager -> REALITY

![](1.png)

That is, the job more-or-less is like a baton, handed on to the next in line.  Obviously this is not true (jobs were handed back and forth) however this was my original mental model.

In fact what happens is everyone talks to everyone - sometimes.  The Engineer clarifies a point with the QA who talks to the BA and the release manager coordinates with the Engineering Director.

![](4.png)

## Teams are fallible and inconsistent

A Product Owner contradicts a prior feature, changing the acceptance criteria on the fly.

The engineer has a bad day, writes a bug which slips through to release because then QA is on vacation.

Four of the team members get together in an adhoc meeting to decide what to do about some ambiguity of design or some impediment that has been discovered, decide on an outcome and don't rewrite the specification.

An automated system *should not* be inconsistent in the manner above.  The prior examples *should* be addressed through modelling.  

It is true that a given LLM will hallucinate and/or output poor quality however the purpose of a COUNCIL is to assess the outcome and decide on the next best step.  

This means task outcomes that result in poor quality are both accepted as a risk and mitigated through the council themselves.

In fact we are not attempting to change a foundational model here - hallucinations are actually baked into the system of LLMs.  Indeed we accept this as a risk and then mitigate it via the decomposition, council and consensus.

## Consensus in teams is what drives progression

What happens is some roles get together in a little council and decide to agree or disagree on all things - the definition of the job, the outcome of work being successful or not.  

This happens in micro-meetings with perhaps some established processes (set ticket FOO-123 status to "in test") - but it also happens adhoc in methods we have not properly captured.

![](2.png)

The progression of any work is as a result of consensus.   Where work proceeds with consensus of 1 (That is, the engineer says: it works, trust me), is where there is high risk.

Where there is NO consensus, escalation occurs. It is possible we will learn that some outcomes cannot be decided autonomously - the system will then escalate to the HUMAN in the loop.  This is not strictly a failure - rather an outcome we will learn from.   This escalation may inform higher quality role descriptions, yielding fewer escalations.

## Work is carried out by a Role

All work (ticket, design) is carried out by a discretely named ROLE - for example a Business Analyst.   

The outcome of that work is then received by other ROLEs with different motivations - those roles will implicitly or explicitly appraise the quality of work, using or rejecting it.

This is as true for a written document as it is for a unit test or a video game.

This means any piece of work would have a history of effort and outcome for all the roles that "worked" on the work.

## There is no implicit memory

There can be no place for tribal knowledge in an automated system.  

Memory in a company is "institutional" - stored in the employees.  The challenge is that the HUMAN currently IS the ROLE - meaning they are not differentiated effectively, so "tribal knowledge" remains in the head of a specific human.

This means there is no such thing as a "QA Engineer" - rather there is a "QA Engineer called Sally".

Here is the conundrum - Sally has years of experience and value which are what contributes to the high-quality outcome of completed tasks with rigour.  Sally is the memory.  Sally "knows" the different ROLES to assume or to go find in the company to achieve the outcome.

This "institutional" memory needs to be modelled within our system:

A ROLE containing as much of the institutional memory as possible - describes the job function, motivation, objectives, boundaries and relationships between other ROLES. The expectation of the performance, inputs and outputs.

SKILLs - reusable technology capabilities that a ROLE calls upon to achieve a given objective. 

STORY - the objective itself.  A task contains the objective itself, its history, which roles did-what to it.   The data model for a task then is significant as it will contain sections that are effectively a living document of change.

The theory is the state - the memory - is the aggregation of ROLES, SKILLS and a STORY.

## Work is a living document

For any piece of work then, it will have lineage - the effort taken, by who and when, the outcome.  This state represents the "history of the story" and can be useful informing future decisions.

Caution will be needed as there is no upperbound to the history of a given problem - so using contemporary LLMs we will need to address this with regards to the relevant context sizes.   My assumption is that as a story itself presents its own challenge, another task is then created to "prepare this story for the role" which may act as a summarisation agent.

## A council is required to move a piece of work

As WORK is completed by a ROLE, the outcome will then be assessed by a COUNCIL.   This is effectively another role - a role "of" roles, which acts as the motivation to decide what is the appropriate next step.

It is NOT that the work is DONE, it is more that the work of the ROLE is finished for now and the decision to progress the work in any direction is the result of the council decision.

A story itself will be part of an EPIC, so there are hints in the story itself as to what the intent is.  Additionally, an initial WORKFLOW structure will be provided by the OPERATOR which will act as advice on the available roles and various acceptance criteria that must be met.

This means that the outcome of a given piece of work by a given ROLE is just some claimed state - "I'm finished".  But really the proof and quality of work is decided by the recipient - the COUNCIL.

So really work looks like this

ROLE1+WORK -> COMPLETED -> COUNCIL -> OK -> ROLE2+WORK

ROLE1+WORK -> COMPLETED -> COUNCIL -> NOT OK -> ROLE3+WORK

Until the COUNCIL decides the work is complete.   The other thing is a COUNCIL is an appropriate set of ROLES in any given moment - so it is not a fixed decision making process - different roles and motivations will apply.

## LLMs are the critical dependency

Critical to the whole endeavour then is use of large language models.  The decomposition of work into ROLES, SKILLS, TASKs and the nonlinearity of the work means I expect we will see in the first outcomes:

- high token usage (back-and-forth)
- a ratio of consensus:code to be high:low (delivers trusted outcomes)
- many roles and role compositions

The introduction of evolutionary roles in a subsequent phase may then allow for novel role and skills descriptions however I think they are beyond the scope of our V1.

The use of LLMs then ARE the whole decision making process.  This endeavour is about strictly controlling the entire context for any given prompt by modelling roles, skills, tasks.   

## What is success?

Referring to the question at the top of the document, the outcomes will be directional:

- Can a requirement go from input to working code without human intervention mid-process?
- Does the council mechanism catch defects that a single-agent system would miss?
- What's the token cost per story point (or per unit of output) compared to a human team's cost?
- Does the living document provide sufficient auditability that a human reviewer can understand why decisions were made?

## Cost

- Removing linearity creates the risk of cost increase; to mitigate this we will introduce a circuit breaker and cost agent who will be part of the council.

## Evolution

An SDLC process is not cast in stone - it changes.  The method of change is based on the outcomes, objectives, learning, experimentation, prior experience.   

We will model this adaptive system with:

- a ticket tracking system
- the history of the work
- separate roles with their objectives

In V1, we won't address evolution but prepare for it.  Once we have built up a corpus of tasks and their outcomes, we will then be able to use this as the learning material to see if we can create new ROLE descriptions or SKILLs that will yield a better outcome.

## Conclusion

This document proposed an approach to automate SDLC. The challenges of Trust, Repeatability and Provability are addressed through the decomposition of ROLEs, TASKs and COUNCILs.

The first version, tsf-1, will attempt to implement a reference factory and stop as it arrives at evolutionary roles.   This means tsf1 will be

- a ticket engine
- tickets are living documents
- a CLI to interact
- a website to interface and view work
- a fleet of workers 


![](3.png)

----

## Notes on Bootstrapping

In the first iterations of tsf, a HUMAN operator will need to describe the ROLEs, SKILLS, conditions under which a COUNCIL operates, so as to bootstrap the system - as well as provide requirements.

This means the first version of the system will need simple to use tools (CLI, Website) to allow an operator to define, refine and experiment.

Consensus loosely can be described as "all ROLES in the current state of the work accept the outcome meets their objectives".   This assumes each ROLE is of a high enough quality that we are not collapsing in a house of cards of false positives.  Hence iteration in bootstrapping.

