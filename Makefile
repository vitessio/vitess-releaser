# Copyright 2024 The Vitess Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# build local vitess-releaser for development purposes
build:
	go build -o vitess-releaser ./main.go
	./vitess-releaser -v

# install the released vitess-releaser package, for production use
install:
	go install github.com/vitessio/vitess-releaser@latest

# serves for local testing with mods, as well as quick documentation on how to use the tool
test: build
	./vitess-releaser --date 2024-10-23 --rc 2 -r 21 # --live --vtop-release 2.14
