echo "Starting cleanup for OLM -> No OLM migration..."
operatorNamespace={{ .Values.global.certmgrNamespace }}
servicesNamespace={{ .Values.global.instanceNamespace }}
namespaces=$(oc get cm namespace-scope -n $operatorNamespace -o jsonpath="{.data.namespaces}")

certSub=$(oc get subscription.operators.coreos.com -n "$operatorNamespace" -o jsonpath="{.items[?(@.spec.name=='ibm-cert-manager-operator')].metadata.name}")
if [[ -z $certSub ]]; then
    echo "IBM Cert Manager Subscription not present in namespace $operatorNamespace, aborting."
else
    certCSV=$(oc get --ignore-not-found subscription.operators.coreos.com $certSub -n $operatorNamespace -o jsonpath='{.status.currentCSV}')
    echo "Deleting IBM Cert Manager CSV and Subscription in namespace $operatorNamespace..."
    oc delete --ignore-not-found csv $certCSV -n $operatorNamespace && oc delete --ignore-not-found subscription.operators.coreos.com $certSub -n $operatorNamespace

    #deployments are operands right? Do they need to be deleted?
    # echo "Cleaning up IBM Cert Manager deployments"
    # oc delete --ignore-not-found deploy cert-manager-cainjector cert-manager-controller cert-manager-webhook -n $operatorNamespace

    echo "Cleaning up IBM Cert Manager RBAC"
    oc delete --ignore-not-found sa ibm-cert-manager-operator -n $operatorNamespace
    roles=$(oc get roles -n $operatorNamespace | grep ibm-cert-manager-op | awk '{print $1}' | tr "\n" " ")
    rolebindings=$(oc get rolebindings -n $operatorNamespace | grep ibm-cert-manager-op | awk '{print $1}' | tr "\n" " ")
    clusterroles=$(oc get clusterroles | grep ibm-cert-manager-op | awk '{print $1}' | tr "\n" " ")
    clusterrolebindings=$(oc get clusterrolebindings | grep ibm-cert-manager-op | awk '{print $1}' | tr "\n" " ")
    oc delete --ignore-not-found roles $roles -n $operatorNamespace
    oc delete --ignore-not-found rolebindings $rolebindings -n $operatorNamespace
    oc delete --ignore-not-found clusterroles $clusterroles -n $operatorNamespace
    oc delete --ignore-not-found clusterrolebindings $clusterrolebindings -n $operatorNamespace

    echo "IBM Cert Manager OLM install cleaned up."
    echo "Ready for No OLM install in namespace $operatorNamespace."
fi