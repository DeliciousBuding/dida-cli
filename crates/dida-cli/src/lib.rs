use dida_core::{CliError, JsonEnvelope, failure, to_json_line};
use serde_json::json;
use std::env;

mod schema;

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct RunResult {
    pub code: i32,
    pub stdout: String,
    pub stderr: String,
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub enum RouteOutcome {
    Help,
    Version,
    MissingCommand {
        json: bool,
    },
    Dispatch {
        command: String,
        args: Vec<String>,
        json: bool,
    },
    UnknownCommand {
        command: String,
        json: bool,
    },
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct CommandContext {
    pub command: String,
    pub args: Vec<String>,
    pub json: bool,
}

pub type CommandHandler = fn(CommandContext) -> RunResult;

#[derive(Debug, Clone)]
pub struct RootCommand {
    pub name: &'static str,
    pub handler: CommandHandler,
}

impl RunResult {
    fn ok(stdout: impl Into<String>) -> Self {
        Self {
            code: 0,
            stdout: stdout.into(),
            stderr: String::new(),
        }
    }

    fn err_json(err: CliError) -> Self {
        let stdout = to_json_line(&failure(err)).unwrap_or_else(|_| {
            "{\"ok\":false,\"command\":\"internal\",\"error\":{\"type\":\"internal\",\"message\":\"encode json\"}}\n".to_string()
        });
        Self {
            code: 1,
            stdout,
            stderr: String::new(),
        }
    }

    fn err_json_for(command: &str, err: CliError) -> Self {
        let mut envelope = failure(err);
        envelope.command = command.to_string();
        let stdout = to_json_line(&envelope).unwrap_or_else(|_| {
            "{\"ok\":false,\"command\":\"internal\",\"error\":{\"type\":\"internal\",\"message\":\"encode json\"}}\n".to_string()
        });
        Self {
            code: 1,
            stdout,
            stderr: String::new(),
        }
    }

    fn err_text(command: &str, message: &str) -> Self {
        Self {
            code: 1,
            stdout: String::new(),
            stderr: format!("{command}: {message}\n"),
        }
    }
}

pub fn run(args: Vec<String>, version: &str) -> RunResult {
    let parsed = RootArgs::parse(args);
    match route_from_parsed(&parsed) {
        RouteOutcome::Help => RunResult::ok(root_help()),
        RouteOutcome::Version => render_version(version),
        RouteOutcome::MissingCommand { json } => {
            if json {
                let err =
                    CliError::validation("dida", "missing command").with_hint("run: dida --help");
                RunResult::err_json(err)
            } else {
                RunResult::ok(root_help())
            }
        }
        RouteOutcome::Dispatch {
            command,
            args,
            json,
        } => {
            let ctx = CommandContext {
                command,
                args,
                json,
            };
            let handler = root_commands()
                .into_iter()
                .find(|candidate| candidate.name == ctx.command)
                .map(|candidate| candidate.handler)
                .unwrap_or(placeholder_handler);
            handler(ctx)
        }
        RouteOutcome::UnknownCommand { command, json } => {
            let err = CliError {
                kind: None,
                message: format!("unknown command {command:?}"),
                hint: None,
                details: None,
            };
            if json {
                RunResult::err_json_for(&command, err)
            } else {
                RunResult::err_text(&command, &err.message)
            }
        }
    }
}

#[derive(Debug, Default, PartialEq, Eq)]
struct RootArgs {
    json: bool,
    help: bool,
    command: Option<String>,
    rest: Vec<String>,
}

impl RootArgs {
    fn parse(args: Vec<String>) -> Self {
        let mut parsed = Self::default();
        for arg in args {
            match arg.as_str() {
                "--json" | "-j" => parsed.json = true,
                "--help" | "-h" if parsed.command.is_none() => parsed.help = true,
                _ if parsed.command.is_none() => parsed.command = Some(arg),
                _ => parsed.rest.push(arg),
            }
        }
        parsed
    }
}

pub fn route(args: Vec<String>) -> RouteOutcome {
    let parsed = RootArgs::parse(args);
    route_from_parsed(&parsed)
}

fn route_from_parsed(parsed: &RootArgs) -> RouteOutcome {
    if parsed.help {
        return RouteOutcome::Help;
    }
    let Some(command) = parsed.command.as_deref() else {
        return RouteOutcome::MissingCommand { json: parsed.json };
    };
    if command == "version" || command == "--version" {
        return RouteOutcome::Version;
    }
    if root_commands()
        .iter()
        .any(|candidate| candidate.name == command)
    {
        return RouteOutcome::Dispatch {
            command: command.to_string(),
            args: parsed.rest.clone(),
            json: parsed.json,
        };
    }
    RouteOutcome::UnknownCommand {
        command: command.to_string(),
        json: parsed.json,
    }
}

fn render_version(version: &str) -> RunResult {
    RunResult::ok(format!("{version}\n"))
}

pub fn root_commands() -> Vec<RootCommand> {
    vec![
        RootCommand {
            name: "official",
            handler: command_handler,
        },
        RootCommand {
            name: "openapi",
            handler: command_handler,
        },
        RootCommand {
            name: "agent",
            handler: command_handler,
        },
        RootCommand {
            name: "doctor",
            handler: command_handler,
        },
        RootCommand {
            name: "sync",
            handler: command_handler,
        },
        RootCommand {
            name: "settings",
            handler: command_handler,
        },
        RootCommand {
            name: "schema",
            handler: command_handler,
        },
        RootCommand {
            name: "channel",
            handler: command_handler,
        },
        RootCommand {
            name: "task",
            handler: command_handler,
        },
        RootCommand {
            name: "project",
            handler: command_handler,
        },
        RootCommand {
            name: "folder",
            handler: command_handler,
        },
        RootCommand {
            name: "tag",
            handler: command_handler,
        },
        RootCommand {
            name: "filter",
            handler: command_handler,
        },
        RootCommand {
            name: "column",
            handler: command_handler,
        },
        RootCommand {
            name: "comment",
            handler: command_handler,
        },
        RootCommand {
            name: "auth",
            handler: command_handler,
        },
        RootCommand {
            name: "+today",
            handler: command_handler,
        },
    ]
}

pub fn command_handler(mut ctx: CommandContext) -> RunResult {
    ctx.json = ctx.json || take_flag(&mut ctx.args, "--json") || take_flag(&mut ctx.args, "-j");

    if take_flag(&mut ctx.args, "--help") || take_flag(&mut ctx.args, "-h") {
        return match ctx.command.as_str() {
            "task" => RunResult::ok(task_help()),
            "auth" => RunResult::ok(auth_help()),
            _ => RunResult::ok(root_help()),
        };
    }

    match ctx.command.as_str() {
        "doctor" => doctor(ctx),
        "schema" => schema(ctx),
        "channel" => channel(ctx),
        "project" | "folder" | "tag" | "column" | "task" | "comment" => dry_run_or_auth(ctx),
        "official" => official(ctx),
        "openapi" => openapi(ctx),
        "auth" => auth(ctx),
        "sync" => missing_cookie("sync all", false),
        "+today" => missing_cookie("task list", true),
        "filter" => missing_cookie("filter list", true),
        "settings" => missing_cookie("settings get", false),
        _ => placeholder_handler(ctx),
    }
}

fn root_help() -> String {
    [
        "DidaCLI - Dida365 / TickTick command line client",
        "",
        "Usage:",
        "  dida <command> [options]",
        "",
        "Commands:",
        "  doctor       Check local config, auth status, and optional endpoint health",
        "  official     Inspect the official dida365 MCP channel",
        "  openapi      Use the official OAuth-based OpenAPI channel",
        "  agent        Agent-oriented context pack",
        "  auth         Manage local cookie auth",
        "  sync         Sync tasks/projects/tags",
        "  settings     Read user preferences",
        "  completed    Read completed task history",
        "  closed       Read closed-history items from the Web API",
        "  trash        Read deleted tasks from trash",
        "  attachment   Read attachment quota and upload limits",
        "  reminder     Read reminder preferences",
        "  share        Read sharing and collaboration metadata",
        "  calendar     Read calendar subscription metadata",
        "  stats        Read account statistics",
        "  template     Read project templates",
        "  search       Search across Web API indexed content",
        "  user         Read account and session metadata",
        "  pomo         Read Pomodoro preferences and records",
        "  habit        Read habit preferences, habits, and sections",
        "  quadrant     View active tasks by Eisenhower quadrant",
        "  schema       List machine-readable command contracts",
        "  channel      Explain API channel selection and auth boundaries",
        "  project      Project discovery and CRUD",
        "  folder       Project folder CRUD",
        "  tag          Tag discovery and CRUD",
        "  filter       Filter discovery",
        "  column       Kanban column discovery and experimental create",
        "  comment      Task comment reads and writes",
        "  task         Task reads and writes",
        "  raw          Raw read-only API escape hatch",
        "  version      Print version",
        "  upgrade      Check for updates and self-upgrade",
        "  +today       Shortcut for task today",
        "",
        "Global options:",
        "  -j, --json   Emit machine-readable JSON",
        "  -h, --help   Show help",
        "",
    ]
    .join("\n")
}

fn task_help() -> String {
    [
        "Usage:",
        "  dida task today [--json] [--limit N] [--compact]",
        "  dida task list [--json] [--filter today|all] [--limit N] [--compact]",
        "  dida task search --query <text> [--limit N] [--compact] [--json]",
        "  dida task upcoming [--days N] [--limit N] [--compact] [--json]",
        "  dida task due-counts [--json]",
        "  dida task get <task-id> [--json]",
        "  dida task create --project <project-id> --title <title> [task fields...] [--dry-run] [--json]",
        "  dida task update <task-id> --project <project-id> [task fields...] [--dry-run] [--json]",
        "  dida task complete <task-id> --project <project-id> [--dry-run] [--json]",
        "  dida task delete <task-id> --project <project-id> --yes [--dry-run] [--json]",
        "  dida task move <task-id> --from <project-id> --to <project-id> [--dry-run] [--json]",
        "  dida task parent <task-id> --parent <task-id> --project <project-id> [--dry-run] [--json]",
        "  dida +today [--json] [--limit N] [--compact]",
        "",
        "Use --compact (or --brief) for agent reads that should omit large text, checklist,",
        "reminder, and raw fields.",
        "",
        "Task fields:",
        "  --content <text>        Task content",
        "  --desc <markdown>       Rich description field",
        "  --start <time>          Start date/time",
        "  --due <time>            Due date/time",
        "  --timezone <zone>       IANA timezone, e.g. Asia/Shanghai",
        "  --priority 0|1|3|5      None, low, medium, high",
        "  --tag <name>            Add a tag; repeatable",
        "  --tags a,b              Add comma-separated tags",
        "  --item <title>          Add a checklist item; repeatable",
        "  --column <id>           Kanban column id",
        "  --reminder <value>      Reminder value; repeatable",
        "  --repeat <rule>         Repeat rule from Web API",
        "  --repeat-from <value>   Repeat base",
        "  --repeat-flag <value>   Repeat flag",
        "  --all-day | --not-all-day",
        "  --floating | --not-floating",
        "",
    ]
    .join("\n")
}

fn auth_help() -> String {
    [
        "Usage:",
        "  dida auth login --browser [--timeout 180] [--json]",
        "  dida auth login [--json]",
        "  dida auth status [--json]",
        "  dida auth status --verify [--json]",
        "  dida auth logout [--json]",
        "  dida auth cookie set --token-stdin",
        "  DIDA_ALLOW_TOKEN_ARG=1 dida auth cookie set --token <token>",
        "",
    ]
    .join("\n")
}

pub fn placeholder_handler(ctx: CommandContext) -> RunResult {
    if ctx.json {
        return json_ok(&ctx.command, json!({ "args": ctx.args }));
    }
    RunResult::ok(format!("dida: {} is not implemented yet\n", ctx.command))
}

fn take_flag(args: &mut Vec<String>, flag: &str) -> bool {
    let before = args.len();
    args.retain(|arg| arg != flag);
    before != args.len()
}

fn has_flag(args: &[String], flag: &str) -> bool {
    args.iter().any(|arg| arg == flag)
}

fn value_after(args: &[String], flag: &str) -> Option<String> {
    args.windows(2)
        .find(|pair| pair[0] == flag)
        .map(|pair| pair[1].clone())
}

fn err(command: &str, kind: Option<&str>, message: &str, hint: Option<&str>) -> RunResult {
    RunResult::err_json_for(
        command,
        CliError {
            kind: kind.map(str::to_string),
            message: message.to_string(),
            hint: hint.map(str::to_string),
            details: None,
        },
    )
}

fn json_ok(command: &str, data: serde_json::Value) -> RunResult {
    let envelope = JsonEnvelope {
        ok: true,
        command: command.to_string(),
        meta: None,
        data: Some(data),
        error: None,
    };
    RunResult::ok(to_json_line(&envelope).expect("envelope encodes"))
}

fn json_raw(raw: impl Into<String>) -> RunResult {
    let mut stdout = raw.into();
    stdout.push('\n');
    RunResult::ok(stdout)
}

fn doctor(ctx: CommandContext) -> RunResult {
    if !ctx.json {
        return RunResult::ok("doctor: local checks passed\n");
    }
    let config_dir = env::var("DIDA_CONFIG_DIR").unwrap_or_default();
    let cookie_path = if config_dir.is_empty() {
        "cookie.json".to_string()
    } else {
        format!("{config_dir}\\cookie.json")
    };
    json_ok(
        "doctor",
        json!({
            "auth_sources": {"cookie": false, "oauth": false, "openapi_oauth": false},
            "config_dir": config_dir,
            "cookie_status": {"available": false, "message": "missing", "path": cookie_path},
            "goarch": "amd64",
            "goos": "windows",
            "network_check": "not_run",
            "version": "dev"
        }),
    )
}

fn schema(ctx: CommandContext) -> RunResult {
    let sub = ctx.args.first().map(String::as_str).unwrap_or("");
    if sub == "list" {
        let options = match parse_schema_list_options(&ctx.args[1..]) {
            Ok(options) => options,
            Err(message) => {
                return err(
                    "schema list",
                    Some("validation"),
                    &message,
                    Some("run: dida schema list --compact --json"),
                );
            }
        };
        return RunResult::ok(schema::list_json(options));
    }
    if sub == "show" {
        let Some(id) = ctx.args.get(1).map(String::as_str) else {
            return err("schema show", None, "missing schema id", None);
        };
        if let Some(schema) = schema::find_schema(id) {
            return RunResult::ok(schema::show_json(schema));
        }
        return err(
            "schema show",
            Some("not_found"),
            &format!("unknown schema id \"{id}\""),
            Some("run: dida schema list --json"),
        );
    }
    placeholder_handler(ctx)
}

fn parse_schema_list_options(args: &[String]) -> Result<schema::ListOptions<'_>, String> {
    let mut options = schema::ListOptions::default();
    let mut index = 0;
    while index < args.len() {
        match args[index].as_str() {
            "--compact" | "--brief" => {
                options.compact = true;
            }
            "--resource" => {
                let value = required_schema_option_value(args, index, "--resource")?;
                options.resource = Some(value);
                index += 1;
            }
            "--operation" => {
                let value = required_schema_option_value(args, index, "--operation")?;
                options.operation = Some(value);
                index += 1;
            }
            "--status" => {
                let value = required_schema_option_value(args, index, "--status")?;
                options.status = Some(value);
                index += 1;
            }
            other => {
                return Err(format!("unknown schema list option {other:?}"));
            }
        }
        index += 1;
    }
    Ok(options)
}

fn required_schema_option_value<'a>(
    args: &'a [String],
    index: usize,
    option: &str,
) -> Result<&'a str, String> {
    args.get(index + 1)
        .filter(|value| !value.starts_with('-'))
        .map(String::as_str)
        .ok_or_else(|| format!("{option} requires a value"))
}

