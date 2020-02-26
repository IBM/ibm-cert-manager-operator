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
				Description: "CreationTimestamp is a timestamp representing time when this object was created. It is not guaranteed to be set in happens-before order across separate operations. Clients may not set this value. It is represented in RFC3339 form and is in UTC. Populated by the system. Read-only. Null for lists.",
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
												Description: "Status of the condition one of ('True', 'False', 'Unknown')",
												Enum: []apiext.JSON{
													{Raw: []byte("\"True\"")},
													{Raw: []byte("\"False\"")},
													{Raw: []byte("\"Unknown\"")},
												},
												Type: "string",
											},
											"type": apiext.JSONSchemaProps{
												Description: "Type of the condition, currently ('Ready')",
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
															"azuredns":      apiext.JSONSchemaProps{},
															"clouddns":      apiext.JSONSchemaProps{},
															"cloudfalre":    apiext.JSONSchemaProps{},
															"cnameStrategy": apiext.JSONSchemaProps{},
															"digitalocean":  apiext.JSONSchemaProps{},
															"name":          apiext.JSONSchemaProps{},
															"rfc2136":       apiext.JSONSchemaProps{},
															"router53":      apiext.JSONSchemaProps{},
															"webhook":       apiext.JSONSchemaProps{},
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
									"email":               apiext.JSONSchemaProps{},
									"http01":              apiext.JSONSchemaProps{},
									"privateKeySecretRef": apiext.JSONSchemaProps{},
									"server":              apiext.JSONSchemaProps{},
									"skipTLSVerify":       apiext.JSONSchemaProps{},
									"solvers":             apiext.JSONSchemaProps{},
								},
								Required: []string{"privateKeySecretRef", "server"},
								Type:     "object",
							},
							"ca":         apiext.JSONSchemaProps{},
							"selfSigned": apiext.JSONSchemaProps{},
							"vault":      apiext.JSONSchemaProps{},
							"venafi":     apiext.JSONSchemaProps{},
						},
					},
					"status": apiext.JSONSchemaProps{
						Description: "",
						Properties: map[string]apiext.JSONSchemaProps{
							"acme":       apiext.JSONSchemaProps{},
							"conditions": apiext.JSONSchemaProps{},
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
