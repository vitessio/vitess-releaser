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

package release

func VtopManualUpdateMessage(latest bool) []string {
	msg := []string{
		"On the release branch:",
		"\t- The 'upgrade_test.sh', 'backup_restore_test.sh' and 'vtorc_vtadmin_test.sh' files must be updated with the proper release increment.",
		"\t- Change the 'verifyVtGateVersion' function calls to use the proper version (current version being released and latest previous version (only used in 'upgrade_test.sh')).",
	}
	if latest {
		msg = append(msg, []string{
			"",
			"On main:",
			"\t- The vitess/lite image tag must be changed in '101_initial_cluster.yaml'. The latest Vitess release tag must be used.",
			"\t- We must copy the 'operator-latest.yaml' file we created during the release onto main's 'operator.yaml' file.",
			"\t- Once copied, remove the change that adds 'imagePullPolicy: Never' and update the image: 'vitess-operator-pr:latest' to use the docker image of latest vitess-operator patch like image: planetscale/vitess-operator:v2.10.0.",
			"\t- The 'upgrade_test.sh', 'backup_restore_test.sh' and 'vtorc_vtadmin_test.sh' files must be updated with the proper release increment.",
			"\t- Change the 'verifyVtGateVersion' function calls to use the proper version (new snapshot Vitess version and current version being released (only used in 'upgrade_test.sh')).",
		}...)
	}
	return msg
}
