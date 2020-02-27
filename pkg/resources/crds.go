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
	apiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CRDMap a map from the crd name to the definition of that crd
var CRDMap = map[string]*apiext.CustomResourceDefinition{
	"certificates":   certificateCRD,
	"issuers":        issuerCRD,
	"clusterissuers": clusterIssuerCRD,
	"orders":         orderCRD,
	"challenges":     challengeCRD,
}

var certificateCRD = &apiext.CustomResourceDefinition{
	ObjectMeta: metav1.ObjectMeta{Name: "certificates.certmanager.k8s.io", Labels: ControllerLabelMap},
	Spec: apiext.CustomResourceDefinitionSpec{
		Group:   GroupVersion,
		Version: CRDVersion,
		Scope:   apiext.NamespaceScoped,
		Names: apiext.CustomResourceDefinitionNames{
			Plural:     "certificates",
			Kind:       "Certificate",
			ShortNames: []string{"cert", "certs"},
		},
		AdditionalPrinterColumns: []apiext.CustomResourceColumnDefinition{
			{
				JSONPath: `.status.conditions[?(@.type=="Ready")].status`,
				Name:     "Ready",
				Type:     "string",
			},
			{
				JSONPath: ".spec.secretName",
				Name:     "Secret",
				Type:     "string",
			},
			{
				JSONPath: ".spec.issuerRef.name",
				Name:     "Issuer",
				Type:     "string",
				Priority: 1,
			},
			{
				JSONPath: `.status.conditions[?(@.type=="Ready")].message`,
				Name:     "Status",
				Type:     "string",
				Priority: 1,
			},
			{
				JSONPath:    ".metadata.creationTimestamp",
				Description: "CreationTimestamp is a timestamp representing time when this object was created. It is not guaranteed to be set in happens-before order across separate operations. Clients may not set this value. It is represented in RFC3339 form and is in UTC. Populated by the system. Read-only. Null for lists. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata",
				Name:        "Age",
				Type:        "date",
			},
			{
				JSONPath: ".status.notAfter",
				Name:     "Expiration",
				Type:     "string",
			},
		},
		Validation: &apiext.CustomResourceValidation{
			OpenAPIV3Schema: &apiext.JSONSchemaProps{
				Description: "Certificate is a type to represent a Certificate from ACME",
				Properties: map[string]apiext.JSONSchemaProps{
					"apiVersion": apiext.JSONSchemaProps{
						Description: "APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources",
						Type:        "string",
					},
					"kind": apiext.JSONSchemaProps{
						Description: "Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds",
						Type:        "string",
					},
					"metadata": apiext.JSONSchemaProps{
						Type: "object",
					},
					"spec": apiext.JSONSchemaProps{
						Description: "CertificateSpec defines the desired state of Certificate",
						Properties: map[string]apiext.JSONSchemaProps{
							"acme": apiext.JSONSchemaProps{
								Description: "ACME contains configuration specific to ACME Certificates. Notably, this contains details on how the domain names listed on this Certificate resource should be 'solved', i.e. mapping HTTP01 and DNS01 providers to DNS names.",
								Properties: map[string]apiext.JSONSchemaProps{
									"config": apiext.JSONSchemaProps{
										Items: &apiext.JSONSchemaPropsOrArray{
											Schema: &apiext.JSONSchemaProps{
												Description: "DomainSolverConfig contains solver configuration for a set of domains.",
												Properties: map[string]apiext.JSONSchemaProps{
													"dns01": apiext.JSONSchemaProps{
														Description: "DNS01 contains DNS01 challenge solving configuration",
														Properties: map[string]apiext.JSONSchemaProps{
															"provider": apiext.JSONSchemaProps{
																Description: "Provider is the name of the DNS01 challenge provider to use, as configure on the referenced Issuer or ClusterIssuer resource.",
																Type:        "string",
															},
														},
														Required: []string{"provider"},
														Type:     "object",
													},
													"domains": apiext.JSONSchemaProps{
														Description: "Domains is the list of domains that this Solverconfig applies to.",
														Items: &apiext.JSONSchemaPropsOrArray{
															Schema: &apiext.JSONSchemaProps{
																Type: "string",
															},
														},
														Type: "array",
													},
													"http01": apiext.JSONSchemaProps{
														Description: "HTTP01 contains HTTP01 challenge solving configuration",
														Properties: map[string]apiext.JSONSchemaProps{
															"ingress": apiext.JSONSchemaProps{
																Description: "Ingress is the name of an Ingress resource that will be edited to include the ACME HTTP01 'well-known' challenge path in order to solve HTTP01 challenges. If this field is specified, 'ingressClass' **must not** be specified.",
																Type:        "string",
															},
															"ingressClass": apiext.JSONSchemaProps{
																Description: "IngressClass is the ingress class that should be set on new ingress resources that are created in order to solve HTTP01 challenges. This field should be used when using an ingress controller such as nginx, which 'flattens' ingress configuration instead of maintaining a 1:1 mapping between loadbalancer IP:ingress resources. If this field is not set, and 'ingress' is not set, then ingresses without an ingress class set will be created to solve HTTP01 challenges. If this field is specified, 'ingress' **must not** be specified.",
																Type:        "string",
															},
														},
														Type: "object",
													},
												},
												Required: []string{"domains"},
												Type:     "object",
											},
										},
										Type: "array",
									},
								},
								Required: []string{"config"},
								Type:     "object",
							},
							"commonName": apiext.JSONSchemaProps{
								Description: "CommonName is a common name to be used on the Certificate. If no CommonName is given, then the first entry in DNSNames is used as the CommonName. The CommonName should have a length of 64 characters or fewer to avoid generating invalid CSRs; in order to have longer domain names, set the CommonName (or first DNSNames entry) to have 64 characters or fewer, and then add the longer domain name to DNSNames.",
								Type:        "string",
							},
							"dnsNames": apiext.JSONSchemaProps{
								Description: "DNSNames is a list of subject alt names to be used on the Certificate. If no CommonName is given, then the first entry in DNSNames is used as the CommonName and must have a length of 64 characters or fewer.",
								Items: &apiext.JSONSchemaPropsOrArray{
									Schema: &apiext.JSONSchemaProps{
										Type: "string",
									},
								},
								Type: "array",
							},
							"duration": apiext.JSONSchemaProps{
								Description: "Certificate default Duration",
								Type:        "string",
							},
							"ipAddresses": apiext.JSONSchemaProps{
								Description: "IPAddresses is a list of IP addresses to be used on the Certificate",
								Items: &apiext.JSONSchemaPropsOrArray{
									Schema: &apiext.JSONSchemaProps{
										Type: "string",
									},
								},
								Type: "array",
							},
							"isCA": apiext.JSONSchemaProps{
								Description: "IsCA will mark this Certificate as valid for signing. This implies that the 'cert sign' usage is set",
								Type:        "boolean",
							},
							"issuerRef": apiext.JSONSchemaProps{
								Description: "IssuerRef is a reference to the issuer for this certificate. If the 'kind' field is not set, or set to 'Issuer', an Issuer resource with the given name in the same namespace as the Certificate will be used. If the 'kind' field is set to 'ClusterIssuer', a ClusterIssuer with the provided name will be used. The 'name' field in this stanza is required at all times.",
								Properties: map[string]apiext.JSONSchemaProps{
									"group": apiext.JSONSchemaProps{
										Type: "string",
									},
									"kind": apiext.JSONSchemaProps{
										Type: "string",
									},
									"name": apiext.JSONSchemaProps{
										Type: "string",
									},
								},
								Required: []string{"name"},
								Type:     "object",
							},
							"keyAlgorithm": apiext.JSONSchemaProps{
								Description: "KeyAlgorithm is the private key algorithm of the corresponding private key for this certificate. If provided, allowed values are either \"rsa\" or \"ecdsa\" If KeyAlgorithm is specified and KeySize is not provided, key size of 256 will be used for \"ecdsa\" key algorithm and key size of 2048 will be used for \"rsa\" key algorithm.",
								Enum: []apiext.JSON{
									{
										Raw: []byte("\"rsa\""),
									},
									{
										Raw: []byte("\"ecdsa\""),
									},
								},
								Type: "string",
							},
							"keyEncoding": apiext.JSONSchemaProps{
								Description: "KeyEncoding is the private key cryptography standards (PKCS) for this certificate's private key to be encoded in. If provided, allowed values are \"pkcs1\" and \"pkcs8\" standing for PKCS#1 and PKCS#8, respectively. If KeyEncoding is not specified, then PKCS#1 will be used by default.",
								Enum: []apiext.JSON{
									{
										Raw: []byte("\"pkcs1\""),
									},
									{
										Raw: []byte("\"pkcs8\""),
									},
								},
								Type: "string",
							},
							"keySize": apiext.JSONSchemaProps{
								Description: "KeySize is the key bit size of the corresponding private key for this certificate. If provided, value must be between 2048 and 8192 inclusive when KeyAlgorithm is empty or is set to \"rsa\", and value must be one of (256, 384, 521) when KeyAlgorithm is set to \"ecdsa\".",
								Type:        "integer",
							},
							"organization": apiext.JSONSchemaProps{
								Description: "Organization is the organization to be used on the Certificate",
								Items: &apiext.JSONSchemaPropsOrArray{
									Schema: &apiext.JSONSchemaProps{
										Type: "string",
									},
								},
								Type: "array",
							},
							"renewBefore": apiext.JSONSchemaProps{
								Description: "Certificate renew before expiration duration",
								Type:        "string",
							},
							"secretName": apiext.JSONSchemaProps{
								Description: "SecretName is the name of the secret resource to store this secret in",
								Type:        "string",
							},
							"usages": apiext.JSONSchemaProps{
								Description: "Usages is the set of x509 actions that are enabled for a given key. Defaults are ('digital signature', 'key encipherment') if empty",
								Items: &apiext.JSONSchemaPropsOrArray{
									Schema: &apiext.JSONSchemaProps{
										Description: "KeyUsage specifies valid usage contexts for keys. See: https://tools.ietf.org/html/rfc5280#section-4.2.1.3 https://tools.ietf.org/html/rfc5280#section-4.2.1.12",
										Enum: []apiext.JSON{
											{Raw: []byte("\"signing\"")},
											{Raw: []byte("\"digital signature\"")},
											{Raw: []byte("\"content commitment\"")},
											{Raw: []byte("\"key encipherment\"")},
											{Raw: []byte("\"key agreement\"")},
											{Raw: []byte("\"data encipherment\"")},
											{Raw: []byte("\"cert sign\"")},
											{Raw: []byte("\"crl sign\"")},
											{Raw: []byte("\"encipher only\"")},
											{Raw: []byte("\"decipher only\"")},
											{Raw: []byte("\"any\"")},
											{Raw: []byte("\"server auth\"")},
											{Raw: []byte("\"client auth\"")},
											{Raw: []byte("\"code signing\"")},
											{Raw: []byte("\"email protection\"")},
											{Raw: []byte("\"s/mime\"")},
											{Raw: []byte("\"ipsec end system\"")},
											{Raw: []byte("\"ipsec tunnel\"")},
											{Raw: []byte("\"ipsec user\"")},
											{Raw: []byte("\"timestamping\"")},
											{Raw: []byte("\"ocsp signing\"")},
											{Raw: []byte("\"microsoft sgc\"")},
											{Raw: []byte("\"netscape sgc\"")},
										},
										Type: "string",
									},
								},
								Type: "array",
							},
						},
						Required: []string{"issuerRef", "secretName"},
						Type:     "object",
					},
					"status": apiext.JSONSchemaProps{
						Description: "CertificateStatus defines the observed state of Certificate",
						Properties: map[string]apiext.JSONSchemaProps{
							"conditions": apiext.JSONSchemaProps{
								Items: &apiext.JSONSchemaPropsOrArray{
									Schema: &apiext.JSONSchemaProps{
										Description: "CertificateCondition contains condition information for an Certificate.",
										Properties: map[string]apiext.JSONSchemaProps{
											"lastTransitionTime": apiext.JSONSchemaProps{
												Description: "LastTransitionTime is the timestamp corresponding to the last status change of this condition.",
												Format:      "date-time",
												Type:        "string",
											},
											"message": apiext.JSONSchemaProps{
												Description: "Message is a human readable description of the details of the last transition, complementing reason.",
												Type:        "string",
											},
											"reason": apiext.JSONSchemaProps{
												Description: "Reason is a brief machine readable explanation for the condition's last transition.",
												Type:        "string",
											},
											"status": apiext.JSONSchemaProps{
												Description: "Status of the condition one of ('True', 'False', 'Unknown').",
												Enum: []apiext.JSON{
													{Raw: []byte("\"True\"")},
													{Raw: []byte("\"False\"")},
													{Raw: []byte("\"Unknown\"")},
												},
												Type: "string",
											},
											"type": apiext.JSONSchemaProps{
												Description: "Type of the condition, currently ('Ready').",
												Type:        "string",
											},
										},
										Required: []string{"status", "type"},
										Type:     "object",
									},
								},
								Type: "array",
							},
							"lastFailureTime": apiext.JSONSchemaProps{
								Format: "date-time",
								Type:   "string",
							},
							"notAfter": apiext.JSONSchemaProps{
								Description: "The expiration time of the certificate stored in the secret named by this resource in spec.secretName.",
								Format:      "date-time",
								Type:        "string",
							},
						},
						Type: "object",
					},
				},
				Type: "object",
			},
		},
	},
	Status: apiext.CustomResourceDefinitionStatus{
		Conditions: []apiext.CustomResourceDefinitionCondition{},
		AcceptedNames: apiext.CustomResourceDefinitionNames{
			Kind:   "",
			Plural: "",
		},
		StoredVersions: []string{},
	},
}

