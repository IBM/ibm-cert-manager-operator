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
			Plural: "certificates",
			Kind:   "Certificate",
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
		Versions: []apiext.CustomResourceDefinitionVersion{
			{
				Name: CRDVersion,
				Served: TrueVar,
				Storage: TrueVar,
				Schema: &apiext.CustomResourceValidation{
					OpenAPIV3Schema: &apiext.JSONSchemaProps{
						Description: "Certificate is a type to represent a Certificate from ACME",
						Properties: map[string]apiext.JSONSchemaProps{
							"apiVersion": apiext.JSONSchemaProps{
								Description: "APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources",
								Type: "string",
							},
							"kind": apiext.JSONSchemaProps{
								Description: "Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds",
								Type: "string",
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
																		Type: "string",
																	},
																},
																Required: []string{"provider"},
																Type: "object",
															},
															"domains": apiext.JSONSchemaProps{
																Description: "Domains is the list of domains that this Solverconfig applies to.",
																Items:  &apiext.JSONSchemaPropsOrArray{
																	Type: "string",
																},
																Type: "array",
															},
															"http01": apiext.JSONSchemaProps{
																Description: "HTTP01 contains HTTP01 challenge solving configuration",
																Properties: map[string]apiext.JSONSchemaProps{
																	"ingress": apiext.JSONSchemaProps{
																		Description: "Ingress is the name of an Ingress resource that will be edited to include the ACME HTTP01 'well-known' challenge path in order to solve HTTP01 challenges. If this field is specified, 'ingressClass' **must not** be specified.",
																		Type: "string",
																	},
																	"ingressClass": apiext.JSONSchemaProps{
																		Description: "IngressClass is the ingress class that should be set on new ingress resources that are created in order to solve HTTP01 challenges. This field should be used when using an ingress controller such as nginx, which 'flattens' ingress configuration instead of maintaining a 1:1 mapping between loadbalancer IP:ingress resources. If this field is not set, and 'ingress' is not set, then ingresses without an ingress class set will be created to solve HTTP01 challenges. If this field is specified, 'ingress' **must not** be specified.",
																		Type: "string",
																	},
																},
																Type: "object",
															},
														},
														Required: []string{"domains"},
														Type: "object",
													},
												},
												Type: "array",
											},
										},
										Required: []string{"config"},
										Type: "object",
									},
									"commonName": apiext.JSONSchemaProps{
									},
									"dnsNames": apiext.JSONSchemaProps{
									},
									"duration": apiext.JSONSchemaProps{
									},
									"ipAddresses": apiext.JSONSchemaProps{
									},
									"isCA": apiext.JSONSchemaProps{
									},
									"issuerRef": apiext.JSONSchemaProps{
									},
									"keyAlgorithm": apiext.JSONSchemaProps{
									},
									"keyEncoding": apiext.JSONSchemaProps{
									},
									"keySize": apiext.JSONSchemaProps{
									},
									"organization": apiext.JSONSchemaProps{
									},
									"renewBefore": apiext.JSONSchemaProps{
									},
									"secretName": apiext.JSONSchemaProps{
									},
									"usages": apiext.JSONSchemaProps{
									},
								},

							},
						},
					},
				},
			},
		},
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
}

