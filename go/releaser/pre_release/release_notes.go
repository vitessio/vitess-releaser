/*
Copyright 2023 The Vitess Authors.

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
	"log"
	"os"
	"path"
	"regexp"
	"sort"
	"strings"
	"text/template"

	"vitess.io/vitess-releaser/go/releaser"
	"vitess.io/vitess-releaser/go/releaser/github"
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
		ctx                        *releaser.Context
		Version, VersionUnderscore string
		Announcement               string
		KnownIssues                string
		AddDetails                 string
		PathToChangeLogFileOnGH    string
		ChangeLog                  string
		ChangeMetrics              string
		SubDirPath                 string
	}
)

var (
	releaseNotesPath = `changelog/`
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

func generateReleaseNotes(ctx *releaser.Context, version string) {
	// There should be 4 sub-matches, input: "14.0.0", output: ["14.0.0", "14", "0", "0"].
	rx := regexp.MustCompile(`([0-9]+)\.([0-9]+)\.([0-9]+)`)
	versionMatch := rx.FindStringSubmatch(version)
	if len(versionMatch) != 4 {
		log.Fatal("could not parse the release version when generating the release notes")
	}

	majorVersion := versionMatch[1] + "." + versionMatch[2]
	patchVersion := versionMatch[1] + "." + versionMatch[2] + "." + versionMatch[3]
	releaseNotesPath = path.Join(releaseNotesPath, majorVersion, patchVersion)
	summaryFile := path.Join(releaseNotesPath, "summary.md")

	version = "v" + version

	err := os.MkdirAll(releaseNotesPath, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}

	releaseNotes := releaseNote{
		ctx:               ctx,
		Version:           version,
		VersionUnderscore: fmt.Sprintf("%s_%s_%s", versionMatch[1], versionMatch[2], versionMatch[3]), // v14.0.0 -> 14_0_0, this is used to format filenames.
		SubDirPath:        releaseNotesPath,
	}

	// summary of the release
	if summaryFile != "" {
		releaseNotes.Announcement = releaseSummary(summaryFile)
	}

	// known issues
	knownIssues := github.LoadKnownIssues(ctx.VitessRepo, ctx.MajorRelease)
	releaseNotes.KnownIssues = getStringForKnownIssues(knownIssues)

	// changelog with pull requests
	prs, authors := github.GetMergedPRsAndAuthorsByMilestone(ctx.VitessRepo, version)

	releaseNotes.ChangeLog = groupAndStringifyPullRequest(ctx.VitessRepo, prs)

	// changelog metrics
	if len(prs) > 0 && len(authors) > 0 {
		releaseNotes.ChangeMetrics = fmt.Sprintf(`
The release includes %d merged Pull Requests.

Thanks to all our contributors: @%s
`, len(prs), strings.Join(authors, ", @"))
	}

	releaseNotes.generate()
}

func (rn *releaseNote) generate() {
	// Generate the release notes
	rn.PathToChangeLogFileOnGH = fmt.Sprintf(releaseNotesPathGitHub, rn.ctx.VitessRepo) + path.Join(rn.SubDirPath, "changelog.md")
	rnFile, err := os.OpenFile(path.Join(rn.SubDirPath, "release_notes.md"), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		log.Fatal(err)
	}

	t := template.Must(template.New("release_notes").Parse(markdownTemplate))
	err = t.ExecuteTemplate(rnFile, "release_notes", rn)
	if err != nil {
		log.Fatal(err)
	}

	// Generate the changelog
	changelogFile, err := os.OpenFile(path.Join(rn.SubDirPath, "changelog.md"), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		log.Fatal(err)
	}
	t = template.Must(template.New("release_notes_changelog").Parse(markdownTemplateChangelog))
	err = t.ExecuteTemplate(changelogFile, "release_notes_changelog", rn)
	if err != nil {
		log.Fatal(err)
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
		log.Fatal(err)
	}
	return string(contentSummary)
}

func getStringForPullRequestInfos(repo string, prPerType prsByType) string {
	data := createSortedPrTypeSlice(prPerType)

	t := template.Must(template.New("markdownTemplatePR").Parse(fmt.Sprintf(markdownTemplatePR, repo)))
	buff := bytes.Buffer{}
	if err := t.ExecuteTemplate(&buff, "markdownTemplatePR", data); err != nil {
		log.Fatal(err)
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
		log.Fatal(err)
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
