package content

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type VIPApplication struct {
	Identity      VIPIdentity       `json:"identity"`
	Headline      string            `json:"headline"`
	Pitch         string            `json:"pitch"`
	DashboardLink string            `json:"dashboardLink"`
	Availability  string            `json:"availability"`
	EvidenceCards []VIPEvidenceCard `json:"evidenceCards"`
	Alignment     []VIPAlignment    `json:"alignment"`
	Links         VIPLinks          `json:"links"`
}
type VIPIdentity struct {
	Name       string `json:"name"`
	Credential string `json:"credential"`
	TitleLine  string `json:"titleLine"`
	Badge      string `json:"badge"`
}
type VIPEvidenceCard struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Text  string `json:"text"`
}
type VIPAlignment struct {
	Requirement string  `json:"requirement"`
	Evidence    string  `json:"evidence"`
	Link        *string `json:"link"`
}
type VIPLinks struct {
	GitHub   string `json:"github"`
	LinkedIn string `json:"linkedin"`
	Site     string `json:"site"`
}

type VIPTimeline struct {
	Entries []VIPTimelineEntry `json:"entries"`
}
type VIPTimelineEntry struct {
	Period       string `json:"period"`
	Role         string `json:"role"`
	Organization string `json:"organization"`
	Summary      string `json:"summary"`
	Type         string `json:"type"`
}

type VIPProjects struct {
	Projects []VIPProject `json:"projects"`
}
type VIPProject struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	URL       string   `json:"url"`
	Domain    string   `json:"domain"`
	Problem   string   `json:"problem"`
	Ownership string   `json:"ownership"`
	Outcome   string   `json:"outcome"`
	Relevance string   `json:"relevance"`
	Link      string   `json:"link"`
	Stack     []string `json:"stack"`
	Evidence  []string `json:"evidence"`
}

type VIPBuildLog struct {
	Entries []VIPBuildEntry `json:"entries"`
}
type VIPBuildEntry struct {
	Phase        string         `json:"phase"`
	Date         string         `json:"date"`
	Title        string         `json:"title"`
	Decision     string         `json:"decision"`
	Tradeoff     string         `json:"tradeoff"`
	Verification string         `json:"verification"`
	Measured     map[string]any `json:"measured"`
}

type VIPContent struct {
	Application VIPApplication
	Dossier     string
	Timeline    VIPTimeline
	Projects    VIPProjects
	BuildLog    VIPBuildLog
}

const maxVIPFileBytes = 256 << 10

// LoadVIP reads the private portal corpus once during startup. It deliberately
// accepts only regular, non-symlink files beneath the configured directory.
func LoadVIP(dir string) (VIPContent, error) {
	var value VIPContent
	root, err := filepath.Abs(filepath.Clean(dir))
	if err != nil || strings.TrimSpace(dir) == "" {
		return value, fmt.Errorf("invalid VIP content directory")
	}
	value.Application, err = loadFile[VIPApplication](root, "application.json")
	if err != nil {
		return value, err
	}
	value.Dossier, err = loadTextFile(root, "dossier.md")
	if err != nil {
		return value, err
	}
	value.Timeline, err = loadFile[VIPTimeline](root, "timeline.json")
	if err != nil {
		return value, err
	}
	value.Projects, err = loadFile[VIPProjects](root, "projects.json")
	if err != nil {
		return value, err
	}
	value.BuildLog, err = loadFile[VIPBuildLog](root, "build-log.json")
	if err != nil {
		return value, err
	}
	if strings.TrimSpace(value.Dossier) == "" || strings.TrimSpace(value.Application.Identity.Name) == "" ||
		strings.TrimSpace(value.Application.Headline) == "" {
		return VIPContent{}, fmt.Errorf("VIP content is missing required application fields")
	}
	return value, nil
}

func loadFile[T any](root, name string) (T, error) {
	var value T
	b, err := readVIPFile(root, name)
	if err != nil {
		return value, err
	}
	dec := json.NewDecoder(bytes.NewReader(b))
	if err := dec.Decode(&value); err != nil {
		return value, fmt.Errorf("parse VIP content %s", name)
	}
	var extra any
	if err := dec.Decode(&extra); err != io.EOF {
		return value, fmt.Errorf("parse VIP content %s: trailing data", name)
	}
	return value, nil
}

func loadTextFile(root, name string) (string, error) {
	b, err := readVIPFile(root, name)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func readVIPFile(root, name string) ([]byte, error) {
	path := filepath.Join(root, name)
	info, err := os.Lstat(path)
	if err != nil || !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
		return nil, fmt.Errorf("VIP content file unavailable: %s", name)
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("VIP content file unavailable: %s", name)
	}
	defer f.Close()
	b, err := io.ReadAll(io.LimitReader(f, maxVIPFileBytes+1))
	if err != nil || len(b) > maxVIPFileBytes {
		return nil, fmt.Errorf("VIP content file too large or unreadable: %s", name)
	}
	return b, nil
}
