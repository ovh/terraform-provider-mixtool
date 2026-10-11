## Unreleased

FEATURES:

* data-source/mixtool_alerts, data-source/mixtool_rules, data-source/mixtool_dashboards: optional `lint` attribute runs `mixtool lint` and fails on any finding.

BUG FIXES:

* data-source/mixtool_alerts, data-source/mixtool_rules, data-source/mixtool_dashboards: `id` is the SHA-256 of the rendered output, not of the `source` path, so it changes when the mixin or an import changes. Existing users see a one-time `id` change on upgrade.

## 0.1.0 (November 15, 2022)

Initial version
