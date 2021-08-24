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

package certificate

import (
	"strings"

	v1 "github.com/ibm/ibm-cert-manager-operator/pkg/apis/certmanager/v1"
	v1alpha1 "github.com/ibm/ibm-cert-manager-operator/pkg/apis/certmanager/v1alpha1"
)

// converUsages converts v1alpha1 Certificate Spec Usages to v1 Usages.
func convertUsages(usages []v1alpha1.KeyUsage) []v1.KeyUsage {
	v1Usages := make([]v1.KeyUsage, len(usages))
	for i, u := range usages {
		v1Usages[i] = v1.KeyUsage(u)
	}
	return v1Usages
}

// convertKeyEncoding converts v1alpha1 Certificate Spec KeyEncoding to v1
// Certificate Spec PrivateKeyEncoding.Encoding. This is necessary because the
// v1 values have been capitalized
func convertKeyEncoding(e v1alpha1.KeyEncoding) v1.PrivateKeyEncoding {
	return v1.PrivateKeyEncoding(strings.ToUpper(string(e)))
}

// converKeyAlgorithm converts v1alpha1 Certificate Spec KeyAlgorithm to v1
// Certificate Spec PrivateKeyEncoding.Algorithm. This is necessary because the
// v1 values have been capitalized
func converKeyAlgorithm(a v1alpha1.KeyAlgorithm) v1.PrivateKeyAlgorithm {
	return v1.PrivateKeyAlgorithm(strings.ToUpper(string(a)))
}