var issuerCRD = &apiext.CustomResourceDefinition{
	ObjectMeta: metav1.ObjectMeta{Name: "issuers.certmanager.k8s.io", Labels: ControllerLabelMap},
	Spec: apiext.CustomResourceDefinitionSpec{
		Group:   GroupVersion,
		Version: CRDVersion,
		Scope:   apiext.NamespaceScoped,
		Names: apiext.CustomResourceDefinitionNames{
			Plural: "issuers",
			Kind:   "Issuer",
		},
		AdditionalPrinterColumns: []apiext.CustomResourceColumnDefinition{
			{
				JSONPath: `.status.conditions[?(@.type=="Ready")].status`,
				Name:     "Ready",
				Type:     "string",
			},
		},
		Validation: &apiext.CustomResourceValidation{
			OpenAPIV3Schema: &apiext.JSONSchemaProps{
				Properties: map[string]apiext.JSONSchemaProps{
					"apiVersion": apiext.JSONSchemaProps{
						Description: "APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources",
						Type:        "string",
					},
					"kind": apiext.JSONSchemaProps{
						Description: "Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds",
						Type:        "string",
					},
					"metadata": apiext.JSONSchemaProps{
						Type: "object",
					},
					"spec": apiext.JSONSchemaProps{
						Type:        "object",
						Description: "IssuerSpec is the specification of an Issuer. This includes any configuration required for the issuer.",
						Properties: map[string]apiext.JSONSchemaProps{
							"acme": apiext.JSONSchemaProps{
								Description: "ACMEIssuer contains the specification for an ACME issuer",
								Properties: map[string]apiext.JSONSchemaProps{
									"dns01": apiext.JSONSchemaProps{
										Description: "DEPRECATED: DNS-01 config",
										Properties: map[string]apiext.JSONSchemaProps{
											"providers": apiext.JSONSchemaProps{
												Items: &apiext.JSONSchemaPropsOrArray{
													Schema: &apiext.JSONSchemaProps{
														Description: "ACMEIssuerDNS01Provider contains configuration for a DNS provider that can be used to solve ACME DNS01 challenges.",
														Properties: map[string]apiext.JSONSchemaProps{
															"acmedns": apiext.JSONSchemaProps{
																Description: "ACMEIssuerDNS01ProviderAcmeDNS is a structure containing the configuration for ACME-DNS servers",
																Properties: map[string]apiext.JSONSchemaProps{
																	"accountSecretRef": apiext.JSONSchemaProps{
																		Properties: map[string]apiext.JSONSchemaProps{
																			"key": apiext.JSONSchemaProps{
																				Description: "The key of the secret to select from. Must be a valid secret key.",
																				Type:        "string",
																			},
																			"name": apiext.JSONSchemaProps{
																				Description: "Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?",
																				Type:        "string",
																			},
																		},
																		Required: []string{"name"},
																		Type:     "object",
																	},
																	"host": apiext.JSONSchemaProps{
																		Type: "string",
																	},
																},
																Required: []string{"accountSecretRef", "host"},
																Type:     "object",
															},
															"akami": apiext.JSONSchemaProps{
																Description: "ACMEIssuerDNS01ProviderAkamai is a structure containing the DNS configuration for Akamai DNS—Zone Record Management API",
																Properties: map[string]apiext.JSONSchemaProps{
																	"accessTokenSecretRef": apiext.JSONSchemaProps{
																		Properties: map[string]apiext.JSONSchemaProps{
																			"key": apiext.JSONSchemaProps{
																				Description: "The key of the secret to select from. Must be a valid secret key.",
																				Type:        "string",
																			},
																			"name": apiext.JSONSchemaProps{
																				Description: "Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?",
																				Type:        "string",
																			},
																		},
																		Required: []string{"name"},
																		Type:     "object",
																	},
																	"clientSecretSecretRef": apiext.JSONSchemaProps{Properties: map[string]apiext.JSONSchemaProps{
																		"key": apiext.JSONSchemaProps{
																			Description: "The key of the secret to select from. Must be a valid secret key.",
																			Type:        "string",
																		},
																		"name": apiext.JSONSchemaProps{
																			Description: "Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?",
																			Type:        "string",
																		},
																	},
																		Required: []string{"name"},
																		Type:     "object",
																	},
																	"clientTokenSecretRef": apiext.JSONSchemaProps{Properties: map[string]apiext.JSONSchemaProps{
																		"key": apiext.JSONSchemaProps{
																			Description: "The key of the secret to select from. Must be a valid secret key.",
																			Type:        "string",
																		},
																		"name": apiext.JSONSchemaProps{
																			Description: "Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?",
																			Type:        "string",
																		},
																	},
																		Required: []string{"name"},
																		Type:     "object",
																	},
																	"serviceConsumerDomain": apiext.JSONSchemaProps{
																		Type: "string",
																	},
																},
																Required: []string{"accessTokenSecretRef", "clientSecretSecretRef", "clientTokenSecretRef", "serviceConsumerDomain"},
																Type:     "object",
															},
															"azuredns": apiext.JSONSchemaProps{
																Description: "ACMEIssuerDNS01ProviderAzureDNS is a structure containing the configuration for Azure DNS",
																Properties: map[string]apiext.JSONSchemaProps{
																	"clientID": apiext.JSONSchemaProps{
																		Type: "string",
																	},
																	"clientSecretSecretRef": apiext.JSONSchemaProps{
																		Properties: map[string]apiext.JSONSchemaProps{
																			"key": apiext.JSONSchemaProps{
																				Description: "The key of the secret to select from. Must be a valid secret key.",
																				Type:        "string",
																			},
																			"name": apiext.JSONSchemaProps{
																				Description: "Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?",
																				Type:        "string",
																			},
																		},
																		Required: []string{"name"},
																		Type:     "object",
																	},
																	"environment": apiext.JSONSchemaProps{
																		Enum: []apiext.JSON{
																			{Raw: []byte("\"AzurePublicCloud\"")},
																			{Raw: []byte("\"AzureChinaCloud\"")},
																			{Raw: []byte("\"AzureGermanCloud\"")},
																			{Raw: []byte("\"AzureUSGovernmentCloud\"")},
																		},
																		Type: "string",
																	},
																	"hostedZoneName": apiext.JSONSchemaProps{
																		Type: "string",
																	},
																	"resourceGroupName": apiext.JSONSchemaProps{
																		Type: "string",
																	},
																	"subscriptionID": apiext.JSONSchemaProps{
																		Type: "string",
																	},
																	"tenantID": apiext.JSONSchemaProps{
																		Type: "string",
																	},
																},
																Required: []string{"clientID", "clientSecretSecretRef", "resourceGroupName", "subscriptionID", "tenantID"},
																Type:     "object",
															},
															"clouddns": apiext.JSONSchemaProps{
																Description: "ACMEIssuerDNS01ProviderCloudDNS is a structure containing the DNS configuration for Google Cloud DNS",
																Properties: map[string]apiext.JSONSchemaProps{
																	"project": apiext.JSONSchemaProps{
																		Type: "string",
																	},
																	"serviceAccountSecretRef": apiext.JSONSchemaProps{
																		Properties: map[string]apiext.JSONSchemaProps{
																			"key": apiext.JSONSchemaProps{
																				Description: "The key of the secret to select from. Must be a valid secret key.",
																				Type:        "string",
																			},
																			"name": apiext.JSONSchemaProps{
																				Description: "Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?",
																				Type:        "string",
																			},
																		},
																		Type:     "object",
																		Required: []string{"name"},
																	},
																},
																Required: []string{"project", "serviceAccountSecretRef"},
																Type:     "object",
															},
															"cloudflare": apiext.JSONSchemaProps{
																Description: "ACMEIssuerDNS01ProviderCloudflare is a structure containing the DNS configuration for Cloudflare",
																Properties: map[string]apiext.JSONSchemaProps{
																	"apiKeySecretRef": apiext.JSONSchemaProps{
																		Properties: map[string]apiext.JSONSchemaProps{
																			"key": apiext.JSONSchemaProps{
																				Description: "The key of the secret to select from. Must be a valid secret key.",
																				Type:        "string",
																			},
																			"name": apiext.JSONSchemaProps{
																				Description: "Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?",
																				Type:        "string",
																			},
																		},
																		Required: []string{"name"},
																		Type:     "object",
																	},
																	"email": apiext.JSONSchemaProps{Type: "string"},
																},
																Required: []string{"apiKeySecretRef", "email"},
																Type:     "object",
															},
															"cnameStrategy": apiext.JSONSchemaProps{
																Description: "CNAMEStrategy configures how the DNS01 provider should handle CNAME records when found in DNS zones.",
																Enum: []apiext.JSON{
																	{Raw: []byte("\"None\"")},
																	{Raw: []byte("\"Follow\"")},
																},
																Type: "string",
															},
															"digitalocean": apiext.JSONSchemaProps{
																Description: "ACMEIssuerDNS01ProviderDigitalOcean is a structure containing the DNS configuration for DigitalOcean Domains",
																Properties: map[string]apiext.JSONSchemaProps{
																	"tokenSecretRef": apiext.JSONSchemaProps{
																		Properties: map[string]apiext.JSONSchemaProps{
																			"key": apiext.JSONSchemaProps{
																				Description: "The key of the secret to select from. Must be a valid secret key.",
																				Type:        "string",
																			},
																			"name": apiext.JSONSchemaProps{
																				Description: "Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?",
																				Type:        "string",
																			},
																		},
																		Required: []string{"name"},
																		Type:     "object",
																	},
																},
																Required: []string{"tokenSecretRef"},
																Type:     "object",
															},
															"name": apiext.JSONSchemaProps{
																Description: "Name is the name of the DNS provider, which should be used to reference this DNS provider configuration on Certificate resources.",
																Type:        "string",
															},
															"rfc2136": apiext.JSONSchemaProps{
																Description: "ACMEIssuerDNS01ProviderRFC2136 is a structure containing the configuration for RFC2136 DNS",
																Properties: map[string]apiext.JSONSchemaProps{
																	"nameserver": apiext.JSONSchemaProps{
																		Description: "The IP address of the DNS supporting RFC2136. Required. Note: FQDN is not a valid value, only IP.",
																		Type:        "string",
																	},
																	"tsigAlgorithm": apiext.JSONSchemaProps{
																		Description: "The TSIG Algorithm configured in the DNS supporting RFC2136. Used only when `tsigSecretSecretRef` and `tsigKeyName` are defined. Supported values are (case-insensitive): `HMACMD5` (default), `HMACSHA1`, `HMACSHA256` or `HMACSHA512`",
																		Type:        "string",
																	},
																	"tsigKeyName": apiext.JSONSchemaProps{
																		Description: "The TSIG Key name configured in the DNS. If `tsigSecretSecretRef` is defined, this field is required.",
																		Type:        "string",
																	},
																	"tsigSecretSecretRef": apiext.JSONSchemaProps{
																		Description: "The name of the secret containing the TSIG value. If `tsigKeyName` is defined, this field is required.",
																		Properties: map[string]apiext.JSONSchemaProps{
																			"key": apiext.JSONSchemaProps{
																				Description: "The key of the secret to select from. Must be a valid secret key.",
																				Type:        "string",
																			},
																			"name": apiext.JSONSchemaProps{
																				Description: "Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?",
																				Type:        "string",
																			},
																		},
																		Required: []string{"name"},
																		Type:     "object",
																	},
																},
																Required: []string{"nameserver"},
																Type:     "object",
															},
															"route53": apiext.JSONSchemaProps{
																Description: "ACMEIssuerDNS01ProviderRoute53 is a structure containing the Route 53 configuration for AWS",
																Properties: map[string]apiext.JSONSchemaProps{
																	"accessKeyID": apiext.JSONSchemaProps{
																		Description: "The AccessKeyID is used for authentication. If not set we fall-back to using env vars, shared credentials file or AWS Instance metadata see: https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html#specifying-credentials",
																		Type:        "string",
																	},
																	"hostedZoneID": apiext.JSONSchemaProps{
																		Description: "If set, the provider will manage only this zone in Route53 and will not do an lookup using the route53:ListHostedZonesByName api call.",
																		Type:        "string",
																	},
																	"region": apiext.JSONSchemaProps{
																		Description: "Always set the region when using AccessKeyID and SecretAccessKey",
																		Type:        "string",
																	},
																	"role": apiext.JSONSchemaProps{
																		Description: "Role is a Role ARN which the Route53 provider will assume using either the explicit credentials AccessKeyID/SecretAccessKey or the inferred credentials from environment variables, shared credentials file or AWS Instance metadata",
																		Type:        "string",
																	},
																	"secretAccessKeySecretRef": apiext.JSONSchemaProps{
																		Description: "The SecretAccessKey is used for authentication. If not set we fall-back to using env vars, shared credentials file or AWS Instance metadata https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html#specifying-credentials",
																		Properties: map[string]apiext.JSONSchemaProps{
																			"key": apiext.JSONSchemaProps{
																				Description: "The key of the secret to select from. Must be a valid secret key.",
																				Type:        "string",
																			},
																			"name": apiext.JSONSchemaProps{
																				Description: "Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?",
																				Type:        "string",
																			},
																		},
																		Required: []string{"name"},
																		Type:     "object",
																	},
																},
																Required: []string{"region"},
																Type:     "object",
															},
															"webhook": apiext.JSONSchemaProps{
																Description: "ACMEIssuerDNS01ProviderWebhook specifies configuration for a webhook DNS01 provider, including where to POST ChallengePayload resources.",
																Properties: map[string]apiext.JSONSchemaProps{
																	"config": apiext.JSONSchemaProps{
																		Description: "Additional configuration that should be passed to the webhook apiserver when challenges are processed. This can contain arbitrary JSON data. Secret values should not be specified in this stanza. If secret values are needed (e.g. credentials for a DNS service), you should use a SecretKeySelector to reference a Secret resource. For details on the schema of this field, consult the webhook provider implementation's documentation.",
																		Type:        "object",
																	},
																	"groupName": apiext.JSONSchemaProps{
																		Description: "The API group name that should be used when POSTing ChallengePayload resources to the webhook apiserver. This should be the same as the GroupName specified in the webhook provider implementation.",
																		Type:        "string",
																	},
																	"solverName": apiext.JSONSchemaProps{
																		Description: "The name of the solve to use as defined in the webhook provider implementation. This will typically be the name of th eprovider, e.g. 'cloudflare'.",
																		Type:        "string",
																	},
																},
																Required: []string{"groupName", "solverName"},
																Type:     "object",
															},
														},
														Required: []string{"name"},
														Type:     "object",
													},
												},
												Type: "array",
											},
										},
										Type: "object",
									},
									"email": apiext.JSONSchemaProps{
										Description: "Email is the email for this account",
										Type:        "string",
									},
									"http01": apiext.JSONSchemaProps{
										Description: "DEPRECATED: HTTP-01 config",
										Properties: map[string]apiext.JSONSchemaProps{
											"serviceType": apiext.JSONSchemaProps{
												Description: "Optional service type for Kubernetes solver service",
												Type:        "string",
											},
										},
										Type: "object",
									},
									"privateKeySecretRef": apiext.JSONSchemaProps{
										Description: "PrivateKey is the name of a secret containing the private key for this user account.",
										Properties: map[string]apiext.JSONSchemaProps{
											"key": apiext.JSONSchemaProps{
												Description: "The key of the secret to select from. Must be a valid secret key.",
												Type:        "string",
											},
											"name": apiext.JSONSchemaProps{
												Description: "Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?",
												Type:        "string",
											},
										},
										Required: []string{"name"},
										Type:     "object",
									},
									"server": apiext.JSONSchemaProps{
										Description: "Server is the ACME server URL",
										Type:        "string",
									},
									"skipTLSVerify": apiext.JSONSchemaProps{
										Description: "If true, skip verifying the ACME server TLS certificate",
										Type:        "boolean",
									},
									"solvers": apiext.JSONSchemaProps{
										Description: "Solvers is a list of challenge solvers that will be used to solve ACME challenges for the matching domains.",
										Items: &apiext.JSONSchemaPropsOrArray{
											Schema: &apiext.JSONSchemaProps{
												Properties: map[string]apiext.JSONSchemaProps{
													"dns01": apiext.JSONSchemaProps{
														Properties: map[string]apiext.JSONSchemaProps{
															"acmedns": apiext.JSONSchemaProps{
																Description: "ACMEIssuerDNS01ProviderAcmeDNS is a structure containing the configuration for ACME-DNS servers",
																Properties: map[string]apiext.JSONSchemaProps{
																	"accountSecretRef": apiext.JSONSchemaProps{
																		Properties: map[string]apiext.JSONSchemaProps{
																			"key": apiext.JSONSchemaProps{
																				Description: "The key of the secret to select from. Must be a valid secret key.",
																				Type:        "string",
																			},
																			"name": apiext.JSONSchemaProps{
																				Description: "Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?",
																				Type:        "string",
																			},
																		},
																		Required: []string{"name"},
																		Type:     "object",
																	},
																	"host": apiext.JSONSchemaProps{
																		Type: "string",
																	},
																},
																Required: []string{"accountSecretRef", "host"},
																Type:     "object",
															},
															"akamai": apiext.JSONSchemaProps{
																Description: "ACMEIssuerDNS01ProviderAkamai is a structure containing the DNS configuration for Akamai DNS—Zone Record Management API",
																Properties: map[string]apiext.JSONSchemaProps{
																	"accessTokenSecretRef": apiext.JSONSchemaProps{
																		Properties: map[string]apiext.JSONSchemaProps{
																			"key": apiext.JSONSchemaProps{
																				Description: "The key of the secret to select from. Must be a valid secret key.",
																				Type:        "string",
																			},
																			"name": apiext.JSONSchemaProps{
																				Description: "Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?",
																				Type:        "string",
																			},
																		},
																		Required: []string{"name"},
																		Type:     "object",
																	},
																	"clientSecretSecretRef": apiext.JSONSchemaProps{Properties: map[string]apiext.JSONSchemaProps{
																		"key": apiext.JSONSchemaProps{
																			Description: "The key of the secret to select from. Must be a valid secret key.",
																			Type:        "string",
																		},
																		"name": apiext.JSONSchemaProps{
																			Description: "Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?",
																			Type:        "string",
																		},
																	},
																		Required: []string{"name"},
																		Type:     "object",
																	},
																	"clientTokenSecretRef": apiext.JSONSchemaProps{Properties: map[string]apiext.JSONSchemaProps{
																		"key": apiext.JSONSchemaProps{
																			Description: "The key of the secret to select from. Must be a valid secret key.",
																			Type:        "string",
																		},
																		"name": apiext.JSONSchemaProps{
																			Description: "Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?",
																			Type:        "string",
																		},
																	},
																		Required: []string{"name"},
																		Type:     "object",
																	},
																	"serviceConsumerDomain": apiext.JSONSchemaProps{
																		Type: "string",
																	},
																},
																Required: []string{"accessTokenSecretRef", "clientSecretSecretRef", "clientTokenSecretRef", "serviceConsumerDomain"},
																Type:     "object",
															},
															"azuredns": apiext.JSONSchemaProps{
																Description: "ACMEIssuerDNS01ProviderAzureDNS is a structure containing the configuration for Azure DNS",
																Properties: map[string]apiext.JSONSchemaProps{
																	"clientID": apiext.JSONSchemaProps{
																		Type: "string",
																	},
																	"clientSecretSecretRef": apiext.JSONSchemaProps{
																		Properties: map[string]apiext.JSONSchemaProps{
																			"key": apiext.JSONSchemaProps{
																				Description: "The key of the secret to select from. Must be a valid secret key.",
																				Type:        "string",
																			},
																			"name": apiext.JSONSchemaProps{
																				Description: "Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?",
																				Type:        "string",
																			},
																		},
																		Required: []string{"name"},
																		Type:     "object",
																	},
																	"environment": apiext.JSONSchemaProps{
																		Enum: []apiext.JSON{
																			{Raw: []byte("\"AzurePublicCloud\"")},
																			{Raw: []byte("\"AzureChinaCloud\"")},
																			{Raw: []byte("\"AzureGermanCloud\"")},
																			{Raw: []byte("\"AzureUSGovernmentCloud\"")},
																		},
																		Type: "string",
																	},
																	"hostedZoneName": apiext.JSONSchemaProps{
																		Type: "string",
																	},
																	"resourceGroupName": apiext.JSONSchemaProps{
																		Type: "string",
																	},
																	"subscriptionID": apiext.JSONSchemaProps{
																		Type: "string",
																	},
																	"tenantID": apiext.JSONSchemaProps{
																		Type: "string",
																	},
																},
																Required: []string{"clientID", "clientSecretSecretRef", "resourceGroupName", "subscriptionID", "tenantID"},
																Type:     "object",
															},
															"clouddns": apiext.JSONSchemaProps{
																Description: "ACMEIssuerDNS01ProviderCloudDNS is a structure containing the DNS configuration for Google Cloud DNS",
																Properties: map[string]apiext.JSONSchemaProps{
																	"project": apiext.JSONSchemaProps{
																		Type: "string",
																	},
																	"serviceAccountSecretRef": apiext.JSONSchemaProps{
																		Properties: map[string]apiext.JSONSchemaProps{
																			"key": apiext.JSONSchemaProps{
																				Description: "The key of the secret to select from. Must be a valid secret key.",
																				Type:        "string",
																			},
																			"name": apiext.JSONSchemaProps{
																				Description: "Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?",
																				Type:        "string",
																			},
																		},
																		Type:     "object",
																		Required: []string{"name"},
																	},
																},
																Required: []string{"project", "serviceAccountSecretRef"},
																Type:     "object",
															},
															"cloudflare": apiext.JSONSchemaProps{
																Description: "ACMEIssuerDNS01ProviderCloudflare is a structure containing the DNS configuration for Cloudflare",
																Properties: map[string]apiext.JSONSchemaProps{
																	"apiKeySecretRef": apiext.JSONSchemaProps{
																		Properties: map[string]apiext.JSONSchemaProps{
																			"key": apiext.JSONSchemaProps{
																				Description: "The key of the secret to select from. Must be a valid secret key.",
																				Type:        "string",
																			},
																			"name": apiext.JSONSchemaProps{
																				Description: "Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?",
																				Type:        "string",
																			},
																		},
																		Required: []string{"name"},
																		Type:     "object",
																	},
																	"email": apiext.JSONSchemaProps{Type: "string"},
																},
																Required: []string{"apiKeySecretRef", "email"},
																Type:     "object",
															},
															"cnameStrategy": apiext.JSONSchemaProps{
																Description: "CNAMEStrategy configures how the DNS01 provider should handle CNAME records when found in DNS zones.",
																Enum: []apiext.JSON{
																	{Raw: []byte("\"None\"")},
																	{Raw: []byte("\"Follow\"")},
																},
																Type: "string",
															},
															"digitalocean": apiext.JSONSchemaProps{
																Description: "ACMEIssuerDNS01ProviderDigitalOcean is a structure containing the DNS configuration for DigitalOcean Domains",
																Properties: map[string]apiext.JSONSchemaProps{
																	"tokenSecretRef": apiext.JSONSchemaProps{
																		Properties: map[string]apiext.JSONSchemaProps{
																			"key": apiext.JSONSchemaProps{
																				Description: "The key of the secret to select from. Must be a valid secret key.",
																				Type:        "string",
																			},
																			"name": apiext.JSONSchemaProps{
																				Description: "Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?",
																				Type:        "string",
																			},
																		},
																		Required: []string{"name"},
																		Type:     "object",
																	},
																},
																Required: []string{"tokenSecretRef"},
																Type:     "object",
															},
															"name": apiext.JSONSchemaProps{
																Description: "Name is the name of the DNS provider, which should be used to reference this DNS provider configuration on Certificate resources.",
																Type:        "string",
															},
															"rfc2136": apiext.JSONSchemaProps{
																Description: "ACMEIssuerDNS01ProviderRFC2136 is a structure containing the configuration for RFC2136 DNS",
																Properties: map[string]apiext.JSONSchemaProps{
																	"nameserver": apiext.JSONSchemaProps{
																		Description: "The IP address of the DNS supporting RFC2136. Required. Note: FQDN is not a valid value, only IP.",
																		Type:        "string",
																	},
																	"tsigAlgorithm": apiext.JSONSchemaProps{
																		Description: "The TSIG Algorithm configured in the DNS supporting RFC2136. Used only when `tsigSecretSecretRef` and `tsigKeyName` are defined. Supported values are (case-insensitive): `HMACMD5` (default), `HMACSHA1`, `HMACSHA256` or `HMACSHA512`",
																		Type:        "string",
																	},
																	"tsigKeyName": apiext.JSONSchemaProps{
																		Description: "The TSIG Key name configured in the DNS. If `tsigSecretSecretRef` is defined, this field is required.",
																		Type:        "string",
																	},
																	"tsigSecretSecretRef": apiext.JSONSchemaProps{
																		Description: "The name of the secret containing the TSIG value. If `tsigKeyName` is defined, this field is required.",
																		Properties: map[string]apiext.JSONSchemaProps{
																			"key": apiext.JSONSchemaProps{
																				Description: "The key of the secret to select from. Must be a valid secret key.",
																				Type:        "string",
																			},
																			"name": apiext.JSONSchemaProps{
																				Description: "Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?",
																				Type:        "string",
																			},
																		},
																		Required: []string{"name"},
																		Type:     "object",
																	},
																},
																Required: []string{"nameserver"},
																Type:     "object",
															},
															"route53": apiext.JSONSchemaProps{
																Description: "ACMEIssuerDNS01ProviderRoute53 is a structure containing the Route 53 configuration for AWS",
																Properties: map[string]apiext.JSONSchemaProps{
																	"accessKeyID": apiext.JSONSchemaProps{
																		Description: "The AccessKeyID is used for authentication. If not set we fall-back to using env vars, shared credentials file or AWS Instance metadata see: https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html#specifying-credentials",
																		Type:        "string",
																	},
																	"hostedZoneID": apiext.JSONSchemaProps{
																		Description: "If set, the provider will manage only this zone in Route53 and will not do an lookup using the route53:ListHostedZonesByName api call.",
																		Type:        "string",
																	},
																	"region": apiext.JSONSchemaProps{
																		Description: "Always set the region when using AccessKeyID and SecretAccessKey",
																		Type:        "string",
																	},
																	"role": apiext.JSONSchemaProps{
																		Description: "Role is a Role ARN which the Route53 provider will assume using either the explicit credentials AccessKeyID/SecretAccessKey or the inferred credentials from environment variables, shared credentials file or AWS Instance metadata",
																		Type:        "string",
																	},
																	"secretAccessKeySecretRef": apiext.JSONSchemaProps{
																		Description: "The SecretAccessKey is used for authentication. If not set we fall-back to using env vars, shared credentials file or AWS Instance metadata https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html#specifying-credentials",
																		Properties: map[string]apiext.JSONSchemaProps{
																			"key": apiext.JSONSchemaProps{
																				Description: "The key of the secret to select from. Must be a valid secret key.",
																				Type:        "string",
																			},
																			"name": apiext.JSONSchemaProps{
																				Description: "Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?",
																				Type:        "string",
																			},
																		},
																		Required: []string{"name"},
																		Type:     "object",
																	},
																},
																Required: []string{"region"},
																Type:     "object",
															},
															"webhook": apiext.JSONSchemaProps{
																Description: "ACMEIssuerDNS01ProviderWebhook specifies configuration for a webhook DNS01 provider, including where to POST ChallengePayload resources.",
																Properties: map[string]apiext.JSONSchemaProps{
																	"config": apiext.JSONSchemaProps{
																		Description: "Additional configuration that should be passed to the webhook apiserver when challenges are processed. This can contain arbitrary JSON data. Secret values should not be specified in this stanza. If secret values are needed (e.g. credentials for a DNS service), you should use a SecretKeySelector to reference a Secret resource. For details on the schema of this field, consult the webhook provider implementation's documentation.",
																		Type:        "object",
																	},
																	"groupName": apiext.JSONSchemaProps{
																		Description: "The API group name that should be used when POSTing ChallengePayload resources to the webhook apiserver. This should be the same as the GroupName specified in the webhook provider implementation.",
																		Type:        "string",
																	},
																	"solverName": apiext.JSONSchemaProps{
																		Description: "The name of the solve to use as defined in the webhook provider implementation. This will typically be the name of th eprovider, e.g. 'cloudflare'.",
																		Type:        "string",
																	},
																},
																Required: []string{"groupName", "solverName"},
																Type:     "object",
															},
														},
														Type: "object",
													},
													"http01": apiext.JSONSchemaProps{
														Description: "ACMEChallengeSolverHTTP01 contains configuration detailing how to solve HTTP01 challenges within a Kubernetes cluster. Typically this is accomplished through creating 'routes' of some description that configure ingress controllers to direct traffic to 'solver pods', which are responsible for responding to the ACME server's HTTP requests.",
														Properties: map[string]apiext.JSONSchemaProps{
															"ingress": apiext.JSONSchemaProps{
																Description: "The ingress based HTTP01 challenge solver will solve challenges by creating or modifying Ingress resources in order to route requests for '/.well-known/acme-challenge/XYZ' to 'challenge solver' pods that are provisioned by cert-manager for each Challenge to be completed.",
																Properties: map[string]apiext.JSONSchemaProps{
																	"class": apiext.JSONSchemaProps{
																		Description: "The ingress class to use when creating Ingress resources to solve ACME challenges that use this challenge solver. Only one of 'class' or 'name' may be specified.",
																		Type:        "string",
																	},
																	"name": apiext.JSONSchemaProps{
																		Description: "The name of the ingress resource that should have ACME challenge solving routes inserted into it in order to solve HTTP01 challenges. This is typically used in conjunction with ingress controllers like ingress-gce, which maintains a 1:1 mapping between external IPs and ingress resources.",
																		Type:        "string",
																	},
																	"podTemplate": apiext.JSONSchemaProps{
																		Description: "Optional pod template used to configure the ACME challenge solver pods used for HTTP01 challenges",
																		Properties: map[string]apiext.JSONSchemaProps{
																			"metadata": apiext.JSONSchemaProps{
																				Description: "ObjectMeta overrides for the pod used to solve HTTP01 challenges. Only the 'labels' and 'annotations' fields may be set. If labels or annotations overlap with in-built values, the values here will override the in-built values.",
																				Type:        "object",
																			},
																			"spec": apiext.JSONSchemaProps{
																				Description: "PodSpec defines overrides for the HTTP01 challenge solver pod. Only the 'nodeSelector', 'affinity' and 'tolerations' fields are supported currently. All other fields will be ignored.",
																				Properties: map[string]apiext.JSONSchemaProps{
																					"affinity": apiext.JSONSchemaProps{
																						Description: "If specified, the pod's scheduling constraints",
																						Properties: map[string]apiext.JSONSchemaProps{
																							"nodeAffinity": apiext.JSONSchemaProps{
																								Description: "Describes node affinity scheduling rules for the pod.",
																								Properties: map[string]apiext.JSONSchemaProps{
																									"preferredDuringSchedulingIgnoredDuringExecution": apiext.JSONSchemaProps{
																										Description: "The scheduler will prefer to schedule pods to nodes that satisfy the affinity expressions specified by this field, but it may choose a node that violates one or more of the expressions. The node that is most preferred is the one with the greatest sum of weights, i.e. for each node that meets all of the scheduling requirements (resource request, requiredDuringScheduling affinity expressions, etc.), compute a sum by iterating through the elements of this field and adding 'weight' to the sum if the node matches the corresponding matchExpressions; the node(s) with the highest sum are the most preferred.",
																										Items: &apiext.JSONSchemaPropsOrArray{
																											Schema: &apiext.JSONSchemaProps{
																												Description: "An empty preferred scheduling term matches all objects with implicit weight 0 (i.e. it's a no-op). A null preferred scheduling term matches no objects (i.e. is also a no-op).",
																												Properties: map[string]apiext.JSONSchemaProps{
																													"preference": apiext.JSONSchemaProps{
																														Description: "A node selector term, associated with the corresponding weight.",
																														Properties: map[string]apiext.JSONSchemaProps{
																															"matchExpressions": apiext.JSONSchemaProps{
																																Description: "A list of node selector requirements by node's labels.",
																																Items: &apiext.JSONSchemaPropsOrArray{
																																	Schema: &apiext.JSONSchemaProps{
																																		Description: "A node selector requirement is a selector that contains values, a key, and an operator that relates the key and values.",
																																		Properties: map[string]apiext.JSONSchemaProps{
																																			"key": apiext.JSONSchemaProps{
																																				Description: "The label key that the selector applies to.",
																																				Type:        "string",
																																			},
																																			"operator": apiext.JSONSchemaProps{
																																				Description: "Represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists, DoesNotExist. Gt, and Lt.",
																																				Type:        "string",
																																			},
																																			"values": apiext.JSONSchemaProps{
																																				Description: "An array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. If the operator is Gt or Lt, the values array must have a single element, which will be intepreted as an integer. This array is replaced during a strategic merge patch.",
																																				Items: &apiext.JSONSchemaPropsOrArray{
																																					Schema: &apiext.JSONSchemaProps{
																																						Type: "string",
																																					},
																																				},
																																				Type: "array",
																																			},
																																		},
																																		Required: []string{"key", "operator"},
																																		Type:     "object",
																																	},
																																},
																																Type: "array",
																															},
																															"matchFields": apiext.JSONSchemaProps{
																																Description: "A list of node selector requirements by node's labels.",
																																Items: &apiext.JSONSchemaPropsOrArray{
																																	Schema: &apiext.JSONSchemaProps{
																																		Description: "A node selector requirement is a selector that contains values, a key, and an operator that relates the key and values.",
																																		Properties: map[string]apiext.JSONSchemaProps{
																																			"key": apiext.JSONSchemaProps{
																																				Description: "The label key that the selector applies to.",
																																				Type:        "string",
																																			},
																																			"operator": apiext.JSONSchemaProps{
																																				Description: "Represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists, DoesNotExist. Gt, and Lt.",
																																				Type:        "string",
																																			},
																																			"values": apiext.JSONSchemaProps{
																																				Description: "An array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. If the operator is Gt or Lt, the values array must have a single element, which will be intepreted as an integer. This array is replaced during a strategic merge patch.",
																																				Items: &apiext.JSONSchemaPropsOrArray{
																																					Schema: &apiext.JSONSchemaProps{
																																						Type: "string",
																																					},
																																				},
																																				Type: "array",
																																			},
																																		},
																																		Required: []string{"key", "operator"},
																																		Type:     "object",
																																	},
																																},
																																Type: "array",
																															},
																														},
																														Type: "object",
																													},
																													"weight": apiext.JSONSchemaProps{
																														Description: "Weight associated with matching the corresponding nodeSelectorTerm, in the range 1-100.",
																														Format:      "int32",
																														Type:        "integer",
																													},
																												},
																												Required: []string{"preference", "weight"},
																												Type:     "object",
																											},
																										},
																										Type: "array",
																									},
																									"requiredDuringSchedulingIgnoredDuringExecution": apiext.JSONSchemaProps{
																										Description: "If the affinity requirements specified by this field are not met at scheduling time, the pod will not be scheduled onto the node. If the affinity requirements specified by this field cease to be met at some point during pod execution (e.g. due to an update), the system may or may not try to eventually evict the pod from its node.",
																										Properties: map[string]apiext.JSONSchemaProps{
																											"nodeSelectorTerms": apiext.JSONSchemaProps{
																												Description: "Required. A list of node selector terms. The terms are ORed.",
																												Items: &apiext.JSONSchemaPropsOrArray{
																													Schema: &apiext.JSONSchemaProps{
																														Description: "A null or empty node selector term matches no objects. The requirements of them are ANDed. The TopologySelectorTerm type implements a subset of the NodeSelectorTerm",
																														Properties: map[string]apiext.JSONSchemaProps{
																															"matchExpressions": apiext.JSONSchemaProps{
																																Description: "A list of node selector requirements by node's labels.",
																																Items: &apiext.JSONSchemaPropsOrArray{
																																	Schema: &apiext.JSONSchemaProps{
																																		Description: "A node selector requirement is a selector that contains values, a key, and an operator that relates the key and values.",
																																		Properties: map[string]apiext.JSONSchemaProps{
																																			"key": apiext.JSONSchemaProps{
																																				Description: "The label key that the selector applies to.",
																																				Type:        "string",
																																			},
																																			"operator": apiext.JSONSchemaProps{
																																				Description: "Represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists, DoesNotExist. Gt, and Lt.",
																																				Type:        "string",
																																			},
																																			"values": apiext.JSONSchemaProps{
																																				Description: "An array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. If the operator is Gt or Lt, the values array must have a single element, which will be intepreted as an integer. This array is replaced during a strategic merge patch.",
																																				Items: &apiext.JSONSchemaPropsOrArray{
																																					Schema: &apiext.JSONSchemaProps{
																																						Type: "string",
																																					},
																																				},
																																				Type: "array",
																																			},
																																		},
																																		Required: []string{"key", "operator"},
																																		Type:     "object",
																																	},
																																},
																																Type: "array",
																															},
																															"matchFields": apiext.JSONSchemaProps{
																																Description: "A list of node selector requirements by node's labels.",
																																Items: &apiext.JSONSchemaPropsOrArray{
																																	Schema: &apiext.JSONSchemaProps{
																																		Description: "A node selector requirement is a selector that contains values, a key, and an operator that relates the key and values.",
																																		Properties: map[string]apiext.JSONSchemaProps{
																																			"key": apiext.JSONSchemaProps{
																																				Description: "The label key that the selector applies to.",
																																				Type:        "string",
																																			},
																																			"operator": apiext.JSONSchemaProps{
																																				Description: "Represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists, DoesNotExist. Gt, and Lt.",
																																				Type:        "string",
																																			},
																																			"values": apiext.JSONSchemaProps{
																																				Description: "An array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. If the operator is Gt or Lt, the values array must have a single element, which will be intepreted as an integer. This array is replaced during a strategic merge patch.",
																																				Items: &apiext.JSONSchemaPropsOrArray{
																																					Schema: &apiext.JSONSchemaProps{
																																						Type: "string",
																																					},
																																				},
																																				Type: "array",
																																			},
																																		},
																																		Required: []string{"key", "operator"},
																																		Type:     "object",
																																	},
																																},
																																Type: "array",
																															},
																														},
																														Type: "object",
																													},
																												},
																												Type: "array",
																											},
																										},
																										Required: []string{"nodeSelectorTerms"},
																										Type:     "object",
																									},
																								},
																								Type: "object",
																							},
																							"podAffinity": apiext.JSONSchemaProps{
																								Description: "Describes pod affinity scheduling rules (e.g. co-locate this pod in the same node, zone, etc. as some other pod(s)).",
																								Properties: map[string]apiext.JSONSchemaProps{
																									"preferredDuringSchedulingIgnoredDuringExecution": apiext.JSONSchemaProps{
																										Description: "The scheduler will prefer to schedule pods to nodes that satisfy the affinity expressions specified by this field, but it may choose a node that violates one or more of the expressions. The node that is most preferred is the one with the greatest sum of weights, i.e. for each node that meets all of the scheduling requirements (resource request, requiredDuringScheduling affinity expressions, etc.), compute a sum by iterating through the elements of this field and adding 'weight' to the sum if the node matches the corresponding matchExpressions; the node(s) with the highest sum are the most preferred.",
																										Items: &apiext.JSONSchemaPropsOrArray{
																											Schema: &apiext.JSONSchemaProps{
																												Description: "An empty preferred scheduling term matches all objects with implicit weight 0 (i.e. it's a no-op). A null preferred scheduling term matches no objects (i.e. is also a no-op).",
																												Properties: map[string]apiext.JSONSchemaProps{
																													"podAffinityTerm": apiext.JSONSchemaProps{
																														Description: "Required. A pod affinity term, associated with the corresponding weight..",
																														Properties: map[string]apiext.JSONSchemaProps{
																															"labelSelector": apiext.JSONSchemaProps{
																																Description: "A label query over a set of resources in this case pods.",
																																Properties: map[string]apiext.JSONSchemaProps{
																																	"matchExpressions": apiext.JSONSchemaProps{
																																		Description: "matchExpressions is a list of label selector requirements. The requirements are ANDed",
																																		Items: &apiext.JSONSchemaPropsOrArray{
																																			Schema: &apiext.JSONSchemaProps{
																																				Description: "A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.",
																																				Properties: map[string]apiext.JSONSchemaProps{
																																					"key": apiext.JSONSchemaProps{
																																						Description: "The label key that the selector applies to.",
																																						Type:        "string",
																																					},
																																					"operator": apiext.JSONSchemaProps{
																																						Description: "Represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists, DoesNotExist. Gt, and Lt.",
																																						Type:        "string",
																																					},
																																					"values": apiext.JSONSchemaProps{
																																						Description: "An array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. If the operator is Gt or Lt, the values array must have a single element, which will be intepreted as an integer. This array is replaced during a strategic merge patch.",
																																						Items: &apiext.JSONSchemaPropsOrArray{
																																							Schema: &apiext.JSONSchemaProps{
																																								Type: "string",
																																							},
																																						},
																																						Type: "array",
																																					},
																																				},
																																				Required: []string{"key", "operator"},
																																				Type:     "object",
																																			},
																																		},
																																		Type: "array",
																																	},
																																	"matchLabels": apiext.JSONSchemaProps{
																																		AdditionalProperties: &apiext.JSONSchemaPropsOrBool{
																																			Schema: &apiext.JSONSchemaProps{
																																				Type: "string",
																																			},
																																		},
																																		Description: "matchLabels is a map of {key, value} pairs. A single {key, value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is 'key', the operator is 'In', and the values array contains only 'value'. The requirements are ANDed.",
																																		Type:        "object",
																																	},
																																},
																																Type: "object",
																															},
																															"namespaces": apiext.JSONSchemaProps{
																																Description: "namespaces specifies which namespaces the labelSelector applies to (matches against); null or empty list means this pod's namespace",
																																Items: &apiext.JSONSchemaPropsOrArray{
																																	Schema: &apiext.JSONSchemaProps{
																																		Type: "string",
																																	},
																																},
																																Type: "array",
																															},
																															"topologyKey": apiext.JSONSchemaProps{
																																Description: "This pod should be co-located (affinity) or not co-located (anti-affinity) with the pods matching the labelSelector in the specified namespaces, where co-located is defined as running on a node whose value of the label with key topologyKey matches taht of any node on which any of the selected pods is running. Empty topologyKey is not allowed.",
																																Type:        "string",
																															},
																														},
																														Required: []string{"topologyKey"},
																														Type:     "object",
																													},
																													"weight": apiext.JSONSchemaProps{
																														Description: "Weight associated with matching the corresponding nodeSelectorTerm, in the range 1-100.",
																														Format:      "int32",
																														Type:        "integer",
																													},
																												},
																												Required: []string{"podAffinityTerm", "weight"},
																												Type:     "object",
																											},
																										},
																										Type: "array",
																									},
																									"requiredDuringSchedulingIgnoredDuringExecution": apiext.JSONSchemaProps{
																										Description: "If the affinity requirements specified by this field are not met at scheduling time, the pod will not be scheduled onto the node. If the affinity requirements specified by this field cease to be met at some point during pod execution (e.g. due to an update), the system may or may not try to eventually evict the pod from its node.",
																										Items: &apiext.JSONSchemaPropsOrArray{
																											Schema: &apiext.JSONSchemaProps{
																												Description: "An empty preferred scheduling term matches all objects with implicit weight 0 (i.e. it's a no-op). A null preferred scheduling term matches no objects (i.e. is also a no-op).",
																												Properties: map[string]apiext.JSONSchemaProps{
																													"podAffinityTerm": apiext.JSONSchemaProps{
																														Description: "Required. A pod affinity term, associated with the corresponding weight..",
																														Properties: map[string]apiext.JSONSchemaProps{
																															"labelSelector": apiext.JSONSchemaProps{
																																Description: "A label query over a set of resources in this case pods.",
																																Properties: map[string]apiext.JSONSchemaProps{
																																	"matchExpressions": apiext.JSONSchemaProps{
																																		Description: "matchExpressions is a list of label selector requirements. The requirements are ANDed",
																																		Items: &apiext.JSONSchemaPropsOrArray{
																																			Schema: &apiext.JSONSchemaProps{
																																				Description: "A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.",
																																				Properties: map[string]apiext.JSONSchemaProps{
																																					"key": apiext.JSONSchemaProps{
																																						Description: "The label key that the selector applies to.",
																																						Type:        "string",
																																					},
																																					"operator": apiext.JSONSchemaProps{
																																						Description: "Represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists, DoesNotExist. Gt, and Lt.",
																																						Type:        "string",
																																					},
																																					"values": apiext.JSONSchemaProps{
																																						Description: "An array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. If the operator is Gt or Lt, the values array must have a single element, which will be intepreted as an integer. This array is replaced during a strategic merge patch.",
																																						Items: &apiext.JSONSchemaPropsOrArray{
																																							Schema: &apiext.JSONSchemaProps{
																																								Type: "string",
																																							},
																																						},
																																						Type: "array",
																																					},
																																				},
																																				Required: []string{"key", "operator"},
																																				Type:     "object",
																																			},
																																		},
																																		Type: "array",
																																	},
																																	"matchLabels": apiext.JSONSchemaProps{
																																		AdditionalProperties: &apiext.JSONSchemaPropsOrBool{
																																			Schema: &apiext.JSONSchemaProps{
																																				Type: "string",
																																			},
																																		},
																																		Description: "matchLabels is a map of {key, value} pairs. A single {key, value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is 'key', the operator is 'In', and the values array contains only 'value'. The requirements are ANDed.",
																																		Type:        "object",
																																	},
																																},
																																Type: "object",
																															},
																															"namespaces": apiext.JSONSchemaProps{
																																Description: "namespaces specifies which namespaces the labelSelector applies to (matches against); null or empty list means this pod's namespace",
																																Items: &apiext.JSONSchemaPropsOrArray{
																																	Schema: &apiext.JSONSchemaProps{
																																		Type: "string",
																																	},
																																},
																																Type: "array",
																															},
																															"topologyKey": apiext.JSONSchemaProps{
																																Description: "This pod should be co-located (affinity) or not co-located (anti-affinity) with the pods matching the labelSelector in the specified namespaces, where co-located is defined as running on a node whose value of the label with key topologyKey matches taht of any node on which any of the selected pods is running. Empty topologyKey is not allowed.",
																																Type:        "string",
																															},
																														},
																														Required: []string{"topologyKey"},
																														Type:     "object",
																													},
																													"weight": apiext.JSONSchemaProps{
																														Description: "Weight associated with matching the corresponding nodeSelectorTerm, in the range 1-100.",
																														Format:      "int32",
																														Type:        "integer",
																													},
																												},
																												Required: []string{"podAffinityTerm", "weight"},
																												Type:     "object",
																											},
																										},
																										Type: "array",
																									},
																								},
																								Type: "object",
																							},
																							"podAntiAffinity": apiext.JSONSchemaProps{
																								Description: "Describes pod anti-affinity scheduling rules (e.g. co-locate this pod in the same node, zone, etc. as some other pod(s)).",
																								Properties: map[string]apiext.JSONSchemaProps{
																									"preferredDuringSchedulingIgnoredDuringExecution": apiext.JSONSchemaProps{
																										Description: "The scheduler will prefer to schedule pods to nodes that satisfy the affinity expressions specified by this field, but it may choose a node that violates one or more of the expressions. The node that is most preferred is the one with the greatest sum of weights, i.e. for each node that meets all of the scheduling requirements (resource request, requiredDuringScheduling affinity expressions, etc.), compute a sum by iterating through the elements of this field and adding 'weight' to the sum if the node matches the corresponding matchExpressions; the node(s) with the highest sum are the most preferred.",
																										Items: &apiext.JSONSchemaPropsOrArray{
																											Schema: &apiext.JSONSchemaProps{
																												Description: "An empty preferred scheduling term matches all objects with implicit weight 0 (i.e. it's a no-op). A null preferred scheduling term matches no objects (i.e. is also a no-op).",
																												Properties: map[string]apiext.JSONSchemaProps{
																													"podAffinityTerm": apiext.JSONSchemaProps{
																														Description: "Required. A pod affinity term, associated with the corresponding weight..",
																														Properties: map[string]apiext.JSONSchemaProps{
																															"labelSelector": apiext.JSONSchemaProps{
																																Description: "A label query over a set of resources in this case pods.",
																																Properties: map[string]apiext.JSONSchemaProps{
																																	"matchExpressions": apiext.JSONSchemaProps{
																																		Description: "matchExpressions is a list of label selector requirements. The requirements are ANDed",
																																		Items: &apiext.JSONSchemaPropsOrArray{
																																			Schema: &apiext.JSONSchemaProps{
																																				Description: "A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.",
																																				Properties: map[string]apiext.JSONSchemaProps{
																																					"key": apiext.JSONSchemaProps{
																																						Description: "The label key that the selector applies to.",
																																						Type:        "string",
																																					},
																																					"operator": apiext.JSONSchemaProps{
																																						Description: "Represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists, DoesNotExist. Gt, and Lt.",
																																						Type:        "string",
																																					},
																																					"values": apiext.JSONSchemaProps{
																																						Description: "An array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. If the operator is Gt or Lt, the values array must have a single element, which will be intepreted as an integer. This array is replaced during a strategic merge patch.",
																																						Items: &apiext.JSONSchemaPropsOrArray{
																																							Schema: &apiext.JSONSchemaProps{
																																								Type: "string",
																																							},
																																						},
																																						Type: "array",
																																					},
																																				},
																																				Required: []string{"key", "operator"},
																																				Type:     "object",
																																			},
																																		},
																																		Type: "array",
																																	},
																																	"matchLabels": apiext.JSONSchemaProps{
																																		AdditionalProperties: &apiext.JSONSchemaPropsOrBool{
																																			Schema: &apiext.JSONSchemaProps{
																																				Type: "string",
																																			},
																																		},
																																		Description: "matchLabels is a map of {key, value} pairs. A single {key, value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is 'key', the operator is 'In', and the values array contains only 'value'. The requirements are ANDed.",
																																		Type:        "object",
																																	},
																																},
																																Type: "object",
																															},
																															"namespaces": apiext.JSONSchemaProps{
																																Description: "namespaces specifies which namespaces the labelSelector applies to (matches against); null or empty list means this pod's namespace",
																																Items: &apiext.JSONSchemaPropsOrArray{
																																	Schema: &apiext.JSONSchemaProps{
																																		Type: "string",
																																	},
																																},
																																Type: "array",
																															},
																															"topologyKey": apiext.JSONSchemaProps{
																																Description: "This pod should be co-located (affinity) or not co-located (anti-affinity) with the pods matching the labelSelector in the specified namespaces, where co-located is defined as running on a node whose value of the label with key topologyKey matches taht of any node on which any of the selected pods is running. Empty topologyKey is not allowed.",
																																Type:        "string",
																															},
																														},
																														Required: []string{"topologyKey"},
																														Type:     "object",
																													},
																													"weight": apiext.JSONSchemaProps{
																														Description: "Weight associated with matching the corresponding nodeSelectorTerm, in the range 1-100.",
																														Format:      "int32",
																														Type:        "integer",
																													},
																												},
																												Required: []string{"podAffinityTerm", "weight"},
																												Type:     "object",
																											},
																										},
																										Type: "array",
																									},
																									"requiredDuringSchedulingIgnoredDuringExecution": apiext.JSONSchemaProps{
																										Description: "If the affinity requirements specified by this field are not met at scheduling time, the pod will not be scheduled onto the node. If the affinity requirements specified by this field cease to be met at some point during pod execution (e.g. due to an update), the system may or may not try to eventually evict the pod from its node.",
																										Items: &apiext.JSONSchemaPropsOrArray{
																											Schema: &apiext.JSONSchemaProps{
																												Description: "An empty preferred scheduling term matches all objects with implicit weight 0 (i.e. it's a no-op). A null preferred scheduling term matches no objects (i.e. is also a no-op).",
																												Properties: map[string]apiext.JSONSchemaProps{
																													"podAffinityTerm": apiext.JSONSchemaProps{
																														Description: "Required. A pod affinity term, associated with the corresponding weight..",
																														Properties: map[string]apiext.JSONSchemaProps{
																															"labelSelector": apiext.JSONSchemaProps{
																																Description: "A label query over a set of resources in this case pods.",
																																Properties: map[string]apiext.JSONSchemaProps{
																																	"matchExpressions": apiext.JSONSchemaProps{
																																		Description: "matchExpressions is a list of label selector requirements. The requirements are ANDed",
																																		Items: &apiext.JSONSchemaPropsOrArray{
																																			Schema: &apiext.JSONSchemaProps{
																																				Description: "A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.",
																																				Properties: map[string]apiext.JSONSchemaProps{
																																					"key": apiext.JSONSchemaProps{
																																						Description: "The label key that the selector applies to.",
																																						Type:        "string",
																																					},
																																					"operator": apiext.JSONSchemaProps{
																																						Description: "Represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists, DoesNotExist. Gt, and Lt.",
																																						Type:        "string",
																																					},
																																					"values": apiext.JSONSchemaProps{
																																						Description: "An array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. If the operator is Gt or Lt, the values array must have a single element, which will be intepreted as an integer. This array is replaced during a strategic merge patch.",
																																						Items: &apiext.JSONSchemaPropsOrArray{
																																							Schema: &apiext.JSONSchemaProps{
																																								Type: "string",
																																							},
																																						},
																																						Type: "array",
																																					},
																																				},
																																				Required: []string{"key", "operator"},
																																				Type:     "object",
																																			},
																																		},
																																		Type: "array",
																																	},
																																	"matchLabels": apiext.JSONSchemaProps{
																																		AdditionalProperties: &apiext.JSONSchemaPropsOrBool{
																																			Schema: &apiext.JSONSchemaProps{
																																				Type: "string",
																																			},
																																		},
																																		Description: "matchLabels is a map of {key, value} pairs. A single {key, value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is 'key', the operator is 'In', and the values array contains only 'value'. The requirements are ANDed.",
																																		Type:        "object",
																																	},
																																},
																																Type: "object",
																															},
																															"namespaces": apiext.JSONSchemaProps{
																																Description: "namespaces specifies which namespaces the labelSelector applies to (matches against); null or empty list means this pod's namespace",
																																Items: &apiext.JSONSchemaPropsOrArray{
																																	Schema: &apiext.JSONSchemaProps{
																																		Type: "string",
																																	},
																																},
																																Type: "array",
																															},
																															"topologyKey": apiext.JSONSchemaProps{
																																Description: "This pod should be co-located (affinity) or not co-located (anti-affinity) with the pods matching the labelSelector in the specified namespaces, where co-located is defined as running on a node whose value of the label with key topologyKey matches taht of any node on which any of the selected pods is running. Empty topologyKey is not allowed.",
																																Type:        "string",
																															},
																														},
																														Required: []string{"topologyKey"},
																														Type:     "object",
																													},
																													"weight": apiext.JSONSchemaProps{
																														Description: "Weight associated with matching the corresponding nodeSelectorTerm, in the range 1-100.",
																														Format:      "int32",
																														Type:        "integer",
																													},
																												},
																												Required: []string{"podAffinityTerm", "weight"},
																												Type:     "object",
																											},
																										},
																										Type: "array",
																									},
																								},
																								Type: "object",
																							},
																						},
																						Type: "object",
																					},
																					"nodeSelector": apiext.JSONSchemaProps{
																						AdditionalProperties: &apiext.JSONSchemaPropsOrBool{
																							Schema: &apiext.JSONSchemaProps{
																								Type: "string",
																							},
																						},
																						Description: "NodeSelector is a selector which must be true for the pod to fit on a node. Selector which must match a node''s labels for the pod to be scheduled on that node. More info: https://kubernetes.io/docs/concepts/configuration/assign-pod-node/",
																						Type:        "object",
																					},
																					"tolerations": apiext.JSONSchemaProps{
																						Description: "If specified, the pod's tolerations.",
																						Items: &apiext.JSONSchemaPropsOrArray{
																							Schema: &apiext.JSONSchemaProps{
																								Description: "The pod this Toleration is attached to tolerates any taint that matches the triple <key,value,effect> using the matching operator <operator>.",
																								Type:        "object",
																								Properties: map[string]apiext.JSONSchemaProps{
																									"effect": apiext.JSONSchemaProps{
																										Description: "Effect indicates the taint effect to match. Empty means match all taint effects. When specified, allowed values are NoSchedule, PreferNoSchedule and Noexecute.",
																										Type:        "string",
																									},
																									"key": apiext.JSONSchemaProps{
																										Description: "Key is the taint key that the toleration applies to. Empty means match all taint keys. If the key is empty, operator must be Exists; this combination means to match all values and all keys.",
																										Type:        "string",
																									},
																									"operator": apiext.JSONSchemaProps{
																										Description: "Operator represents a key's relationship to the value. Valid operators are Exists and Equal. Defaults to Equal. Exists is equivalent to wildcard for value, so that a pod can tolerate all taints of a particular category.",
																										Type:        "string",
																									},
																									"tolerationSeconds": apiext.JSONSchemaProps{
																										Description: "TolerationSeconds represents the period of time the toleration (which must be of effect NoExecute, otherwise this field is ignored) tolerates the taint. By default, it is not set, which means tolerate the taint forever (do not evict). Zero and negative values will be treated as 0 (evict immediately) by the system.",
																										Type:        "integer",
																										Format:      "int64",
																									},
																									"value": apiext.JSONSchemaProps{
																										Description: "Value is the taint value the toleration matches to. If the operator is Exists, the value should be empty, otherwise just a regular string.",
																										Type:        "string",
																									},
																								},
																							},
																						},
																						Type: "array",
																					},
																				},
																				Type: "object",
																			},
																		},
																		Type: "object",
																	},
																	"serviceType": apiext.JSONSchemaProps{
																		Description: "Optional service type for Kubernetes solver service",
																		Type:        "string",
																	},
																},
																Type: "object",
															},
														},
														Type: "object",
													},
													"selector": apiext.JSONSchemaProps{
														Description: "Selector selects a set of DNSNames on the Certificate resource that should be solved using this challenge solver.",
														Properties: map[string]apiext.JSONSchemaProps{
															"dnsNames": apiext.JSONSchemaProps{
																Description: "List of DNSNames that this solver will be used to solve. If specified and a match is found, dnsNames selector will take precedence over a dnsZones selector. If multiple solvers match with the same dnsNames value, the solver with the most matching labels in matchLabels will be selected. If neither has more matches, the solver defined earlier in the list will be selected.",
																Items: &apiext.JSONSchemaPropsOrArray{
																	Schema: &apiext.JSONSchemaProps{
																		Type: "string",
																	},
																},
																Type: "array",
															},
															"dnsZones": apiext.JSONSchemaProps{
																Description: "List of DNSZones that this solver will be used to solve. The most specific DNS zone match specified here will take precedence over other DNS zone matches, so a solver specifying sys.example.com will be selected over one specifying example.com for the domain www.sys.example.com. If multiple solvers match with the same dnsZones value, the solver with the most matching labels in matchLabels will be selected. If neither has more matches, the solver defined earlier in the list will be selected.",
																Items: &apiext.JSONSchemaPropsOrArray{
																	Schema: &apiext.JSONSchemaProps{
																		Type: "string",
																	},
																},
																Type: "array",
															},
															"matchLabels": apiext.JSONSchemaProps{
																AdditionalProperties: &apiext.JSONSchemaPropsOrBool{
																	Schema: &apiext.JSONSchemaProps{
																		Type: "string",
																	},
																},
																Description: "A label selector that is used to refine the set of certificate's that this challenge solver will apply to.",
																Type:        "object",
															},
														},
														Type: "object",
													},
												},
												Type: "object",
											},
										},
										Type: "array",
									},
								},
								Required: []string{"privateKeySecretRef", "server"},
								Type:     "object",
							},
							"ca": apiext.JSONSchemaProps{
								Properties: map[string]apiext.JSONSchemaProps{
									"secretName": apiext.JSONSchemaProps{
										Description: "SecretName is the name of the secret used to sign Certificates issued by this Issuer.",
										Type:        "string",
									},
								},
								Required: []string{"secretName"},
								Type:     "object",
							},
							"selfSigned": apiext.JSONSchemaProps{
								Type: "object",
							},
							"vault": apiext.JSONSchemaProps{
								Properties: map[string]apiext.JSONSchemaProps{
									"auth": apiext.JSONSchemaProps{
										Description: "Vault authentication",
										Properties: map[string]apiext.JSONSchemaProps{
											"appRole": apiext.JSONSchemaProps{
												Description: "This Secret contains an AppRole and Secret",
												Properties: map[string]apiext.JSONSchemaProps{
													"path": apiext.JSONSchemaProps{
														Description: "Where the authentication path is mounted in Vault.",
														Type:        "string",
													},
													"roleId": apiext.JSONSchemaProps{
														Type: "string",
													},
													"secretRef": apiext.JSONSchemaProps{
														Properties: map[string]apiext.JSONSchemaProps{
															"key": apiext.JSONSchemaProps{
																Description: "The key of the secret to select from. Must be a valid secret key.",
																Type:        "string",
															},
															"name": apiext.JSONSchemaProps{
																Description: "Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?",
																Type:        "string",
															},
														},
														Required: []string{"name"},
														Type:     "object",
													},
												},
												Required: []string{"path", "roleId", "secretRef"},
												Type:     "object",
											},
											"tokenSecretRef": apiext.JSONSchemaProps{
												Description: "This Secret contains the Vault token key",
												Properties: map[string]apiext.JSONSchemaProps{
													"key": apiext.JSONSchemaProps{
														Description: "The key of the secret to select from. Must be a valid secret key.",
														Type:        "string",
													},
													"name": apiext.JSONSchemaProps{
														Description: "Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?",
														Type:        "string",
													},
												},
												Required: []string{"name"},
												Type:     "object",
											},
										},
										Type: "object",
									},
									"caBundle": apiext.JSONSchemaProps{
										Description: "Base64 encoded CA bundle to validate Vault server certificate. Only used if the Server URL is using HTTPS protocol. This parameter is ignored for plain HTTP protocol connection. If not set the system root certificates are used to validate the TLS connection.",
										Format:      "byte",
										Type:        "string",
									},
									"path": apiext.JSONSchemaProps{
										Description: "Vault URL path to the certificate role",
										Type:        "string",
									},
									"server": apiext.JSONSchemaProps{
										Description: "Server is the vault connection address",
										Type:        "string",
									},
								},
								Required: []string{"auth", "path", "server"},
								Type:     "object",
							},
							"venafi": apiext.JSONSchemaProps{
								Description: "VenaifIssuer describes issuer configuration details for Venaif Cloud.",
								Properties: map[string]apiext.JSONSchemaProps{
									"cloud": apiext.JSONSchemaProps{
										Description: "Cloud specifies the Venafi cloud configuration settings. Only one of TPP or Cloud may be specified.",
										Properties: map[string]apiext.JSONSchemaProps{
											"apiTokenSecretRef": apiext.JSONSchemaProps{
												Description: "APITokenSecretRef is a secret key selector for the Venafi Cloud API token.",
												Properties: map[string]apiext.JSONSchemaProps{
													"key": apiext.JSONSchemaProps{
														Description: "The key of the secret to select from. Must be a valid secret key.",
														Type:        "string",
													},
													"name": apiext.JSONSchemaProps{
														Description: "Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?",
														Type:        "string",
													},
												},
												Required: []string{"name"},
												Type:     "object",
											},
											"url": apiext.JSONSchemaProps{
												Description: "URL is the base URL for Venafi Cloud",
												Type:        "string",
											},
										},
										Required: []string{"apiTokenSecretRef", "url"},
										Type:     "object",
									},
									"tpp": apiext.JSONSchemaProps{
										Description: "TPP specifies Trust Protection Platform configuration settings. Only one of TPP or Cloud may be specified.",
										Properties: map[string]apiext.JSONSchemaProps{
											"caBundle": apiext.JSONSchemaProps{
												Description: "CABundle is a PEM encoded TLS certifiate to use to verify connections to the TPP instance. If specified, system roots will not be used and the issuing CA for the TPP instance must be verifiable using the provided root. If not specified, the connection will be verified using the cert-manager system root certificates.",
												Format:      "byte",
												Type:        "string",
											},
											"credentialsRef": apiext.JSONSchemaProps{
												Description: "CredentialsRef is a reference to a Secret containing the username and password for the TPP server. The secret must contain two keys, 'username' and 'password'.",
												Properties: map[string]apiext.JSONSchemaProps{
													"name": apiext.JSONSchemaProps{
														Description: "Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?",
														Type:        "string",
													},
												},
												Required: []string{"name"},
												Type:     "object",
											},
											"url": apiext.JSONSchemaProps{
												Description: "URL is the base URL for the Venafi TPP instance",
												Type:        "string",
											},
										},
										Required: []string{"credentialsRef", "url"},
										Type:     "object",
									},
									"zone": apiext.JSONSchemaProps{
										Description: "Zone is the Venafi Policy Zone to use for this issuer. All requests made to the Venafi platform will be restricted by the named zone policy. This field is required.",
										Type:        "string",
									},
								},
								Required: []string{"zone"},
								Type:     "object",
							},
						},
					},
					"status": apiext.JSONSchemaProps{
						Description: "IssuerStatus contains tatus information about an Issuer",
						Properties: map[string]apiext.JSONSchemaProps{
							"acme": apiext.JSONSchemaProps{
								Properties: map[string]apiext.JSONSchemaProps{
									"lastRegisteredEmail": apiext.JSONSchemaProps{
										Description: "LastRegisteredEmail is the email associated with the latest registered ACME account, in order to track changes made to registered account associated with the Issuer",
										Type:        "string",
									},
									"uri": apiext.JSONSchemaProps{
										Description: "URI is the unique account identifier, which can also be used to retrieve account details from the CA",
										Type:        "string",
									},
								},
								Type: "object",
							},
							"conditions": apiext.JSONSchemaProps{
								Items: &apiext.JSONSchemaPropsOrArray{
									Schema: &apiext.JSONSchemaProps{
										Description: "IssuerCondition contains condition information for an Issuer.",
										Properties: map[string]apiext.JSONSchemaProps{
											"lastTransitionTime": apiext.JSONSchemaProps{
												Description: "LastTransitionTime is the timestamp corresponding to the last status change of this condition.",
												Format:      "date-time",
												Type:        "string",
											},
											"message": apiext.JSONSchemaProps{
												Description: "Message is a human readable description of the details of the last transition, complementing reason.",
												Type:        "string",
											},
											"reason": apiext.JSONSchemaProps{
												Description: "Reason is a brief machine readable explanation for the condition's last transition.",
												Type:        "string",
											},
											"status": apiext.JSONSchemaProps{
												Description: "Status of the condition, one of ('True', 'False', 'Unknown').",
												Enum: []apiext.JSON{
													{Raw: []byte("\"True\"")},
													{Raw: []byte("\"False\"")},
													{Raw: []byte("\"Unknown\"")},
												},
												Type: "string",
											},
											"type": apiext.JSONSchemaProps{
												Description: "Type of the condition, currently ('Ready').",
												Type:        "string",
											},
										},
										Required: []string{"status", "type"},
										Type:     "object",
									},
								},
								Type: "array",
							},
						},
						Type: "object",
					},
				},
				Type: "object",
			},
		},
	},
	Status: apiext.CustomResourceDefinitionStatus{
		Conditions: []apiext.CustomResourceDefinitionCondition{},
		AcceptedNames: apiext.CustomResourceDefinitionNames{
			Kind:   "",
			Plural: "",
		},
		StoredVersions: []string{},
	},
}

