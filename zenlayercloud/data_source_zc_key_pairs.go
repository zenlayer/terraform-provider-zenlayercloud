/*
Use this data source to query SSH key pair list.

Example Usage

```hcl
data "zenlayercloud_key_pairs" "all" {
}

data "zenlayercloud_key_pairs" "myname" {
	key_name = "myname"
}
```
*/
package zenlayercloud

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	vm "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/vm20230313"
)

func dataSourceZenlayerCloudKeyPairs() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceZenlayerCloudKeyPairsRead,

		Schema: map[string]*schema.Schema{
			"key_ids": {
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Description: "IDs of the key pair to be queried.",
			},
			"key_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Name of the key pair to be queried.",
			},
			"result_output_file": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Used to save results.",
			},
			// Computed value
			"key_pairs": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "An information list of key pairs. Each element contains the following attributes:",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of the key pair, such as `key-xxxxxxxx`.",
						},
						"key_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Name of the key pair.",
						},
						"public_key": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Public SSH keys in OpenSSH format, such as `ssh-rsa XXXXXXXXXXXX`.",
						},
						"key_description": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Description of the key pair.",
						},
						"create_time": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Create time of the key pair.",
						},
					},
				},
			},
		},
	}
}

func dataSourceZenlayerCloudKeyPairsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer logElapsed(ctx, "data_source.zenlayercloud_key_pairs.read")()

	vmService := VmService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	var errRet error

	request := vm.NewDescribeKeyPairsRequest()
	if v, ok := d.GetOk("key_name"); ok {
		if v != "" {
			request.KeyName = v.(string)
		}
	}
	if v, ok := d.GetOk("key_ids"); ok {
		keyIds := v.(*schema.Set).List()
		if len(keyIds) > 0 {
			request.KeyIds = toStringList(keyIds)
		}
	}

	var keyPairs []*vm.KeyPair
	err := resource.RetryContext(ctx, readRetryTimeout, func() *resource.RetryError {
		keyPairs, errRet = vmService.DescribeKeyPairs(ctx, request)
		if errRet != nil {
			return retryError(ctx, errRet, InternalServerError, ReadTimedOut)
		}
		return nil
	})

	keyPairList := make([]map[string]interface{}, 0, len(keyPairs))
	ids := make([]string, 0, len(keyPairs))
	for _, keyPair := range keyPairs {
		mapping := map[string]interface{}{
			"key_id":          keyPair.KeyId,
			"key_name":        keyPair.KeyName,
			"public_key":      keyPair.PublicKey,
			"key_description": keyPair.KeyDescription,
			"create_time":     keyPair.CreateTime,
		}
		keyPairList = append(keyPairList, mapping)
		ids = append(ids, keyPair.KeyId)
	}

	d.SetId(dataResourceIdHash(ids))
	err = d.Set("key_pairs", keyPairList)
	if err != nil {
		return diag.FromErr(err)
	}

	output, ok := d.GetOk("result_output_file")
	if ok && output.(string) != "" {
		if err := writeToFile(output.(string), keyPairList); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}
