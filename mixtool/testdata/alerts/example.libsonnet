local utils = import 'mixin-utils/utils.libsonnet';

{
  prometheusAlerts+:: {
    groups+: [
      {
        name: 'test.alerting',
        rules: [
          {
            'for': '10m',
            alert: 'InstanceDown',
            expr: 'up{job="node"} == 0',
            labels: {
              severity: 'critical',
            },
            annotations: {
              title: 'Instance {{ $labels.instance }} is down.',
              description: 'Failed to scrape {{ $labels.job }} on {{ $labels.instance }} for more than 10 minutes. Node seems down.',
            },
          },
        ],
      },
    ],
  },
}
