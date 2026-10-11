{
  grafanaDashboards+:
    (import 'dashboards/example.libsonnet') +
    (import 'dashboards/second.libsonnet'),
}
