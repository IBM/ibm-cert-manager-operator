//
// Copyright 2021 IBM Corporation
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

package issuer

import (
	acme "github.com/ibm/ibm-cert-manager-operator/pkg/apis/acme/v1"
	v1 "github.com/ibm/ibm-cert-manager-operator/pkg/apis/certmanager/v1"
	"github.com/ibm/ibm-cert-manager-operator/pkg/apis/certmanager/v1alpha1"
	cmmeta "github.com/ibm/ibm-cert-manager-operator/pkg/apis/meta/v1"
)

func convertACME(a *v1alpha1.ACMEIssuer) *acme.ACMEIssuer {
	if a == nil {
		return nil
	}
	return &acme.ACMEIssuer{
		Email:         a.Email,
		Server:        a.Server,
		SkipTLSVerify: a.SkipTLSVerify,
		PrivateKey: cmmeta.SecretKeySelector{
			LocalObjectReference: cmmeta.LocalObjectReference(a.PrivateKey.LocalObjectReference),
			Key:                  a.PrivateKey.Key,
		},
		Solvers: convertSolvers(a.Solvers),
	}
}

func convertSolvers(a []v1alpha1.ACMEChallengeSolver) []acme.ACMEChallengeSolver {
	if a == nil {
		return nil
	}

	r := make([]acme.ACMEChallengeSolver, len(a))
	for i, s := range a {
		r[i] = acme.ACMEChallengeSolver{
			Selector: convertSelector(s.Selector),
			HTTP01:   convertHTTP01(s.HTTP01),
			DNS01:    convertDNS01(s.DNS01),
		}
	}
	return r
}

// convertSelector converts v1alpha1 ACMEIssuer.Solver.Selector to v1 equivalent
// This function is necessary to check if v1alpha1 Selector is nil first before
// casting
func convertSelector(c *v1alpha1.CertificateDNSNameSelector) *acme.CertificateDNSNameSelector {
	if c == nil {
		return nil
	}
	return (*acme.CertificateDNSNameSelector)(c)
}

func convertHTTP01(a *v1alpha1.ACMEChallengeSolverHTTP01) *acme.ACMEChallengeSolverHTTP01 {
	if a == nil || a.Ingress == nil {
		return nil
	}

	var podTemplate *acme.ACMEChallengeSolverHTTP01IngressPodTemplate
	if a.Ingress.PodTemplate != nil {
		podTemplate = &acme.ACMEChallengeSolverHTTP01IngressPodTemplate{
			ACMEChallengeSolverHTTP01IngressPodObjectMeta: acme.ACMEChallengeSolverHTTP01IngressPodObjectMeta{
				Annotations: a.Ingress.PodTemplate.Annotations,
				Labels:      a.Ingress.PodTemplate.Labels,
			},
		}
	}

	return &acme.ACMEChallengeSolverHTTP01{
		Ingress: &acme.ACMEChallengeSolverHTTP01Ingress{
			ServiceType: a.Ingress.ServiceType,
			Class:       a.Ingress.Class,
			Name:        a.Ingress.Name,
			PodTemplate: podTemplate,
		},
	}
}