var challengeCRD = &apiext.CustomResourceDefinition{
	ObjectMeta: metav1.ObjectMeta{Name: "challenges.certmanager.k8s.io", Labels: ControllerLabelMap},
	Spec: apiext.CustomResourceDefinitionSpec{
		Group:   GroupVersion,
		Version: CRDVersion,
		Versions: []apiext.CustomResourceDefinitionVersion{
			{
				Name: CRDVersion,
				Served: TrueVar,
				Storage: TrueVar,
				Schema: &apiext.CustomResourceValidation{
					OpenAPIV3Schema: &apiext.JSONSchemaProps{
						Description: "Challenge is a type to represent a Challenge request with an ACME server",
						Properties: map[string]apiext.JSONSchemaProps{
							"apiVersion": apiext.JSONSchemaProps{
								Description: "APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources",
								Type: "string",
							},
							"kind": apiext.JSONSchemaProps{
								Description: "Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds",
								Type: "string",
							},
							"metadata": apiext.JSONSchemaProps{
								Type: "object",
							},
							"spec": apiext.JSONSchemaProps{
								Properties: map[string]apiext.JSONSchemaProps{
									"authzURL": apiext.JSONSchemaProps{
										Description: "AuthzURL is the URL to the ACME Authorization resource that this challenge is a part of.",
										Type: "string",
									},
									"config": apiext.JSONSchemaProps{
										Description: "Config specifies the solver configuration for this challenge. Only **one** of ''config'' or ''solver'' may be specified, and if both are specified then no action will be performed on the Challenge resource. DEPRECATED: the ''solver'' field should be specified instead",
										Type: "object",
										Properties: map[string]apiext.JSONSchemaProps{
											"dns01": apiext.JSONSchemaProps{
												Description: "DNS01 contains DNS01 challenge solving configuration",
												Type: "object",
												Properties: map[string]apiext.JSONSchemaProps{
													"provider":  apiext.JSONSchemaProps{
														Description: "Provider is the name of the DNS01 challenge provider to use, as configure on the referenced Issuer or ClusterIssuer resource.",
														Type: "string",
													},
												},
												Required: []string {
													"provider",
												},
											},
											"http01": apiext.JSONSchemaProps{
												Description: "HTTP01 contains HTTP01 challenge solving configuration",
												Type: "object",
												Properties: map[string]apiext.JSONSchemaProps{
													"ingress": apiext.JSONSchemaProps{
														Description: "Ingress is the name of an Ingress resource that will be edited to include the ACME HTTP01 'well-known' challenge path in order to solve HTTP01 challenges. If this field is specified, 'ingressClass' **must not** be specified.",
														Type: "string",
													},
													"ingressClass": apiext.JSONSchemaProps{
														Description: "IngressClass is the ingress class that should be set on new ingress resources that are created in order to solve HTTP01 challenges. This field should be used when using an ingress controller such as nginx, which 'flattens' ingress configuration instead of maintaining a 1:1 mapping between loadbalancer IP:ingress resources. If this field is not set, and 'ingress' is not set, then ingresses without an ingress class set will be created to solve HTTP01 challenges. If this field is specified, 'ingress' **must not** be specified.",
														Type: "string",
													},
												},
											},
										},
									},
									"dnsName": apiext.JSONSchemaProps{
										Description: "DNSName is the identifier that this challenge is for, e.g. example.com.",
										Type: "string",
									},
									"issuerRef": apiext.JSONSchemaProps{
										Description: "IssuerRef references a properly configured ACME-type Issuer which should be used to create this Challenge. If the Issuer does not exist, processing will be retried. If the Issuer is not an 'ACME' Issuer, an error will be returned and the Challenge will be marked as failed.",
										Type: "object",
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
										Required: []string{
											"name",
										},
									},
									"key": apiext.JSONSchemaProps{
										Description: "Key is the ACME challenge key for this challenge",
										Type: "string",
									},
									"solver": apiext.JSONSchemaProps{
										Description: "Solver contains the domain solving configuration that should be used to solve this challenge resource. Only **one** of 'config' or 'solver' may be specified, and if both are specified then no action will be performed on the Challenge resource.",
										Type: "string",
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
																		Type: "string",
																	},
																	"name": apiext.JSONSchemaProps{
																		Description: "Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?",
																		Type: "string",
																	},
																},
																Required: []string {
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
																		Type: "string",
																	},
																	"name": apiext.JSONSchemaProps{
																		Description: "Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?",
																		Type: "string",
																	},
																},
																Required: []string{"name"},
																Type: "object",
															},
															"clientSecretSecretRef": apiext.JSONSchemaProps{
																Properties: map[string]apiext.JSONSchemaProps{
																	"key": apiext.JSONSchemaProps{
																		Description: "The key of the secret to select from. Must be a valid secret key.",
																		Type: "string",
																	},
																	"name": apiext.JSONSchemaProps{
																		Description: "Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?",
																		Type: "string",
																	},
																},
																Required: []string{"name"},
																Type: "object",
															},
															"clientTokenSecretRef": apiext.JSONSchemaProps{
																Properties: map[string]apiext.JSONSchemaProps{
																	"key": apiext.JSONSchemaProps{
																		Description: "The key of the secret to select from. Must be a valid secret key.",
																		Type: "string",
																	},
																	"name": apiext.JSONSchemaProps{
																		Description: "Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?",
																		Type: "string",
																	},
																},
																Required: []string{"name"},
																Type: "object",
															},
															"serviceConsumerDomain": apiext.JSONSchemaProps{
																Type: "string",
															},
														},
														Required: []string{"accessTokenSecretRef", "clientSecretSecretRef", "clientTokenSecretRef", "serviceConsumerDomain"},
														Type: "object",
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
																		Type: "string",
																	},
																	"name": apiext.JSONSchemaProps{
																		Description: "Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?",
																		Type: "string",
																	},
																},
																Required: []string{"name"},
																Type: "object",
															},
															"environment": apiext.JSONSchemaProps{
																Type: "string",
																Enum: []JSON{
																	Raw: []byte{
																		"AzurePublicCloud",
																		"AzureChinaCloud",
																		"AzureGermanCloud",
																		"AzureUSGovernmentCloud",
																	},
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
														Type: "object",
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
																		Type: "string",
																	},
																	"name": apiext.JSONSchemaProps{
																		Description: "Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?",
																		Type: "string",
																	},
																},
																Required: []string{"name"},
																Type: "object",
															},
														},
														Required: []string{"project", "serviceAccountSecretRef"},
														Type: "object",
													},
													"cloudflare": apiext.JSONSchemaProps{
														Description: "ACMEIssuerDNS01ProviderCloudflare is a structure containing the DNS configuration for Cloudflare",
														Properties: map[string]apiext.JSONSchemaProps{
															"apiKeySecretRef": apiext.JSONSchemaProps{
																Properties: map[string]apiext.JSONSchemaProps{
																	"key": apiext.JSONSchemaProps{
																		Description: "The key of the secret to select from. Must be a valid secret key.",
																		Type: "string",
																	},
																	"name": apiext.JSONSchemaProps{
																		Description: "Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?",
																		Type: "string",
																	},
																},
																Required: []string {
																	"name",
																},
																Type: "object",
															},
															"email": apiext.JSONSchemaProps{
																Type: "string",
															},
														},
														Required: []string{"apiKeySecretRef", "email"},
														Type: "object",
													},
												},
											},
										},
									},
									"token": apiext.JSONSchemaProps{
										Description: "AuthzURL is the URL to the ACME Authorization resource that this challenge is a part of.",
										Type: "string",
									},
									"type": apiext.JSONSchemaProps{
										Description: "AuthzURL is the URL to the ACME Authorization resource that this challenge is a part of.",
										Type: "string",
									},
									"url": apiext.JSONSchemaProps{
										Description: "AuthzURL is the URL to the ACME Authorization resource that this challenge is a part of.",
										Type: "string",
									},
									"wildcard": apiext.JSONSchemaProps{
										Description: "Wildcard will be true if this challenge is for a wildcard identifier, for example '*.example.com'",
										Type: "boolean",
									},
								},
							},
						},
					},
				},
				AdditionalPrinterColumns: []apixtensionv1.CustomResourceColumnDefinition{
					{
						Name: "State",
						Type: "string",
						JSONPath: ".status.state",
					},
					{
						Name: "Domain",
						Type: "string",
						JSONPath: ".spec.dnsName",
					},
					{
						Name: "Reason",
						Type: "string",
						JSONPath: ".status.reason",
					},
					{
						Name: "Age",
						Type: "date",
						JSONPath: ".metadata.creationTimestamp",
						Description: "CreationTimestamp is a timestamp representing the server time when this object was created. It is not guaranteed to be set in happens-before order across separate operations. Clients may not set this value. It is represented in RFC3339 form and is in UTC.",
					},
				},
			},
		},
		Scope:   apiext.NamespacedScope,
		Names: apiext.CustomResourceDefinitionNames{
			Plural: "challenges",
			Kind:   "Challenge",
		},
	},
}
