package cli

import (
	"fmt"
	"io"
)

type rootCommand struct {
	Name string
	Run  func(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int
}

func Run(args []string, version string, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		printHelp(stdout)
		return 0
	}
	if args[0] == "--version" || args[0] == "version" {
		fmt.Fprintln(stdout, version)
		return 0
	}

	jsonOut, args := consumeJSONFlag(args)
	if len(args) == 0 {
		if jsonOut {
			return failTyped("dida", "validation", "missing command", "run: dida --help", jsonOut, stdout, stderr)
		}
		printHelp(stdout)
		return 1
	}
	command := args[0]

	for _, cmd := range rootCommands(version) {
		if cmd.Name == command {
			return cmd.Run(args[1:], jsonOut, stdout, stderr)
		}
	}
	return fail(command, fmt.Sprintf("unknown command %q", command), jsonOut, stdout, stderr)
}

func rootCommands(version string) []rootCommand {
	return []rootCommand{
		{
			Name: "+today",
			Run: func(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
				return runTask(append([]string{"today"}, args...), jsonOut, stdout, stderr)
			},
		},
		{
			Name: "doctor",
			Run: func(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
				return runDoctor(args, version, jsonOut, stdout, stderr)
			},
		},
		{Name: "official", Run: runOfficial},
		{Name: "openapi", Run: runOpenAPI},
		{Name: "agent", Run: runAgent},
		{Name: "auth", Run: runAuth},
		{Name: "sync", Run: runSync},
		{Name: "settings", Run: runSettings},
		{Name: "completed", Run: runCompleted},
		{Name: "closed", Run: runClosed},
		{Name: "trash", Run: runTrash},
		{Name: "attachment", Run: runAttachment},
		{Name: "reminder", Run: runReminder},
		{Name: "share", Run: runShare},
		{Name: "calendar", Run: runCalendar},
		{Name: "stats", Run: runStats},
		{Name: "template", Run: runTemplate},
		{Name: "search", Run: runSearch},
		{Name: "user", Run: runUser},
		{Name: "pomo", Run: runPomo},
		{Name: "habit", Run: runHabit},
		{Name: "quadrant", Run: runQuadrant},
		{Name: "schema", Run: runSchema},
		{Name: "channel", Run: runChannel},
		{Name: "raw", Run: runRaw},
		{Name: "project", Run: runProject},
		{Name: "folder", Run: runFolder},
		{Name: "tag", Run: runTag},
		{Name: "filter", Run: runFilter},
		{Name: "column", Run: runColumn},
		{Name: "comment", Run: runComment},
		{Name: "task", Run: runTask},
		{
			Name: "upgrade",
			Run: func(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
				versionFromBuild = version
				return runUpgrade(args, jsonOut, stdout, stderr)
			},
		},
	}
}

func consumeJSONFlag(args []string) (bool, []string) {
	out := args[:0]
	jsonOut := false
	for _, arg := range args {
		if arg == "--json" || arg == "-j" {
			jsonOut = true
			continue
		}
		out = append(out, arg)
	}
	return jsonOut, out
}

func hasHelpFlag(args []string) bool {
	for _, arg := range args {
		if arg == "-h" || arg == "--help" {
			return true
		}
	}
	return false
}