func convertDNS01(a *v1alpha1.ACMEChallengeSolverDNS01) *acme.ACMEChallengeSolverDNS01 {
	if a == nil {
		return nil
	}

	var akamai *acme.ACMEIssuerDNS01ProviderAkamai
	if a.Akamai != nil {
		akamai = &acme.ACMEIssuerDNS01ProviderAkamai{
			ServiceConsumerDomain: a.Akamai.ServiceConsumerDomain,
			ClientToken: cmmeta.SecretKeySelector{
				LocalObjectReference: cmmeta.LocalObjectReference(a.Akamai.ClientToken.LocalObjectReference),
				Key:                  a.Akamai.ClientToken.Key,
			},
			ClientSecret: cmmeta.SecretKeySelector{
				LocalObjectReference: cmmeta.LocalObjectReference(a.Akamai.ClientSecret.LocalObjectReference),
				Key:                  a.Akamai.ClientSecret.Key,
			},
			AccessToken: cmmeta.SecretKeySelector{
				LocalObjectReference: cmmeta.LocalObjectReference(a.Akamai.AccessToken.LocalObjectReference),
				Key:                  a.Akamai.AccessToken.Key,
			},
		}
	}

	var cloudDNS *acme.ACMEIssuerDNS01ProviderCloudDNS
	if a.CloudDNS != nil {
		cloudDNS = &acme.ACMEIssuerDNS01ProviderCloudDNS{
			ServiceAccount: &cmmeta.SecretKeySelector{ // not optional in v1alpha1
				LocalObjectReference: cmmeta.LocalObjectReference(a.CloudDNS.ServiceAccount.LocalObjectReference),
				Key:                  a.CloudDNS.ServiceAccount.Key,
			},
			Project: a.CloudDNS.Project,
		}
	}

	var cloudflare *acme.ACMEIssuerDNS01ProviderCloudflare
	if a.Cloudflare != nil {
		cloudflare = &acme.ACMEIssuerDNS01ProviderCloudflare{
			Email: a.Cloudflare.Email,
			APIKey: &cmmeta.SecretKeySelector{ // not optional in v1alpha1
				LocalObjectReference: cmmeta.LocalObjectReference(a.Cloudflare.APIKey.LocalObjectReference),
				Key:                  a.Cloudflare.APIKey.Key,
			},
		}
	}

	var route53 *acme.ACMEIssuerDNS01ProviderRoute53
	if a.Route53 != nil {
		var key cmmeta.SecretKeySelector
		if a.Route53.SecretAccessKey != (v1alpha1.SecretKeySelector{}) {
			key = cmmeta.SecretKeySelector{
				LocalObjectReference: cmmeta.LocalObjectReference(a.Route53.SecretAccessKey.LocalObjectReference),
				Key:                  a.Route53.SecretAccessKey.Key,
			}
		}
		route53 = &acme.ACMEIssuerDNS01ProviderRoute53{
			AccessKeyID:     a.Route53.AccessKeyID,
			SecretAccessKey: key,
			Role:            a.Route53.Role,
			HostedZoneID:    a.Route53.HostedZoneID,
			Region:          a.Route53.Region,
		}
	}

	var azureDNS *acme.ACMEIssuerDNS01ProviderAzureDNS
	if a.AzureDNS != nil {
		azureDNS = &acme.ACMEIssuerDNS01ProviderAzureDNS{
			ClientID: a.AzureDNS.ClientID,
			ClientSecret: &cmmeta.SecretKeySelector{ // not optional in v1alpha1
				LocalObjectReference: cmmeta.LocalObjectReference(a.AzureDNS.ClientSecret.LocalObjectReference),
				Key:                  a.AzureDNS.ClientSecret.Key,
			},
			SubscriptionID:    a.AzureDNS.SubscriptionID,
			TenantID:          a.AzureDNS.TenantID,
			ResourceGroupName: a.AzureDNS.ResourceGroupName,
			HostedZoneName:    a.AzureDNS.HostedZoneName,
			Environment:       acme.AzureDNSEnvironment(a.AzureDNS.Environment),
		}
	}

	var digitalOcean *acme.ACMEIssuerDNS01ProviderDigitalOcean
	if a.DigitalOcean != nil {
		digitalOcean = &acme.ACMEIssuerDNS01ProviderDigitalOcean{
			Token: cmmeta.SecretKeySelector{
				LocalObjectReference: cmmeta.LocalObjectReference(a.DigitalOcean.Token.LocalObjectReference),
				Key:                  a.DigitalOcean.Token.Key,
			},
		}
	}

	var acmeDNS *acme.ACMEIssuerDNS01ProviderAcmeDNS
	if a.AcmeDNS != nil {
		acmeDNS = &acme.ACMEIssuerDNS01ProviderAcmeDNS{
			Host: a.AcmeDNS.Host,
			AccountSecret: cmmeta.SecretKeySelector{
				LocalObjectReference: cmmeta.LocalObjectReference(a.AcmeDNS.AccountSecret.LocalObjectReference),
				Key:                  a.AcmeDNS.AccountSecret.Key,
			},
		}
	}

	var rfc2136 *acme.ACMEIssuerDNS01ProviderRFC2136
	if a.RFC2136 != nil {
		var secret cmmeta.SecretKeySelector
		if a.RFC2136.TSIGSecret != (v1alpha1.SecretKeySelector{}) {
			secret = cmmeta.SecretKeySelector{
				LocalObjectReference: cmmeta.LocalObjectReference(a.RFC2136.TSIGSecret.LocalObjectReference),
				Key:                  a.RFC2136.TSIGSecret.Key,
			}
		}
		rfc2136 = &acme.ACMEIssuerDNS01ProviderRFC2136{
			Nameserver:    a.RFC2136.Nameserver,
			TSIGSecret:    secret,
			TSIGKeyName:   a.RFC2136.TSIGKeyName,
			TSIGAlgorithm: a.RFC2136.TSIGAlgorithm,
		}
	}

	var webhook *acme.ACMEIssuerDNS01ProviderWebhook
	if a.Webhook != nil {
		webhook = &acme.ACMEIssuerDNS01ProviderWebhook{
			GroupName:  a.Webhook.GroupName,
			SolverName: a.Webhook.SolverName,
			Config:     a.Webhook.Config,
		}
	}

	return &acme.ACMEChallengeSolverDNS01{
		CNAMEStrategy: acme.CNAMEStrategy(a.CNAMEStrategy),
		Akamai:        akamai,
		CloudDNS:      cloudDNS,
		Cloudflare:    cloudflare,
		Route53:       route53,
		AzureDNS:      azureDNS,
		DigitalOcean:  digitalOcean,
		AcmeDNS:       acmeDNS,
		RFC2136:       rfc2136,
		Webhook:       webhook,
	}
}

