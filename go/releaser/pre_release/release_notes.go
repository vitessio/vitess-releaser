/*
Copyright 2024 The Vitess Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package pre_release

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"regexp"
	"sort"
	"strings"
	"text/template"

	"vitess.io/vitess-releaser/go/releaser"
	"vitess.io/vitess-releaser/go/releaser/github"
	"vitess.io/vitess-releaser/go/releaser/utils"
)

type (
	prsByComponent = map[string][]github.PR

	prsByType = map[string]prsByComponent

	sortedPRComponent struct {
		Name    string
		PrInfos []github.PR
	}

	sortedPRType struct {
		Name       string
		Components []sortedPRComponent
	}

	releaseNote struct {
		ctx                     *releaser.State
		Version                 string
		Announcement            string
		KnownIssues             string
		AddDetails              string
		PathToChangeLogFileOnGH string
		ChangeLog               string
		ChangeMetrics           string
		SubDirPath              string
	}
)

var (
	releaseNotesPathPrefix = `changelog/`
)

const (
	releaseNotesPathGitHub = `https://github.com/%s/blob/main/`
	markdownTemplate       = `# Release of Vitess {{.Version}}

{{- if or .Announcement .AddDetails }}
{{ .Announcement }}
{{- end }}

{{- if and (or .Announcement .AddDetails) (or .KnownIssues .ChangeLog) }}
------------
{{- end }}

{{- if .KnownIssues }}
## Known Issues
{{ .KnownIssues }}
{{- end }}

{{- if .ChangeLog }}
The entire changelog for this release can be found [here]({{ .PathToChangeLogFileOnGH }}).
{{- end }}

{{- if .ChangeLog }}
{{ .ChangeMetrics }}
{{- end }}
`

	markdownTemplateChangelog = `# Changelog of Vitess {{.Version}}
{{ .ChangeLog }}
`

	markdownTemplatePR = `
{{- range $type := . }}
### {{ $type.Name }}
{{- range $component := $type.Components }} 
#### {{ $component.Name }}
{{- range $prInfo := $component.PrInfos }}
 * {{ $prInfo.Title }} [#{{ $prInfo.Number }}](https://github.com/%s/pull/{{ $prInfo.Number }})
{{- end }}
{{- end }}
{{- end }}
`

	markdownTemplateKnownIssues = `
{{- range $issue := . }}
 * {{ $issue.Title }} #{{ $issue.Number }} 
{{- end }}
`

	prefixType      = "Type: "
	prefixComponent = "Component: "
)

func GetReleaseNotesDirPath(version string) string {
	prefix, major, patch := getSegmentOfReleaseNotesDir(version)
	return path.Join(prefix, major, patch)
}

func GetReleaseNotesDirPathForMajor(version string) string {
	prefix, major, _ := getSegmentOfReleaseNotesDir(version)
	return path.Join(prefix, major)
}

func getSegmentOfReleaseNotesDir(version string) (prefix string, major string, patch string) {
	// There should be 4 sub-matches, input: "14.0.0", output: ["14.0.0", "14", "0", "0"].
	rx := regexp.MustCompile(`([0-9]+)\.([0-9]+)\.([0-9]+)`)
	versionMatch := rx.FindStringSubmatch(version)
	if len(versionMatch) != 4 {
		utils.LogPanic(nil, "could not parse the release version when generating the release notes")
	}

	majorVersion := versionMatch[1] + "." + versionMatch[2]
	patchVersion := versionMatch[1] + "." + versionMatch[2] + "." + versionMatch[3]
	return releaseNotesPathPrefix, majorVersion, patchVersion
}

func generateReleaseNotes(state *releaser.State, version string) {
	releaseNotesPath := GetReleaseNotesDirPath(version)
	summaryFile := path.Join(releaseNotesPathPrefix, "summary.md")

	version = "v" + version

	err := os.MkdirAll(releaseNotesPath, os.ModePerm)
	if err != nil {
		utils.LogPanic(err, "could not create the directory: %s", releaseNotesPath)
	}

	releaseNotes := releaseNote{
		ctx:        state,
		Version:    version,
		SubDirPath: releaseNotesPath,
	}

	// summary of the release
	_, err = os.Stat(summaryFile)
	if err != nil && !os.IsNotExist(err) {
		utils.LogPanic(err, "could not stat the summary file: %s", summaryFile)
	} else if err == nil {
		releaseNotes.Announcement = releaseSummary(summaryFile)
	}

	// known issues
	knownIssues := github.LoadKnownIssues(state.VitessRelease.Repo, state.VitessRelease.MajorRelease)
	releaseNotes.KnownIssues = getStringForKnownIssues(knownIssues)

	// changelog with pull requests
	prs, authors := github.GetMergedPRsAndAuthorsByMilestone(state.VitessRelease.Repo, version)

	releaseNotes.ChangeLog = groupAndStringifyPullRequest(state.VitessRelease.Repo, prs)

	// changelog metrics
	if len(prs) > 0 && len(authors) > 0 {
		releaseNotes.ChangeMetrics = fmt.Sprintf(`
The release includes %d merged Pull Requests.

Thanks to all our contributors: @%s
`, len(prs), strings.Join(authors, ", @"))
	}

	releaseNotes.generate()

	// update the entire changelog directory
	// go run ./go/tools/releases/releases.go
	utils.Exec("go", "run", "./go/tools/releases/releases.go")
}

func (rn *releaseNote) generate() {
	// Generate the release notes
	changeLogPath := path.Join(rn.SubDirPath, "changelog.md")
	releaseNotesPath := path.Join(rn.SubDirPath, "release_notes.md")
	rn.PathToChangeLogFileOnGH = fmt.Sprintf(releaseNotesPathGitHub, rn.ctx.VitessRelease.Repo) + changeLogPath
	rnFile, err := os.OpenFile(releaseNotesPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		utils.LogPanic(err, "could not open file %s", releaseNotesPath)
	}

	t := template.Must(template.New("release_notes").Parse(markdownTemplate))
	err = t.ExecuteTemplate(rnFile, "release_notes", rn)
	if err != nil {
		utils.LogPanic(err, "could not execute the release notes template")
	}

	// Generate the changelog
	changelogFile, err := os.OpenFile(changeLogPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		utils.LogPanic(err, "could not open changelog file %s", changeLogPath)
	}
	t = template.Must(template.New("release_notes_changelog").Parse(markdownTemplateChangelog))
	err = t.ExecuteTemplate(changelogFile, "release_notes_changelog", rn)
	if err != nil {
		utils.LogPanic(err, "could not execute the changelog template")
	}
}

func groupPRs(prs []github.PR) prsByType {
	prPerType := prsByType{}

	for _, info := range prs {
		var typ, component string
		for _, lbl := range info.Labels {
			switch {
			case strings.HasPrefix(lbl.Name, prefixType):
				typ = strings.TrimPrefix(lbl.Name, prefixType)
			case strings.HasPrefix(lbl.Name, prefixComponent):
				component = strings.TrimPrefix(lbl.Name, prefixComponent)
			}
		}
		switch typ {
		case "":
			typ = "Other"
		case "Bug":
			typ = "Bug fixes"
		}

		if component == "" {
			component = "Other"
		}
		components, exists := prPerType[typ]
		if !exists {
			components = prsByComponent{}
			prPerType[typ] = components
		}

		prsPerComponentAndType := components[component]
		components[component] = append(prsPerComponentAndType, info)
	}
	return prPerType
}

func createSortedPrTypeSlice(prPerType prsByType) []sortedPRType {
	var data []sortedPRType
	for typeKey, typeElem := range prPerType {
		newPrType := sortedPRType{
			Name: typeKey,
		}
		for componentKey, prInfos := range typeElem {
			newComponent := sortedPRComponent{
				Name:    componentKey,
				PrInfos: prInfos,
			}
			sort.Slice(newComponent.PrInfos, func(i, j int) bool {
				return newComponent.PrInfos[i].Number < newComponent.PrInfos[j].Number
			})
			newPrType.Components = append(newPrType.Components, newComponent)
		}
		sort.Slice(newPrType.Components, func(i, j int) bool {
			return newPrType.Components[i].Name < newPrType.Components[j].Name
		})
		data = append(data, newPrType)
	}
	sort.Slice(data, func(i, j int) bool {
		return data[i].Name < data[j].Name
	})
	return data
}

func releaseSummary(summaryFile string) string {
	contentSummary, err := os.ReadFile(summaryFile)
	if err != nil {
		utils.LogPanic(err, "failed to read file %s", summaryFile)
	}
	return string(contentSummary)
}

func getStringForPullRequestInfos(repo string, prPerType prsByType) string {
	data := createSortedPrTypeSlice(prPerType)

	t := template.Must(template.New("markdownTemplatePR").Parse(fmt.Sprintf(markdownTemplatePR, repo)))
	buff := bytes.Buffer{}
	if err := t.ExecuteTemplate(&buff, "markdownTemplatePR", data); err != nil {
		utils.LogPanic(err, "failed to execute the pull request list template")
	}
	return buff.String()
}

func getStringForKnownIssues(issues []github.Issue) string {
	if len(issues) == 0 {
		return ""
	}
	t := template.Must(template.New("markdownTemplateKnownIssues").Parse(markdownTemplateKnownIssues))
	buff := bytes.Buffer{}
	if err := t.ExecuteTemplate(&buff, "markdownTemplateKnownIssues", issues); err != nil {
		utils.LogPanic(err, "failed to execute the known issues template")
	}
	return buff.String()
}

func groupAndStringifyPullRequest(repo string, prs []github.PR) string {
	if len(prs) == 0 {
		return ""
	}

	prPerType := groupPRs(prs)
	return getStringForPullRequestInfos(repo, prPerType)
}
