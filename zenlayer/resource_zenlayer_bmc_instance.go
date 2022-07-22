package zenlayer

import (
        "context"
        "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
        "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
        "github.com/zenlayer/terraform-provider-zenlayer/zenlayer/connectivity"
        "github.com/zenlayer/zenlayer-go-sdk/services/bmc"
)

func resourceZenlayerBmcInstance() *schema.Resource {
        return &schema.Resource{
                CreateContext: resourceZenlayerBmcInstanceCreate,
                ReadContext: resourceZenlayerBmcInstanceRead,
                UpdateContext: resourceZenlayerBmcInstanceUpdate,
                DeleteContext: resourceZenlayerBmcInstanceDelete,
                Schema: map[string]*schema.Schema{
                        "uuid": {
                                Type: schema.TypeString,
                                Computed: true,
                        },
                        "zone_uuid": {
                                Type: schema.TypeString,
                                Required: true,
                        },
                        "model_uuid": {
                                Type: schema.TypeString,
                                Required: true,
                        },
                        "os_uuid": {
                                Type: schema.TypeString,
                                Optional: true,
                        },
                        "hostname": {
                                Type: schema.TypeString,
                                Required: true,
                        },
                        "create_time": {
                                Type: schema.TypeString,
                                Computed: true,
                        },
                        "label_name": {
                                Type: schema.TypeString,
                                Optional: true,
                        },
                },
        }
}

func resourceZenlayerBmcInstanceCreate(ctx context.Context, d *schema.ResourceData, m interface{}) (diags diag.Diagnostics) {
        client := m.(*connectivity.ZenlayerClient)
        bmcClient, err := client.NewBmcClient()
        if err != nil {
                return diag.FromErr(err)
        }

        var labelName string
        zoneUuid := d.Get("zone_uuid").(string)
        modelUuid := d.Get("model_uuid").(string)
        hostname := d.Get("hostname").(string)

        if v, ok := d.GetOk("label_name"); ok {
                labelName = v.(string)
        }

        request := bmc.CreateCreateInstancePostpaidRequest(zoneUuid, modelUuid, bmc.Name{
                Hostname:  hostname,
                LabelName: labelName,
                Mark:      "",
        })

        if v, ok := d.GetOk("os_uuid"); ok {
                osUuid := v.(string)
                request.WithOSUuid(osUuid)
        }

        response, err := bmcClient.CreateInstancePostpaid(request)
        if err != nil {
                return diag.FromErr(err)
        }

        d.SetId(response.Uuid)
        err = d.Set("uuid", response.Uuid)
        if err != nil {
                return diag.FromErr(err)
        }

        return
}

func resourceZenlayerBmcInstanceRead(ctx context.Context, d *schema.ResourceData, m interface{}) (diags diag.Diagnostics) {
        client := m.(*connectivity.ZenlayerClient)
        bmcClient, err := client.NewBmcClient()
        if err != nil {
                return diag.FromErr(err)
        }

        instanceUuid := d.Id()
        request := bmc.CreateGetInstanceRequest(instanceUuid)
        response, err := bmcClient.GetInstance(request)
        if err != nil {
                return diag.FromErr(err)
        }

        instance := response.Data
        err = d.Set("create_time", instance.CreateTime)
        if err != nil {
               return diag.FromErr(err)
        }

        return
}

func resourceZenlayerBmcInstanceUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) (diags diag.Diagnostics) {
        return
}

func resourceZenlayerBmcInstanceDelete(ctx context.Context, d *schema.ResourceData, m interface{}) (diags diag.Diagnostics) {
        client := m.(*connectivity.ZenlayerClient)
        bmcClient, err := client.NewBmcClient()
        if err != nil {
                return diag.FromErr(err)
        }

        instanceUuid := d.Id()
        request := bmc.CreateDeleteInstanceRequest(instanceUuid)
        _, err = bmcClient.DeleteInstance(request)
        if err != nil {
                return diag.FromErr(err)
        }

        d.SetId("")
        return
}
