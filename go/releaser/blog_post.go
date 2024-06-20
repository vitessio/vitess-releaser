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

package releaser

func DraftBlogPost() []string {
	return []string{
		"The release announcement blog post must be finished before the release day.",
		"You may take example on previous release announcement blog post.",
		"",
		"The blog post should at least have the following:",
		"\t- Highlight of the major changes.",
		"\t- A link to the release notes.",
		"",
		"During the pre-release, we will create a Pull Request on the website repository with the draft.",
	}
}

func RequestCrossPostBlogPost() []string {
	return []string{
		"You must coordinate with the different organizations (CNCF and PlanetScale) to cross-post the blog post.",
		"Send out an email to both parties to request a cross-post a couple weeks before the GA release.",
		"",
		"CNCF:",
		"\tThe title of the blog post Google Docs usually contains 'DRAFT', and we usually let CNCF know that until the title does not say 'FINAL',",
		"\tthe blog post cannot be posted. We usually ask them to publish the cross-post on the same day as the release, morning time PST.",
		"",
		"PlanetScale:",
		"\t Send a request to a Vitess maintainers working at PlanetScale to let them handle the cross-post request internally.",
		"\t PlanetScale usually publish the cross-post the same days as the release, morning time PST.",
		"",
		"Important:",
		"\tIf for whatever reason the GA release is delayed, all parties must be informed ASAP so they don't release the blog post before the release.",
	}
}

func CreateBlogPostPR() []string {
	return []string{
		"Open a Pull Request on the website repository that contains the new blog post.",
	}
}

func MergeBlogPostPR() []string {
	return []string{
		"Merge the blog post Pull Request you have created during the pre-release.",
	}
}
