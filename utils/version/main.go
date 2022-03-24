/*
 * Cadence - The resource-oriented smart contract programming language
 *
 * Copyright 2019 Dapper Labs, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const template string = `// Code generated by utils/version. DO NOT EDIT.
/*
 * Cadence - The resource-oriented smart contract programming language
 *
 * Copyright 2019 Dapper Labs, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

%s

package cadence

const Version = "%s"
`

// NOTE: must be formatted/injected , as otherwise
// it will be detected itself as a go generate invocation itself
const goGenerateComment = "//go:generate go run ./utils/version/main.go"

const target = "version.go"

func main() {
	version := getLastTag()

	f, err := os.Create(target)
	if err != nil {
		panic(fmt.Errorf("could not create file %s: %w\n", target, err))
	}
	defer func() {
		_ = f.Close()
	}()

	_, err = f.WriteString(fmt.Sprintf(template, goGenerateComment, version))
	if err != nil {
		panic(fmt.Errorf("could not write to %s: %w\n", target, err))
	}
}

func getLastTag() string {
	gitOutput, err := exec.Command("git", "describe", "--tags", "--match", `v*`, "--abbrev=0").Output()
	if err != nil {
		panic(fmt.Errorf("could not get last tag: %w", err))
	}

	return strings.TrimSpace(string(gitOutput))
}