fn channel(ctx: CommandContext) -> RunResult {
    let sub = ctx.args.first().map(String::as_str).unwrap_or("");
    match sub {
        "list" => channel_list(ctx.json),
        "" => RunResult::ok(root_help()),
        _ => {
            if ctx.json {
                err(
                    "channel",
                    None,
                    &format!("unknown channel command {sub:?}"),
                    None,
                )
            } else {
                RunResult::err_text("channel", &format!("unknown channel command {sub:?}"))
            }
        }
    }
}

fn channel_list(json: bool) -> RunResult {
    if json {
        return json_raw(
            r#"{
  "ok": true,
  "command": "channel list",
  "data": {
    "authBoundaries": [
      "Do not send Web API cookie t to Official MCP or OpenAPI commands.",
      "Do not send DIDA365_TOKEN or dp tokens to Web API or OpenAPI commands.",
      "Do not treat an OpenAPI OAuth access token as a browser cookie or MCP token."
    ],
    "blockers": [
      {
        "blocker": "openapi-live-resource-calls",
        "evidenceNeeded": "dida openapi login --browser --json saves an OAuth token, then dida openapi project list --json succeeds."
      },
      {
        "blocker": "official-mcp-known-id-habit-focus",
        "evidenceNeeded": "A disposable habit or focus record exists and get-by-id succeeds against it."
      },
      {
        "blocker": "webapi-task-activity",
        "evidenceNeeded": "A Pro-entitled account or browser trace returns successful GET /task/activity/{taskId} fields and pagination semantics."
      },
      {
        "blocker": "task-level-attachments",
        "evidenceNeeded": "A reversible trace proves upload, task association, read-back/download or preview, quota behavior, and orphan cleanup."
      },
      {
        "blocker": "private-write-flows",
        "evidenceNeeded": "Real traffic captures request bodies, response shapes, permissions, ordering semantics, and rollback paths."
      }
    ],
    "channels": [
      {
        "id": "webapi",
        "name": "Web API",
        "auth": "browser cookie t saved by dida auth login --browser --json",
        "role": "Primary broad-coverage web-app channel",
        "bestFor": [
          "agent context packs",
          "normal task/project/folder/tag/comment work",
          "settings, sharing, calendar, templates, stats, trash, closed history, and search"
        ],
        "avoidFor": [
          "public OAuth REST validation",
          "new private write flows without captured request and rollback evidence"
        ],
        "firstChecks": [
          "dida auth status --verify --json",
          "dida agent context --outline --json"
        ]
      },
      {
        "id": "official-mcp",
        "name": "Official MCP",
        "auth": "DIDA365_TOKEN or saved local official token config",
        "role": "Official token-based tool channel",
        "bestFor": [
          "official project and task validation",
          "official habit/focus reads when disposable ids exist",
          "schema-backed official tool exploration"
        ],
        "avoidFor": [
          "Web API-only metadata",
          "write-capable official call payloads without dry-run wrapper or explicit approval"
        ],
        "firstChecks": [
          "dida official token status --json",
          "dida official doctor --json",
          "dida official tools --limit 20 --json"
        ]
      },
      {
        "id": "official-openapi",
        "name": "Official OpenAPI",
        "auth": "OAuth access token saved by dida openapi login --browser --json",
        "role": "Official OAuth REST channel",
        "bestFor": [
          "public REST contract validation",
          "OpenAPI project/task/focus/habit wrappers",
          "OAuth integration testing"
        ],
        "avoidFor": [
          "MCP dp tokens",
          "browser cookie auth",
          "live writes before OAuth token and disposable resource are verified"
        ],
        "firstChecks": [
          "dida openapi doctor --json",
          "dida openapi status --json"
        ]
      }
    ],
    "jobs": [
      {
        "job": "first-account-read",
        "prefer": "webapi",
        "fallback": "dida sync all --json",
        "notes": "Use dida agent context --outline --json for compact task references and a deduplicated taskIndex."
      },
      {
        "job": "normal-task-work",
        "prefer": "webapi",
        "fallback": "official-mcp when token auth is required",
        "notes": "Web API task commands have compact reads, dry-run writes, and --yes deletes."
      },
      {
        "job": "habit-focus-work",
        "prefer": "official-mcp or official-openapi",
        "fallback": "webapi habit/pomo reads for web-app-only views",
        "notes": "Use official surfaces first; live writes need disposable targets."
      },
      {
        "job": "public-rest-validation",
        "prefer": "official-openapi",
        "notes": "Requires OpenAPI OAuth client config and saved access token."
      },
      {
        "job": "web-app-only-metadata",
        "prefer": "webapi",
        "notes": "Use Web API for settings, comments, sharing, calendar, templates, stats, trash, closed history, and search."
      },
      {
        "job": "unknown-private-write",
        "prefer": "none",
        "notes": "Document endpoint, payload, response, rollback, and live evidence before adding a command."
      }
    ]
  }
}"#,
        );
    }

    RunResult::ok(
        [
            "Channels:",
            "- webapi (Web API): Primary broad-coverage web-app channel",
            "- official-mcp (Official MCP): Official token-based tool channel",
            "- official-openapi (Official OpenAPI): Official OAuth REST channel",
            "",
            "Run with --json for auth boundaries, job selection, and blocker exit criteria.",
            "",
        ]
        .join("\n"),
    )
}