func convertCA(c *v1alpha1.CAIssuer) *v1.CAIssuer {
	if c == nil {
		return nil
	}
	return &v1.CAIssuer{
		SecretName:            c.SecretName,
		CRLDistributionPoints: nil,
		OCSPServers:           nil,
	}
}

func convertVault(v *v1alpha1.VaultIssuer) *v1.VaultIssuer {
	if v == nil {
		return nil
	}

	var ref *cmmeta.SecretKeySelector
	if v.Auth.TokenSecretRef != (v1alpha1.SecretKeySelector{}) {
		ref = &cmmeta.SecretKeySelector{
			LocalObjectReference: cmmeta.LocalObjectReference(v.Auth.TokenSecretRef.LocalObjectReference),
			Key:                  v.Auth.TokenSecretRef.Key,
		}
	}

	var role *v1.VaultAppRole
	if v.Auth.AppRole != (v1alpha1.VaultAppRole{}) {
		role = &v1.VaultAppRole{
			Path:   v.Auth.AppRole.Path,
			RoleId: v.Auth.AppRole.RoleId,
			SecretRef: cmmeta.SecretKeySelector{
				LocalObjectReference: cmmeta.LocalObjectReference(v.Auth.AppRole.SecretRef.LocalObjectReference),
				Key:                  v.Auth.AppRole.SecretRef.Key,
			},
		}
	}
	return &v1.VaultIssuer{
		Auth: v1.VaultAuth{
			TokenSecretRef: ref,
			AppRole:        role,
			Kubernetes:     nil,
		},
		Server:   v.Server,
		Path:     v.Path,
		CABundle: v.CABundle,
	}
}

func convertSelfSigned(s *v1alpha1.SelfSignedIssuer) *v1.SelfSignedIssuer {
	if s == nil {
		return nil
	}
	return &v1.SelfSignedIssuer{
		CRLDistributionPoints: nil,
	}
}

func convertVenafi(v *v1alpha1.VenafiIssuer) *v1.VenafiIssuer {
	if v == nil {
		return nil
	}
	var tpp *v1.VenafiTPP
	if v.TPP != nil {
		tpp = &v1.VenafiTPP{
			URL: v.TPP.URL,
			CredentialsRef: cmmeta.LocalObjectReference{
				Name: v.TPP.CredentialsRef.Name,
			},
			CABundle: v.TPP.CABundle,
		}
	}
	var cloud *v1.VenafiCloud
	if v.Cloud != nil {
		cloud = &v1.VenafiCloud{
			URL: v.Cloud.URL,
			APITokenSecretRef: cmmeta.SecretKeySelector{
				LocalObjectReference: cmmeta.LocalObjectReference(v.Cloud.APITokenSecretRef.LocalObjectReference),
				Key:                  v.Cloud.APITokenSecretRef.Key,
			},
		}
	}
	return &v1.VenafiIssuer{
		Zone:  v.Zone,
		TPP:   tpp,
		Cloud: cloud,
	}
}

func convertStatus(s v1.IssuerStatus) v1alpha1.IssuerStatus {
	conditions := make([]v1alpha1.IssuerCondition, len(s.Conditions))
	if s.Conditions != nil && len(conditions) > 0 {
		for i, c := range s.Conditions {
			conditions[i] = v1alpha1.IssuerCondition{
				Type:               v1alpha1.IssuerConditionType(c.Type),
				Status:             v1alpha1.ConditionStatus(c.Status),
				LastTransitionTime: c.LastTransitionTime,
				Reason:             c.Reason,
				Message:            c.Message,
			}
		}
	}

	return v1alpha1.IssuerStatus{
		Conditions: conditions,
		ACME:       (*v1alpha1.ACMEIssuerStatus)(s.ACME),
	}
}
