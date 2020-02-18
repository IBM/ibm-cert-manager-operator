# Operator Flow

1. User creates `CertManager` CR
1. Operator picks up the CR spec
1. If the CR is not named "default", it is not allowed to pass through
1. Operator performs reconcile loop for cert manager service
    1. Checks the instance of the CertManager CR exists
        - if it doesn't, exit
        - if it does, continue
    1. Check cert-manager CRDs (certificates, issuers, etc.) exist
        - if they exist, continue
        - if they do not exist, try to create them
            - if there's errors creating them, requeue and try again
    1. Check the cert-manager namespace exists
        - if it exists, continue
        - if it does not exist, create it
            - if there's an error creating it, requeue and try again
    1. Check RBAC is in place for cert-manager
        - Check image pull secret
            - Check if the secret is already in the cert-manager namespace
            - Check if the pull secret is in the namespace specified on the CR aka CopyPullSecret
                - if no namespace is specified, skip
            - Define a new pull secret using
                - the name of the image pull secret if specified
                - the cert-manager namespace
                - the data of the pull secret in the namespace of the CR (CopyPullSecret) if it exists
            - If both the pull secret exists and the CopyPullSecret exists, update the existing pull secret so it gets the data of the CopyPullSecret
            - if the CopyPullSecret exists but the pull secret is not already in the cert-manager namespace, create it (effectively copying it over from one namespace to the cert-manager namespace)
            - If neither the pull secret in the cert-manager namespace or the CopyPullSecret exist, throw an error and requeue
            - If the CopyPullSecret does not exist, but the image pull secret exists in the cert-manager namespace, do nothing and continue
        - Check roles
            - Clusterrole
                - if it exists, continue
                - if it doesn't exist, create it from default spec
                    - if there's an error creating it, requeue and try again
            - ClusterRoleBinding
                - if it exists, continue
                - if it doesn't exist, create it from default spec
                    - if there's an error creating it, requeue and try again
            - Create the service account ignoring errors if it already exists
                - if there's an error that's not related to it already existing, requeue and try again
    1. Check the deployment
        - If the cert-manager deployment exists
            - if it does, check if anything differs from the template
                - if it does send an update using the template
            - if it does not exist, create it
        - If the configmap-watcher deployment exists
            - if it does, check if anything differs from the template
                - if it does send an update using the template
            - if it does not exist, create it
        - If webhook is enabled
            - Check the prerequisites
                - if the service exists, do nothing, otherwise create it
                    - if there's an error creating it, requeue and try again
                - if the apiservice exists, do nothing, otherwise create it
                    - if there's an error creating it, requeue and try again
                - if the mutatingwebhookconfiguration exists, do nothing, otherwise create it
                    - if there's an error creating it, requeue and try again
                - if the validatingwebhookconfiguration exists, do nothing, otherwise create it
                    - if there's an error creating it, requeue and try again
                - if the webhook rolebinding exists, do nothing, otherwise create it
                    - if there's an error creating it, requeue and try again
            - If cainjector deploy exists
                - if it does, check if anything differs from the template
                    - if it does send an update using the template
                - if it does not exist, create it
            - if webhook deploy exists
                - if it does, check if anything differs from the template
                    - if it does send an update using the template
                - if it does not exist, create it
        - If webhook is not enabled
            - Delete cainjector deploy if it exists
            - Delete webhook deploy if it exists
            - Delete webhook prerequisites if they exist
1. If the CR is deleted
    1. If there is a finalizer on the CR, all resources created by this operator are deleted automatically
        - RBAC
            - Removes clusterrolebinding
            - Removes the clusterrole
            - Removes the imagepullsecret - if it was copied over by this operator
            - Removes webhook rolebinding
        - Deployments
            - webhook deploy
            - cainjector deploy
            - controller deploy
            - configmap-watcher deploy
        - Validating Webhook Configuration
        - Mutating Webhook Configuration
        - Service
        - API Service
        - NOTE: the finalizer automatically removes resources created by this operator. If any of these were not created by operator and were already present in the system then the operator will not remove them upon CR removal.
    1. Remove the finalizer from list of finalizers on CR

## Log levels

1. Level 0 logging
    - Error messages that prevent continuing through the flow
    - Basic status
    - Always log
1. Level 1 logging
    - Errors that are not critical such as missing deployment for deletion
1. Level 2 logging
    - Status logging seeing where we are in the process of things such as "finished checking crds", etc
1. Debugging information
    - Arguments for the deployment, deployment itself, etc
