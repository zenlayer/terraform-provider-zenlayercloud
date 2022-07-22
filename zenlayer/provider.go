package zenlayer

import (
        "context"
        "fmt"
        "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
        "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
        "github.com/zenlayer/terraform-provider-zenlayer/zenlayer/connectivity"
        "os"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"access_key_id": {
				Type: schema.TypeString,
				Optional: true,
				DefaultFunc: schema.EnvDefaultFunc("ZENLAYER_ACCESS_KEY_ID", os.Getenv("ZENLAYER_ACCESS_KEY_ID")),
				Description: "Access Key Id",
			},
			"access_key_password": {
				Type: schema.TypeString,
				Optional: true,
				DefaultFunc: schema.EnvDefaultFunc("ZENLAYER_ACCESS_KEY_PASSWORD", os.Getenv("ZENLAYER_ACCESS_KEY_PASSWORD")),
                                Description: "Access Key Password",
			},
		},
		DataSourcesMap: map[string]*schema.Resource{
			"zenlayer_zones": dataSourceZenlayerZones(),
                        "zenlayer_models": dataSourceZenlayerModels(),
                        "zenlayer_oss": dataSourceZenlayerOSs(),
		},
		ResourcesMap: map[string]*schema.Resource{
			"zenlayer_bmc_instance": resourceZenlayerBmcInstance(),
		},
                ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (zenlayerClient interface{}, diags diag.Diagnostics) {
        accessKeyId := d.Get("access_key_id").(string)
        accessKeyPassword := d.Get("access_key_password").(string)

        if accessKeyId == "" {
                err := fmt.Errorf("Empty zenlayer access key id")
                diags = diag.FromErr(err)
                return
        }

        if accessKeyPassword == "" {
                err := fmt.Errorf("Empty zenlayer access key password")
                diags = diag.FromErr(err)
                return
        }

        if (accessKeyId != "") && (accessKeyPassword != "") {
                zenlayerClient = &connectivity.ZenlayerClient{
                        AccessKeyId:       accessKeyId,
                        AccessKeyPassword: accessKeyPassword,
                }
        }

        return
}
