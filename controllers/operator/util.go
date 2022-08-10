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

package operator

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"reflect"
	"strconv"
	"strings"

	utilyaml "github.com/ghodss/yaml"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/apimachinery/pkg/runtime/serializer/streaming"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/client"

	secretshare "github.com/IBM/ibm-secretshare-operator/api/v1"
	certmanagerv1 "github.com/ibm/ibm-cert-manager-operator/v1apis/cert-manager/v1"
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

// YamlToObjects convert YAML content to unstructured objects
func YamlToObjects(yamlContent []byte) ([]*unstructured.Unstructured, error) {
	var objects []*unstructured.Unstructured

	// This step is for converting large yaml file, we can remove it after using "apimachinery" v0.19.0
	if len(yamlContent) > 1024*64 {
		object, err := YamlToObject(yamlContent)
		if err != nil {
			return nil, err
		}
		objects = append(objects, object)
		return objects, nil
	}

	yamlDecoder := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)

	reader := json.YAMLFramer.NewFrameReader(ioutil.NopCloser(bytes.NewReader(yamlContent)))
	decoder := streaming.NewDecoder(reader, yamlDecoder)
	for {
		obj, _, err := decoder.Decode(nil, nil)
		if err != nil {
			if err == io.EOF {
				break
			}
			klog.Infof("error convert object: %v", err)
			continue
		}

		switch t := obj.(type) {
		case *unstructured.Unstructured:
			objects = append(objects, t)
		default:
			return nil, fmt.Errorf("failed to convert object %s", reflect.TypeOf(obj))
		}
	}

	return objects, nil
}

// YamlToObject converting large yaml file, we can remove it after using "apimachinery" v0.19.0
func YamlToObject(yamlContent []byte) (*unstructured.Unstructured, error) {
	obj := &unstructured.Unstructured{}
	jsonSpec, err := utilyaml.YAMLToJSON(yamlContent)
	if err != nil {
		return nil, fmt.Errorf("could not convert yaml to json: %v", err)
	}

	if err := obj.UnmarshalJSON(jsonSpec); err != nil {
		return nil, fmt.Errorf("could not unmarshal resource: %v", err)
	}

	return obj, nil
}

// CompareVersion takes vx.y.z, vx.y.z -> bool: if v1 is larger than v2
func CompareVersion(v1, v2 string) (v1IsLarger bool, err error) {
	if v1 == "" {
		v1 = "0.0.0"
	}
	v1Slice := strings.Split(v1, ".")
	if len(v1Slice) == 1 {
		v1 = "0.0." + v1
	}

	if v2 == "" {
		v2 = "0.0.0"
	}
	v2Slice := strings.Split(v2, ".")
	if len(v2Slice) == 1 {
		v2 = "0.0." + v2
	}

	v1Slice = strings.Split(v1, ".")
	v2Slice = strings.Split(v2, ".")
	for index := range v1Slice {
		v1SplitInt, e1 := strconv.Atoi(v1Slice[index])
		if e1 != nil {
			return false, e1
		}
		v2SplitInt, e2 := strconv.Atoi(v2Slice[index])
		if e2 != nil {
			return false, e2
		}

		if v1SplitInt > v2SplitInt {
			return true, nil
		} else if v1SplitInt == v2SplitInt {
			continue
		} else {
			return false, nil
		}
	}
	return false, nil
}

func Namespacelize(resource, placeholder, ns string) string {
	return strings.ReplaceAll(resource, placeholder, ns)
}

func IssuerHasCondition(issuer *certmanagerv1.Issuer) bool {

	existingConditions := issuer.Status.Conditions
	return existingConditions != nil
}
