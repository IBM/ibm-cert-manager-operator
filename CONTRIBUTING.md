<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**  *generated with [DocToc](https://github.com/thlorenz/doctoc)*

- [Contributing guidelines](#contributing-guidelines)
    - [Developer Certificate of Origin](#developer-certificate-of-origin)
    - [Contributing A Patch](#contributing-a-patch)
    - [Issue and Pull Request Management](#issue-and-pull-request-management)
    - [Pre-check before submitting a PR](#pre-check-before-submitting-a-pr)
    - [Build images](#build-images)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

# Contributing guidelines

## Developer Certificate of Origin

This repository built with [probot](https://github.com/probot/probot) that enforces the [Developer Certificate of Origin](https://developercertificate.org/) (DCO) on Pull Requests. It requires all commit messages to contain the `Signed-off-by` line with an email address that matches the commit author.

## Contributing A Patch

1. Submit an issue describing your proposed change to the repo in question.
1. The [repo owners](OWNERS) will respond to your issue promptly.
1. Fork the desired repo, develop and test your code changes.
1. Commit your changes with DCO
1. Submit a pull request.

## Issue and Pull Request Management

Anyone may comment on issues and submit reviews for pull requests. However, in
order to be assigned an issue or pull request, you must be a member of the
[IBM](https://github.com/ibm) GitHub organization.

Repo maintainers can assign you an issue or pull request by leaving a
`/assign <your Github ID>` comment on the issue or pull request.

## Developing

### Pre-requisite

1. [operator-sdk CLI](https://github.com/operator-framework/operator-sdk) v1.23.0 or above

### Version bump

1. Edit the `PREV_VERSION` and `VERSION` values in the Makefile.
1. Edit the image tags in [manager.yaml](config/manager/manager.yaml).
1. Edit the image tags in [base csv](config/manifests/bases/ibm-cert-manager-operator.clusterserviceversion.yaml).
1. Re-generate the bundle.

    ```
    make bundle
    ```

1. Verify CSV has all the edits that were made in previous steps.

## Testing on Open Shift cluster

### Pre-requisites

1. [operator-sdk CLI](https://github.com/operator-framework/operator-sdk) v1.23.0 or above

### Testing bundle with OLM

A bundle is a packaging format for the operator, which mainly consists of the CSV and CRDs. Bundles are understood by OLM. The operator-sdk CLI has the capability to create everything necessary to run this bundle on the cluster.

Running the bundle involves ephemerally creating all the necessary OLM objects to ultimately have the operator's deployment running, such as temporary CatalogSource, OperatorGroup, Subscription, etc.

This type of testing is as close as possible to how IBM Foundational services installs `ibm-cert-manager-operator` without creating a complete IBM Foundational services' CatalogSource and using ODLM.

1. Verify you can build and push the operator's image to a registry. Check the `REGISTRY` variable in Makefile to see what is the default. Recommended to use your own personal registry that your Open Shift cluster has access to.

    ```
    make push-image-amd64

    ```

1. Temporarily edit the `image` field in [manager.yaml](config/manager/manager.yaml) file to be the operator image you pushed in step 1.
1. Verify you can generate the CSV in `bundle/`. The `image` field in the CSV should be the image you pushed in step 1.

    ```
    make bundle
    ```

1. Verify you can build the image for the operator bundle.

    ```
    make bundle-build
    ```

1. Push the bundle up to a registry. Check the `REGISTRY` variable in Makefile to see what is the default. Recommended to use your own personal registry that your Open Shift cluster has access to.

    ```
    make bundle-push
    ```

1. Use the built-in operator-sdk feature to [run the bundle](https://sdk.operatorframework.io/docs/olm-integration/tutorial-bundle/#deploying-an-operator-with-olm)
 
    ```
    make bundle-run
    ```

1. Verify operator is running, and you can create the operands by creating a new CertManager object
1. Revert the `image` change in [manager.yaml](config/manager/manager.yaml) file, and re-generate the bundle before opening PR

    ```
    make bundle
    ```

## Pre-check before submitting a PR
