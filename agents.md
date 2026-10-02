# AGENTS.md — myPersonalTechRadar

These rules apply to every AI agent working in this repository.

The project is a personal technology intelligence platform that discovers,
understands, filters, and personalizes technology information for the user.

The repository is also a learning project. The agent must optimize for both:
1. building a clean, maintainable system, and
2. helping the user understand the engineering decisions and code.

---

# 1. Highest Priority

## Plan Before Implementation

Never jump directly from a request to code.

For every non-trivial task:

1. Understand the request.
2. Inspect the relevant repository files.
3. Identify the existing architecture and constraints.
4. Explain the proposed approach to the user.
5. Produce a concrete implementation plan.
6. Wait for explicit user approval before making implementation changes.
7. After approval, break the plan into ordered subtasks.
8. Complete subtasks sequentially.
9. Verify each subtask before moving to the next.

Do not treat a task as too small to plan if it changes architecture,
data models, APIs, persistence, or project structure.

If the correct approach is unclear, ask the user instead of guessing.

---

# 2. Single-Agent Rule

Work as a single agent.

Do not create:
- worker agents
- sub-agents
- parallel implementation agents
- delegated review agents

All exploration, planning, implementation, testing, debugging,
and review must happen in the main agent session.

---

# 3. Learning-First Development

This is not a code-generation-only project.

The user is intentionally using this project to learn:

- Go
- backend engineering
- PostgreSQL
- networking
- concurrency
- workers
- Docker
- Kubernetes
- distributed systems
- system design
- system architecture
- observability
- production engineering

Therefore:

## Explain Before Abstracting

When introducing a new concept, briefly explain:

- what it is
- why this project needs it
- where it fits
- what problem it solves

Do not dump large amounts of code without explanation.

Prefer:

    concept
      ↓
    design
      ↓
    small implementation
      ↓
    run it
      ↓
    verify it
      ↓
    next step

over generating the entire application at once.

## Do Not Hide Complexity

If something is intentionally simplified for V0, say so.

Example:

"V0 uses PostgreSQL full-text search. We can introduce pgvector later
when semantic search becomes necessary."

The user should understand what is being built and why.

---

# 4. Project Identity

Project name:

    myPersonalTechRadar

The project is a personal technology intelligence system.

Its purpose is NOT to become a generic news aggregator.

The system should eventually:

1. discover high-quality technology information
2. normalize different content types
3. remove duplicates
4. understand the content
5. identify topics
6. determine personal relevance
7. explain why content was recommended
8. learn from user feedback
9. identify longer-term technology trends
10. connect discoveries to the user's learning roadmap

The system should optimize for:

    useful information
    over
    maximum information

and:

    attention protection
    over
    engagement maximization

---

# 5. V0 Scope Discipline

V0 must remain intentionally small.

The initial system should focus on:

- source ingestion
- content normalization
- basic deduplication
- topic classification
- basic summaries
- explicit user interests
- learning roadmap
- read/save/dismiss interactions
- basic relevance ranking
- simple web feed

Do NOT introduce the following unless explicitly approved:

- microservices
- Kafka
- event buses
- complex agent frameworks
- Kubernetes
- distributed databases
- multi-region infrastructure
- billing
- subscriptions
- multi-tenancy
- social features
- notifications
- mobile applications
- Chrome extensions
- WhatsApp integrations
- complex recommendation models
- large-scale vector infrastructure
- unnecessary abstractions

Start simple.

Complexity must be justified by an actual requirement.

---

# 6. Architecture

Use a monorepo.

Initial structure:

    myPersonalTechRadar/
    ├── apps/
    │   ├── api/
    │   ├── worker/
    │   └── web/
    │
    ├── internal/
    │   ├── database/
    │   ├── content/
    │   ├── sources/
    │   ├── topics/
    │   ├── interests/
    │   └── roadmap/
    │
    ├── migrations/
    ├── deployments/
    ├── docs/
    ├── go.mod
    ├── docker-compose.yml
    ├── Dockerfile
    └── README.md

This structure may evolve as the project grows.

Do not reorganize the repository without a concrete reason.

---

# 7. Backend

The backend is written in Go.

Prefer the Go standard library when it is sufficient.

Do not introduce a framework simply because one exists.

If a framework such as Gin is proposed:

1. explain why it is useful
2. compare it with the standard library for the current requirement
3. get user approval before introducing it

Backend applications:

    apps/api
    apps/worker

The API and worker are separate executables but remain inside the same
repository.

