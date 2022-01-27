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

package certmanager

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	secretshare "github.com/IBM/ibm-secretshare-operator/api/v1"
)

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
	return result
}

//copySecret copies the secret from one namespace to another
func copySecret(client client.Client, secretToCopy string, srcNamespace string, destNamespace string, secretShareCRName string) error {
	// create a secretshare CR instance

	var secretShareCR = &secretshare.SecretShare{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretShareCRName,
			Namespace: srcNamespace,
		},
		Spec: secretshare.SecretShareSpec{
			Secretshares: []secretshare.Secretshare{
				{
					Secretname: secretToCopy,
					Sharewith: []secretshare.TargetNamespace{
						{
							Namespace: destNamespace,
						},
					},
				},
			},
		},
	}

	// Create the secretshare CR to copy the secret
	err := client.Create(context.TODO(), secretShareCR)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("could not create resource: %v", err)
	}

	return nil

}
