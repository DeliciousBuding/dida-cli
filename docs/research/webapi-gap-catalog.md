# Web API Gap Catalog

This document tracks what is already covered in the private Dida365 Web API
surface, what is blocked, and what still needs deeper reverse-engineering.

It complements:

- [api-coverage.md](../api-coverage.md)
- [web-api.md](../web-api.md)
- [api-surfaces.md](api-surfaces.md)

## Coverage Snapshot

The current Web API channel is already broad and is still the widest surface in
the repository.

Major areas already implemented:

- sync and checkpoint
- settings and web-side settings
- task CRUD and advanced task fields
- project, folder, tag CRUD
- comments
- closed history
- search
- statistics
- templates
- sharing metadata
- calendar metadata
- Pomodoro and habit read surfaces

## Confirmed Gaps

These are known private surfaces that exist but are not yet sufficiently
understood or safely wrapped.

### Column Management

- `POST /batch/columnProject`

Known status:

- endpoint is visible in reverse-engineering notes
- payload shapes for update, delete, and reorder are still not verified
- rollback semantics are still unclear

Current implication:

- `column create` can exist as experimental
- full kanban column lifecycle is still incomplete

### Filter Writes

- `POST /batch/filter`

Known status:

- endpoint is visible
- create/update/delete bodies are not mapped
- there is no reliable write-safe command surface yet

### Task Activity Detail Stream

- `GET /task/activity/{taskId}`
- `GET /task/activity/{taskId}?skip=<n>`
- `GET /task/activity/{taskId}?lastId=<id>`

Known status:

- the path is visible in the webapp bundle and uses the legacy v1 client
- optional cursor-like query values `skip` and `lastId` are visible
- v2-style `/api/v1/task/activity/{taskId}` returned 404 when sent through the
  v2 base
- v1 `/task/activity/{taskId}` reached the route but returned `need_pro` on the
  observed account
- 2026-05-10 raw CLI probes against a real task returned HTTP 500 with
  `errorCode=need_pro` in the body snippet for the no-query form, `skip=0`, and
  `skip=0&lastId=`

Current implication:

- task activity remains a read-gap because response fields and cursor semantics
  are not verified
- raw probe JSON now exposes a short `error.details.bodySnippet`, which is
  enough to distinguish this entitlement failure from a path-shape failure
- this should not be promoted into a first-class command until a successful
  Pro-account read or browser-traced request shape is captured

### Attachments

- attachment upload
- attachment association with comments
- comment attachment flow
- task-level attachment upload/download

Known status:

- quota reads are implemented
- 2026-05-10 live `attachment quota` returned a valid quota envelope
- comment attachment paths and create payload shape are mapped in
  [webapi-attachment-flow-notes.md](webapi-attachment-flow-notes.md)
- 2026-05-10 reversible live probe confirmed comment attachment upload with
  multipart field `file`, successful PNG upload response keys, attachment id
  read-back through `comment list`, and cleanup by deleting the disposable
  comment/task
- logical `projectId=inbox` is not valid for upload; use the real inbox/list id
  from `dida agent context --json`
- task-level upload is bundle-mapped as
  `POST /api/v1/attachment/upload/{projectId}/{taskId}/{attachmentId}` with a
  generated local attachment id/refId
- task-level render/download/preview path shape is bundle-mapped as
  `/api/v1/attachment/{projectId}/{taskId}/{attachmentId}` with optional
  `action=download` or `action=preview`
- task-level association/persistence, accepted file matrix, and
  uploaded-but-not-attached cleanup behavior still need live evidence

Current implication:

- comment attachments are implemented because upload, comment payload, read-back,
  and cleanup are live-confirmed
- task-level attachment commands should stay out of the public command surface
  until a reversible trace proves the task mutation semantics

### Collaboration Writes

- invite creation / deletion
- share permission mutation
- multi-user collaboration updates

Known status:

- read-only share metadata is implemented
- write semantics are still not mapped and not verified for rollback

### Trash Pagination

- `GET /project/all/trash/page?...`

Known status:

- live-smoked on 2026-05-10
- `GET /project/all/trash/page` returns the first page and a `next` cursor
- `GET /project/all/trash/page?from=20` returns the next page and a later `next` cursor
- `type=task` returned HTTP 500 and should not be sent

## Confirmed Probe Failures Or Uncertain Surfaces

These are useful to keep visible so later sessions do not waste time repeating
the same dead-end assumptions.

- `GET /project/{id}/data` on the observed CN Web API returned 404
- `GET /project/{id}/columns` returned 404
- `GET /project/{id}` returned 405
- `GET /api/v1/task/activity/{taskId}` returned 404 through the v2 base
- `GET /task/activity/{taskId}` returned HTTP 500 with `need_pro` through the
  v1 base on the observed account
- `GET /project/all/trash/page?type=task` returned 500; use `from=<cursor>` for pagination
- `POST /column` produced responses that looked successful but did not yet prove
  full semantic correctness of the write
- `PUT /tag/merge` can return success while the source tag still remains

## Priority Order

If the goal is to keep deepening the Web API channel, the most valuable next
targets are:

1. task activity detail
2. task-level attachments and attachment download/preview
3. columnProject update/delete/order
4. filter writes
5. collaboration writes

## Documentation Direction

This repo will be easier to maintain if the Web API docs are split more clearly:

- `docs/webapi/overview.md`
- `docs/webapi/coverage-matrix.md`
- `docs/webapi/gaps.md`
- `docs/webapi/probe-log.md`
- `docs/webapi/resources/...`

For now, this file serves as the working gap ledger.
