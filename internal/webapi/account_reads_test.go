package webapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAccountReadsUseExpectedEndpoints(t *testing.T) {
	var seen []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = append(seen, r.Method+" "+r.URL.RequestURI())
		switch r.URL.Path {
		case "/attachment/isUnderQuota":
			_ = json.NewEncoder(w).Encode(true)
		case "/attachment/dailyLimit", "/project/p1/share/check-quota":
			_ = json.NewEncoder(w).Encode(1)
		case "/project/p1/shares", "/calendar/subscription", "/calendar/archivedEvent", "/project/share/recentProjectUsers", "/user/sessions":
			_ = json.NewEncoder(w).Encode([]map[string]any{{"id": "x1", "name": "item"}})
		default:
			_ = json.NewEncoder(w).Encode(map[string]any{"enabled": true})
		}
	}))
	defer server.Close()

	client := NewClient("test-token")
	client.BaseURL = server.URL
	client.BaseURLV1 = server.URL
	ctx := context.Background()

	calls := []func() error{
		func() error { _, err := client.AttachmentQuota(ctx); return err },
		func() error { _, err := client.DailyReminderPreferences(ctx); return err },
		func() error { _, err := client.ShareContacts(ctx); return err },
		func() error { _, err := client.RecentProjectUsers(ctx); return err },
		func() error { _, err := client.ProjectShares(ctx, "p1"); return err },
		func() error { _, err := client.ProjectShareQuota(ctx, "p1"); return err },
		func() error { _, err := client.ProjectInviteURL(ctx, "p1"); return err },
		func() error { _, err := client.CalendarSubscriptions(ctx); return err },
		func() error { _, err := client.CalendarArchivedEvents(ctx); return err },
		func() error { _, err := client.CalendarThirdAccounts(ctx); return err },
		func() error { _, err := client.StatisticsGeneral(ctx); return err },
		func() error { _, err := client.ProjectTemplates(ctx, 1234); return err },
		func() error { _, err := client.SearchAll(ctx, "hello world"); return err },
		func() error { _, err := client.UserStatus(ctx); return err },
		func() error { _, err := client.UserProfile(ctx); return err },
		func() error { _, err := client.UserSessions(ctx, "zh_CN"); return err },
	}
	for _, call := range calls {
		if err := call(); err != nil {
			t.Fatalf("call failed: %v", err)
		}
	}

	want := []string{
		"GET /attachment/isUnderQuota",
		"GET /attachment/dailyLimit",
		"GET /user/preferences/dailyReminder",
		"GET /share/shareContacts",
		"GET /project/share/recentProjectUsers",
		"GET /project/p1/shares",
		"GET /project/p1/share/check-quota",
		"GET /project/p1/collaboration/invite-url",
		"GET /calendar/subscription",
		"GET /calendar/archivedEvent",
		"GET /calendar/third/accounts",
		"GET /statistics/general",
		"GET /projectTemplates/all?timestamp=1234",
		"GET /search/all?keywords=hello+world",
		"GET /user/status",
		"GET /user/profile",
		"GET /user/sessions?lang=zh_CN",
	}
	if strings.Join(seen, "\n") != strings.Join(want, "\n") {
		t.Fatalf("seen endpoints:\n%s\nwant:\n%s", strings.Join(seen, "\n"), strings.Join(want, "\n"))
	}
}

func TestAccountReadsHandleHTTPErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"errorMessage":"internal error"}`))
	}))
	defer server.Close()

	client := NewClient("test-token")
	client.BaseURL = server.URL
	client.BaseURLV1 = server.URL
	ctx := context.Background()

	// Each of these should return an error, not panic
	cases := []struct {
		name string
		call func() error
	}{
		{"AttachmentQuota", func() error { _, err := client.AttachmentQuota(ctx); return err }},
		{"DailyReminderPreferences", func() error { _, err := client.DailyReminderPreferences(ctx); return err }},
		{"ShareContacts", func() error { _, err := client.ShareContacts(ctx); return err }},
		{"RecentProjectUsers", func() error { _, err := client.RecentProjectUsers(ctx); return err }},
		{"ProjectShares", func() error { _, err := client.ProjectShares(ctx, "p1"); return err }},
		{"ProjectShareQuota", func() error { _, err := client.ProjectShareQuota(ctx, "p1"); return err }},
		{"ProjectInviteURL", func() error { _, err := client.ProjectInviteURL(ctx, "p1"); return err }},
		{"CalendarSubscriptions", func() error { _, err := client.CalendarSubscriptions(ctx); return err }},
		{"CalendarArchivedEvents", func() error { _, err := client.CalendarArchivedEvents(ctx); return err }},
		{"CalendarThirdAccounts", func() error { _, err := client.CalendarThirdAccounts(ctx); return err }},
		{"StatisticsGeneral", func() error { _, err := client.StatisticsGeneral(ctx); return err }},
		{"ProjectTemplates", func() error { _, err := client.ProjectTemplates(ctx, 0); return err }},
		{"SearchAll", func() error { _, err := client.SearchAll(ctx, "test"); return err }},
		{"UserStatus", func() error { _, err := client.UserStatus(ctx); return err }},
		{"UserProfile", func() error { _, err := client.UserProfile(ctx); return err }},
		{"UserSessions", func() error { _, err := client.UserSessions(ctx, "en"); return err }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.call()
			if err == nil {
				t.Fatalf("%s: error = nil, want HTTP error", tc.name)
			}
			if !strings.Contains(err.Error(), "500") {
				t.Fatalf("%s: error = %v, want 500", tc.name, err)
			}
		})
	}
}
