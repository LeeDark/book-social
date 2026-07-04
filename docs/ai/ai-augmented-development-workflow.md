# AI-Augmented Development Workflow

## Purpose

This document describes a practical workflow for using AI tools in software development
without losing engineering ownership, review discipline, or learning value.

The goal is not to treat AI output as a finished result. The goal is to use AI as a planning,
implementation, review, documentation, and learning assistant while keeping decisions,
verification, and accountability with the developer.

An AI-augmented developer:

- uses AI intentionally for scoped tasks;
- keeps ownership of architecture and code;
- writes clear prompts with enough project context;
- reviews generated changes before accepting them;
- runs local checks when possible;
- documents meaningful decisions;
- uses AI to support learning, not replace practice;
- uses Git as a review and safety checkpoint.

## Scope

This workflow is useful for:

- backend development;
- repository maintenance;
- documentation updates;
- code review;
- testing practice;
- structured learning;
- public project presentation.

This workflow does not replace:

- source control discipline;
- local builds and tests;
- security review;
- domain understanding;
- architectural judgment;
- manual debugging.

## Core Principles

### 1. Ownership Stays With The Developer

AI can suggest, draft, implement, review, and explain.

The developer remains responsible for:

- choosing the problem to solve;
- deciding which solution fits the current project;
- accepting or rejecting generated changes;
- verifying behavior locally;
- keeping documentation accurate;
- understanding the final result.

### 2. AI Output Is A Draft

Generated code, text, tests, or plans should be treated as draft material until reviewed.

AI output becomes useful only when it is:

- understandable;
- scoped;
- consistent with the repository;
- tested or otherwise verified;
- documented when needed;
- committed intentionally.

### 3. Context Improves Output Quality

AI tools work better when the task includes:

- repository purpose;
- current stage;
- relevant files or folders;
- constraints;
- forbidden changes;
- expected output;
- verification steps;
- stop condition.

Vague prompts usually produce vague or oversized changes. Explicit task context reduces
accidental rewrites and makes review easier.

### 4. Small Tasks Are Easier To Review

Prefer tasks that fit one clear goal:

- one feature;
- one bug fix;
- one document;
- one test area;
- one refactor;
- one architectural question.

Avoid broad requests such as:

- improve the whole project;
- refactor everything;
- make it production-ready;
- clean up unrelated code and documentation in one pass.

### 5. Learning Requires Practice

AI can explain concepts and generate exercises, but skill still comes from implementation,
testing, debugging, and review.

Recommended learning loop:

```text
concept -> explanation -> exercise -> implementation -> test -> review -> summary
```

## Standard Workflow

Use this loop for AI-assisted development tasks:

```text
idea
  -> clarify the goal
  -> convert the goal into a scoped task
  -> choose the working mode
  -> provide repository context
  -> execute the task
  -> review the diff
  -> run checks
  -> update documentation when needed
  -> commit focused changes
```

## Working Modes

Different tasks need different operating modes. Stating the mode at the start of a prompt
helps control scope and expected behavior.

### 1. Manager Mode

Formula:

```text
Manager Mode: execute this scoped task step by step and report the result.
```

Use when:

- the task is clear enough to delegate;
- the expected output is concrete;
- implementation or editing support is useful;
- the change should remain small and controlled.

Good for:

- focused implementation tasks;
- documentation updates;
- README cleanup;
- small refactors;
- mechanical repository cleanup;
- moving or renaming files when scope is explicit.

Rules:

- explain the approach before editing when the change is meaningful;
- change only files in scope;
- preserve existing project style;
- run safe relevant checks when possible;
- summarize changed files, checks, and remaining risks;
- do not commit without explicit instruction.

Main risk:

- delegation can reduce learning if used before understanding the task.

### 2. Pair Programmer Mode

Formula:

```text
Pair Programmer Mode: review an existing implementation and suggest targeted fixes.
```

Use when:

- the first version was implemented manually;
- the goal is learning through feedback;
- the task is architecture-sensitive;
- review is more useful than replacement.

Good for:

- Go exercises;
- handlers;
- tests;
- repository logic;
- important first implementations;
- naming and structure feedback.

Rules:

- do not rewrite the whole solution by default;
- review the diff;
- prioritize correctness and project fit;
- suggest minimal changes;
- explain what should be practiced next when useful.

Main value:

- the developer keeps implementation practice while still getting a second review pass.

### 3. Tutor Mode

Formula:

```text
Tutor Mode: explain the concept, guide practice, and avoid writing the final solution first.
```

Use when:

- the topic needs explanation;
- the task is not ready for implementation;
- an exercise or learning path is needed;
- passive copy-paste should be avoided.

Good for:

- Go language topics;
- concurrency concepts;
- web development fundamentals;
- API design;
- testing;
- profiling and observability;
- architecture discussions.

