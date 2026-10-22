# Goals

These seven goals say what hive is for. An agent uses them to resolve a judgment
call that no rule covers.

1. **One home coordinates coding agents across all projects of the user**, in one
   herdr session on one machine.
2. **The user talks to one coordinator.** A manager owns each project. Workers do
   one task each in their own checkout.
3. **Work outlives agents.** Tasks, roles, decisions, memory, and handoffs live in
   durable records, never only in a conversation.
4. **Hive builds hive.** The checkout is the home and the first registered project.
5. **Everything is observable.** Every action is a structured log event. The user
   can view the events live and query them later.
6. **Nothing lands unverified.** An independent agent verifies a task before
   delivery. Delivery follows a standing grant or an explicit approval.
7. **Right-sized models.** A strong model plans and verifies. A decent model
   codes. A small model does small tasks.

## Definition of v1

v1 is complete when all of these are true on the machine of the user:

- The coordinator runs in the hive session, registers this checkout as the hive
  project, and spawns the hive manager.
- The hive manager plans a change to hive, spawns workers in task checkouts,
  gets an independent verification, and lands the change under a grant.
- The TUI shows the live log while this happens.
- Every step of it is in `.hive/logs/`.

## Non-goals of v1

- Hive does not serve more than one user, one machine, or one herdr session for
  each home.
- Hive does not release binaries, update itself, or promise compatibility to
  other installations. The source is public. The tool is personal.
- Hive does not enforce grants in the backend. A grant is instruction policy.