Do not turn them into separate repositories.

---

# 8. Frontend

The web application is part of the same monorepo.

The frontend should remain intentionally lightweight.

Do not introduce a large frontend architecture solely for the sake of
using a popular framework.

The frontend exists to provide the product experience:

- curated feed
- item details
- save
- dismiss
- topic filtering
- roadmap
- personalization explanations

Frontend technology decisions must be made separately rather than
assuming a framework in advance.

---

# 9. Database

Primary database:

    PostgreSQL

Database schema must be designed before implementing database-dependent
application logic.

Follow this sequence:

    Domain model
        ↓
    Relationships
        ↓
    Database schema
        ↓
    Migration
        ↓
    Go database access
        ↓
    Application logic

Do not randomly create tables while implementing features.

Every table must have a clear domain purpose.

Prefer PostgreSQL-native features where appropriate.

Use:

- primary keys
- foreign keys
- unique constraints
- appropriate indexes
- timestamps
- explicit relationships

Avoid premature database complexity.

---

# 10. Initial Domain Model

The initial V0 domain contains:

- Source
- ContentItem
- Topic
- ContentItemTopic
- Interest
- RoadmapItem
- UserItemInteraction

The model must remain generic enough to support different content types:

- article
- video
- GitHub release
- conference talk
- research paper
- discussion

Do not model the system exclusively around "articles".

---

# 11. API Design

API design must follow the domain model.

Do not create endpoints before understanding the underlying data model.

Initial API direction may include:

    GET  /api/health
    GET  /api/feed
    GET  /api/items/:id

and later:

    POST /api/items/:id/read
    POST /api/items/:id/save
    POST /api/items/:id/dismiss

These are examples, not immutable requirements.

API changes must be discussed before implementation.

---

# 12. Data Ingestion

Initial sources should prioritize high-quality and original sources.

Examples:

- official engineering blogs
- official project blogs
- GitHub releases
- selected technical communities
- RSS feeds

Community sources are useful for discovery and discussion but should not
automatically be treated as authoritative sources.

Do not build every possible integration in V0.

Add sources incrementally.

---

# 13. Personalization

The system must distinguish:

## Topic

What a piece of content is about.

Example:

    Kubernetes

## Interest

How much the user cares about that topic.

Example:

    Kubernetes → 1.0

## Roadmap

What the user is currently learning or plans to learn.

Example:

    Kubernetes → CURRENT
    Rust → NEXT

These are different concepts and must not be collapsed into one field.

---

# 14. Explainable Recommendations

Personalization should eventually answer:

> Why was this shown to me?

Examples:

    Matches your Kubernetes interest
    Relevant to your current infrastructure roadmap
    From a high-quality primary source
    New information
    Similar to content you previously saved

Do not build a black-box recommendation system when a simple,
explainable scoring model is sufficient.

---

# 15. AI Usage

AI should be introduced only where it provides meaningful value.

Potential V0 AI tasks:

- summarization
- topic extraction
- basic classification
- "why it matters" generation

Do not introduce autonomous agents merely because the project is
AI-related.

AI-generated information must be treated as derived data, not unquestioned
truth.

Where practical, preserve the original source and allow the user to inspect
it.

---

# 16. Deduplication

Content may appear through multiple sources.

V0 can use simple techniques such as:

- canonical URLs
- normalized URLs
- content hashes
- title/source/date comparison

Later versions may introduce semantic similarity or embeddings.

Do not introduce vector databases or embeddings until the simpler
approaches are insufficient.

---

# 17. Code Quality

Prefer:

- simple code
- explicit dependencies
- small functions
- clear naming
- idiomatic Go
- minimal abstractions
- readable SQL
- deterministic behavior

Avoid:

- clever abstractions
- unnecessary interfaces
- premature generics
- unnecessary design patterns
- speculative infrastructure
- framework-driven architecture

Do not optimize code before there is a measured problem.

---

# 18. Go Standards

Write idiomatic Go.

Prefer:

- standard library
- explicit error handling
- small packages
- context propagation
- clear ownership of dependencies
- meaningful errors
- table-driven tests where appropriate

Do not create interfaces solely for mocking.

Introduce an interface when there is a real abstraction boundary,
multiple implementations, or a clear testing benefit.

---

# 19. Error Handling

Never silently ignore errors.

Bad:

    result, _ := operation()

Prefer:

    result, err := operation()
    if err != nil {
        return err
    }

Errors should contain enough context to diagnose the failure.

