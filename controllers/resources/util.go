//
// Copyright 2022 IBM Corporation
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

package resources

import (
	"os"
	"strings"

	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var log = logf.Log.WithName("resource_utils")

// GetImageID constructs image IDs for operands: either <IMAGE_NAME>:<IMAGE_TAG> or <IMAGE_NAME>@<IMAGE_SHA>
func GetImageID(imageRegistry, imageName, defaultImageVersion, imagePostfix, envVarName string) string {

	//Check if the env var exists, if yes, check whether it's a SHA or tag and use accordingly; if no, use the default image version
	imageID := os.Getenv(envVarName)

	if len(imageID) > 0 {
		log.V(2).Info("Using env var for operand image: " + imageName)

		if !strings.Contains(imageID, "sha256:") {
			// if tag, append imagePostfix to the tag if set in CR
			if imagePostfix != "" {
				imageID += imagePostfix
			}
		}
	} else {
		//Use default value
		log.V(2).Info("Using default tag value for operand image " + imageName)
		imageID = imageRegistry + "/" + imageName + ":" + defaultImageVersion

		if imagePostfix != "" {
			imageID += imagePostfix
		}
	}

	return imageID
}

// GetDeployNamespace returns the namespace cert manager operator is deployed in
func GetDeployNamespace() string {
	ns, _ := os.LookupEnv("DEPLOYED_NAMESPACE")
	if ns == "" {
		return DefaultNamespace
	}
	return ns
}