Rules:

- explain the concept clearly;
- propose a small exercise;
- ask checking questions when useful;
- give hints before final answers when appropriate;
- connect theory with practical code.

### 4. Review Mode

Formula:

```text
Review Mode: inspect the change as a reviewer and do not edit files.
```

Use when:

- a diff needs review;
- generated changes need validation;
- the work is close to commit;
- feedback is needed without automatic fixes.

Good for:

- code review;
- test review;
- documentation review;
- architecture review;
- public-facing text review;
- scope review.

Rules:

- list findings first;
- order findings by severity;
- include concrete file and line references when possible;
- separate critical issues from suggestions;
- mention test gaps and residual risk;
- do not rewrite unless explicitly asked.

### 5. Planning Mode

Formula:

```text
Planning Mode: decompose the task into stages without implementation.
```

Use when:

- the task is large;
- the order of work is unclear;
- several approaches are possible;
- a Definition of Done is needed;
- implementation should not start yet.

Good for:

- roadmaps;
- learning tracks;
- version planning;
- project stages;
- selecting the next few tasks.

Rules:

- define the goal and constraints;
- propose stages;
- define outputs and Definition of Done;
- explicitly name what to postpone;
- avoid editing files unless the mode changes.

### 6. Documentation Mode

Formula:

```text
Documentation Mode: update only the requested documentation.
```

Use when:

- the output is a README, plan, roadmap, task history, or repository context file;
- code changes would be distracting;
- written structure is the main deliverable.

Good for:

- README files;
- plan files;
- repository context;
- public project descriptions;
- technical notes;
- task summaries.

Rules:

- touch only requested documents;
- do not change implementation files;
- preserve technical meaning;
- avoid broad cosmetic rewrites;
- keep public and internal documentation clearly separated.

### 7. Book Or Source Study Mode

Formula:

```text
Book Study Mode: understand the source and convert it into small practical tasks.
```

Use when:

- studying a book, article, course, or technical documentation;
- concepts need to be mapped to current practice;
- useful exercises should be separated from irrelevant examples;
- source material should not blindly dictate project architecture.

Good for:

- Go books and documentation;
- web development material;
- API design material;
- testing resources;
- profiling and observability resources.

Rules:

- summarize ideas in original wording;
- avoid copying large source fragments;
- extract key concepts;
- map concepts to current projects;
- classify ideas as apply now, postpone, or ignore;
- switch to Tutor, Pair Programmer, or Manager Mode when implementation starts.

## Default Mode Selection

When no mode is specified, choose the safest interpretation:

- questions and learning: Tutor Mode or Planning Mode;
- inspect this: Review Mode;
- do this: Manager Mode with a short plan first;
- documents: Documentation Mode;
- books, articles, or docs: Book Study Mode;
- ambiguous requests: clarify scope before editing.

## Mode Checklist

Every significant AI-assisted task should answer:

- What mode is this?
- What is the goal?
- What is the scope?
- What must not change?
- What is the expected output?
- How will the result be verified?
- Should the assistant stop before editing or before the next logical step?

## Task Design

A strong AI task includes:

- working mode;
- repository context;
- current state;
- goal;
- files or areas to inspect;
- allowed files or areas;
- forbidden files or areas;
- constraints;
- expected output;
- verification steps;
- documentation expectations;
- stop condition.

Good task design helps prevent:

- unrelated changes;
- oversized diffs;
- accidental rewrites;
- invented requirements;
- documentation that no longer matches the code.

## Implementation Loop

Use this loop for AI-assisted code changes:

```text
prepare task
  -> inspect relevant files
  -> plan the smallest safe change
  -> edit within scope
  -> run formatting/checks/tests
  -> review diff
  -> document behavior changes
  -> commit only after review
```

Useful checks may include:

- `go test ./...`;
- `go test` for a specific package;
- `go vet`;
- formatter checks;
- linter checks when already part of the project;
- manual `curl` checks for APIs;
- browser checks for web interfaces.

If a check cannot be run, record:

- the command attempted;
- why it failed;
- whether the failure was environmental or code-related;
- what should be run locally later.

## Learning Loop

Use this loop for AI-assisted study:

```text
topic
  -> short explanation
  -> small exercise
  -> manual implementation
  -> run or test
  -> review
  -> short summary
```

Learning outputs should include at least one of:

- code exercise;
- test;
- note in original wording;
- reusable checklist;
- comparison of approaches;
- practical example connected to a project.

## Documentation Loop

Use this loop for AI-assisted documentation:

```text
identify audience
  -> define purpose
  -> inspect current files
  -> update only relevant sections
  -> verify commands and claims
  -> review tone and scope
```

Public documentation should:

