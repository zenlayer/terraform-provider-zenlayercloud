package zenlayer

import (
        "context"
        "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
        "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
        "github.com/zenlayer/terraform-provider-zenlayer/zenlayer/connectivity"
)

func dataSourceZenlayerModels() *schema.Resource {
        return &schema.Resource{
                ReadContext: dataSourceZenlayerModelsRead,
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
                        "zone_uuid": {
                                Type: schema.TypeString,
                                ForceNew: true,
                                Required: true,
                        },
                },
        }
}

func dataSourceZenlayerModelsRead(ctx context.Context, d *schema.ResourceData, m interface{}) (diags diag.Diagnostics) {
        client := m.(*connectivity.ZenlayerClient)
        bmcService := BmcService{client: client}
        zoneUuid := d.Get("zone_uuid").(string)
        models, err := bmcService.ListModels(zoneUuid)

        if err != nil {
                return diag.FromErr(err)
        }

        var uuids []string
        var names []string
        for _, model := range models {
                if model.Stock < 0 {
                        uuids = append(uuids, model.Uuid)
                        names = append(names, model.Name)
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
