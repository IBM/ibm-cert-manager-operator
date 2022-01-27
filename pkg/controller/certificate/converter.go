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

package certificate

import (
	"net/url"
	"strings"

	v1 "github.com/ibm/ibm-cert-manager-operator/pkg/apis/certmanager/v1"
	v1alpha1 "github.com/ibm/ibm-cert-manager-operator/pkg/apis/certmanager/v1alpha1"
	cmmeta "github.com/ibm/ibm-cert-manager-operator/pkg/apis/meta/v1"
)

func sanitizeDNSNames(names []string) ([]string, []string) {
	dnsNames := make([]string, 0)
	ipAddresses := make([]string, 0)
	for _, n := range names {
		if isURL(n) {
			dnsNames = append(dnsNames, n)
		} else {
			ipAddresses = append(ipAddresses, n)
		}
	}
	return dnsNames, ipAddresses
}

func isURL(s string) bool {
	if _, err := url.Parse(s); err != nil {
		return false
	}
	return true
}

func convertCommonName(s string, names []string) string {
	if s == "" && names != nil && len(names) > 0 {
		return names[0]
	}
	return s
}

func convertIPAddresses(fromCR, fromDNS []string) []string {
	if fromCR == nil {
		return fromDNS
	}
	return append(fromCR, fromDNS...)
}

func convertSubject(o []string) *v1.X509Subject {
	if o == nil {
		return nil
	}
	return &v1.X509Subject{
		Organizations: o,
	}
}

func convertIssuerRef(o v1alpha1.ObjectReference) cmmeta.ObjectReference {
	return cmmeta.ObjectReference{
		Name:  o.Name,
		Kind:  o.Kind,
		Group: "cert-manager.io",
	}
}

// convertPrivateKey converts v1alpha1 Certificate KeyEncoding, KeyAlgorithm,
// and KeySize to a v1 PrivateKey object if the v1alpha1 fields exist
func convertPrivateKey(s v1alpha1.CertificateSpec) *v1.CertificatePrivateKey {
	if s.KeyEncoding == "" && s.KeyAlgorithm == "" && s.KeySize == 0 {
		return nil
	}
	r := &v1.CertificatePrivateKey{
		Encoding:  convertKeyEncoding(s.KeyEncoding),
		Algorithm: converKeyAlgorithm(s.KeyAlgorithm),
		Size:      s.KeySize,
	}
	return r
}

// converUsages converts v1alpha1 Certificate Spec Usages to v1 Usages if
// v1alpha1 Usages exists
func convertUsages(usages []v1alpha1.KeyUsage) []v1.KeyUsage {
	if usages == nil {
		return nil
	}
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

func convertStatus(s v1.CertificateStatus) v1alpha1.CertificateStatus {
	conditions := make([]v1alpha1.CertificateCondition, len(s.Conditions))
	if s.Conditions != nil && len(conditions) > 0 {
		for i, c := range s.Conditions {
			conditions[i] = v1alpha1.CertificateCondition{
				Type:               v1alpha1.CertificateConditionType(c.Type),
				Status:             v1alpha1.ConditionStatus(c.Status),
				LastTransitionTime: c.LastTransitionTime,
				Reason:             c.Reason,
				Message:            c.Message,
			}
		}
	}
	return v1alpha1.CertificateStatus{
		Conditions:      conditions,
		LastFailureTime: s.LastFailureTime,
		NotAfter:        s.NotAfter,
	}
}
