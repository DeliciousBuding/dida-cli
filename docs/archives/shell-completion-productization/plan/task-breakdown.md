# Task Breakdown

## Overview

- Total phases: 1
- Total tasks: 4
- Effort: S

## Phase 1: Shell Completion Productization

| # | Task | Priority | Effort | S.U.P.E.R | Test Expectation | Acceptance Criteria |
|:--|:--|:--|:--|:--|:--|:--|
| 1 | Add CLI tests for completion generation and `--json` rejection | P0 | S | P, E | Go unit tests | Tests fail before implementation and pass after implementation. |
| 2 | Implement `dida completion <shell>` | P0 | S | S, U, E | Go unit tests | Bash, zsh, fish, and PowerShell scripts are generated without auth or network access. |
| 3 | Register help and schema contract | P0 | S | P, R | Schema/docs tests | `dida --help`, `schema list`, and `schema show completion` expose the command. |
| 4 | Update user docs and roadmap | P1 | S | P, R | `TestCommandReferenceMentionsSchemaCommands` | README, Chinese README, command reference, roadmap, and changelog mention completion. |

## Telemetry

| Date | Task | Effort | S.U.P.E.R Score | Unplanned Dependencies | Notes |
|:--|:--|:--|:--|--:|:--|
| 2026-07-07 | Shell completion productization | S | S/U/P/E/R pass | 0 | Local-only feature. No external auth, network, or release tag required. |
