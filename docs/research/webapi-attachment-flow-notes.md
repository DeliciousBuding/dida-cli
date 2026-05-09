# Web API Attachment Flow Notes

This note records the currently mapped attachment flow from the saved webapp
bundle and live read-only probes. It is not a command spec yet.

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

## Task Attachment Flow

Bundle evidence also shows task-level attachment references under:

```text
GET /attachment/{encodedAttachmentPath}
GET /attachment/{encodedAttachmentPath}?action=preview
GET /attachment/upload/countdown?attachmentId={attachmentId}
GET /api/v1/attachment/countdown?id={id}&attachmentId={attachmentId}
```

The exact `encodedAttachmentPath` helper still needs to be decoded before the
CLI can safely expose task attachment download or preview helpers.

## Remaining Gaps

- Multipart request format for upload helper.
- Full upload response shape.
- Accepted file types and size failure responses.
- Task-level attachment creation and association payloads.
- Delete behavior for uploaded-but-not-attached files.
- Whether comment attachment ids are reusable across comment create/update.

## Implementation Gate

Do not add attachment upload/download commands until a reversible live test
records:

1. upload request method, headers, and multipart field names
2. upload response fields
3. comment create payload with attachment id
4. read-back through comment list
5. cleanup or rollback behavior
