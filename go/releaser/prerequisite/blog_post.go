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

package prerequisite

func DraftBlogPost() []string {
	return []string{
		"The release announcement blog post must be finished before the release day.",
		"A highlight of the major changes must be included in the blog post, along with a link to the release notes.",
		"",
		"A Pull Request on the website repository of Vitess has to be created so we can easily publish the blog during the release day.",
	}
}

func RequestCrossPostBlogPost() []string {
	return []string{
		"You must coordinate with the different organizations (CNCF and PlanetScale) for cross-posting the blog post.",
		"Send out an email to both parties to request a cross-post a couple weeks before the GA release.",
	}
}
