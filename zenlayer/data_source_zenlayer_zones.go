package zenlayer

import (
        "context"
        "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
        "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
        "github.com/zenlayer/terraform-provider-zenlayer/zenlayer/connectivity"
)

func dataSourceZenlayerZones() *schema.Resource {
        return &schema.Resource{
                ReadContext: dataSourceZenlayerZonesRead,
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
                        "titles": {
                                Type: schema.TypeList,
                                Elem: &schema.Schema{Type: schema.TypeString},
                                Computed: true,
                                ForceNew: true,
                        },
                },
        }
}

func dataSourceZenlayerZonesRead(ctx context.Context, d *schema.ResourceData, m interface{}) (diags diag.Diagnostics) {
        client := m.(*connectivity.ZenlayerClient)
        bmcService := BmcService{client: client}
        zones, err := bmcService.ListZones()
        if err != nil {
                return
        }

        var uuids []string
        var names []string
        var titles []string
        for _, zone := range zones {
                if zone.Enable {
                        uuids = append(uuids, zone.Uuid)
                        names = append(names, zone.Name)
                        titles = append(titles, zone.Title)
                }
        }

        if err := d.Set("uuids", uuids); err != nil {
                return diag.FromErr(err)
        }
        if err := d.Set("names", names); err != nil {
                return diag.FromErr(err)
        }
        if err := d.Set("titles", titles); err != nil {
                return diag.FromErr(err)
        }

        d.SetId(dataResourceIdHash(uuids))
        return diags
}