- explain purpose, status, and usage clearly;
- avoid overclaiming;
- separate planned work from completed work;
- include runnable commands when possible;
- avoid exposing sensitive paths, internal planning, or unrelated project strategy.

## Evaluation

AI assistance should be evaluated by outcomes, not by volume of generated text.

Useful indicators:

- completed scoped tasks;
- reviewed diffs;
- tests or checks run;
- rejected or revised generated changes;
- documentation improved;
- learning exercises completed manually;
- recurring failure patterns identified;
- follow-up tasks reduced rather than multiplied.

The workflow is improving when:

- tasks become smaller and clearer;
- generated diffs become easier to review;
- verification becomes more consistent;
- documentation matches actual behavior;
- fewer unrelated changes appear in a task.

## Tool Usage Guidelines

### Planning And Discussion Assistants

Best for:

- clarifying vague ideas;
- decomposing work;
- architecture discussion;
- learning explanations;
- writing and rewriting;
- preparing implementation tasks;
- summarizing progress.

Avoid using them for:

- final technical authority;
- accepting code without verification;
- security-sensitive decisions without additional review;
- replacing hands-on practice.

### Code Agents

Best for:

- scoped implementation;
- tests;
- small refactors;
- documentation tied to code;
- repository-aware cleanup;
- review support.

Rules:

- provide repository context;
- keep tasks focused;
- review the diff;
- run relevant checks;
- reject code that cannot be explained or maintained.

### Git

Git is the control point for AI-assisted work.

Rules:

- use focused branches when appropriate;
- inspect `git diff`;
- keep commits small;
- avoid mixing unrelated changes;
- do not commit generated artifacts accidentally;
- slow down when a diff becomes too large to review confidently.

## Prompt Templates

### Mode-Based Task Prompt

```text
Mode:
- Manager / Pair Programmer / Tutor / Review / Planning / Documentation / Book Study

Context:
- Project/repository:
- Current state:
- Why this task matters:

Scope:
- Files/areas allowed:
- Files/areas forbidden:
- What must not change:

Task:
- Exact request:

Verification:
- Checks to run:
- What will be verified locally:

Stop condition:
- Stop before editing / after plan / after one logical step / after summary.

Expected output:
- Plan, diff, review, document, exercise, or summary.
```

### Planning Prompt

```text
Context:
- Project:
- Current state:
- Goal:
- Constraints:
- What to avoid:

Task:
Break this into small, realistic development steps.

Expected output:
- recommended sequence;
- first 1-3 tasks;
- risks;
- what to postpone.
```

### Implementation Prompt

```text
Mode:
- Manager Mode

Context:
- Repository:
- Current goal:
- Relevant files/areas:
- Current stage:

Task:
Implement only the following change:

Constraints:
- Do not change unrelated files.
- Do not redesign the architecture.
- Keep the current style.
- Update docs only if directly relevant.

Verification:
- Run the safest relevant tests/checks if possible.
- If the environment prevents a check, explain exactly what failed and what should be run locally.

Expected summary:
- files changed;
- behavior changed;
- tests/checks run;
- risks or follow-up notes.
```

### Review Prompt

```text
Review this change as a backend reviewer.

Focus on:
- correctness;
- simplicity;
- project fit;
- error handling;
- naming;
- test gaps;
- overengineering;
- hidden behavior changes.

Do not rewrite the whole solution.
Give concrete comments and a short verdict.
```

### Pair Programmer Review Prompt

```text
Pair Programmer Mode.

Review this diff without replacing the whole solution.

Focus on:
- correctness;
- simplicity;
- idiomatic code;
- test gaps;
- hidden regressions;
- naming and structure;
- lessons for future practice.

Expected output:
- findings by severity;
- minimal suggested fixes;
- relevant learning notes;
- what to practice next.
```

### Learning Prompt

```text
Topic:
Current level:
Goal:

Please provide:
1. short explanation;
2. common mistakes;
3. small exercise;
4. expected behavior;
5. hints before the full solution.
```

## Anti-Patterns

Avoid:

- using a learning prompt but accepting generated code without practice;
- asking for review and allowing automatic rewrites;
- asking for implementation without defining scope;
- mixing planning, teaching, implementation, refactoring, documentation, and validation in one task;
- asking AI to improve everything;
- accepting large diffs without review;
- creating roadmaps faster than they can be executed;
- treating explanations as proof of understanding;
- polishing documentation while avoiding implementation;
- adding new tools before finishing current tasks;
- trusting broad production-ready claims without verification;
- turning every idea into a permanent document.

## Practical Rule

Use AI to:

- clarify;
- accelerate scoped work;
- review;
- document;
- learn;
- ship verified changes.

Do not use AI to:

- avoid thinking;
- avoid coding;
- avoid debugging;
- avoid responsibility.
