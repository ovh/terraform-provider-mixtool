package mixtool

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/monitoring-mixins/mixtool/pkg/mixer"
)

func dataSourceMixtoolDashboards() *schema.Resource {
	return &schema.Resource{
		Description: "Read mixin and render dashboards as json using mixtool",

		ReadContext: dataSourceMixtoolDashboardsRead,

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
			"dashboards": {
				Computed: true,
				Type:     schema.TypeMap,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func dataSourceMixtoolDashboardsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	dashboards, err := mixer.GenerateDashboards(source, opts)
	if err != nil {
		return diag.FromErr(err)
	}

	rendered := make(map[string]string)

	// Convert dashboardContent from json.RawMessage to string
	for dashboardName, dashboardContent := range dashboards {
		rendered[dashboardName] = strings.TrimSpace(string(dashboardContent))
	}

	d.Set("dashboards", rendered)
	d.SetId(hash(source))
	return nil
}

func hash(s string) string {
	sha := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sha[:])
}
