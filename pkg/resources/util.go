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

package resources

import (
	"os"
	"strings"

	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var log = logf.Log.WithName("resource_utils")

//GetImageID constructs image IDs for operands: either <IMAGE_NAME>:<IMAGE_TAG> or <IMAGE_NAME>@<IMAGE_SHA>
func GetImageID(imageRegistry, imageName, defaultImageVersion, imagePostfix, envVarName string) string {
	reqLogger := log.WithValues("Func", "GetImageID")

	var imageSuffix string

	//Check if the env var exists, if yes, check whether it's a SHA or tag and use accordingly; if no, use the default image version
	imageTagOrSHA := os.Getenv(envVarName)

	if len(imageTagOrSHA) > 0 {
		//check if it is a SHA or tag and prepend appropriately
		if strings.HasPrefix(imageTagOrSHA, "sha256:") {
			reqLogger.Info("Using SHA digest value from environment variable for image " + imageName)
			imageSuffix = "@" + imageTagOrSHA
		} else {
			reqLogger.Info("Using tag value from environment variable for image " + imageName)
			imageSuffix = ":" + imageTagOrSHA
			if imagePostfix != "" {
				imageSuffix += imagePostfix
			}
		}
	} else {
		//Use default value
		reqLogger.Info("Using default tag value for image " + imageName)
		imageSuffix = ":" + defaultImageVersion
		if imagePostfix != "" {
			imageSuffix += imagePostfix
		}
	}

	imageID := imageRegistry + "/" + imageName + imageSuffix

	reqLogger.Info("imageID: " + imageID)

	return imageID
}
