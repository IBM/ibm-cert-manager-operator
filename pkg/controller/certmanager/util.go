//
// Copyright 2020 IBM Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package certmanager

func containsString(source []string, str string) bool {
	for _, searchString := range source {
		if searchString == str {
			return true
		}
	}
	return false
}

func removeString(source []string, str string) (result []string) {
	for _, sourceString := range source {
		if sourceString == str {
			continue
		}
		result = append(result, sourceString)
	}
	return
}