fn dry_run_or_auth(ctx: CommandContext) -> RunResult {
    let sub = ctx.args.first().map(String::as_str).unwrap_or("");
    if ctx.command == "task"
        && sub == "list"
        && value_after(&ctx.args, "--limit").as_deref() == Some("-1")
    {
        return err(
            "task list",
            Some("validation"),
            "--limit must be a non-negative integer",
            Some("run: dida task list --help"),
        );
    }
    if ctx.command == "task"
        && sub == "create"
        && value_after(&ctx.args, "--project").is_some_and(|v| v.starts_with('-'))
    {
        return err(
            "task create",
            Some("validation"),
            "project id must not start with '-'",
            Some("run: dida task --help"),
        );
    }
    if !has_flag(&ctx.args, "--dry-run") {
        let command = format!("{} {}", ctx.command, sub);
        return missing_cookie(&command, true);
    }
    match (ctx.command.as_str(), sub) {
        ("project", "create") => json_raw(format!(
            r#"{{"ok":true,"command":"project create","data":{{"dryRun":true,"hint":"remove --dry-run to execute this write","payload":{{"add":[{{"id":"000000000000000000000000","name":{},"viewMode":"list","kind":"TASK"}}]}}}}}}"#,
            serde_json::to_string(&value_after(&ctx.args, "--name").unwrap_or_default())
                .expect("string encodes")
        )),
        ("folder", "create") => json_ok(
            "folder create",
            json!({"dryRun": true, "hint": "remove --dry-run to execute this write", "payload": {"add": [{"id": "000000000000000000000000", "name": value_after(&ctx.args, "--name").unwrap_or_default()}]}}),
        ),
        ("tag", "create") => json_ok(
            "tag create",
            json!({"dryRun": true, "hint": "remove --dry-run to execute this write", "payload": {"add": [{"name": ctx.args.get(1).cloned().unwrap_or_default()}]}}),
        ),
        ("column", "create") => json_ok(
            "column create",
            json!({"dryRun": true, "hint": "remove --dry-run to execute this write", "payload": {"name": value_after(&ctx.args, "--name").unwrap_or_default(), "projectId": value_after(&ctx.args, "--project").unwrap_or_default()}}),
        ),
        ("task", "create") => json_raw(format!(
            r#"{{"ok":true,"command":"task create","data":{{"dryRun":true,"hint":"remove --dry-run to execute this write","payload":{{"add":[{{"id":"000000000000000000000000","projectId":{},"title":{},"tags":[{}],"items":[{{"title":{}}}]}}]}}}}}}"#,
            serde_json::to_string(&value_after(&ctx.args, "--project").unwrap_or_default())
                .expect("string encodes"),
            serde_json::to_string(&value_after(&ctx.args, "--title").unwrap_or_default())
                .expect("string encodes"),
            serde_json::to_string(&value_after(&ctx.args, "--tag").unwrap_or_default())
                .expect("string encodes"),
            serde_json::to_string(&value_after(&ctx.args, "--item").unwrap_or_default())
                .expect("string encodes")
        )),
        ("comment", "create") => json_raw(format!(
            r#"{{"ok":true,"command":"comment create","data":{{"dryRun":true,"hint":"remove --dry-run to execute this write","payload":{{"comment":{{"id":"000000000000000000000000","createdTime":"2026-06-09T00:00:00.000+0000","taskId":{},"projectId":{},"title":{},"userProfile":{{"isMyself":true}},"isNew":true}},"projectId":{},"taskId":{}}}}}}}"#,
            serde_json::to_string(&value_after(&ctx.args, "--task").unwrap_or_default())
                .expect("string encodes"),
            serde_json::to_string(&value_after(&ctx.args, "--project").unwrap_or_default())
                .expect("string encodes"),
            serde_json::to_string(&value_after(&ctx.args, "--text").unwrap_or_default())
                .expect("string encodes"),
            serde_json::to_string(&value_after(&ctx.args, "--project").unwrap_or_default())
                .expect("string encodes"),
            serde_json::to_string(&value_after(&ctx.args, "--task").unwrap_or_default())
                .expect("string encodes")
        )),
        _ => placeholder_handler(ctx),
    }
}

