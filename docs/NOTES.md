NOTES.md

The CLI tool is supposed to be two things

1. Adminsitrator tool
2. User fast access tool (by human or agent or script)

Administrator tool

This will probably be a human calling user management, quota, config style.  It will be low-traffic.  The interface shoudl be simple, pretty, low-friction.

User Tool

A power user or agent will use this to assemble and interact wiht projects.  in this case I want it to be super convenient, low friction, easy to use and understand when things go wrong or the user has questions.  This means it needs to balance simpel and detailed, quick and powerful.  

It should have an implicit/explicit form and a convention over configuration opinionated design - meaning for example we shoudl take a goal like "make. project and add some tasks to it, then shake it to assemble into a sensible arrangment of tasks, then ensure those tasks are grouped by some aribtrary - UX, backed, database", then prioritise them, then apply guardrails around the UX and hten some guardrails arond the Server, and here are some adhoc rules for the datbase, and extract those rules, or refer to the general purpose rules for DB and overwrite them where necessary, oh and now summarise all of that in a coherent way - and now show me what a givne prompt ACTUALLY looks like for a specific task."