Do not expose internal implementation details or secrets through API
responses.

---

# 20. Configuration and Secrets

Secrets must never be committed.

Never commit:

- API keys
- database passwords
- tokens
- credentials
- private keys

Use environment variables or an appropriate local configuration
mechanism.

Provide safe examples through files such as:

    .env.example

Never put real credentials in `.env.example`.

---

# 21. Testing

Every meaningful feature should have appropriate tests.

Prioritize:

- domain logic
- ranking logic
- deduplication
- parsing
- database behavior where practical
- API behavior

Do not write meaningless tests merely to increase coverage.

After each implementation subtask, run the relevant checks.

At minimum, for Go changes where applicable:

    gofmt
    go test ./...
    go vet ./...

If a check cannot be run, state why.

---

# 22. Verification

Never assume code works because it compiles.

Verification should include the smallest useful real-world check.

Example:

    Build API
       ↓
    Start API
       ↓
    Request /api/health
       ↓
    Verify response

For ingestion:

    Start worker
       ↓
    Fetch source
       ↓
    Store content
       ↓
    Query database
       ↓
    Verify duplicate handling

---

# 23. Documentation

Document decisions that affect the architecture or product.

Do not document obvious code line-by-line.

Prefer documentation that explains:

- why a technology was selected
- why an architectural decision was made
- why a simpler approach was rejected
- important domain assumptions
- known limitations

---

# 24. Repository Memory

Maintain project memory under:

    .ai/

Structure:

    .ai/
    ├── decisions/
    │   └── <topic>/
    │       └── decision.md
    │
    └── sessions/
        └── <topic>/
            └── <session-file>.md

## Before Every Session

Read only the relevant existing memory:

1. relevant `.ai/decisions/` topic
2. relevant recent `.ai/sessions/` topic

Do not load the entire memory tree unnecessarily.

## Decisions

`.ai/decisions/` stores durable project decisions.

Record:

- the decision
- why it was made
- important constraints
- alternatives considered when useful

Do not store implementation details that belong in code.

## Sessions

`.ai/sessions/` stores AI work history.

Create a new session file for every working session.

Never append to an old session file.

Record:

- task
- plan
- work completed
- decisions confirmed by the user
- tests/checks performed
- unresolved issues
- next steps

Never invent historical information.

If something is unknown, ask the user.

---

# 25. User Approval

The user must explicitly approve implementation after the plan.

Examples of approval:

    "go ahead"
    "build it"
    "implement this"
    "proceed"

Questions, brainstorming, architecture discussion, or requests for an
explanation are NOT implementation approval.

Do not modify project files during the planning stage unless the user
explicitly asks for a documentation-only change.

---

# 26. Sequential Development

Work in small milestones.

Example:

    Milestone 1
    Go API foundation
        ↓
    verify
        ↓
    Milestone 2
    PostgreSQL connection
        ↓
    verify
        ↓
    Milestone 3
    Initial schema
        ↓
    verify
        ↓
    Milestone 4
    Source ingestion
        ↓
    verify

Do not implement multiple unrelated milestones at once.

---

# 27. No Guessing

Never invent:

- requirements
- APIs
- database relationships
- business rules
- source behavior
- user preferences
- architecture decisions

If the repository, documentation, or user has not established something,
ask.

When making a reasonable technical assumption, explicitly state it.

---

# 28. Scope Control

Before adding a dependency, service, table, abstraction, or infrastructure
component, ask:

1. What problem does this solve?
2. Do we have that problem now?
3. Is there a simpler solution?
4. Does it help V0?
5. What complexity does it introduce?

If the answer is unclear, do not add it yet.

---

# 29. Definition of Done

A task is not complete merely because code was written.

A subtask is complete when:

- implementation is finished
- relevant tests/checks pass
- formatting is correct
- behavior is verified
- documentation is updated when necessary
- relevant project memory is updated
- remaining limitations are known

---

# 30. Current Development Philosophy

Build the simplest system that can teach and validate the next important
engineering concept.

The project should evolve approximately as:

    Go
      ↓
    HTTP
      ↓
    PostgreSQL
      ↓
    Data modeling
      ↓
    Ingestion
      ↓
    Workers
      ↓
    AI enrichment
      ↓
    Personalization
      ↓
    Docker
      ↓
    Observability
      ↓
    Kubernetes
      ↓
    Distributed systems
      ↓
    Advanced ranking / semantic search

Do not skip directly to the final architecture.

The goal is to build a useful product while progressively learning the
engineering underneath it.