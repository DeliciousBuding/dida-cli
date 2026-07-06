package cli

import (
	"fmt"
	"io"
	"sort"
	"strings"
)

type completionCommand struct {
	Name        string
	Description string
}

func runCompletion(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		printCompletionHelp(stdout)
		return 0
	}
	if jsonOut {
		return failTyped("completion", "validation", "completion emits shell script and does not support --json", "run: dida completion <bash|zsh|fish|powershell>", true, stdout, stderr)
	}
	if len(args) != 1 {
		return failTyped("completion", "validation", "usage: dida completion <bash|zsh|fish|powershell>", "run: dida completion --help", jsonOut, stdout, stderr)
	}
	switch args[0] {
	case "bash":
		fmt.Fprint(stdout, bashCompletionScript())
	case "zsh":
		fmt.Fprint(stdout, zshCompletionScript())
	case "fish":
		fmt.Fprint(stdout, fishCompletionScript())
	case "powershell":
		fmt.Fprint(stdout, powerShellCompletionScript())
	default:
		return failTyped("completion", "validation", fmt.Sprintf("unknown completion shell %q", args[0]), "supported shells: bash, zsh, fish, powershell", jsonOut, stdout, stderr)
	}
	return 0
}

func completionCommands() []completionCommand {
	commands := []completionCommand{
		{Name: "+today", Description: "Shortcut for task today"},
		{Name: "agent", Description: "Build an agent context pack"},
		{Name: "attachment", Description: "Read attachment quota and download task attachments"},
		{Name: "auth", Description: "Manage local cookie auth"},
		{Name: "calendar", Description: "Read calendar subscription metadata"},
		{Name: "channel", Description: "Explain auth channel selection"},
		{Name: "closed", Description: "Read closed-history items"},
		{Name: "column", Description: "Read and create Kanban columns"},
		{Name: "comment", Description: "Read and write task comments"},
		{Name: "completed", Description: "Read completed task history"},
		{Name: "completion", Description: "Generate shell completion script"},
		{Name: "doctor", Description: "Check local config and auth health"},
		{Name: "filter", Description: "Read saved filters"},
		{Name: "folder", Description: "Manage project folders"},
		{Name: "habit", Description: "Read habit preferences and records"},
		{Name: "official", Description: "Use the official MCP channel"},
		{Name: "openapi", Description: "Use the official OpenAPI channel"},
		{Name: "pomo", Description: "Read Pomodoro records"},
		{Name: "project", Description: "Read and manage projects"},
		{Name: "quadrant", Description: "Group tasks by Eisenhower quadrant"},
		{Name: "raw", Description: "Run read-only Web API probes"},
		{Name: "reminder", Description: "Read reminder preferences"},
		{Name: "schema", Description: "List machine-readable command contracts"},
		{Name: "search", Description: "Search indexed content"},
		{Name: "settings", Description: "Read user preferences"},
		{Name: "share", Description: "Read sharing metadata"},
		{Name: "stats", Description: "Read account statistics"},
		{Name: "sync", Description: "Sync tasks and projects"},
		{Name: "tag", Description: "Read and manage tags"},
		{Name: "task", Description: "Read and write tasks"},
		{Name: "template", Description: "Read project templates"},
		{Name: "trash", Description: "Read deleted tasks"},
		{Name: "upgrade", Description: "Check for updates and self-upgrade"},
		{Name: "user", Description: "Read user profile and sessions"},
		{Name: "version", Description: "Print version"},
	}
	sort.Slice(commands, func(i, j int) bool { return commands[i].Name < commands[j].Name })
	return commands
}

func completionCommandNames() string {
	names := make([]string, 0, len(completionCommands()))
	for _, command := range completionCommands() {
		names = append(names, command.Name)
	}
	return strings.Join(names, " ")
}

func completionShellNames() string {
	return "bash zsh fish powershell"
}

