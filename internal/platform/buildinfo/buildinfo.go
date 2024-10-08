// Copyright © 2018 Banzai Cloud
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package buildinfo

import (
	"runtime"
)

// BuildInfo represents all available build information.
type BuildInfo struct {
	Version    string `json:"version"`
	CommitHash string `json:"commit_hash"`
	BuildDate  string `json:"build_date"`
	Branch     string `json:"branch"`
	GoVersion  string `json:"go_version"`
	Os         string `json:"os"`
	Arch       string `json:"arch"`
	Compiler   string `json:"compiler"`
}

// New returns all available build information.
func New(version string, commitHash string, buildDate string, branch string) BuildInfo {
	return BuildInfo{
		Version:    version,
		CommitHash: commitHash,
		BuildDate:  buildDate,
		Branch:     branch,
		GoVersion:  runtime.Version(),
		Os:         runtime.GOOS,
		Arch:       runtime.GOARCH,
		Compiler:   runtime.Compiler,
	}
}

func (bi BuildInfo) Context() []interface{} {
	return []interface{}{
		"version", bi.Version,
		"commit_hash", bi.CommitHash,
		"build_date", bi.BuildDate,
		"branch", bi.Branch,
		"go_version", bi.GoVersion,
		"os", bi.Os,
		"arch", bi.Arch,
		"compiler", bi.Compiler,
	}
}
