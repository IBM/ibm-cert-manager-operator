# Onboarding to cert-manager

## Background

Previously, when all the services were deployed as helm charts, it was easy to use cert-manager by specifying a yaml file with your cert-manager resource in your chart. When your chart was installed, the cert-manager resources were created.

## How to do it

There are two ways to create cert-manager resources like in the background now that we've switched to operators:
1. [As Go code](#go)
1. [As yaml](#yaml)

## Prerequisites

{: #pre}

1. You may need to add the following additional permissions in your operator's `Role` in `deploy/role.yaml`

    ````
    ...
    rules:
    - apiGroups:
      - certmanager.k8s.io
      resources:
      - certificates
      - issuers
      verbs:
      - create
    ````

## Go Code

{: #go}

1. Complete the [prerequisites](#pre)
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
                        Name: "icp-ca-issuer",
                        Kind: "ClusterIssuer",
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

### Live Example

Can be found in [ibm-cert-manager-operator](http://github.com/Crystal-Chun/ibm-cert-manager-operator/tree/test-certmanager)

## Yaml

{: #yaml}

This way will be most similar to how it's done in the helm chart.
Credits to @chenzhiwei for coming up with this.

1. Complete the [prerequisites](#pre)
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

### Example

Courtesy of @chenzhiwei: [ibm-mongodb-operator](https://github.com/IBM/ibm-mongodb-operator/pull/28/files)

## Resources

1. [Knowledge Center Documents]()
