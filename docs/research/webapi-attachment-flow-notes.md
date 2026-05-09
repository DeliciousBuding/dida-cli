# Web API Attachment Flow Notes

This note records the currently mapped attachment flow from the saved webapp
bundle and reversible live probes. It is not a full media command spec yet.

## Confirmed Read Surfaces

- `GET /attachment/isUnderQuota`
- `GET /attachment/dailyLimit`

DidaCLI already exposes these through `dida attachment quota --json`.

## Comment Attachment Flow

Bundle evidence shows comment image/file attachment handling uses the legacy v1
host.

Observed paths:

```text
POST /attachment/upload/comment/{projectId}/{taskId}
GET /attachment/comment/{projectId}/{taskId}/{attachmentId}
```

Observed client behavior:

- Before upload, the webapp checks attachment quota/daily limit.
- The UI uploads a local file to
  `/attachment/upload/comment/{projectId}/{taskId}`.
- The upload response returns an id that the UI stores as `pathId`.
- When creating a new comment, the UI appends:

```json
{
  "attachments": [
    {
      "id": "<pathId>"
    }
  ]
}
```

- Existing comment attachments are displayed through
  `/attachment/comment/{projectId}/{taskId}/{attachmentId}`.

Reversible live probe evidence from 2026-05-10 confirmed the comment attachment
write path:

- Upload endpoint:
  `POST /api/v1/attachment/upload/comment/{projectId}/{taskId}`
- The `projectId` must be the real inbox/list id. The logical `inbox` project
  alias returned `task_not_found` during upload.
- Multipart field name: `file`.
- A PNG probe with `Content-Type: image/png` uploaded successfully.
- A plain `.txt` probe returned `unsupported_file_type`.
- Successful upload response included these keys:
  `id`, `refId`, `path`, `size`, `fileName`, `fileType`, `createdTime`.
- Comment create accepted this attachment payload:

```json
{
  "attachments": [
    {
      "id": "<uploaded attachment id>"
    }
  ]
}
```

- `comment list` read-back included the uploaded attachment id.
- Cleanup was verified by deleting the created comment and disposable task.

## Task Attachment Flow

Bundle evidence also shows task-level attachments use a separate flow from
comment attachments.

Observed task attachment upload path:

```text
POST /api/v1/attachment/upload/{projectId}/{taskId}/{attachmentId}
```

Observed local pre-upload metadata:

- `id`: generated local attachment id
- `refId`: same generated id
- `projectId`: task project id
- `taskId`: task id
- `fileName`: local file name
- `fileType`: webapp file type classification from extension or MIME type
- `size`: local file size
- `isLocal`: `true`
- `previewUrl`: local object URL while upload is pending

Observed upload response handling:

- The server response includes a `refId`.
- The webapp matches uploaded attachments by `refId`.
- The response object is merged back into the matching local attachment.
- Response fields `id` and `taskId` are removed before the merge.
- After merge, `isLocal` becomes `false` and transient render state is removed.
- If a failed or inactive attachment has no server path, the webapp preserves
  inactive status and marks the object as updated for sync.
- An `exceed_quota` upload failure disables upload capability and triggers
  failure cleanup events.

Observed task attachment render/download paths:

```text
GET /api/v1/attachment/{projectId}/{taskId}/{attachmentId}
GET /api/v1/attachment/{projectId}/{taskId}/{attachmentId}?action=download
GET /api/v1/attachment/{projectId}/{taskId}/{attachmentId}?action=preview
```

Image display paths append the lower-cased file extension to the attachment
path. PDF preview uses `action=preview`. Download uses `action=download` and
expects a binary response.

Countdown-specific attachment paths are separate from task attachments:

```text
POST /api/v1/attachment/upload/countdown?attachmentId={attachmentId}
GET /api/v1/attachment/countdown?id={id}&attachmentId={attachmentId}
```

Current confidence:

- Upload path, local metadata model, response merge behavior, and
  render/download path shape are bundle-mapped.
- Live creation/association is not yet verified.
- The task write payload or sync mutation that persists the uploaded attachment
  on a task still needs a reversible trace.

## Remaining Gaps

- Full accepted file type matrix and size failure responses.
- Task-level attachment creation and association payloads.
- Whether task-level attachment persistence is only a `/batch/task`
  `attachments` mutation or requires additional batch semantics.
- Whether download/preview paths need any extra query parameters for non-image
  files or Pro-gated file types.
- Delete behavior for uploaded-but-not-attached files.
- Whether comment attachment ids are reusable across comment create/update.

## Implementation Gate

Comment attachment upload/create is eligible for first-class CLI support because
the reversible live test recorded:

1. upload request method, headers, and multipart field names
2. upload response fields
3. comment create payload with attachment id
4. read-back through comment list
5. cleanup or rollback behavior

Do not add task-level attachment download, preview, or task attachment mutation
commands until those flows have the same level of evidence.
