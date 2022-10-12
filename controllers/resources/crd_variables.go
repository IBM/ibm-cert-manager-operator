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

import _ "embed"

//go:embed v1crds/certificaterequests.yaml
var CertificaterequestsCRD string

//go:embed v1crds/certificates.yaml
var CertificatesCRD string

//go:embed v1crds/challenges.yaml
var ChallengesCRD string

//go:embed v1crds/clusterissuers.yaml
var ClusterissuersCRD string

//go:embed v1crds/issuers.yaml
var IssuersCRD string

//go:embed v1crds/orders.yaml
var OrdersCRD string
