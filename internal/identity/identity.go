package identity

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/DeliciousBuding/dida-cli/internal/config"
)

const (
	ChannelWebAPI      = "webapi"
	ChannelOpenAPI     = "openapi"
	ChannelOfficialMCP = "official-mcp"
)

// ChannelIdentity is non-secret account metadata for one auth channel.
type ChannelIdentity struct {
	Channel    string `json:"channel"`
	UserID     string `json:"userId,omitempty"`
	Name       string `json:"name,omitempty"`
	ProjectFP  string `json:"projectFingerprint,omitempty"`
	VerifiedAt int64  `json:"verifiedAt"`
	Source     string `json:"source,omitempty"`
}

// Store holds identities for all channels.
type Store struct {
	Channels map[string]ChannelIdentity `json:"channels"`
}

func Path() string {
	return filepath.Join(config.DefaultDir(), "identity.json")
}

func Load() (*Store, error) {
	data, err := os.ReadFile(Path())
	if err != nil {
		if os.IsNotExist(err) {
			return &Store{Channels: map[string]ChannelIdentity{}}, nil
		}
		return nil, err
	}
	var s Store
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("decode identity: %w", err)
	}
	if s.Channels == nil {
		s.Channels = map[string]ChannelIdentity{}
	}
	return &s, nil
}

func (s *Store) Save() error {
	if s.Channels == nil {
		s.Channels = map[string]ChannelIdentity{}
	}
	dir := filepath.Dir(Path())
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	payload, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(Path(), append(payload, '\n'), 0o600)
}

func (s *Store) Put(id ChannelIdentity) {
	if s.Channels == nil {
		s.Channels = map[string]ChannelIdentity{}
	}
	if id.VerifiedAt == 0 {
		id.VerifiedAt = time.Now().UnixMilli()
	}
	s.Channels[id.Channel] = id
}

func (s *Store) Get(channel string) (ChannelIdentity, bool) {
	if s == nil || s.Channels == nil {
		return ChannelIdentity{}, false
	}
	id, ok := s.Channels[channel]
	return id, ok
}

// ProjectFingerprint hashes sorted project IDs into a short stable fingerprint.
func ProjectFingerprint(projectIDs []string) string {
	clean := make([]string, 0, len(projectIDs))
	for _, id := range projectIDs {
		id = strings.TrimSpace(id)
		if id != "" {
			clean = append(clean, id)
		}
	}
	sort.Strings(clean)
	if len(clean) == 0 {
		return ""
	}
	sum := sha256.Sum256([]byte(strings.Join(clean, "\n")))
	return hex.EncodeToString(sum[:8])
}

// MatchResult describes whether configured channels refer to the same account.
type MatchResult struct {
	Match    *bool                      `json:"match"` // nil = unknown/insufficient data
	Reason   string                     `json:"reason,omitempty"`
	Channels map[string]ChannelIdentity `json:"channels"`
}

// EvaluateMatch compares webapi and openapi identities when both are present.
// Official MCP is informational only when project fingerprint is missing.
func EvaluateMatch(s *Store) MatchResult {
	out := MatchResult{Channels: map[string]ChannelIdentity{}}
	if s != nil {
		for k, v := range s.Channels {
			out.Channels[k] = v
		}
	}
	web, hasWeb := out.Channels[ChannelWebAPI]
	oa, hasOA := out.Channels[ChannelOpenAPI]
	if !hasWeb || !hasOA {
		out.Reason = "need both webapi and openapi identities to compare"
		return out
	}
	if web.ProjectFP == "" || oa.ProjectFP == "" {
		// Fall back to userId only if both have it (OpenAPI usually won't).
		if web.UserID != "" && oa.UserID != "" {
			match := web.UserID == oa.UserID
			out.Match = &match
			if !match {
				out.Reason = fmt.Sprintf("userId mismatch: webapi=%s openapi=%s", web.UserID, oa.UserID)
			} else {
				out.Reason = "userId match"
			}
			return out
		}
		out.Reason = "missing project fingerprints; run: dida account verify --json"
		return out
	}
	match := web.ProjectFP == oa.ProjectFP
	out.Match = &match
	if !match {
		out.Reason = fmt.Sprintf("project fingerprint mismatch: webapi=%s openapi=%s", web.ProjectFP, oa.ProjectFP)
	} else {
		out.Reason = "project fingerprint match"
	}
	return out
}

// GuardMultiChannelWrite returns an error when write would touch multiple channels
// that do not appear to be the same account. Env DIDA_ALLOW_CROSS_ACCOUNT=1 bypasses.
func GuardMultiChannelWrite(channels []string) error {
	if strings.TrimSpace(os.Getenv("DIDA_ALLOW_CROSS_ACCOUNT")) == "1" {
		return nil
	}
	needed := uniqueNonEmpty(channels)
	if len(needed) < 2 {
		return nil
	}
	// Only guard when both webapi and openapi are in the write plan.
	hasWeb, hasOA := false, false
	for _, c := range needed {
		if c == ChannelWebAPI {
			hasWeb = true
		}
		if c == ChannelOpenAPI {
			hasOA = true
		}
	}
	if !hasWeb || !hasOA {
		return nil
	}
	s, err := Load()
	if err != nil {
		return fmt.Errorf("load identity store: %w", err)
	}
	res := EvaluateMatch(s)
	if res.Match == nil {
		return fmt.Errorf("cannot verify webapi and openapi are the same account (%s); run: dida account verify --json", res.Reason)
	}
	if !*res.Match {
		return fmt.Errorf("identity mismatch between webapi and openapi (%s); re-login the same account on both channels or set DIDA_ALLOW_CROSS_ACCOUNT=1", res.Reason)
	}
	return nil
}

func uniqueNonEmpty(in []string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, v := range in {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}

// ExtractProjectIDs pulls id/projectId from a list of project maps.
func ExtractProjectIDs(projects []map[string]any) []string {
	ids := make([]string, 0, len(projects))
	for _, p := range projects {
		if id, ok := p["id"].(string); ok && strings.TrimSpace(id) != "" {
			ids = append(ids, id)
			continue
		}
		if id, ok := p["projectId"].(string); ok && strings.TrimSpace(id) != "" {
			ids = append(ids, id)
		}
	}
	return ids
}

func StringField(m map[string]any, keys ...string) string {
	for _, k := range keys {
		if v, ok := m[k]; ok {
			switch t := v.(type) {
			case string:
				if strings.TrimSpace(t) != "" {
					return strings.TrimSpace(t)
				}
			case float64:
				// JSON numbers sometimes decode as float64.
				return strings.TrimSpace(fmt.Sprintf("%.0f", t))
			case json.Number:
				return strings.TrimSpace(t.String())
			case fmt.Stringer:
				s := strings.TrimSpace(t.String())
				if s != "" {
					return s
				}
			default:
				s := strings.TrimSpace(fmt.Sprint(t))
				if s != "" && s != "<nil>" {
					return s
				}
			}
		}
	}
	return ""
}
