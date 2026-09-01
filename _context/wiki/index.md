# ibm-cert-manager-operator Wiki

Operator that deploys and manages the `icp-cert-manager` operand on IBM Cloud Pak foundational services (CPfs). Installed standalone as a prerequisite before other CPfs components.

## Contents

| File | What's in it |
|------|-------------|
| [project.md](project.md) | Project overview, goals, relationship to icp-cert-manager |
| [preferences.md](preferences.md) | Working standards, coding style, AI collaboration preferences |

## Quick orientation

- This operator has **one job**: deploy and lifecycle-manage the `icp-cert-manager` operand.
- It is installed **standalone** — it is no longer installed by `ibm-common-service-operator` (that was a deprecated, out-of-support version).
- It is a **prerequisite** for `ibm-common-service-operator` and other CPfs components.
- Certificate rotation controllers were moved **out** of this operator into `ibm-common-service-operator` to simplify scope.
- Most work here is **bug fixes and version bumps** to track `icp-cert-manager` releases.
