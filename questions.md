- question about image pull secret
    - security hole if we can copy a secret from any namespace to that certmanager namespace - nefarious person that can take any secret from a restricted namespace and gain access
    - two scenarios
        1. prereq that user provide their own image pull secret in cert-manager namespace no matter what
        2. need to copy it over for installer deployed image pull secret - where is it placing common location?
    - is there a way to consider restricting to only pull from one namespace 
    - no namespace level control with rbac for operator
