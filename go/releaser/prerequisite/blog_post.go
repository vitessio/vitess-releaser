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

func BlogPost() []string {
	return []string{
		"The release announcement blog post must be finished before the release day.",
		"You must coordinate with the different organizations for cross-posting. Usually CNCF and PlanetScale.",
		"",
		"A Pull Request on the website repository of Vitess has to be created so we can easily publish the blog during the release day.",
	}
}
