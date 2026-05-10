package cli

import (
	"fmt"
	"io"
	"time"

	"github.com/DeliciousBuding/dida-cli/internal/model"
)

type agentContextOptions struct {
	Days    int
	Limit   int
	Compact bool
	Outline bool
}

type agentProject struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Closed bool   `json:"closed,omitempty"`
	Color  string `json:"color,omitempty"`
}

type agentQuadrant struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Tasks       any    `json:"tasks"`
	Count       int    `json:"count"`
	Total       int    `json:"total"`
}

type agentQuadrantRefs struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	TaskIDs     []string `json:"taskIds"`
	Count       int      `json:"count"`
	Total       int      `json:"total"`
}

func runAgent(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		printAgentHelp(stdout)
		return 0
	}
	switch args[0] {
	case "context":
		return runAgentContext(args[1:], jsonOut, stdout, stderr)
	default:
		return fail("agent", fmt.Sprintf("unknown agent command %q", args[0]), jsonOut, stdout, stderr)
	}
}

func runAgentContext(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	if len(args) > 0 && (args[0] == "-h" || args[0] == "--help") {
		printAgentHelp(stdout)
		return 0
	}
	opts, err := parseAgentContextFlags(args)
	if err != nil {
		return failTyped("agent context", "validation", err.Error(), "run: dida agent --help", jsonOut, stdout, stderr)
	}
	view, err := loadSyncView()
	if err != nil {
		return failTyped("agent context", "auth", err.Error(), "run: dida auth login --browser --json", jsonOut, stdout, stderr)
	}
	data, meta := buildAgentContext(view, opts, time.Now())
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "agent context", Meta: meta, Data: data})
	}
	fmt.Fprintf(stdout, "Projects: %d\n", meta["projects"])
	fmt.Fprintf(stdout, "Today: %d of %d task(s)\n", meta["today"], meta["todayTotal"])
	fmt.Fprintf(stdout, "Upcoming: %d of %d task(s)\n", meta["upcoming"], meta["upcomingTotal"])
	return 0
}

func parseAgentContextFlags(args []string) (agentContextOptions, error) {
	opts := agentContextOptions{Days: 14, Limit: 50, Compact: true}
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--days":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("--days requires a value")
			}
			if _, err := fmt.Sscanf(args[i+1], "%d", &opts.Days); err != nil || opts.Days <= 0 {
				return opts, fmt.Errorf("--days must be a positive integer")
			}
			i++
		case "--limit":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("--limit requires a value")
			}
			if _, err := fmt.Sscanf(args[i+1], "%d", &opts.Limit); err != nil || opts.Limit < 0 {
				return opts, fmt.Errorf("--limit must be a non-negative integer")
			}
			i++
		case "--compact", "--brief":
			opts.Compact = true
			opts.Outline = false
		case "--full":
			opts.Compact = false
			opts.Outline = false
		case "--outline", "--refs":
			opts.Compact = true
			opts.Outline = true
		default:
			return opts, fmt.Errorf("unknown flag %q", args[i])
		}
	}
	return opts, nil
}

