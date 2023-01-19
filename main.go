/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"flag"
	"os"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	admRegv1 "k8s.io/api/admissionregistration/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiextensionv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	apiRegv1 "k8s.io/kube-aggregator/pkg/apis/apiregistration/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlpkg "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	cache "github.com/IBM/controller-filtered-cache/filteredcache"
	secretshare "github.com/IBM/ibm-secretshare-operator/api/v1"

	acmecertmanagerv1 "github.com/ibm/ibm-cert-manager-operator/apis/acme.cert-manager/v1"
	certmanagerv1 "github.com/ibm/ibm-cert-manager-operator/apis/cert-manager/v1"
	certmanagerv1alpha1 "github.com/ibm/ibm-cert-manager-operator/apis/certmanager/v1alpha1"
	metacertmanagerv1 "github.com/ibm/ibm-cert-manager-operator/apis/meta.cert-manager/v1"
	operatorv1alpha1 "github.com/ibm/ibm-cert-manager-operator/apis/operator/v1alpha1"
	certmanagerv1controllers "github.com/ibm/ibm-cert-manager-operator/controllers/cert-manager"
	certmanagercontrollers "github.com/ibm/ibm-cert-manager-operator/controllers/certmanager"
	operatorcontrollers "github.com/ibm/ibm-cert-manager-operator/controllers/operator"
	constants "github.com/ibm/ibm-cert-manager-operator/controllers/resources"
	//+kubebuilder:scaffold:imports
)

var log = logf.Log.WithName("cmd")

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(operatorv1alpha1.AddToScheme(scheme))
	utilruntime.Must(certmanagerv1alpha1.AddToScheme(scheme))
	utilruntime.Must(metacertmanagerv1.AddToScheme(scheme))
	utilruntime.Must(acmecertmanagerv1.AddToScheme(scheme))
	utilruntime.Must(certmanagerv1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	gvkLabelMap := map[schema.GroupVersionKind]cache.Selector{
		corev1.SchemeGroupVersion.WithKind("Secret"): {
			LabelSelector: constants.SecretWatchLabel,
		},
	}

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "1557e857.ibm.com",
		NewCache:               cache.NewFilteredCacheBuilder(gvkLabelMap),
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	// Setup Scheme for all resources
	if err := clientgoscheme.AddToScheme(mgr.GetScheme()); err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	if err := apiextensionv1.AddToScheme(mgr.GetScheme()); err != nil {
		log.Error(err, "")
		os.Exit(1)
	}
	if err := apiRegv1.AddToScheme(mgr.GetScheme()); err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	if err := admRegv1.AddToScheme(mgr.GetScheme()); err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	if err := secretshare.AddToScheme(mgr.GetScheme()); err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	// build index for cache that controller client reads from
	// necessary for filtering objects by FieldSelector
	cache := mgr.GetCache()
	indexFunc := func(obj ctrlpkg.Object) []string {
		return []string{obj.(*appsv1.Deployment).Name}
	}
	if err := cache.IndexField(context.Background(), &appsv1.Deployment{}, "metadata.name", indexFunc); err != nil {
		panic(err)
	}

	kubeclient, _ := kubernetes.NewForConfig(mgr.GetConfig())
	apiextclient, _ := apiextensionclientset.NewForConfig(mgr.GetConfig())
	ns, _ := os.LookupEnv("OPERATOR_NAMESPACE")
	if err = (&operatorcontrollers.CertManagerReconciler{
		Client:       mgr.GetClient(),
		Reader:       mgr.GetAPIReader(),
		Kubeclient:   kubeclient,
		APIextclient: apiextclient,
		Scheme:       mgr.GetScheme(),
		Recorder:     mgr.GetEventRecorderFor("ibm-cert-manager-operator"),
		NS:           ns,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "CertManager")
		os.Exit(1)
	}
	if err = (&certmanagercontrollers.IssuerReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Issuer")
		os.Exit(1)
	}
	if err = (&certmanagercontrollers.CertificateReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Certificate")
		os.Exit(1)
	}
	if err = (&certmanagercontrollers.ChallengeReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Challenge")
		os.Exit(1)
	}
	if err = (&certmanagercontrollers.OrderReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Order")
		os.Exit(1)
	}
	if err = (&certmanagercontrollers.CertificateRequestReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "CertificateRequest")
		os.Exit(1)
	}
	if err = (&certmanagerv1controllers.CertificateRefreshReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "CertificateRefresh")
		os.Exit(1)
	}
	if err = (&certmanagerv1controllers.PodRefreshReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "PodRefresh")
		os.Exit(1)
	}
	if err = (&certmanagerv1controllers.V1AddLabelReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "V1AddLabel")
		os.Exit(1)
	}
	if err = (&certmanagerv1controllers.V1Alpha1AddLabelReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "V1Alpha1AddLabel")
		os.Exit(1)
	}
	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
