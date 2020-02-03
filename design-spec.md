# Operator Flow
1. User creates `CertManager` CR
2. Operator picks up the CR spec
3. Operator performs reconcile loop for cert manager service
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
            - Remove preexisting clusterrole and clusterrolebinding
                - if they don't exist, ignore the error 
            - Recreate clusterrole from default spec
                - if there's an error creating it, requeue and try again
            - Recreate clusterrolebinding from default spec
                - if there's an error creating it, requeue and try again
            - Create the service account ignoring errors if it already exists
                - if there's an error that's not related to it already existing, requeue and try again
    4. Check the deployment
        - If the cert-manager deployment exists 
            - if it does, check if anything differs from the template
                - if it does send an update using the template
            - if it does not exist, create it 
        - If webhook is enabled
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
4. If the CR is deleted
    1. If there is a finalizer on the CR, delete all resources created by this operator
        - Delete the RBAC
            - Removes clusterrolebinding
            - Removes the clusterrole
            - Removes the imagepullsecret
        - Delete webhook deploy
        - Delete cainjector deploy
        - Delete controller deploy
    2. Remove the finalizer from list of finalizers on CR
            
## Log levels
0. Always log
    - Error messages that prevent continuing through the flow
    - Basic status such as 
1. Level 1 logging
    - Errors that are not critical such as missing deployment for deletion
    - 
2. Level 2 logging
    - Status logging seeing where we are in the process of things such as "finished checking crds", etc
3. Debugging information
    - Arguments for the deployment, deployment itself, etc
