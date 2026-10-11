## Unreleased

BREAKING CHANGES:

* Terraform 1.6 is the oldest tested version. Earlier versions are no longer tested.

FEATURES:

* data-source/mixtool_alerts, data-source/mixtool_rules, data-source/mixtool_dashboards: optional `lint` attribute runs `mixtool lint` and fails on any finding.
* function/alerts, function/rules, function/dashboards: render a mixin without a data block. Requires Terraform 1.8 or later.

BUG FIXES:

* data-source/mixtool_alerts, data-source/mixtool_rules, data-source/mixtool_dashboards: `id` is the SHA-256 of the rendered output, not of the `source` path, so it changes when the mixin or an import changes. Existing users see a one-time `id` change on upgrade.

## 0.1.0 (November 15, 2022)

Initial version
