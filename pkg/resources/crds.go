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
	apiextensionv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CRDMap a map from the crd name to the definition of that crd
var CRDMap = map[string]*apiextensionv1.CustomResourceDefinition{
	"certificates":   certificateCRD,
	"issuers":        issuerCRD,
	"clusterissuers": clusterIssuerCRD,
	"orders":         orderCRD,
	"challenges":     challengeCRD,
}

var certificateCRD = &apiextensionv1.CustomResourceDefinition{
	ObjectMeta: metav1.ObjectMeta{Name: "certificates.certmanager.k8s.io", Labels: ControllerLabelMap},
	Spec: apiextensionv1.CustomResourceDefinitionSpec{
		Group:   GroupVersion,
		Version: CRDVersion,
		Scope:   apiextensionv1.NamespaceScoped,
		Names: apiextensionv1.CustomResourceDefinitionNames{
			Plural: "certificates",
			Kind:   "Certificate",
		},
	},
}

var issuerCRD = &apiextensionv1.CustomResourceDefinition{
	ObjectMeta: metav1.ObjectMeta{Name: "issuers.certmanager.k8s.io", Labels: ControllerLabelMap},
	Spec: apiextensionv1.CustomResourceDefinitionSpec{
		Group:   GroupVersion,
		Version: CRDVersion,
		Scope:   apiextensionv1.NamespaceScoped,
		Names: apiextensionv1.CustomResourceDefinitionNames{
			Plural: "issuers",
			Kind:   "Issuer",
		},
	},
}

var clusterIssuerCRD = &apiextensionv1.CustomResourceDefinition{
	ObjectMeta: metav1.ObjectMeta{Name: "clusterissuers.certmanager.k8s.io", Labels: ControllerLabelMap},
	Spec: apiextensionv1.CustomResourceDefinitionSpec{
		Group:   GroupVersion,
		Version: CRDVersion,
		Scope:   apiextensionv1.ClusterScoped,
		Names: apiextensionv1.CustomResourceDefinitionNames{
			Plural: "clusterissuers",
			Kind:   "ClusterIssuer",
		},
	},
}

var orderCRD = `

`

