package identity

import (
	"os"
	"path/filepath"
	"testing"
)

func TestProjectFingerprintStable(t *testing.T) {
	t.Parallel()
	a := ProjectFingerprint([]string{"b", "a"})
	b := ProjectFingerprint([]string{"a", "b"})
	if a == "" || a != b {
		t.Fatalf("fingerprint mismatch: %q vs %q", a, b)
	}
	if ProjectFingerprint(nil) != "" {
		t.Fatalf("empty ids should yield empty fingerprint")
	}
}

func TestEvaluateMatch(t *testing.T) {
	t.Parallel()
	s := &Store{Channels: map[string]ChannelIdentity{
		ChannelWebAPI:  {Channel: ChannelWebAPI, ProjectFP: "abc", UserID: "1"},
		ChannelOpenAPI: {Channel: ChannelOpenAPI, ProjectFP: "abc"},
	}}
	res := EvaluateMatch(s)
	if res.Match == nil || !*res.Match {
		t.Fatalf("expected match, got %#v", res)
	}
	s.Channels[ChannelOpenAPI] = ChannelIdentity{Channel: ChannelOpenAPI, ProjectFP: "zzz"}
	res = EvaluateMatch(s)
	if res.Match == nil || *res.Match {
		t.Fatalf("expected mismatch, got %#v", res)
	}
}

func TestGuardMultiChannelWrite(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DIDA_CONFIG_DIR", dir)
	t.Setenv("DIDA_ALLOW_CROSS_ACCOUNT", "")

	// Single channel: always ok.
	if err := GuardMultiChannelWrite([]string{ChannelWebAPI}); err != nil {
		t.Fatalf("single channel: %v", err)
	}

	// No identity yet: dual channel should fail closed.
	if err := GuardMultiChannelWrite([]string{ChannelWebAPI, ChannelOpenAPI}); err == nil {
		t.Fatalf("expected error without identity")
	}

	s := &Store{Channels: map[string]ChannelIdentity{
		ChannelWebAPI:  {Channel: ChannelWebAPI, ProjectFP: "same"},
		ChannelOpenAPI: {Channel: ChannelOpenAPI, ProjectFP: "same"},
	}}
	if err := s.Save(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "identity.json")); err != nil {
		t.Fatal(err)
	}
	if err := GuardMultiChannelWrite([]string{ChannelWebAPI, ChannelOpenAPI}); err != nil {
		t.Fatalf("matched identity: %v", err)
	}

	s.Channels[ChannelOpenAPI] = ChannelIdentity{Channel: ChannelOpenAPI, ProjectFP: "other"}
	if err := s.Save(); err != nil {
		t.Fatal(err)
	}
	if err := GuardMultiChannelWrite([]string{ChannelWebAPI, ChannelOpenAPI}); err == nil {
		t.Fatalf("expected mismatch error")
	}

	t.Setenv("DIDA_ALLOW_CROSS_ACCOUNT", "1")
	if err := GuardMultiChannelWrite([]string{ChannelWebAPI, ChannelOpenAPI}); err != nil {
		t.Fatalf("allow env should bypass: %v", err)
	}
}