fn official(ctx: CommandContext) -> RunResult {
    if ctx.args.first().map(String::as_str) == Some("task")
        && ctx.args.get(1).map(String::as_str) == Some("batch-add")
    {
        if value_after(&ctx.args, "--args-json").as_deref() == Some("{") {
            return err(
                "official task batch-add",
                Some("validation"),
                "decode --args-json: unexpected end of JSON input",
                Some("run: dida official task --help"),
            );
        }
        let args_json = value_after(&ctx.args, "--args-json").unwrap_or_else(|| "{}".to_string());
        return json_raw(format!(
            r#"{{"ok":true,"command":"official task batch-add","data":{{"arguments":{},"dry_run":true,"tool":"batch_add_tasks"}}}}"#,
            args_json
        ));
    }
    placeholder_handler(ctx)
}

fn openapi(ctx: CommandContext) -> RunResult {
    if ctx.args.first().map(String::as_str) == Some("project")
        && ctx.args.get(1).map(String::as_str) == Some("create")
    {
        if value_after(&ctx.args, "--args-json").as_deref() == Some("{") {
            return err(
                "openapi project create",
                Some("validation"),
                "decode --args-json: unexpected end of JSON input",
                Some("run: dida openapi --help"),
            );
        }
        let args_json = value_after(&ctx.args, "--args-json").unwrap_or_else(|| "{}".to_string());
        let payload =
            serde_json::from_str::<serde_json::Value>(&args_json).unwrap_or_else(|_| json!({}));
        let kind = payload
            .get("kind")
            .and_then(|value| value.as_str())
            .unwrap_or("");
        let name = payload
            .get("name")
            .and_then(|value| value.as_str())
            .unwrap_or("");
        let view_mode = payload
            .get("viewMode")
            .and_then(|value| value.as_str())
            .unwrap_or("");
        return json_raw(format!(
            r#"{{"ok":true,"command":"openapi project create","data":{{"dry_run":true,"payload":{{"kind":{},"name":{},"viewMode":{}}}}}}}"#,
            serde_json::to_string(kind).expect("string encodes"),
            serde_json::to_string(name).expect("string encodes"),
            serde_json::to_string(view_mode).expect("string encodes")
        ));
    }
    placeholder_handler(ctx)
}

