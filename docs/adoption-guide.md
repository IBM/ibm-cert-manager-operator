# Cert-manager Adoption Guide for Operators

- [Background](#back)
- [How to use cert-manager generally](#how)
- [Guidance with the CA](#ca)
- [Resources and Additional Links](#res)

## Background

{: #back}

Previously, when all the services were deployed as helm charts, it was easy to use cert-manager by specifying a yaml file with your cert-manager resource in your chart. When your chart was installed, the cert-manager resources were created.

## How to do it

Fulfill the prerequisites:
1. [Preqrequisites](#pre)

There are two options to create cert-manager resources in your operator:
1. [As Go code](#go)
1. [As yaml](#yaml)

### Prerequisites

{: #pre}

1. In your operator's `Role` (most commonly found in `deploy/role.yaml` in your operator's directory), add the following so that your operator has permission to create/read/update/delete cert-manager Certificate resources.

    ````
    rules:
    - apiGroup:
      - certmanager.k8s.io
      resources:
      - certificates
      verbs:
      - create
      - get
      - update
      - delete
      - patch
      - list
      - watch
    ````

### Go Code

{: #go}

1. In the `require` section of your operator's go.mod file
    - add:

        ````
        require (
            github.com/jetstack/cert-manager v0.10.0
        ````

    - Change

        ````
        require (
            k8s.io/api v0.0.0
            k8s.io/apiextensions-apiserver v0.0.0
            k8s.io/apimachinery v0.0.0
        ````

        to

        ````
        require (
            k8s.io/api v0.17.0
            k8s.io/apiextensions-apiserver v0.17.0
            k8s.io/apimachinery v0.17.0
        ````

1. In the `replace` section of your operator's go.mod file add:

    ````
    replace (
        github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.3.4-0.20200207053602-7439e774c9e9+incompatible
    ````

1. Run `go mod tidy`
1. Add in cmd/manager/main.go

    ````
    import (
        certmgr "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha1"
    )
    ...
    func main() {
    ...
        // Setup Scheme for all resources
        if err := apis.AddToScheme(mgr.GetScheme()); err != nil {
            log.Error(err, "")
            os.Exit(1)
        }
        if err := certmgr.AddToScheme(mgr.GetScheme()); err != nil {
            log.Error(err, "")
            os.Exit(1)
        }
    ...

    ````

1. Add in your code to create your cert-manager resource:
    - Certificate example:

        ````
        import (
            certmgr "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha1"

            logf "sigs.k8s.io/controller-runtime/pkg/log"
            metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
        )
        func (r *ReconcileCertManager) CreateCertificate(request reconcile.Request) error {
            log.Info("Creating cert manager certificate")
            crt := &certmgr.Certificate{
                ObjectMeta: metav1.ObjectMeta{
                    Name:      "test-certificate",
                    Namespace: "default",
                },
                Spec: certmgr.CertificateSpec{
                    SecretName: "test-secret",
                    IssuerRef: certmgr.ObjectReference{
                        Name: "test-issuer",
                        Kind: "Issuer",
                    },
                    CommonName: "test",
                },
            }

            if err := r.client.Create(context.TODO(), crt); err != nil {
                    log.Error(err, "Error creating cert-manager certificate")
                    return err
            }
            return nil
        }
        ````

    - Issuer example:

        ````
        import (
            certmgr "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha1"

            logf "sigs.k8s.io/controller-runtime/pkg/log"
            metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
        )
        func (r *ReconcileCertManager) CreateIssuer(request reconcile.Request) error {
            log.Info("Creating cert manager issuer")
            issuer := &certmgr.Issuer{
            ObjectMeta: metav1.ObjectMeta{
                Name:      "test-issuer",
                Namespace: "default",
            },
            Spec: certmgr.IssuerSpec{
                        IssuerConfig: certmgr.IssuerConfig{
                            CA: &certmgr.CAIssuer{
                                SecretName: "my-ca-secret",
                            },
                        },
                    },
            }

            if err := r.client.Create(context.TODO(), issuer); err != nil {
                    log.Error(err, "Error creating cert-manager issuer")
                    return err
            }
            return nil
        }
        ````

#### Live Example

Can be found in [ibm-cert-manager-operator](http://github.com/Crystal-Chun/ibm-cert-manager-operator/tree/test-certmanager)

### Yaml

{: #yaml}

This way will be most similar to how it's done in the helm chart.
Credits to @chenzhiwei for coming up with this.

1. In a go file, define your cert-manager resource yaml spec
    - Example certificate in `pkg/controller/certManagerResource/resource.go`:

        ````
        package certManagerResource

        const certYaml = `
        apiVersion: certmanager.k8s.io/v1alpha1
        kind: Certificate
        metadata:
        name: my-certificate
        namespace: my-namespace
        spec:
        issuerRef:
            name: my-issuer
            kind: Issuer
        secretName: my-cert-secret
        dnsNames:
        - foo1.bar1
        `
        ````

    - Example issuer in `pkg/controller/certManagerResource/resource.go`:

        ````
        package certManagerResource

        const issuerYaml = `
        apiVersion: certmanager.k8s.io/v1alpha1
        kind: Issuer
        metadata:
        name: my-issuer
        namespace: my-namespace
        spec:
        selfSigned: {}
        `
        ````

1. In a go function, create it:

    ````
    import (
        "context"

        "github.com/ghodss/yaml"
        "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
        "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

        operatorv1alpha1 "github.com/IBM/ibm-operator/pkg/apis/operator/v1alpha1"
    )

    // Creates a yaml spec
    // yamlSpec is the yaml itself
    // client is your Reconciler's client (r.client)
    func (r *ReconcileOperandCRD) create(instance *operatorv1alpha1.OperandCRD, yamlSpec []byte) error{
    ````

    1. Converting from YAML to JSON

        ````
        json, err := yaml.YAMLTOJSON(yamlSpec)
        if err != nil {
            return err
        }
        ````

    1. Unmarshalling from JSON

        ````
        obj := &unstructured.Unstructured{}
        if err = obj.UnmarshalJSON(json) ; err != nil {
            return err
        }
        ````

    1. Set the controller reference using the `controllerutil`

        ````
        if err = controllerutil.SetControllerReference(instance, obj, r.scheme) ; err != nil {
            return err
        }
        ````

    1. Creating it

        ````
        if err = r.client.Create(context.TODO(), obj) ; err != nil && !errors.IsAlreadyExists(err) {
            return err
        }
        ````

    ````
        return nil
    }
    ````

1. In your `Reconcile` function or another go function, use your `create` function to create your cert-manager resources:

    ````
    import (
        certMgrRes "pkg/controller/certManagerResource"

        "sigs.k8s.io/controller-runtime/pkg/reconcile"
        "sigs.k8s.io/controller-runtime/pkg/log"

    ...

    func (r *ReconcileOperandCRD) Reconcile(request reconcile.Request) (reconcile.Result, error) {
        ...
        if err = r.create(instance, []byte(certMgrRes.issuerYaml)) ; err != nil {
            log.Error(err, "Error creating cert-manager issuer")
            return reconcile.Result{}, err
        }
        if err = r.create(instance, []byte(certMgrRes.certYaml)) ; err != nil {
            log.Error(err, "Error creating cert-manager certificate")
            return reconcile.Result{}, err
        }
        ...
    }
    ````

#### Example

Courtesy of @chenzhiwei: [ibm-mongodb-operator](https://github.com/IBM/ibm-mongodb-operator/pull/28/files)

## Guidance for Common Services Adopters requiring common CA

{: #ca}

- [Background](#ca-back)
- [The Problem](#problem)
- [Proposed Solution](#proposal)
- [Your Steps to Adopt](#steps)

### CA Background

{: #ca-back}

Previously in ICP and common services 4Q 2019 release, the icp-inception installer created a Root CA (self-signed CA certificate) that was used to create the ClusterIssuer `icp-ca-issuer`. From there, all the services were able to create Certificate yaml specs in their helm charts that were issued by the `icp-ca-issuer`.

This scenario was fine because:
1. The icp inception installer created the Root CA certificate
1. The icp inception installer also installed all of the helm charts and had ClusterAdmin permissions to do so

### The Problem

{: #problem}

Now with moving to operators:
1. The ODLM doesn't create this Root CA certificate anymore

### Proposed Solution

{: #proposal}

We've thought of multiple ways to handle the problems faced above and this is our proposed solution to it.

1. We (cert-manager) will take responsibility of creating the common CA ClusterIssuer shared amongst the common services.
    - This is generated DIFFERENTLY than how the icp-inception installer created it
        - Essentially, we will create a self-signed Issuer (cert-manager resource), create a Certificate (cert-manager resource) that is a CA certificate with a well-known Secret name, and then create a CA ClusterIssuer from that CA certificate.
    - The result is a ClusterIssuer called `cs-ca-clusterissuer` which is identical to the `icp-ca-issuer`.
    - This has some benefits such as
        - Support from cert-manager for automatic refreshing of the CA certificate when it expires
        - The ability to manually refresh it easily by deleting the secret.
        - The ability to BYO CA by replacing this secret and deleting the cert-manager Certificate
1. Each operator will just need to create their Certificates (cert-manager resource) signed by the common CA ClusterIssuer.
    - See steps below

### Steps

{: #steps}

To adopt the solution, each operator must:

1. Create your Certificate (cert-manager resource) using the common CA ClusterIssuer that's predefined.
    - Example go code (see yaml example above if you wish to do it that way):

        ````
            crt := &certmgr.Certificate{
                ObjectMeta: metav1.ObjectMeta{
                    Name:      "my-certificate",
                    Namespace: "ibm-common-services",
                },
                Spec: certmgr.CertificateSpec{
                    SecretName: "my-secret",
                    IssuerRef: certmgr.ObjectReference{
                        Name: "cs-ca-clusterissuer",
                        Kind: "ClusterIssuer",
                    },
                    CommonName: "my-service-name",
                },
            }

            if err := r.client.Create(context.TODO(), crt); err != nil {
                    log.Error(err, "Error creating cert-manager certificate")
                    return err
            }
            return nil
        ````

        - Notice the `Name` of the Spec.IssuerRef.Name matches the Issuer I defined in step 1.

## Resources

{: #res}

1. [Cert-Manager Knowledge Center Documents](https://www.ibm.com/support/knowledgecenter/SSBS6K_3.2.1/manage_applications/cert_manager.html?pos=2)
