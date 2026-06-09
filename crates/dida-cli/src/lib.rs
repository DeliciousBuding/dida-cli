use dida_core::{CliError, JsonEnvelope, failure, to_json_line};
use serde_json::json;

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
            let err = CliError::validation(&command, format!("unknown command {command:?}"))
                .with_hint("run: dida --help");
            if json {
                RunResult::err_json(err)
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
            name: "doctor",
            handler: placeholder_handler,
        },
        RootCommand {
            name: "task",
            handler: placeholder_handler,
        },
        RootCommand {
            name: "project",
            handler: placeholder_handler,
        },
        RootCommand {
            name: "auth",
            handler: placeholder_handler,
        },
    ]
}

pub fn placeholder_handler(ctx: CommandContext) -> RunResult {
    if ctx.json {
        let envelope = JsonEnvelope {
            ok: true,
            command: ctx.command,
            meta: Some(json!({ "status": "placeholder" })),
            data: Some(json!({ "args": ctx.args })),
            error: None,
        };
        return RunResult::ok(to_json_line(&envelope).expect("placeholder envelope encodes"));
    }
    RunResult::ok(format!("dida: {} is not implemented yet\n", ctx.command))
}

fn root_help() -> String {
    [
        "DidaCLI",
        "",
        "Usage:",
        "  dida [--json] <command>",
        "",
        "Commands:",
        "  version    Print CLI version",
        "  task       Manage tasks",
        "  project    Manage projects",
        "  auth       Manage authentication",
        "",
    ]
    .join("\n")
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
        assert!(result.stdout.contains("\"type\": \"validation\""));
        assert!(result.stderr.is_empty());
    }
}