func bashCompletionScript() string {
	return fmt.Sprintf(`# bash completion for dida
_dida_completion() {
  local cur cword
  COMPREPLY=()
  cur="${COMP_WORDS[COMP_CWORD]}"
  cword="${COMP_CWORD}"

  if [[ "${cword}" -eq 1 ]]; then
    COMPREPLY=( $(compgen -W "%s" -- "${cur}") )
    return 0
  fi

  if [[ "${COMP_WORDS[1]}" == "completion" && "${cword}" -eq 2 ]]; then
    COMPREPLY=( $(compgen -W "%s" -- "${cur}") )
    return 0
  fi

  if [[ "${cur}" == -* ]]; then
    COMPREPLY=( $(compgen -W "--json --help" -- "${cur}") )
  fi
}
complete -F _dida_completion dida
`, completionCommandNames(), completionShellNames())
}

func zshCompletionScript() string {
	return fmt.Sprintf(`#compdef dida

_dida() {
  local -a commands shells
  commands=(%s)
  shells=(%s)

  if (( CURRENT == 2 )); then
    _describe 'command' commands
    return
  fi

  if [[ "${words[2]}" == "completion" && CURRENT == 3 ]]; then
    _describe 'shell' shells
    return
  fi

  _arguments '*: :_files'
}

_dida "$@"
`, zshWords(completionCommands()), zshShellWords())
}

func fishCompletionScript() string {
	var b strings.Builder
	b.WriteString("# fish completion for dida\n")
	for _, command := range completionCommands() {
		fmt.Fprintf(&b, "complete -c dida -f -n '__fish_use_subcommand' -a '%s' -d '%s'\n", fishQuote(command.Name), fishQuote(command.Description))
	}
	for _, shell := range strings.Fields(completionShellNames()) {
		fmt.Fprintf(&b, "complete -c dida -f -n '__fish_seen_subcommand_from completion' -a '%s'\n", shell)
	}
	b.WriteString("complete -c dida -f -l json -d 'Emit JSON when supported'\n")
	b.WriteString("complete -c dida -f -l help -d 'Show help'\n")
	return b.String()
}

func powerShellCompletionScript() string {
	return fmt.Sprintf(`# PowerShell completion for dida
$DidaCommands = @(%s)
$DidaCompletionShells = @(%s)
Register-ArgumentCompleter -Native -CommandName dida -ScriptBlock {
  param($wordToComplete, $commandAst, $cursorPosition)
  $words = $commandAst.CommandElements | ForEach-Object { $_.ToString() }
  $values = if ($words.Count -ge 2 -and $words[1] -eq 'completion') { $DidaCompletionShells } else { $DidaCommands }
  $values |
    Where-Object { $_ -like "$wordToComplete*" } |
    ForEach-Object { [System.Management.Automation.CompletionResult]::new($_, $_, 'ParameterValue', $_) }
}
`, powerShellArray(completionCommands()), powerShellShellArray())
}

func zshWords(commands []completionCommand) string {
	words := make([]string, 0, len(commands))
	for _, command := range commands {
		words = append(words, fmt.Sprintf(`'%s:%s'`, zshQuote(command.Name), zshQuote(command.Description)))
	}
	return strings.Join(words, " ")
}

func zshShellWords() string {
	words := []string{}
	for _, shell := range strings.Fields(completionShellNames()) {
		words = append(words, fmt.Sprintf("'%s'", zshQuote(shell)))
	}
	return strings.Join(words, " ")
}

func powerShellArray(commands []completionCommand) string {
	words := make([]string, 0, len(commands))
	for _, command := range commands {
		words = append(words, fmt.Sprintf("'%s'", powerShellQuote(command.Name)))
	}
	return strings.Join(words, ", ")
}

func powerShellShellArray() string {
	words := []string{}
	for _, shell := range strings.Fields(completionShellNames()) {
		words = append(words, fmt.Sprintf("'%s'", powerShellQuote(shell)))
	}
	return strings.Join(words, ", ")
}

func zshQuote(value string) string {
	return strings.ReplaceAll(value, `'`, `'\''`)
}

func fishQuote(value string) string {
	return strings.ReplaceAll(value, `'`, `\'`)
}

func powerShellQuote(value string) string {
	return strings.ReplaceAll(value, `'`, `''`)
}
