package content

import (
	"encoding/json"
	"fmt"
	"sync"

	blogcontent "goth/content"
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
	Timeline    VIPTimeline
	Projects    VIPProjects
	BuildLog    VIPBuildLog
}

var vipContent struct {
	sync.Once
	value VIPContent
	err   error
}

func VIP() (VIPContent, error) {
	vipContent.Do(func() {
		vipContent.value.Application, vipContent.err = load[VIPApplication]("vip/application.json")
		if vipContent.err != nil {
			return
		}
		vipContent.value.Timeline, vipContent.err = load[VIPTimeline]("vip/timeline.json")
		if vipContent.err != nil {
			return
		}
		vipContent.value.Projects, vipContent.err = load[VIPProjects]("vip/projects.json")
		if vipContent.err != nil {
			return
		}
		vipContent.value.BuildLog, vipContent.err = load[VIPBuildLog]("vip/build-log.json")
	})
	return vipContent.value, vipContent.err
}

func load[T any](path string) (T, error) {
	var value T
	b, err := blogcontent.FS.ReadFile(path)
	if err != nil {
		return value, fmt.Errorf("read VIP content %s: %w", path, err)
	}
	if err := json.Unmarshal(b, &value); err != nil {
		return value, fmt.Errorf("parse VIP content %s: %w", path, err)
	}
	return value, nil
}
