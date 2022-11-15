package mixtool

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/monitoring-mixins/mixtool/pkg/mixer"
	"gopkg.in/yaml.v3"
)

func dataSourceMixtoolAlerts() *schema.Resource {
	return &schema.Resource{
		Description: "Read mixin and render alerts as json using mixtool",

		ReadContext: dataSourceMixtoolAlertsRead,

		Schema: map[string]*schema.Schema{
			"source": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Source jsonnet file.",
			},
			"jsonnet_path": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "External variables providing value as a string.",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"alerts": {
				Computed:    true,
				Description: "Generated Prometheus alerts based on the given mixins",
				Type:        schema.TypeString,
			},
		},
	}
}

func dataSourceMixtoolAlertsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	source := d.Get("source").(string)
	jsonnetPath := d.Get("jsonnet_path").([]interface{})

	// Convert jsonnetPath from []interface{} to []string
	jpath := make([]string, len(jsonnetPath))
	for i, v := range jsonnetPath {
		jpath[i] = fmt.Sprint(v)
	}

	opts := mixer.GenerateOptions{
		JPaths: jpath,
	}

	alerts, err := mixer.GenerateAlerts(source, opts)
	if err != nil {
		return diag.FromErr(err)
	}

	var i interface{}
	if err := json.Unmarshal(alerts, &i); err != nil {
		return diag.FromErr(err)
	}

	alertsCompiled, err := yaml.Marshal(i)
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("alerts", string(alertsCompiled))
	d.SetId(hash(source))
	return nil
}
