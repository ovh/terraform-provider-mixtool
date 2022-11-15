local utils = import 'mixin-utils/utils.libsonnet';

{
  prometheusRules+:: {
    groups+: [
      // See: https://prometheus.io/docs/practices/rules/#examples
      {
        name: 'test.rules',
        rules: [
          {
            record: 'instance_path:requests:rate5m',
            expr: 'rate(requests_total{job="myjob"}[5m])',
          },
        ],
      },
    ],
  },
}
