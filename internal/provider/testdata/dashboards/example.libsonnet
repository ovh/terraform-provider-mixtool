local utils = import 'mixin-utils/utils.libsonnet';
local filename = 'example.json';

(import 'grafana-builder/grafana.libsonnet') {
  [filename]:
    ($.dashboard('Example') + { uid: std.md5(filename) })
    .addRow(
      ($.row('1st row') +
       {
         showTitle: true,
       })
      .addPanel(
        local title = 'Scrape duration';
        $.panel(title) +
        $.queryPanel(
          |||
            scrape_duration_seconds{}
          ||| % {
          },
          'scrape',
        )
      )
    ),
}
