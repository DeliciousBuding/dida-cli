package cli

import (
	"testing"

	"github.com/DeliciousBuding/dida-cli/internal/model"
)

func TestTaskQuadrant(t *testing.T) {
	cases := []struct {
		name string
		task model.Task
		want string
	}{
		{"high priority", model.Task{Priority: 5}, "Q1"},
		{"medium scheduled", model.Task{Priority: 3, DueUnix: 1}, "Q2"},
		{"low scheduled", model.Task{Priority: 1, DueUnix: 1}, "Q3"},
		{"none", model.Task{Priority: 0}, "Q4"},
		{"medium unscheduled", model.Task{Priority: 3}, "Q4"},
	}
	for _, tc := range cases {
		if got := taskQuadrant(tc.task); got != tc.want {
			t.Fatalf("%s quadrant = %s, want %s", tc.name, got, tc.want)
		}
	}
}