var challengeCRD = &apiextensionv1.CustomResourceDefinition{
	ObjectMeta: metav1.ObjectMeta{Name: "challenges.certmanager.k8s.io", Labels: ControllerLabelMap},
	Spec: apiextensionv1.CustomResourceDefinitionSpec{
		Group:   GroupVersion,
		Version: []apiextensionv1.CustomResourceDefinitionVersion{
			{
				Name: CRDVersion,
				Served: TrueVar,
				Storage: TrueVar,
				Schema: &apiextensionv1.CustomResourceValidation{
					OpenAPIV3Schema: &apiextensionv1.JSONSchemaProps{
						Description: "Challenge is a type to represent a Challenge request with an ACME server",
						Properties: map[string]apiextensionv1.JSONSchemaProps{
							"apiVersion": apiextensionv1.JSONSchemaProps{
								Description: "APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources",
								Type: "string",
							},
							"kind": apiextensionv1.JSONSchemaProps{
								Description: "Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds",
								Type: "string",
							},
							"metadata": apiextensionv1.JSONSchemaProps{
								Type: "object",
							},
							"spec": apiextensionv1.JSONSchemaProps{
								Properties: map[string]apiextensionv1.JSONSchemaProps{
									"authzURL": apiextensionv1.JSONSchemaProps{
										Description: "AuthzURL is the URL to the ACME Authorization resource that this challenge is a part of.",
										Type: "string",
									},
									"config": apiextensionv1.JSONSchemaProps{
										Description: "Config specifies the solver configuration for this challenge. Only **one** of ''config'' or ''solver'' may be specified, and if both are specified then no action will be performed on the Challenge resource. DEPRECATED: the ''solver'' field should be specified instead",
										Type: "object",
										Properties: map[string]apiextensionv1.JSONSchemaProps{
											"dns01": apiextensionv1.JSONSchemaProps{
												Description: "DNS01 contains DNS01 challenge solving configuration",
												Type: "object",
												Properties: map[string]apiextensionv1.JSONSchemaProps{
													"provider":  apiextensionv1.JSONSchemaProps{
														Description: "Provider is the name of the DNS01 challenge provider to use, as configure on the referenced Issuer or ClusterIssuer resource.",
														Type: "string",
													},
												},
												Required: []string {
													"provider",
												},
											},
											"http01": apiextensionv1.JSONSchemaProps{
												Description: "HTTP01 contains HTTP01 challenge solving configuration",
												Type: "object",
												Properties: map[string]apiextensionv1.JSONSchemaProps{
													"ingress": apiextensionv1.JSONSchemaProps{
														Description: "Ingress is the name of an Ingress resource that will be edited to include the ACME HTTP01 'well-known' challenge path in order to solve HTTP01 challenges. If this field is specified, 'ingressClass' **must not** be specified.",
														Type: "string",
													},
													"ingressClass": apiextensionv1.JSONSchemaProps{
														Description: "IngressClass is the ingress class that should be set on new ingress resources that are created in order to solve HTTP01 challenges. This field should be used when using an ingress controller such as nginx, which 'flattens' ingress configuration instead of maintaining a 1:1 mapping between loadbalancer IP:ingress resources. If this field is not set, and 'ingress' is not set, then ingresses without an ingress class set will be created to solve HTTP01 challenges. If this field is specified, 'ingress' **must not** be specified.",
														Type: "string",
													},
												},
											},
										},
									},
									"dnsName": apiextensionv1.JSONSchemaProps{
										Description: "DNSName is the identifier that this challenge is for, e.g. example.com.",
										Type: "string",
									},
									"issuerRef": apiextensionv1.JSONSchemaProps{
										Description: "IssuerRef references a properly configured ACME-type Issuer which should be used to create this Challenge. If the Issuer does not exist, processing will be retried. If the Issuer is not an 'ACME' Issuer, an error will be returned and the Challenge will be marked as failed.",
										Type: "object",
										Properties: map[string]apiextensionv1.JSONSchemaProps{
											"group": apiextensionv1.JSONSchemaProps{
												Type: "string",
											},
											"name": apiextensionv1.JSONSchemaProps{
												Type: "string",
											},
											"kind": apiextensionv1.JSONSchemaProps{
												Type: "string",
											},
										Required: []string{
											"name",
										},
									},
									"key": apiextensionv1.JSONSchemaProps{
										Description: "Key is the ACME challenge key for this challenge",
										Type: "string",
									},
									"solver": apiextensionv1.JSONSchemaProps{
										Description: "Solver contains the domain solving configuration that should be used to solve this challenge resource. Only **one** of 'config' or 'solver' may be specified, and if both are specified then no action will be performed on the Challenge resource.",
										Type: "string",
										Properties: map[string]apiextensionv1.JSONSchemaProps{
											"dns01": apiextensionv1.JSONSchemaProps{
												Properties: map[string]apiextensionv1.JSONSchemaProps{
													"acmedns": apiextensionv1.JSONSchemaProps{
														Description: "ACMEIssuerDNS01ProviderAcmeDNS is a structure containing the configuration for ACME-DNS servers",
														Properties: map[string]apiextensionv1.JSONSchemaProps{
															"accountSecretRef": apiextensionv1.JSONSchemaProps{
																Properties: map[string]apiextensionv1.JSONSchemaProps{
																	"key": apiextensionv1.JSONSchemaProps{
																		Description: "The key of the secret to select from. Must be a valid secret key.",
																		Type: "string",
																	},
																	"name": apiextensionv1.JSONSchemaProps{
																		Description: "Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?",
																		Type: "string",
																	},
																},
																Required: []string {
																	"name",
																},
																Type: "object",
															},
															"host": apiextensionv1.JSONSchemaProps{
																Type: "string",
															},
														},
														Required: []string 
													},
												},
											},
										},
									},
									"token": apiextensionv1.JSONSchemaProps{
										Description: "AuthzURL is the URL to the ACME Authorization resource
										that this challenge is a part of.",
										Type: "string",
									},
									"type": apiextensionv1.JSONSchemaProps{
										Description: "AuthzURL is the URL to the ACME Authorization resource
										that this challenge is a part of.",
										Type: "string",
									},
									"url": apiextensionv1.JSONSchemaProps{
										Description: "AuthzURL is the URL to the ACME Authorization resource
										that this challenge is a part of.",
										Type: "string",
									},
									"wildcard": apiextensionv1.JSONSchemaProps{
										Description: "Wildcard will be true if this challenge is for a wildcard
										identifier, for example '*.example.com'",
										Type: "boolean",
									},
								},
							},
						},
					},
				},
				Subresources: &apiextensionv1.CustomResourceSubresources{

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
						Description: "CreationTimestamp is a timestamp representing the server time when
						this object was created. It is not guaranteed to be set in happens-before order
						across separate operations. Clients may not set this value. It is represented
						in RFC3339 form and is in UTC.",
					},
				},
			},
		},
		Scope:   apiextensionv1.ClusterScoped,
		Names: apiextensionv1.CustomResourceDefinitionNames{
			Plural: "challenges",
			Kind:   "Challenge",
		},
	},
}
