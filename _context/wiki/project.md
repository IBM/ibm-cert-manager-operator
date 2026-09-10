# Project Overview

## What this project is

`ibm-cert-manager-operator` is an operator-sdk based operator whose sole responsibility is to deploy and lifecycle-manage the `icp-cert-manager` operand on IBM Cloud Pak foundational services (CPfs) clusters.

It is a **prerequisite** for installing `ibm-common-service-operator` and other CPfs components. It is installed standalone — it is no longer installed by `ibm-common-service-operator` (that integration was deprecated and is out of support).

Certificate rotation controllers previously in this operator have been **moved to `ibm-common-service-operator`** as a deliberate simplification. This operator's scope is intentionally narrow.

## Main goals

1. **Deploy `icp-cert-manager`** — install and manage the cert-manager operand on CPfs clusters.
2. **Track `icp-cert-manager` releases** — bump versions as new `icp-cert-manager` releases are cut.
3. **Minimal scope** — keep this operator simple; complexity belongs in `icp-cert-manager` or `ibm-common-service-operator`.

## Key stakeholders / users

- **CPfs platform team** — maintainers.
- **IBM Cloud Pak teams** — consume certificate management through CPfs.
- **`ibm-common-service-operator`** — depends on cert-manager being available before it installs.

## Relationship to `icp-cert-manager`

| Repo | Role |
|------|------|
| `icp-cert-manager` | The cert-manager fork — actual cert issuance, renewal, rotation logic |
| `ibm-cert-manager-operator` | Operator that installs and manages `icp-cert-manager` on the cluster |

Version bumps in this operator track releases of `icp-cert-manager`.

## Key workflows

- **Bug fixes** — most common work.
- **Version bumps** — updating operand image references when a new `icp-cert-manager` release is cut.

## What no longer lives here

- Certificate rotation controllers → moved to `ibm-common-service-operator`.
