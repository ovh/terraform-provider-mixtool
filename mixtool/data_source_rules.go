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

func dataSourceMixtoolRules() *schema.Resource {
	return &schema.Resource{
		Description: "Read mixin and render rules as json using mixtool",

		ReadContext: dataSourceMixtoolRulesRead,

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
			"rules": {
				Computed:    true,
				Description: "Generated Prometheus alerts based on the given mixins",
				Type:        schema.TypeString,
			},
		},
	}
}

func dataSourceMixtoolRulesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	rules, err := mixer.GenerateRules(source, opts)
	if err != nil {
		return diag.FromErr(err)
	}

	var i interface{}
	if err := json.Unmarshal(rules, &i); err != nil {
		return diag.FromErr(err)
	}

	rulesCompiled, err := yaml.Marshal(i)
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("rules", string(rulesCompiled))
	d.SetId(hash(source))
	return nil
}