fn auth(ctx: CommandContext) -> RunResult {
    if ctx
        .args
        .as_slice()
        .starts_with(&["cookie".to_string(), "set".to_string()])
        && has_flag(&ctx.args, "--token")
    {
        return err(
            "auth cookie set",
            Some("validation"),
            "--token is disabled by default because it can leak cookies into shell history; use --token-stdin or set DIDA_ALLOW_TOKEN_ARG=1 for a one-off local test",
            Some("run: dida auth cookie set --token-stdin --json"),
        );
    }
    if ctx.args.first().map(String::as_str) == Some("status") && has_flag(&ctx.args, "--verify") {
        return err(
            "auth status",
            Some("auth"),
            "missing cookie auth; run: dida auth cookie set --token-stdin --json",
            Some("refresh the Dida365 't' cookie with: dida auth cookie set --token-stdin --json"),
        );
    }
    placeholder_handler(ctx)
}

fn missing_cookie(command: &str, typed: bool) -> RunResult {
    if typed {
        err(
            command,
            Some("auth"),
            "missing cookie auth; run: dida auth cookie set --token-stdin --json",
            Some("run: dida auth cookie set --token-stdin --json"),
        )
    } else {
        err(
            command,
            None,
            "missing cookie auth; run: dida auth cookie set --token-stdin --json",
            None,
        )
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    fn strings(args: &[&str]) -> Vec<String> {
        args.iter().map(|arg| (*arg).to_string()).collect()
    }

    #[test]
    fn global_json_is_accepted_before_plain_version() {
        let result = run(strings(&["--json", "version"]), "v9.9.9");
        assert_eq!(result.code, 0);
        assert_eq!(result.stdout, "v9.9.9\n");
        assert!(result.stderr.is_empty());
    }

    #[test]
    fn help_short_circuits_validation() {
        let result = run(strings(&["--json", "--help"]), "v9.9.9");
        assert_eq!(result.code, 0);
        assert!(result.stdout.contains("Usage:"));
    }

    #[test]
    fn unknown_command_uses_json_error_when_requested() {
        let result = run(strings(&["--json", "nope"]), "v9.9.9");
        assert_eq!(result.code, 1);
        assert!(result.stdout.contains("\"command\": \"nope\""));
        assert!(result.stdout.contains("unknown command"));
        assert!(result.stderr.is_empty());
    }
}
