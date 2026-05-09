# Dida365 OpenAPI Reference

Source: local saved page `C:\Users\Ding\Downloads\TickTick Developer_files\index.html`

Saved from `https://developer.dida365.com/docs/index.html#/openapi`.

# Dida365 Open API

## Introduction

Welcome to the Dida365 Open API documentation. Dida365 is a powerful task management application that allows users to easily manage and organize their daily tasks, deadlines, and projects. With Dida365 Open API, developers can integrate Dida365's powerful task management features into their own applications and create a seamless user experience.

## Getting Started

To get started using the Dida365 Open API, you will need to register your application and obtain a client ID and client secret. You can register your application by visiting the [Dida365 Developer Center](https://developer.dida365.com/manage). Once registered, you will receive a client ID and client secret which you will use to authenticate your requests.

## Authorization

### Get Access Token

In order to call Dida365's Open API, it is necessary to obtain an access token for the corresponding user. Dida365 uses the OAuth2 protocol to obtain the access token.

#### First Step

Redirect the user to the Dida365 authorization page, [https://dida365.com/oauth/authorize](https://dida365.com/oauth/authorize). The required parameters are as follows:

| Name | Description |
| --- | --- |
| client_id | Application unique id |
| scope | Spaces-separated permission scope. The currently available scopes are tasks:write tasks:read |
| state | Passed to redirect url as is |
| redirect_uri | User-configured redirect url |
| response_type | Fixed as code |

Example:
[https://dida365.com/oauth/authorize?scope=scope&client_id=client_id&state=state&redirect_uri=redirect_uri&response_type=code](https://dida365.com/oauth/authorize?scope=scope&client_id=client_id&state=state&redirect_uri=redirect_uri&response_type=code)

#### Second Step

After the user grants access, Dida365 will redirect the user back to your application's `redirect_uri` with an authorization code as a query parameter.

| Name | Description |
| --- | --- |
| code | Authorization code for subsequent access tokens |
| state | state parameter passed in the first step |

#### Third Step

To exchange the authorization code for an access token, make a POST request to `https://dida365.com/oauth/token` with the following parameters(Content-Type: application/x-www-form-urlencoded):

| Name | Description |
| --- | --- |
| client_id | The username is located in the **HEADER** using the **Basic Auth** authentication method |
| client_secret | The password is located in the **HEADER** using the **Basic Auth** authentication method |
| code | The code obtained in the second step |
| grant_type | grant type, now only authorization_code |
| scope | spaces-separated permission scope. The currently available scopes are tasks: write, tasks: read |
| redirect_uri | user-configured redirect url |

Access_token for openapi request authentication in the request response

```
 {
...
"access_token": "access token value"
...
}
```

#### Request OpenAPI

Set **Authorization** in the header, the value is **Bearer** `access token value`

```
Authorization: Bearer e*****b
```

## API Reference

The Dida365 Open API provides a RESTful interface for accessing and managing user tasks, lists, and other related resources. The API is based on the standard HTTP protocol and supports JSON data formats.

### Task

#### Get Task By Project ID And Task ID

```
GET /open/v1/project/{projectId}/task/{taskId}
```

##### Parameters

| Type | Name | Description | Schema |
| --- | --- | --- | --- |
| **Path** | **projectId** *required* | Project identifier | string |
| **Path** | **taskId** *required* | Task identifier | string |

##### Responses

| HTTP Code | Description | Schema |
| --- | --- | --- |
| **200** | OK | [Task](https://developer.dida365.com/docs/index.html#/openapi?id=task) |
| **401** | Unauthorized | No Content |
| **403** | Forbidden | No Content |
| **404** | Not Found | No Content |

##### Example

###### Request

```http
GET /open/v1/project/{{projectId}}/task/{{taskId}} HTTP/1.1
Host: api.dida365.com
Authorization: Bearer {{token}}
```

###### Response

```json
{
"id" : "63b7bebb91c0a5474805fcd4",
"isAllDay" : true,
"projectId" : "6226ff9877acee87727f6bca",
"title" : "Task Title",
"content" : "Task Content",
"desc" : "Task Description",
"timeZone" : "America/Los_Angeles",
"repeatFlag" : "RRULE:FREQ=DAILY;INTERVAL=1",
"startDate" : "2019-11-13T03:00:00+0000",
"dueDate" : "2019-11-14T03:00:00+0000",
"reminders" : [ "TRIGGER:P0DT9H0M0S", "TRIGGER:PT0S" ],
"tags" : [ "work", "urgent" ],
"priority" : 1,
"status" : 0,
"completedTime" : "2019-11-13T03:00:00+0000",
"sortOrder" : 12345,
"items" : [ {
    "id" : "6435074647fd2e6387145f20",
    "status" : 0,
    "title" : "Item Title",
    "sortOrder" : 12345,
    "startDate" : "2019-11-13T03:00:00+0000",
    "isAllDay" : false,
    "timeZone" : "America/Los_Angeles",
    "completedTime" : "2019-11-13T03:00:00+0000"
    } ]
}
```

#### Create Task

```
POST /open/v1/task
```

##### Parameters

| **Type** | **Name** | **Description** | **Schema** |
| --- | --- | --- | --- |
| **Body** | title   *required* | Task title | string |
| **Body** | projectId *required* | Project id | string |
| **Body** | content | Task content | string |
| **Body** | desc | Description of checklist | string |
| **Body** | isAllDay | All day | boolean |
| **Body** | startDate | Start date and time in `"yyyy-MM-dd'T'HH:mm:ssZ"` format<br> **Example** : `"2019-11-13T03:00:00+0000"` | date |
| **Body** | dueDate | Due date and time in `"yyyy-MM-dd'T'HH:mm:ssZ"` format<br> **Example** : `"2019-11-13T03:00:00+0000"` | date |
| **Body** | timeZone | The time zone in which the time is specified | String |
| **Body** | reminders | Lists of reminders specific to the task | list |
| **Body** | tags | Tags of the task | list |
| **Body** | repeatFlag | Recurring rules of task | string |
| **Body** | priority | The priority of task, default is "0" | integer |
| **Body** | sortOrder | The order of task | integer |
| **Body** | items | The list of subtasks | list |
| **Body** | items.title | Subtask title | string |
| **Body** | items.startDate | Start date and time in `"yyyy-MM-dd'T'HH:mm:ssZ"` format | date |
| **Body** | items.isAllDay | All day | boolean |
| **Body** | items.sortOrder | The order of subtask | integer |
| **Body** | items.timeZone | The time zone in which the Start time is specified | string |
| **Body** | items.status | The completion status of subtask | integer |
| **Body** | items.completedTime | Completed time in `"yyyy-MM-dd'T'HH:mm:ssZ"` format<br> **Example** : `"2019-11-13T03:00:00+0000"` | date |

##### Responses

| HTTP Code | Description | Schema |
| --- | --- | --- |
| **200** | OK | [Task](https://developer.dida365.com/docs/index.html#/Task) |
| **201** | Created | No Content |
| **401** | Unauthorized | No Content |
| **403** | Forbidden | No Content |
| **404** | Not Found | No Content |

##### Example

###### Request

```http
POST /open/v1/task HTTP/1.1
Host: api.dida365.com
Content-Type: application/json
Authorization: Bearer {{token}}
{
    ...
    "title":"Task Title",
    "projectId":"6226ff9877acee87727f6bca",
    "tags":["work","urgent"]
    ...
}
```

###### Response

```json
{
"id" : "63b7bebb91c0a5474805fcd4",
"projectId" : "6226ff9877acee87727f6bca",
"title" : "Task Title",
"content" : "Task Content",
"desc" : "Task Description",
"isAllDay" : true,
"startDate" : "2019-11-13T03:00:00+0000",
"dueDate" : "2019-11-14T03:00:00+0000",
"timeZone" : "America/Los_Angeles",
"reminders" : [ "TRIGGER:P0DT9H0M0S", "TRIGGER:PT0S" ],
"tags" : [ "work", "urgent" ],
"repeatFlag" : "RRULE:FREQ=DAILY;INTERVAL=1",
"priority" : 1,
"status" : 0,
"completedTime" : "2019-11-13T03:00:00+0000",
"sortOrder" : 12345,
"items" : [ {
    "id" : "6435074647fd2e6387145f20",
    "status" : 1,
    "title" : "Subtask Title",
    "sortOrder" : 12345,
    "startDate" : "2019-11-13T03:00:00+0000",
    "isAllDay" : false,
    "timeZone" : "America/Los_Angeles",
    "completedTime" : "2019-11-13T03:00:00+0000"
    } ]
}
```

#### Update Task

```
POST /open/v1/task/{taskId}
```

##### Parameters

| **Type** | **Name** | **Description** | **Schema** |
| --- | --- | --- | --- |
| **Path** | **taskId** *required* | Task identifier | string |
| **Body** | id     *required* | Task id. | string |
| **Body** | projectId     *required* | Project id. | string |
| **Body** | title | Task title | string |
| **Body** | content | Task content | string |
| **Body** | desc | Description of checklist | string |
| **Body** | isAllDay | All day | boolean |
| **Body** | startDate | Start date and time in `"yyyy-MM-dd'T'HH:mm:ssZ"` format<br> **Example** : `"2019-11-13T03:00:00+0000"` | date |
| **Body** | dueDate | Due date and time in `"yyyy-MM-dd'T'HH:mm:ssZ"` format<br> **Example** : `"2019-11-13T03:00:00+0000"` | date |
| **Body** | timeZone | The time zone in which the time is specified | String |
| **Body** | reminders | Lists of reminders specific to the task | list |
| **Body** | tags | Tags of the task | list |
| **Body** | repeatFlag | Recurring rules of task | string |
| **Body** | priority | The priority of task, default is "normal" | integer |
| **Body** | sortOrder | The order of task | integer |
| **Body** | items | The list of subtasks | list |
| **Body** | items.title | Subtask title | string |
| **Body** | items.startDate | Start date and time in `"yyyy-MM-dd'T'HH:mm:ssZ"` format | date |
| **Body** | items.isAllDay | All day | boolean |
| **Body** | items.sortOrder | The order of subtask | integer |
| **Body** | items.timeZone | The time zone in which the Start time is specified | string |
| **Body** | items.status | The completion status of subtask | integer |
| **Body** | items.completedTime | Completed time in `"yyyy-MM-dd'T'HH:mm:ssZ"` format<br> **Example** : `"2019-11-13T03:00:00+0000"` | date |

##### Responses

| HTTP Code | Description | Schema |
| --- | --- | --- |
| **200** | OK | [Task](https://developer.dida365.com/docs/index.html#/Task) |
| **201** | Created | No Content |
| **401** | Unauthorized | No Content |
| **403** | Forbidden | No Content |
| **404** | Not Found | No Content |

##### Example

###### Request

```http
POST /open/v1/task/{{taskId}} HTTP/1.1
Host: api.dida365.com
Content-Type: application/json
Authorization: Bearer {{token}}
{
    "id": "{{taskId}}",
    "projectId": "{{projectId}}",
    "title": "Task Title",
    "priority": 1,
    ...
}
```

###### Response

```json
{
"id" : "63b7bebb91c0a5474805fcd4",
"projectId" : "6226ff9877acee87727f6bca",
"title" : "Task Title",
"content" : "Task Content",
"desc" : "Task Description",
"isAllDay" : true,
"startDate" : "2019-11-13T03:00:00+0000",
"dueDate" : "2019-11-14T03:00:00+0000",
"timeZone" : "America/Los_Angeles",
"reminders" : [ "TRIGGER:P0DT9H0M0S", "TRIGGER:PT0S" ],
"tags" : [ "work", "urgent" ],
"repeatFlag" : "RRULE:FREQ=DAILY;INTERVAL=1",
"priority" : 1,
"status" : 0,
"completedTime" : "2019-11-13T03:00:00+0000",
"sortOrder" : 12345,
"items" : [ {
    "id" : "6435074647fd2e6387145f20",
    "status" : 1,
    "title" : "Item Title",
    "sortOrder" : 12345,
    "startDate" : "2019-11-13T03:00:00+0000",
    "isAllDay" : false,
    "timeZone" : "America/Los_Angeles",
    "completedTime" : "2019-11-13T03:00:00+0000"
    } ],
"kind": "CHECKLIST"
}
```

#### Complete Task

```
POST /open/v1/project/{projectId}/task/{taskId}/complete
```

##### Parameters

| Type | Name | Description | Schema |
| --- | --- | --- | --- |
| **Path** | **projectId** *required* | Project identifier | string |
| **Path** | **taskId** *required* | Task identifier | string |

##### Responses

| HTTP Code | Description | Schema |
| --- | --- | --- |
| **200** | OK | No Content |
| **201** | Created | No Content |
| **401** | Unauthorized | No Content |
| **403** | Forbidden | No Content |
| **404** | Not Found | No Content |

##### Example

###### Request

```http
POST /open/v1/project/{{projectId}}/task/{{taskId}}/complete HTTP/1.1
Host: api.dida365.com
Authorization: Bearer {{token}}
```

#### Delete Task

```
DELETE /open/v1/project/{projectId}/task/{taskId}
```

##### Parameters

| Type | Name | Description | Schema |
| --- | --- | --- | --- |
| **Path** | **projectId** *required* | Project identifier | string |
| **Path** | **taskId** *required* | Task identifier | string |

##### Responses

| HTTP Code | Description | Schema |
| --- | --- | --- |
| **200** | OK | No Content |
| **201** | Created | No Content |
| **401** | Unauthorized | No Content |
| **403** | Forbidden | No Content |
| **404** | Not Found | No Content |

##### Example

###### Request

```http
DELETE /open/v1/project/{{projectId}}/task/{{taskId}} HTTP/1.1
Host: api.dida365.com
Authorization: Bearer {{token}}
```

#### Move Task

```
POST /open/v1/task/move
```

Moves one or more tasks between projects.

##### Request Body

A JSON array containing task move operations.

| Type | Name | Description | Schema |
| --- | --- | --- | --- |
| **Body** | **fromProjectId** *required* | The ID of the source project | string |
| **Body** | **toProjectId** *required* | The ID of the destination project | string |
| **Body** | **taskId** *required* | The ID of the task to move | string |

##### Responses

| HTTP Code | Description | Schema |
| --- | --- | --- |
| **200** | OK | Returns an array of move results, including the task ID and its new etag) |
| **201** | Created | No Content |
| **401** | Unauthorized | No Content |
| **403** | Forbidden | No Content |
| **404** | Not Found | No Content |

##### Example

###### Request

```http
POST /open/v1/task/move HTTP/1.1
Host: api.dida365.com
Authorization: Bearer {{token}}
[
  {
    "fromProjectId":"69a850ef1c20d2030e148fdd",
    "toProjectId":"69a850f41c20d2030e148fdf",
    "taskId":"69a850f8b9061f374d54a046"
  }
]
```

###### Response

```json
[
  {
    "id": "69a850f8b9061f374d54a046",
    "etag": "43p2zso1"
  }
]
```

#### List Completed Tasks

```
POST /open/v1/task/completed
```

Retrieves a list of tasks marked as completed within specific projects and a given time range.

##### Request Body

A JSON object containing filter criteria. All fields are optional, but at least one filter is recommended to narrow down results.

| Type | Name | Description | Schema |
| --- | --- | --- | --- |
| **Body** | **projectIds** | List of project identifier | list |
| **Body** | **startDate** | The start of the time range (inclusive). Filters tasks where completedTime ≥ startDate | date |
| **Body** | **endDate** | The end of the time range (inclusive). Filters tasks where completedTime ≤ endDate | date |

##### Responses

| HTTP Code | Description | Schema |
| --- | --- | --- |
| **200** | OK | < [Task](https://developer.dida365.com/docs/index.html#/Task) > array |
| **201** | Created | No Content |
| **401** | Unauthorized | No Content |
| **403** | Forbidden | No Content |
| **404** | Not Found | No Content |

##### Example

###### Request

```http
POST /open/v1/task/completed HTTP/1.1
Host: api.dida365.com
Authorization: Bearer {{token}}
{
  "projectIds": [
    "69a850f41c20d2030e148fdf"
  ],
  "startDate":"2026-03-01T00:58:20.000+0000",
  "endDate":"2026-03-05T10:58:20.000+0000"
}
```

###### Response

```json
[
  {
    "id": "69a850f8b9061f374d54a046",
    "projectId": "69a850f41c20d2030e148fdf",
    "sortOrder": -1099511627776,
    "title": "update",
    "content": "",
    "timeZone": "America/Los_Angeles",
    "isAllDay": false,
    "priority": 0,
    "completedTime": "2026-03-04T23:58:20.000+0000",
    "status": 2,
    "etag": "t3kc5m5f",
    "kind": "TEXT"
  }
]
```

#### Filter Tasks

```
POST /open/v1/task/filter
```

Retrieves a list of tasks based on advanced filtering criteria, including project scope, date ranges, priority levels, tags, and status.

##### Parameters

| Type | Name | Description | Schema |
| --- | --- | --- | --- |
| **Body** | **projectIds** | Filters tasks belonging to the specified project ID | list |
| **Body** | **startDate** | Filters tasks where the task's startDate ≥ startDate | date |
| **Body** | **endDate** | Filters tasks where the task's startDate ≤ endDate | date |
| **Body** | **proiority** | Filters tasks by specific priority levels, Valid Values: None(0), Low(1), Mediunm(3), High(5) | list |
| **Body** | **tag** | Filters tasks that contain all of the specified tags | list |
| **Body** | **status** | Filters tasks by their current status codes (e.g., [0] for Open, [2] for Completed) | list |

##### Responses

| HTTP Code | Description | Schema |
| --- | --- | --- |
| **200** | OK | < [Task](https://developer.dida365.com/docs/index.html#/Task) > array |
| **201** | Created | No Content |
| **401** | Unauthorized | No Content |
| **403** | Forbidden | No Content |
| **404** | Not Found | No Content |

##### Example

###### Request

```http
POST /open/v1/task/filter HTTP/1.1
Host: api.dida365.com
Authorization: Bearer {{token}}
{
  "projectIds": [
    "69a850f41c20d2030e148fdf"
  ],
  "startDate":"2026-03-01T00:58:20.000+0000",
  "endDate":"2026-03-06T10:58:20.000+0000",
  "priority": [0],
  "tag": ["urgent"],
  "status": [0]
}
```

###### Response

```json
[
  {
    "id": "69a85785b9061f3c217e9de6",
    "projectId": "69a850f41c20d2030e148fdf",
    "sortOrder": -2199023255552,
    "title": "task1",
    "content": "",
    "desc": "",
    "startDate": "2026-03-05T00:00:00.000+0000",
    "dueDate": "2026-03-05T00:00:00.000+0000",
    "timeZone": "America/Los_Angeles",
    "isAllDay": false,
    "priority": 0,
    "status": 0,
    "tags": [
      "tag"
    ],
    "etag": "cic6e3cg",
    "kind": "TEXT"
  },
  {
    "id": "69a8ea79b9061f4d803f6b32",
    "projectId": "69a850f41c20d2030e148fdf",
    "sortOrder": -3298534883328,
    "title": "task2",
    "content": "",
    "startDate": "2026-03-05T00:00:00.000+0000",
    "dueDate": "2026-03-05T00:00:00.000+0000",
    "timeZone": "America/Los_Angeles",
    "isAllDay": false,
    "priority": 0,
    "status": 0,
    "tags": [
      "tag"
    ],
    "etag": "0nvpcxzh",
    "kind": "TEXT"
  }
]
```

### Project

#### Get User Project

```
GET /open/v1/project
```

##### Responses

| HTTP Code | Description | Schema |
| --- | --- | --- |
| **200** | OK | < [Project](https://developer.dida365.com/docs/index.html#/Project) > array |
| **401** | Unauthorized | No Content |
| **403** | Forbidden | No Content |
| **404** | Not Found | No Content |

##### Example

###### Request

```http
GET /open/v1/project HTTP/1.1
Host: api.dida365.com
Authorization: Bearer {{token}}
```

###### Response

```json
[{
"id": "6226ff9877acee87727f6bca",
"name": "project name",
"color": "#F18181",
"closed": false,
"groupId": "6436176a47fd2e05f26ef56e",
"viewMode": "list",
"permission": "write",
"kind": "TASK"
}]
```

#### Get Project By ID

```
GET /open/v1/project/{projectId}
```

##### Parameters

| Type | Name | Description | Schema |
| --- | --- | --- | --- |
| **Path** | **project** *required* | Project identifier | string |

##### Responses

| HTTP Code | Description | Schema |
| --- | --- | --- |
| **200** | OK | [Project](https://developer.dida365.com/docs/index.html#/Project) |
| **401** | Unauthorized | No Content |
| **403** | Forbidden | No Content |
| **404** | Not Found | No Content |

##### Example

###### Request path

```http
GET /open/v1/project/{{projectId}} HTTP/1.1
Host: api.dida365.com
Authorization: Bearer {{token}}
```

###### Response

```json
{
    "id": "6226ff9877acee87727f6bca",
    "name": "project name",
    "color": "#F18181",
    "closed": false,
    "groupId": "6436176a47fd2e05f26ef56e",
    "viewMode": "list",
    "kind": "TASK"
}
```

#### Get Project With Data

```
GET /open/v1/project/{projectId}/data
```

##### Parameters

| Type | Name | Description | Schema |
| --- | --- | --- | --- |
| **Path** | **projectId** *required* | Project identifier, "inbox" | string |

##### Responses

| HTTP Code | Description | Schema |
| --- | --- | --- |
| **200** | OK | [ProjectData](https://developer.dida365.com/docs/index.html#/ProjectData) |
| **401** | Unauthorized | No Content |
| **403** | Forbidden | No Content |
| **404** | Not Found | No Content |

##### Example

###### Request

```http
GET /open/v1/project/{{projectId}}/data HTTP/1.1
Host: api.dida365.com
Authorization: Bearer {{token}}
```

###### Response

```json
{
"project": {
    "id": "6226ff9877acee87727f6bca",
    "name": "project name",
    "color": "#F18181",
    "closed": false,
    "groupId": "6436176a47fd2e05f26ef56e",
    "viewMode": "list",
    "kind": "TASK"
},
"tasks": [{
    "id": "6247ee29630c800f064fd145",
    "isAllDay": true,
    "projectId": "6226ff9877acee87727f6bca",
    "title": "Task Title",
    "content": "Task Content",
    "desc": "Task Description",
    "timeZone": "America/Los_Angeles",
    "repeatFlag": "RRULE:FREQ=DAILY;INTERVAL=1",
    "startDate": "2019-11-13T03:00:00+0000",
    "dueDate": "2019-11-14T03:00:00+0000",
    "reminders": [
        "TRIGGER:P0DT9H0M0S",
        "TRIGGER:PT0S"
    ],
    "priority": 1,
    "status": 0,
    "completedTime": "2019-11-13T03:00:00+0000",
    "sortOrder": 12345,
    "items": [{
        "id": "6435074647fd2e6387145f20",
        "status": 0,
        "title": "Subtask Title",
        "sortOrder": 12345,
        "startDate": "2019-11-13T03:00:00+0000",
        "isAllDay": false,
        "timeZone": "America/Los_Angeles",
        "completedTime": "2019-11-13T03:00:00+0000"
    }]
}],
"columns": [{
    "id": "6226ff9e76e5fc39f2862d1b",
    "projectId": "6226ff9877acee87727f6bca",
    "name": "Column Name",
    "sortOrder": 0
}]
}
```

#### Create Project

```
POST /open/v1/project
```

##### Parameters

| **Type** | **Name** | **Description** | **Schema** |
| --- | --- | --- | --- |
| **Body** | name  *required* | name of the project | string |
| **Body** | color | color of project, eg. "#F18181" | string |
| **Body** | sortOrder | sort order value of the project | integer (int64) |
| **Body** | viewMode | view mode, "list", "kanban", "timeline" | string |
| **Body** | kind | project kind, "TASK", "NOTE" | string |

##### Responses

| HTTP Code | Description | Schema |
| --- | --- | --- |
| **200** | OK | [Project](https://developer.dida365.com/docs/index.html#/Project) |
| **201** | Created | No Content |
| **401** | Unauthorized | No Content |
| **403** | Forbidden | No Content |
| **404** | Not Found | No Content |

##### Example

###### Request

```http
POST /open/v1/project HTTP/1.1
Host: api.dida365.com
Content-Type: application/json
Authorization: Bearer {{token}}
{
    "name": "project name",
    "color": "#F18181",
    "viewMode": "list",
    "kind": "task"
}
```

###### Response

```json
{
"id": "6226ff9877acee87727f6bca",
"name": "project name",
"color": "#F18181",
"sortOrder": 0,
"viewMode": "list",
"kind": "TASK"
}
```

#### Update Project

```
POST /open/v1/project/{projectId}
```

##### Parameters

| **Type** | **Parameter** | **Description** | Schema |
| --- | --- | --- | --- |
| **Path** | projectId *required* | project identifier | string |
| **Body** | name | name of the project | string |
| **Body** | color | color of the project | string |
| **Body** | sortOrder | sort order value, default 0 | integer (int64) |
| **Body** | viewMode | view mode, "list", "kanban", "timeline" | string |
| **Body** | kind | project kind, "TASK", "NOTE" | string |

##### Responses

| HTTP Code | Description | Schema |
| --- | --- | --- |
| **200** | OK | [Project](https://developer.dida365.com/docs/index.html#/Project) |
| **201** | Created | No Content |
| **401** | Unauthorized | No Content |
| **403** | Forbidden | No Content |
| **404** | Not Found | No Content |

##### Example

###### Request

```http
POST /open/v1/project/{{projectId}} HTTP/1.1
Host: api.dida365.com
Content-Type: application/json
Authorization: Bearer {{token}}

{
    "name": "Project Name",
    "color": "#F18181",
    "viewMode": "list",
    "kind": "TASK"
}
```

###### Response

```json
{
"id": "6226ff9877acee87727f6bca",
"name": "Project Name",
"color": "#F18181",
"sortOrder": 0,
"viewMode": "list",
"kind": "TASK"
}
```

#### Delete Project

```
DELETE /open/v1/project/{projectId}
```

##### Parameters

| Type | Name | Description | Schema |
| --- | --- | --- | --- |
| Path | **projectId** *required* | Project identifier | string |

##### Responses

| HTTP Code | Description | Schema |
| --- | --- | --- |
| **200** | OK | No Content |
| **401** | Unauthorized | No Content |
| **403** | Forbidden | No Content |
| **404** | Not Found | No Content |

##### Example

###### Request

```http
DELETE /open/v1/project/{{projectId}} HTTP/1.1
Host: api.dida365.com
Authorization: Bearer {{token}}
```

### Focus

#### Get Focus By Focus ID

```
GET /open/v1/focus/{focusId}
```

##### Parameters

| Type | Name | Description | Schema |
| --- | --- | --- | --- |
| **Path** | **focusId** *required* | Focus identifier | string |
| **Query** | **type** *required* | Focus type<br> **Value** : Pomodoro: `0`, Timing: `1` | integer (int32) |

##### Responses

| HTTP Code | Description | Schema |
| --- | --- | --- |
| **200** | OK | [OpenFocus](https://developer.dida365.com/docs/index.html#/openapi?id=openfocus) |
| **401** | Unauthorized | No Content |
| **403** | Forbidden | No Content |
| **404** | Not Found | No Content |

##### Example

###### Request

```http
GET /open/v1/focus/{{focusId}}?type=0 HTTP/1.1
Host: api.dida365.com
Authorization: Bearer {{token}}
```

###### Response

```json
{
  "id": "focus-1",
  "type": 0,
  "taskId": "task-1",
  "note": "Deep work",
  "status": 2,
  "startTime": "2026-04-07T09:00:00+0800",
  "endTime": "2026-04-07T09:25:00+0800",
  "pauseDuration": 0,
  "adjustTime": 0,
  "added": true,
  "createdTime": "2026-04-07T09:25:00+0800",
  "modifiedTime": "2026-04-07T09:25:00+0800",
  "etimestamp": 1712453100000,
  "etag": "focus-etag",
  "duration": 1500,
  "relationType": [0]
}
```

#### Get Focuses By Time Range

```
GET /open/v1/focus
```

##### Parameters

| **Type** | **Name** | **Description** | **Schema** |
| --- | --- | --- | --- |
| **Query** | **from** *required* | Range start time in `"yyyy-MM-dd'T'HH:mm:ssZ"` format<br> **Example** : `"2026-04-01T00:00:00+0800"` | string (date-time) |
| **Query** | **to** *required* | Range end time in `"yyyy-MM-dd'T'HH:mm:ssZ"` format<br> **Example** : `"2026-04-02T00:00:00+0800"` | string (date-time) |
| **Query** | **type** *required* | Focus type<br> **Value** : Pomodoro: `0`, Timing: `1` | integer (int32) |

If the time range exceeds 30 days, the server automatically adjusts the start time to 30 days before `to`.

##### Responses

| HTTP Code | Description | Schema |
| --- | --- | --- |
| **200** | OK | < [OpenFocus](https://developer.dida365.com/docs/index.html#/openapi?id=openfocus) > array |
| **401** | Unauthorized | No Content |
| **403** | Forbidden | No Content |
| **404** | Not Found | No Content |

##### Example

###### Request

```http
GET /open/v1/focus?from=2026-04-01T00:00:00+0800&to=2026-04-02T00:00:00+0800&type=1 HTTP/1.1
Host: api.dida365.com
Authorization: Bearer {{token}}
```

###### Response

```json
[
  {
    "id": "focus-1",
    "type": 1,
    "note": "Timing",
    "startTime": "2026-04-01T09:00:00+0800",
    "endTime": "2026-04-01T10:30:00+0800",
    "duration": 5400
  }
]
```

#### Delete Focus

```
DELETE /open/v1/focus/{focusId}
```

##### Parameters

| **Type** | **Name** | **Description** | **Schema** |
| --- | --- | --- | --- |
| **Path** | **focusId** *required* | Focus identifier | string |
| **Query** | **type** *required* | Focus type<br> **Value** : Pomodoro: `0`, Timing: `1` | integer (int32) |

##### Responses

| HTTP Code | Description | Schema |
| --- | --- | --- |
| **200** | OK | [OpenFocus](https://developer.dida365.com/docs/index.html#/openapi?id=openfocus) |
| **401** | Unauthorized | No Content |
| **403** | Forbidden | No Content |
| **404** | Not Found | No Content |

##### Example

###### Request

```http
DELETE /open/v1/focus/{{focusId}}?type=0 HTTP/1.1
Host: api.dida365.com
Authorization: Bearer {{token}}
```

###### Response

```json
{
  "id": "focus-1",
  "type": 0
}
```

### Habit

#### Get Habit By Habit ID

```
GET /open/v1/habit/{habitId}
```

##### Parameters

| Type | Name | Description | Schema |
| --- | --- | --- | --- |
| **Path** | **habitId** *required* | Habit identifier | string |

##### Responses

| HTTP Code | Description | Schema |
| --- | --- | --- |
| **200** | OK | [OpenHabit](https://developer.dida365.com/docs/index.html#/openapi?id=openhabit) |
| **401** | Unauthorized | No Content |
| **403** | Forbidden | No Content |
| **404** | Not Found | No Content |

##### Example

###### Request

```http
GET /open/v1/habit/{{habitId}} HTTP/1.1
Host: api.dida365.com
Authorization: Bearer {{token}}
```

###### Response

```json
{
  "id": "habit-1",
  "name": "Read",
  "iconRes": "habit_reading",
  "color": "#4D8CF5",
  "sortOrder": 12345,
  "status": 0,
  "encouragement": "Keep going",
  "totalCheckIns": 12,
  "createdTime": "2024-01-01T00:00:00+0000",
  "modifiedTime": "2024-01-08T00:00:00+0000",
  "type": "Boolean",
  "goal": 1.0,
  "step": 1.0,
  "unit": "Count",
  "etag": "habit-etag",
  "repeatRule": "RRULE:FREQ=DAILY;INTERVAL=1",
  "reminders": [],
  "recordEnable": false,
  "sectionId": "section-1",
  "targetDays": 0,
  "targetStartDate": 20240101,
  "completedCycles": 12,
  "exDates": [],
  "style": 0
}
```

#### Get All Habits

```
GET /open/v1/habit
```

##### Parameters

No parameters.

##### Responses

| HTTP Code | Description | Schema |
| --- | --- | --- |
| **200** | OK | < [OpenHabit](https://developer.dida365.com/docs/index.html#/openapi?id=openhabit) > array |
| **401** | Unauthorized | No Content |
| **403** | Forbidden | No Content |

##### Example

###### Request

```http
GET /open/v1/habit HTTP/1.1
Host: api.dida365.com
Authorization: Bearer {{token}}
```

###### Response

```json
[
  {
    "id": "habit-1",
    "name": "Read",
    "repeatRule": "RRULE:FREQ=DAILY;INTERVAL=1",
    "status": 0
  },
  {
    "id": "habit-2",
    "name": "Exercise",
    "repeatRule": "RRULE:FREQ=WEEKLY;BYDAY=MO,WE,FR",
    "status": 0
  }
]
```

#### Create Habit

```
POST /open/v1/habit
```

##### Parameters

| **Type** | **Name** | **Description** | **Schema** |
| --- | --- | --- | --- |
| **Body** | name *required* | Habit name. Maximum length is 1000 characters | string |
| **Body** | iconRes | Habit icon resource | string |
| **Body** | color | Habit color | string |
| **Body** | sortOrder | Habit sort order | integer (int64) |
| **Body** | status | Habit status | integer (int32) |
| **Body** | encouragement | Habit encouragement message | string |
| **Body** | type | Habit type | string |
| **Body** | goal | Habit goal | number (double) |
| **Body** | step | Habit step | number (double) |
| **Body** | unit | Habit unit | string |
| **Body** | repeatRule | Habit repeat rule<br> **Example** : `"RRULE:FREQ=DAILY;INTERVAL=1"` | string |
| **Body** | reminders | Habit reminders | < string > array |
| **Body** | recordEnable | Whether record is enabled | boolean |
| **Body** | sectionId | Habit section identifier | string |
| **Body** | targetDays | Target days | integer (int32) |
| **Body** | targetStartDate | Target start date in `YYYYMMDD` format<br> **Example** : `20240101` | integer (int32) |
| **Body** | completedCycles | Completed cycles | integer (int32) |
| **Body** | exDates | Excluded dates | < string > array |
| **Body** | style | Habit style | integer (int32) |

##### Responses

| HTTP Code | Description | Schema |
| --- | --- | --- |
| **200** | OK | [OpenHabit](https://developer.dida365.com/docs/index.html#/openapi?id=openhabit) |
| **201** | Created | No Content |
| **401** | Unauthorized | No Content |
| **403** | Forbidden | No Content |
| **404** | Not Found | No Content |

##### Example

###### Request

```http
POST /open/v1/habit HTTP/1.1
Host: api.dida365.com
Content-Type: application/json
Authorization: Bearer {{token}}

{
  "name": "Read",
  "iconRes": "habit_reading",
  "color": "#4D8CF5",
  "type": "Boolean",
  "goal": 1.0,
  "step": 1.0,
  "unit": "Count",
  "repeatRule": "RRULE:FREQ=DAILY;INTERVAL=1",
  "recordEnable": false
}
```

###### Response

```json
{
  "id": "habit-1",
  "name": "Read",
  "iconRes": "habit_reading",
  "color": "#4D8CF5",
  "type": "Boolean",
  "goal": 1.0,
  "step": 1.0,
  "unit": "Count",
  "repeatRule": "RRULE:FREQ=DAILY;INTERVAL=1",
  "recordEnable": false
}
```

#### Update Habit

```
POST /open/v1/habit/{habitId}
```

##### Parameters

| **Type** | **Name** | **Description** | **Schema** |
| --- | --- | --- | --- |
| **Path** | **habitId** *required* | Habit identifier | string |
| **Body** | name | Habit name. If empty, it will be treated as null. Maximum length is 1000 characters | string |
| **Body** | iconRes | Habit icon resource | string |
| **Body** | color | Habit color | string |
| **Body** | sortOrder | Habit sort order | integer (int64) |
| **Body** | status | Habit status | integer (int32) |
| **Body** | encouragement | Habit encouragement message | string |
| **Body** | type | Habit type | string |
| **Body** | goal | Habit goal | number (double) |
| **Body** | step | Habit step | number (double) |
| **Body** | unit | Habit unit | string |
| **Body** | repeatRule | Habit repeat rule<br> **Example** : `"RRULE:FREQ=DAILY;INTERVAL=1"` | string |
| **Body** | reminders | Habit reminders | < string > array |
| **Body** | recordEnable | Whether record is enabled | boolean |
| **Body** | sectionId | Habit section identifier | string |
| **Body** | targetDays | Target days | integer (int32) |
| **Body** | targetStartDate | Target start date in `YYYYMMDD` format<br> **Example** : `20240101` | integer (int32) |
| **Body** | completedCycles | Completed cycles | integer (int32) |
| **Body** | exDates | Excluded dates | < string > array |
| **Body** | style | Habit style | integer (int32) |

##### Responses

| HTTP Code | Description | Schema |
| --- | --- | --- |
| **200** | OK | [OpenHabit](https://developer.dida365.com/docs/index.html#/openapi?id=openhabit) |
| **201** | Created | No Content |
| **401** | Unauthorized | No Content |
| **403** | Forbidden | No Content |
| **404** | Not Found | No Content |

##### Example

###### Request

```http
POST /open/v1/habit/{{habitId}} HTTP/1.1
Host: api.dida365.com
Content-Type: application/json
Authorization: Bearer {{token}}

{
  "name": "Read more",
  "goal": 2.0,
  "repeatRule": "RRULE:FREQ=DAILY;INTERVAL=1"
}
```

###### Response

```json
{
  "id": "habit-1",
  "name": "Read more",
  "goal": 2.0,
  "repeatRule": "RRULE:FREQ=DAILY;INTERVAL=1"
}
```

#### Create Or Update Habit Check-In

```
POST /open/v1/habit/{habitId}/checkin
```

##### Parameters

| **Type** | **Name** | **Description** | **Schema** |
| --- | --- | --- | --- |
| **Path** | **habitId** *required* | Habit identifier | string |
| **Body** | stamp *required* | Date stamp in `YYYYMMDD` format<br> **Example** : `20260407` | integer (int32) |
| **Body** | time | Check-in time in `"yyyy-MM-dd'T'HH:mm:ssZ"` format<br> **Example** : `"2026-04-07T08:00:00+0000"` | string (date-time) |
| **Body** | opTime | Operation time in `"yyyy-MM-dd'T'HH:mm:ssZ"` format<br> **Example** : `"2026-04-07T08:00:00+0000"` | string (date-time) |
| **Body** | value | Check-in value. Default is `1.0` | number (double) |
| **Body** | goal | Check-in goal. Default is `1.0` | number (double) |
| **Body** | status | Check-in status | integer (int32) |

##### Responses

| HTTP Code | Description | Schema |
| --- | --- | --- |
| **200** | OK | [OpenHabitCheckin](https://developer.dida365.com/docs/index.html#/openapi?id=openhabitcheckin) |
| **201** | Created | No Content |
| **401** | Unauthorized | No Content |
| **403** | Forbidden | No Content |
| **404** | Not Found | No Content |

##### Example

###### Request

```http
POST /open/v1/habit/{{habitId}}/checkin HTTP/1.1
Host: api.dida365.com
Content-Type: application/json
Authorization: Bearer {{token}}

{
  "stamp": 20260407,
  "value": 1.0,
  "goal": 1.0
}
```

###### Response

```json
{
  "habitId": "habit-1",
  "year": 2026,
  "checkins": [
    {
      "stamp": 20260407,
      "value": 1.0,
      "goal": 1.0
    }
  ]
}
```

#### Get Habit Check-Ins

```
GET /open/v1/habit/checkins
```

##### Parameters

| **Type** | **Name** | **Description** | **Schema** |
| --- | --- | --- | --- |
| **Query** | **habitIds** *required* | Habit identifiers, separated by commas<br> **Example** : `habit-1,habit-2` | string |
| **Query** | **from** *required* | Start date stamp in `YYYYMMDD` format<br> **Example** : `20260401` | integer (int32) |
| **Query** | **to** *required* | End date stamp in `YYYYMMDD` format<br> **Example** : `20260407` | integer (int32) |

##### Responses

| HTTP Code | Description | Schema |
| --- | --- | --- |
| **200** | OK | < [OpenHabitCheckin](https://developer.dida365.com/docs/index.html#/openapi?id=openhabitcheckin) > array |
| **401** | Unauthorized | No Content |
| **403** | Forbidden | No Content |
| **404** | Not Found | No Content |

##### Example

###### Request

```http
GET /open/v1/habit/checkins?habitIds=habit-1,habit-2&from=20260401&to=20260407 HTTP/1.1
Host: api.dida365.com
Authorization: Bearer {{token}}
```

###### Response

```json
[
  {
    "habitId": "habit-1",
    "year": 2026,
    "checkins": [
      {
        "stamp": 20260407,
        "value": 1.0,
        "goal": 1.0
      }
    ]
  }
]
```

## Definitions

### ChecklistItem

| Name | Description | Schema |
| --- | --- | --- |
| **id** | Subtask identifier | string |
| **title** | Subtask title | string |
| **status** | The completion status of subtask<br> **Value** : Normal: `0`, Completed: `1` | integer (int32) |
| **completedTime** | Subtask completed time in `"yyyy-MM-dd'T'HH:mm:ssZ"`<br> **Example** : `"2019-11-13T03:00:00+0000"` | string (date-time) |
| **isAllDay** | All day | boolean |
| **sortOrder** | Subtask sort order<br> **Example** : `234444` | integer (int64) |
| **startDate** | Subtask start date time in `"yyyy-MM-dd'T'HH:mm:ssZ"`<br> **Example** : `"2019-11-13T03:00:00+0000"` | string (date-time) |
| **timeZone** | Subtask timezone<br> **Example** : `"America/Los_Angeles"` | string |

### Task

| Name | Description | Schema |
| --- | --- | --- |
| **id** | Task identifier | string |
| **projectId** | Task project id | string |
| **title** | Task title | string |
| **isAllDay** | All day | boolean |
| **completedTime** | Task completed time in `"yyyy-MM-dd'T'HH:mm:ssZ"`<br> **Example** : `"2019-11-13T03:00:00+0000"` | string (date-time) |
| **content** | Task content | string |
| **desc** | Task description of checklist | string |
| **dueDate** | Task due date time in `"yyyy-MM-dd'T'HH:mm:ssZ"`<br> **Example** : `"2019-11-13T03:00:00+0000"` | string (date-time) |
| **items** | Subtasks of Task | < [ChecklistItem](https://developer.dida365.com/docs/index.html#/openapi?id=checklistitem) > array |
| **priority** | Task priority<br> **Value** : None:`0`, Low:`1`, Medium:`3`, High`5` | integer (int32) |
| **reminders** | List of reminder triggers<br> **Example** : `[ "TRIGGER:P0DT9H0M0S", "TRIGGER:PT0S" ]` | < string > array |
| **tags** | Task tags<br> **Example** : `[ "work", "urgent" ]` | < string > array |
| **repeatFlag** | Recurring rules of task<br> **Example** : `"RRULE:FREQ=DAILY;INTERVAL=1"` | string |
| **sortOrder** | Task sort order<br> **Example** : `12345` | integer (int64) |
| **startDate** | Start date time in `"yyyy-MM-dd'T'HH:mm:ssZ"`<br> **Example** : `"2019-11-13T03:00:00+0000"` | string (date-time) |
| **status** | Task completion status<br> **Value** : Normal: `0`, Completed: `2` | integer (int32) |
| **timeZone** | Task timezone<br> **Example** : `"America/Los_Angeles"` | string |
| **kind** | "TEXT", "NOTE", "CHECKLIST" | string |

### Project

| Name | Description | Schema |
| --- | --- | --- |
| **id** | Project identifier | string |
| **name** | Project name | string |
| **color** | Project color | string |
| **sortOrder** | Order value | integer (int64) |
| **closed** | Projcet closed | boolean |
| **groupId** | Project group identifier | string |
| **viewMode** | view mode, "list", "kanban", "timeline" | string |
| **permission** | "read", "write" or "comment" | string |
| **kind** | "TASK" or "NOTE" | string |

### Column

| Name | Description | Schema |
| --- | --- | --- |
| id | Column identifier | string |
| projectId | Project identifier | string |
| name | Column name | string |
| sortOrder | Order value | integer (int64) |

### ProjectData

| Name | Description | Schema |
| --- | --- | --- |
| project | Project info | [Project](https://developer.dida365.com/docs/index.html#/Project) |
| tasks | Undone tasks under project | <[Task](https://developer.dida365.com/docs/index.html#/Task)> array |
| columns | Columns under project | <[Column](https://developer.dida365.com/docs/index.html#/Column)> array |

### OpenPomodoroTaskBrief

| Name | Description | Schema |
| --- | --- | --- |
| **taskId** | Task id | string |
| **title** | Task title | string |
| **habitId** | Habit id | string |
| **timerId** | Timer id | string |
| **timerName** | Timer name | string |
| **startTime** | Task focus start time in `"yyyy-MM-dd'T'HH:mm:ssZ"`<br> **Example** : `"2026-04-07T09:00:00+0800"` | string (date-time) |
| **endTime** | Task focus end time in `"yyyy-MM-dd'T'HH:mm:ssZ"`<br> **Example** : `"2026-04-07T09:25:00+0800"` | string (date-time) |

### OpenFocus

| Name | Description | Schema |
| --- | --- | --- |
| **id** | Focus unique id | string |
| **userId** | User id | integer (int64) |
| **type** | Focus type<br> **Value** : Pomodoro: `0`, Timing: `1` | integer (int32) |
| **taskId** | Task id | string |
| **note** | Focus note | string |
| **tasks** | Related task briefs | < [OpenPomodoroTaskBrief](https://developer.dida365.com/docs/index.html#/openapi?id=openpomodorotaskbrief) > array |
| **status** | Pomodoro status | integer (int32) |
| **startTime** | Focus start time in `"yyyy-MM-dd'T'HH:mm:ssZ"`<br> **Example** : `"2026-04-07T09:00:00+0800"` | string (date-time) |
| **endTime** | Focus end time in `"yyyy-MM-dd'T'HH:mm:ssZ"`<br> **Example** : `"2026-04-07T09:25:00+0800"` | string (date-time) |
| **pauseDuration** | Pause duration in seconds | integer (int32) |
| **adjustTime** | Adjusted time in seconds | integer (int64) |
| **added** | Whether record was added | boolean |
| **createdTime** | Created time in `"yyyy-MM-dd'T'HH:mm:ssZ"`<br> **Example** : `"2026-04-07T09:25:00+0800"` | string (date-time) |
| **modifiedTime** | Modified time in `"yyyy-MM-dd'T'HH:mm:ssZ"`<br> **Example** : `"2026-04-07T09:25:00+0800"` | string (date-time) |
| **etimestamp** | Entity timestamp | integer (int64) |
| **etag** | Entity tag | string |
| **duration** | Focus duration | integer (int64) |
| **relationType** | Relation types | < integer > array |

### OpenHabit

| Name | Description | Schema |
| --- | --- | --- |
| **id** | Habit unique id | string |
| **name** | Habit name | string |
| **iconRes** | Habit icon resource | string |
| **color** | Habit color | string |
| **sortOrder** | Habit sort order | integer (int64) |
| **status** | Habit status | integer (int32) |
| **encouragement** | Habit encouragement message | string |
| **totalCheckIns** | Total check-ins | integer (int32) |
| **createdTime** | Created time in `"yyyy-MM-dd'T'HH:mm:ssZ"`<br> **Example** : `"2024-01-01T00:00:00+0000"` | string (date-time) |
| **modifiedTime** | Modified time in `"yyyy-MM-dd'T'HH:mm:ssZ"`<br> **Example** : `"2024-01-08T00:00:00+0000"` | string (date-time) |
| **archivedTime** | Archived time in `"yyyy-MM-dd'T'HH:mm:ssZ"`<br> **Example** : `"2024-01-08T00:00:00+0000"` | string (date-time) |
| **type** | Habit type | string |
| **goal** | Habit goal | number (double) |
| **step** | Habit step | number (double) |
| **unit** | Habit unit | string |
| **etag** | Habit etag | string |
| **repeatRule** | Habit repeat rule<br> **Example** : `"RRULE:FREQ=DAILY;INTERVAL=1"` | string |
| **reminders** | Habit reminders | < string > array |
| **recordEnable** | Whether record is enabled | boolean |
| **sectionId** | Habit section identifier | string |
| **targetDays** | Target days | integer (int32) |
| **targetStartDate** | Target start date in `YYYYMMDD` format<br> **Example** : `20240101` | integer (int32) |
| **completedCycles** | Completed cycles | integer (int32) |
| **exDates** | Excluded dates | < string > array |
| **style** | Habit style | integer (int32) |

### OpenHabitCheckinData

| Name | Description | Schema |
| --- | --- | --- |
| **id** | Check-in id | string |
| **stamp** | Date stamp in `YYYYMMDD` format<br> **Example** : `20260407` | integer (int32) |
| **time** | Check-in time in `"yyyy-MM-dd'T'HH:mm:ssZ"`<br> **Example** : `"2026-04-07T08:00:00+0000"` | string (date-time) |
| **opTime** | Operation time in `"yyyy-MM-dd'T'HH:mm:ssZ"`<br> **Example** : `"2026-04-07T08:00:00+0000"` | string (date-time) |
| **value** | Check-in value. Default is `1.0` | number (double) |
| **goal** | Check-in goal. Default is `1.0` | number (double) |
| **status** | Check-in status | integer (int32) |

### OpenHabitCheckin

| Name | Description | Schema |
| --- | --- | --- |
| **id** | Check-in document id | string |
| **habitId** | Habit id | string |
| **createdTime** | Created time in `"yyyy-MM-dd'T'HH:mm:ssZ"`<br> **Example** : `"2026-04-07T08:00:00+0000"` | string (date-time) |
| **modifiedTime** | Modified time in `"yyyy-MM-dd'T'HH:mm:ssZ"`<br> **Example** : `"2026-04-07T08:00:00+0000"` | string (date-time) |
| **etag** | Check-in etag | string |
| **year** | Year | integer (int32) |
| **checkins** | Check-in entries | < [OpenHabitCheckinData](https://developer.dida365.com/docs/index.html#/openapi?id=openhabitcheckindata) > array |

## Feedback and Support

If you have any questions or feedback regarding the Dida365 Open API documentation, please contact us at [support@dida365.com](mailto:support@dida365.com). We appreciate your input and will work to address any concerns or issues as quickly as possible. Thank you for choosing Dida!
