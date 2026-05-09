package cli

import "github.com/DeliciousBuding/dida-cli/internal/model"

type compactTask struct {
	ID          string   `json:"id"`
	ProjectID   string   `json:"projectId"`
	ProjectName string   `json:"projectName,omitempty"`
	ParentID    string   `json:"parentId,omitempty"`
	Title       string   `json:"title"`
	DueDate     string   `json:"dueDate,omitempty"`
	StartDate   string   `json:"startDate,omitempty"`
	Priority    int      `json:"priority"`
	Status      int      `json:"status"`
	ColumnID    string   `json:"columnId,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Overdue     bool     `json:"overdue,omitempty"`
}

func taskOutput(tasks []model.Task, compact bool) any {
	if !compact {
		return stripTaskRaw(tasks)
	}
	out := make([]compactTask, 0, len(tasks))
	for _, task := range tasks {
		out = append(out, compactTask{
			ID:          task.ID,
			ProjectID:   task.ProjectID,
			ProjectName: task.ProjectName,
			ParentID:    task.ParentID,
			Title:       task.Title,
			DueDate:     task.DueDate,
			StartDate:   task.StartDate,
			Priority:    task.Priority,
			Status:      task.Status,
			ColumnID:    task.ColumnID,
			Tags:        task.Tags,
			Overdue:     task.Overdue,
		})
	}
	return out
}