var clusterIssuerCRD = &apiext.CustomResourceDefinition{
	ObjectMeta: metav1.ObjectMeta{Name: "clusterissuers.certmanager.k8s.io", Labels: ControllerLabelMap},
	Spec: apiext.CustomResourceDefinitionSpec{
		Group:   GroupVersion,
		Version: CRDVersion,
		Scope:   apiext.ClusterScoped,
		Names: apiext.CustomResourceDefinitionNames{
			Plural: "clusterissuers",
			Kind:   "ClusterIssuer",
		},
		AdditionalPrinterColumns: []apiext.CustomResourceColumnDefinition{
			{
				JSONPath: `.status.conditions[?(@.type=="Ready")].status`,
				Name:     "Ready",
				Type:     "string",
			},
		},
	},
	Status: apiext.CustomResourceDefinitionStatus{
		Conditions: []apiext.CustomResourceDefinitionCondition{},
		AcceptedNames: apiext.CustomResourceDefinitionNames{
			Kind:   "",
			Plural: "",
		},
		StoredVersions: []string{},
	},
}

var orderCRD = &apiext.CustomResourceDefinition{
	ObjectMeta: metav1.ObjectMeta{Name: "orders.certmanager.k8s.io", Labels: ControllerLabelMap},
	Spec: apiext.CustomResourceDefinitionSpec{
		Group:   GroupVersion,
		Version: CRDVersion,
		Scope:   apiext.NamespaceScoped,
		Names: apiext.CustomResourceDefinitionNames{
			Plural: "orders",
			Kind:   "Order",
		},
	},
	Status: apiext.CustomResourceDefinitionStatus{
		Conditions: []apiext.CustomResourceDefinitionCondition{},
		AcceptedNames: apiext.CustomResourceDefinitionNames{
			Kind:   "",
			Plural: "",
		},
		StoredVersions: []string{},
	},
}

