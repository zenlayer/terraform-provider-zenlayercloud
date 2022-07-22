package zenlayer

import (
        "context"
        "fmt"
        "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
        "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
        "github.com/zenlayer/terraform-provider-zenlayer/zenlayer/connectivity"
)

func dataSourceZenlayerOSs() *schema.Resource {
        return &schema.Resource{
                ReadContext: dataSourceZenlayerOSsRead,
                Schema: map[string]*schema.Schema{
                        "uuids": {
                                Type: schema.TypeList,
                                Elem: &schema.Schema{Type: schema.TypeString},
                                Computed: true,
                                ForceNew: true,
                        },
                        "names": {
                                Type: schema.TypeList,
                                Elem: &schema.Schema{Type: schema.TypeString},
                                Computed: true,
                                ForceNew: true,
                        },
                        "model_uuid": {
                                Type: schema.TypeString,
                                ForceNew: true,
                                Required: true,
                        },
                },
        }
}

func dataSourceZenlayerOSsRead(ctx context.Context, d *schema.ResourceData, m interface{}) (diags diag.Diagnostics) {
        client := m.(*connectivity.ZenlayerClient)
        bmcService := BmcService{client: client}
        modelUuid := d.Get("model_uuid").(string)
        oss, err := bmcService.ListOSs(modelUuid)
        if err != nil {
                return
        }

        var uuids []string
        var names []string
        for _, os := range oss {
                for _, version := range os.Versions {
                        uuids = append(uuids, version.Uuid)
                        names = append(names, fmt.Sprintf("%s:%s", os.Catalog, version.Name))
                }
        }

        if err := d.Set("uuids", uuids); err != nil {
                return diag.FromErr(err)
        }
        if err := d.Set("names", names); err != nil {
                return diag.FromErr(err)
        }

        d.SetId(dataResourceIdHash(uuids))
        return diags
}