func buildAgentContext(view model.SyncView, opts agentContextOptions, now time.Time) (map[string]any, map[string]any) {
	todayAll := model.TodayTasks(view.Tasks, now)
	upcomingAll := model.UpcomingTasks(view.Tasks, now, opts.Days)
	today := limitTasks(todayAll, opts.Limit)
	upcoming := limitTasks(upcomingAll, opts.Limit)
	active := model.ActiveTasks(view.Tasks)
	quadrants := buildAgentQuadrants(active, opts)
	data := map[string]any{
		"inboxId":         view.InboxID,
		"compact":         opts.Compact,
		"outline":         opts.Outline,
		"days":            opts.Days,
		"limit":           opts.Limit,
		"projects":        agentProjects(view.Projects),
		"projectGroups":   compactMapList(view.ProjectGroups),
		"tags":            compactMapList(view.Tags),
		"filters":         compactMapList(view.Filters),
		"today":           taskOutput(today, opts.Compact),
		"upcoming":        taskOutput(upcoming, opts.Compact),
		"quadrants":       quadrants,
		"recommendedNext": []string{"dida task get <task-id> --json", "dida schema show task.update --json"},
	}
	if opts.Outline {
		quadrantRefs, quadrantTasks := buildAgentQuadrantRefs(active, opts)
		indexTasks := append([]model.Task{}, today...)
		indexTasks = append(indexTasks, upcoming...)
		indexTasks = append(indexTasks, quadrantTasks...)
		data["taskIndex"] = compactTaskIndex(indexTasks)
		data["today"] = taskIDs(today)
		data["upcoming"] = taskIDs(upcoming)
		data["quadrants"] = quadrantRefs
	}
	meta := map[string]any{
		"projects":      len(view.Projects),
		"projectGroups": len(view.ProjectGroups),
		"tags":          len(view.Tags),
		"filters":       len(view.Filters),
		"today":         len(today),
		"todayTotal":    len(todayAll),
		"upcoming":      len(upcoming),
		"upcomingTotal": len(upcomingAll),
		"tasks":         len(view.Tasks),
		"activeTasks":   len(active),
		"outline":       opts.Outline,
	}
	if opts.Outline {
		if taskIndex, ok := data["taskIndex"].(map[string]compactTask); ok {
			meta["taskIndex"] = len(taskIndex)
		}
	}
	return data, meta
}

func agentProjects(projects []model.Project) []agentProject {
	out := make([]agentProject, 0, len(projects))
	for _, project := range projects {
		out = append(out, agentProject{ID: project.ID, Name: project.Name, Closed: project.Closed, Color: project.Color})
	}
	return out
}

func buildAgentQuadrants(tasks []model.Task, opts agentContextOptions) []agentQuadrant {
	buckets := orderedQuadrants(buildQuadrants(tasks))
	out := make([]agentQuadrant, 0, len(buckets))
	for _, bucket := range buckets {
		tasks := limitTasks(bucket.Tasks, opts.Limit)
		out = append(out, agentQuadrant{
			ID:          bucket.ID,
			Name:        bucket.Name,
			Description: bucket.Description,
			Tasks:       taskOutput(tasks, opts.Compact),
			Count:       len(tasks),
			Total:       len(bucket.Tasks),
		})
	}
	return out
}

func buildAgentQuadrantRefs(tasks []model.Task, opts agentContextOptions) ([]agentQuadrantRefs, []model.Task) {
	buckets := orderedQuadrants(buildQuadrants(tasks))
	out := make([]agentQuadrantRefs, 0, len(buckets))
	indexTasks := make([]model.Task, 0)
	for _, bucket := range buckets {
		tasks := limitTasks(bucket.Tasks, opts.Limit)
		indexTasks = append(indexTasks, tasks...)
		out = append(out, agentQuadrantRefs{
			ID:          bucket.ID,
			Name:        bucket.Name,
			Description: bucket.Description,
			TaskIDs:     taskIDs(tasks),
			Count:       len(tasks),
			Total:       len(bucket.Tasks),
		})
	}
	return out, indexTasks
}

func taskIDs(tasks []model.Task) []string {
	out := make([]string, 0, len(tasks))
	for _, task := range tasks {
		out = append(out, task.ID)
	}
	return out
}

func compactTaskIndex(tasks []model.Task) map[string]compactTask {
	out := make(map[string]compactTask)
	for _, task := range tasks {
		if task.ID == "" {
			continue
		}
		if _, exists := out[task.ID]; exists {
			continue
		}
		out[task.ID] = compactTaskFromTask(task)
	}
	return out
}

func compactMapList(items []map[string]any) []map[string]any {
	keys := []string{"id", "name", "title", "label", "color", "parent", "parentId", "sortOrder", "sortType", "kind", "viewMode", "status", "projectId", "groupId"}
	out := make([]map[string]any, 0, len(items))
	for _, item := range items {
		compact := make(map[string]any)
		for _, key := range keys {
			if value, ok := item[key]; ok {
				compact[key] = value
			}
		}
		if len(compact) == 0 {
			continue
		}
		out = append(out, compact)
	}
	return out
}

func limitTasks(tasks []model.Task, limit int) []model.Task {
	if limit > 0 && len(tasks) > limit {
		return tasks[:limit]
	}
	return tasks
}