var challengeCRD = &apiext.CustomResourceDefinition{
	ObjectMeta: metav1.ObjectMeta{Name: "challenges.certmanager.k8s.io", Labels: ControllerLabelMap},
	Spec: apiext.CustomResourceDefinitionSpec{
		Group:   GroupVersion,
		Version: CRDVersion,
		Scope:   apiext.NamespaceScoped,
		Names: apiext.CustomResourceDefinitionNames{
			Plural: "challenges",
			Kind:   "Challenge",
		},
		AdditionalPrinterColumns: []apiext.CustomResourceColumnDefinition{
			{
				Name:     "State",
				Type:     "string",
				JSONPath: ".status.state",
			},
			{
				Name:     "Domain",
				Type:     "string",
				JSONPath: ".spec.dnsName",
			},
			{
				Name:     "Reason",
				Type:     "string",
				JSONPath: ".status.reason",
			},
			{
				Name:        "Age",
				Type:        "date",
				JSONPath:    ".metadata.creationTimestamp",
				Description: "CreationTimestamp is a timestamp representing the server time when this object was created. It is not guaranteed to be set in happens-before order across separate operations. Clients may not set this value. It is represented in RFC3339 form and is in UTC.",
			},
		},
		Validation: &apiext.CustomResourceValidation{
			OpenAPIV3Schema: &apiext.JSONSchemaProps{
				Description: "Challenge is a type to represent a Challenge request with an ACME server",
				Properties: map[string]apiext.JSONSchemaProps{
					"apiVersion": apiext.JSONSchemaProps{
						Description: "APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources",
						Type:        "string",
					},
					"kind": apiext.JSONSchemaProps{
						Description: "Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds",
						Type:        "string",
					},
					"metadata": apiext.JSONSchemaProps{
						Type: "object",
					},
					"spec": apiext.JSONSchemaProps{
						Properties: map[string]apiext.JSONSchemaProps{
							"authzURL": apiext.JSONSchemaProps{
								Description: "AuthzURL is the URL to the ACME Authorization resource that this challenge is a part of.",
								Type:        "string",
							},
							"config": apiext.JSONSchemaProps{
								Description: "Config specifies the solver configuration for this challenge. Only **one** of 'config' or 'solver' may be specified, and if both are specified then no action will be performed on the Challenge resource. DEPRECATED: the 'solver' field should be specified instead",
								Type:        "object",
								Properties: map[string]apiext.JSONSchemaProps{
									"dns01": apiext.JSONSchemaProps{
										Description: "DNS01 contains DNS01 challenge solving configuration",
										Type:        "object",
										Properties: map[string]apiext.JSONSchemaProps{
											"provider": apiext.JSONSchemaProps{
												Description: "Provider is the name of the DNS01 challenge provider to use, as configure on the referenced Issuer or ClusterIssuer resource.",
												Type:        "string",
											},
										},
										Required: []string{
											"provider",
										},
									},
									"http01": apiext.JSONSchemaProps{
										Description: "HTTP01 contains HTTP01 challenge solving configuration",
										Type:        "object",
										Properties: map[string]apiext.JSONSchemaProps{
											"ingress": apiext.JSONSchemaProps{
												Description: "Ingress is the name of an Ingress resource that will be edited to include the ACME HTTP01 'well-known' challenge path in order to solve HTTP01 challenges. If this field is specified, 'ingressClass' **must not** be specified.",
												Type:        "string",
											},
											"ingressClass": apiext.JSONSchemaProps{
												Description: "IngressClass is the ingress class that should be set on new ingress resources that are created in order to solve HTTP01 challenges. This field should be used when using an ingress controller such as nginx, which 'flattens' ingress configuration instead of maintaining a 1:1 mapping between loadbalancer IP:ingress resources. If this field is not set, and 'ingress' is not set, then ingresses without an ingress class set will be created to solve HTTP01 challenges. If this field is specified, 'ingress' **must not** be specified.",
												Type:        "string",
											},
										},
									},
								},
							},
							"dnsName": apiext.JSONSchemaProps{
								Description: "DNSName is the identifier that this challenge is for, e.g. example.com.",
								Type:        "string",
							},
							"issuerRef": apiext.JSONSchemaProps{
								Description: "IssuerRef references a properly configured ACME-type Issuer which should be used to create this Challenge. If the Issuer does not exist, processing will be retried. If the Issuer is not an 'ACME' Issuer, an error will be returned and the Challenge will be marked as failed.",
								Type:        "object",
								Properties: map[string]apiext.JSONSchemaProps{
									"group": apiext.JSONSchemaProps{
										Type: "string",
									},
									"name": apiext.JSONSchemaProps{
										Type: "string",
									},
									"kind": apiext.JSONSchemaProps{
										Type: "string",
									},
								},
								Required: []string{
									"name",
								},
							},
							"key": apiext.JSONSchemaProps{
								Description: "Key is the ACME challenge key for this challenge",
								Type:        "string",
							},
							"solver": apiext.JSONSchemaProps{
								Description: "Solver contains the domain solving configuration that should be used to solve this challenge resource. Only **one** of 'config' or 'solver' may be specified, and if both are specified then no action will be performed on the Challenge resource.",
								Type:        "string",
								Properties: map[string]apiext.JSONSchemaProps{
									"dns01": apiext.JSONSchemaProps{
										Properties: map[string]apiext.JSONSchemaProps{
											"acmedns": apiext.JSONSchemaProps{
												Description: "ACMEIssuerDNS01ProviderAcmeDNS is a structure containing the configuration for ACME-DNS servers",
												Properties: map[string]apiext.JSONSchemaProps{
													"accountSecretRef": apiext.JSONSchemaProps{
														Properties: map[string]apiext.JSONSchemaProps{
															"key": apiext.JSONSchemaProps{
																Description: "The key of the secret to select from. Must be a valid secret key.",
																Type:        "string",
															},
															"name": apiext.JSONSchemaProps{
																Description: "Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?",
																Type:        "string",
															},
														},
														Required: []string{
															"name",
														},
														Type: "object",
													},
													"host": apiext.JSONSchemaProps{
														Type: "string",
													},
												},
												Required: []string{
													"accountSecretRef",
													"host",
												},
												Type: "object",
											},
											"akamai": apiext.JSONSchemaProps{
												Description: "ACMEIssuerDNS01ProviderAkamai is a structure containing the DNS configuration for Akamai DNS—Zone Record Management API",
												Properties: map[string]apiext.JSONSchemaProps{
													"accessTokenSecretRef": apiext.JSONSchemaProps{
														Properties: map[string]apiext.JSONSchemaProps{
															"key": apiext.JSONSchemaProps{
																Description: "The key of the secret to select from. Must be a valid secret key.",
																Type:        "string",
															},
															"name": apiext.JSONSchemaProps{
																Description: "Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?",
																Type:        "string",
															},
														},
														Required: []string{"name"},
														Type:     "object",
													},
													"clientSecretSecretRef": apiext.JSONSchemaProps{
														Properties: map[string]apiext.JSONSchemaProps{
															"key": apiext.JSONSchemaProps{
																Description: "The key of the secret to select from. Must be a valid secret key.",
																Type:        "string",
															},
															"name": apiext.JSONSchemaProps{
																Description: "Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?",
																Type:        "string",
															},
														},
														Required: []string{"name"},
														Type:     "object",
													},
													"clientTokenSecretRef": apiext.JSONSchemaProps{
														Properties: map[string]apiext.JSONSchemaProps{
															"key": apiext.JSONSchemaProps{
																Description: "The key of the secret to select from. Must be a valid secret key.",
																Type:        "string",
															},
															"name": apiext.JSONSchemaProps{
																Description: "Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?",
																Type:        "string",
															},
														},
														Required: []string{"name"},
														Type:     "object",
													},
													"serviceConsumerDomain": apiext.JSONSchemaProps{
														Type: "string",
													},
												},
												Required: []string{"accessTokenSecretRef", "clientSecretSecretRef", "clientTokenSecretRef", "serviceConsumerDomain"},
												Type:     "object",
											},
											"azuredns": apiext.JSONSchemaProps{
												Description: "",
												Properties: map[string]apiext.JSONSchemaProps{
													"clientID": apiext.JSONSchemaProps{
														Type: "string",
													},
													"clientSecretSecretRef": apiext.JSONSchemaProps{
														Properties: map[string]apiext.JSONSchemaProps{
															"key": apiext.JSONSchemaProps{
																Description: "The key of the secret to select from. Must be a valid secret key.",
																Type:        "string",
															},
															"name": apiext.JSONSchemaProps{
																Description: "Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?",
																Type:        "string",
															},
														},
														Required: []string{"name"},
														Type:     "object",
													},
													"environment": apiext.JSONSchemaProps{
														Type: "string",
														Enum: []apiext.JSON{
															{Raw: []byte("\"AzurePublicCloud\"")},
															{Raw: []byte("\"AzureChinaCloud\"")},
															{Raw: []byte("\"AzureGermanCloud\"")},
															{Raw: []byte("\"AzureUSGovernmentCloud\"")},
														},
													},
													"hostedZoneName": apiext.JSONSchemaProps{
														Type: "string",
													},
													"resourceGroupNAme": apiext.JSONSchemaProps{
														Type: "string",
													},
													"subscriptionID": apiext.JSONSchemaProps{
														Type: "string",
													},
													"tenantID": apiext.JSONSchemaProps{
														Type: "string",
													},
												},
												Required: []string{"clientID", "clientSecretSecretRef", "resourceGroupName", "subscriptionID", "tenantID"},
												Type:     "object",
											},
											"clouddns": apiext.JSONSchemaProps{
												Description: "ACMEIssuerDNS01ProviderCloudDNS is a structure containing the DNS configuration for Google Cloud DNS",
												Properties: map[string]apiext.JSONSchemaProps{
													"project": apiext.JSONSchemaProps{
														Type: "string",
													},
													"serviceAccountSecretRef": apiext.JSONSchemaProps{
														Properties: map[string]apiext.JSONSchemaProps{
															"key": apiext.JSONSchemaProps{
																Description: "The key of the secret to select from. Must be a valid secret key.",
																Type:        "string",
															},
															"name": apiext.JSONSchemaProps{
																Description: "Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?",
																Type:        "string",
															},
														},
														Required: []string{"name"},
														Type:     "object",
													},
												},
												Required: []string{"project", "serviceAccountSecretRef"},
												Type:     "object",
											},
											"cloudflare": apiext.JSONSchemaProps{
												Description: "ACMEIssuerDNS01ProviderCloudflare is a structure containing the DNS configuration for Cloudflare",
												Properties: map[string]apiext.JSONSchemaProps{
													"apiKeySecretRef": apiext.JSONSchemaProps{
														Properties: map[string]apiext.JSONSchemaProps{
															"key": apiext.JSONSchemaProps{
																Description: "The key of the secret to select from. Must be a valid secret key.",
																Type:        "string",
															},
															"name": apiext.JSONSchemaProps{
																Description: "Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?",
																Type:        "string",
															},
														},
														Required: []string{
															"name",
														},
														Type: "object",
													},
													"email": apiext.JSONSchemaProps{
														Type: "string",
													},
												},
												Required: []string{"apiKeySecretRef", "email"},
												Type:     "object",
											},
										},
									},
								},
							},
							"token": apiext.JSONSchemaProps{
								Description: "AuthzURL is the URL to the ACME Authorization resource that this challenge is a part of.",
								Type:        "string",
							},
							"type": apiext.JSONSchemaProps{
								Description: "AuthzURL is the URL to the ACME Authorization resource that this challenge is a part of.",
								Type:        "string",
							},
							"url": apiext.JSONSchemaProps{
								Description: "AuthzURL is the URL to the ACME Authorization resource that this challenge is a part of.",
								Type:        "string",
							},
							"wildcard": apiext.JSONSchemaProps{
								Description: "Wildcard will be true if this challenge is for a wildcard identifier, for example '*.example.com'",
								Type:        "boolean",
							},
						},
					},
				},
			},
		},
	},
	Status: apiext.CustomResourceDefinitionStatus{
		Conditions: []apiext.CustomResourceDefinitionCondition{},
		AcceptedNames: apiext.CustomResourceDefinitionNames{
			Kind:   "",
			Plural: "",
		},
		StoredVersions: []string{},
	},
}
