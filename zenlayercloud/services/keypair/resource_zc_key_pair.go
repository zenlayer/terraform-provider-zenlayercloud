/*
Provides a resource to manage key pair.

~> **NOTE:** This request is to import an SSH key pair to be used for later instance login..

~> **NOTE:** A key pair name and several public SSH keys are required.

Example Usage

```hcl
resource "zenlayercloud_key_pair" "foo" {
  key_name       	= "my_key"
  public_key    	= "ssh-rsa XXXXXXXXXXXX key"
  key_description	= "create a key pair"
}
```

Import

Key pair can be imported, e.g.

```
$ terraform import zenlayercloud_key_pair.foo key-xxxxxxx
```
*/
package keypair

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	common2 "github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/common"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	ccs "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/ccs20250901"
	"github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/common"
	"time"
)

func ResourceZenlayerCloudKeyPair() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceZenlayerCloudKeyPairCreate,
		ReadContext:   resourceZenlayerCloudKeyPairRead,
		UpdateContext: resourceZenlayerCloudKeyPairUpdate,
		DeleteContext: resourceZenlayerCloudKeyPairDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"key_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(2, 32),
				Description:  "Key pair name. Up to 32 characters in length are supported, containing letters, digits and special character -_. The names cannot be duplicated.",
			},
			"public_key": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Public SSH keys in OpenSSH format. Up to 5 public keys are allowed, separated by pressing ENTER key.",
			},
			"key_description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 255),
				Description:  "Description of key pair. The length should be less than 256 characters.",
			},
		},
	}
}

func resourceZenlayerCloudKeyPairDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common2.LogElapsed(ctx, "resource.zenlayercloud_key_pair.delete")()

	ccsService := CcsService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}
	keyId := d.Id()

	// delete key pair
	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
		errRet := ccsService.DeleteKeyPair(keyId)
		if errRet != nil {
			return common2.RetryError(ctx, errRet)
		}
		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceZenlayerCloudKeyPairUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	ccsService := CcsService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}
	keyId := d.Id()
	if d.HasChange("key_description") {
		err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate)-time.Minute, func() *resource.RetryError {
			keyDesc := common.String(d.Get("key_description").(string))
			err := ccsService.ModifyKeyPair(ctx, keyId, keyDesc)
			if err != nil {
				return common2.RetryError(ctx, err, common2.InternalServerError, common.NetworkError)
			}
			return nil
		})

		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceZenlayerCloudKeyPairRead(ctx, d, meta)
}

func resourceZenlayerCloudKeyPairCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common2.LogElapsed(ctx, "resource.zenlayercloud_key_pair.create")()

	request := ccs.NewImportKeyPairRequest()
	request.KeyName = common.String(d.Get("key_name").(string))
	request.PublicKey = common.String(d.Get("public_key").(string))
	if v, ok := d.GetOk("key_description"); ok {
		request.KeyDescription = common.String(v.(string))
	}
	keyId := ""

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		response, err := meta.(*connectivity.ZenlayerCloudClient).WithCcsClient().ImportKeyPair(request)
		if err != nil {
			tflog.Info(ctx, "Fail to import key pair.", map[string]interface{}{
				"action":  request.GetAction(),
				"request": common2.ToJsonString(request),
				"err":     err.Error(),
			})
			return common2.RetryError(ctx, err)
		}

		tflog.Info(ctx, "Import key pair success", map[string]interface{}{
			"action":   request.GetAction(),
			"request":  common2.ToJsonString(request),
			"response": common2.ToJsonString(response),
		})

		if response.Response.KeyId == nil {
			err = fmt.Errorf("keyId is nil")
			return resource.NonRetryableError(err)
		}
		keyId = *response.Response.KeyId

		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(keyId)

	return resourceZenlayerCloudKeyPairRead(ctx, d, meta)
}

func resourceZenlayerCloudKeyPairRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common2.LogElapsed(ctx, "resource.zenlayer_key_pair.read")()

	var diags diag.Diagnostics

	keyId := d.Id()

	ccsService := CcsService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	var keyPair *ccs.KeyPair
	var errRet error

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead)-time.Minute, func() *resource.RetryError {
		keyPair, errRet = ccsService.DescribeKeyPairById(ctx, keyId)
		if errRet != nil {
			return common2.RetryError(ctx, errRet, common2.InternalServerError)
		}
		if keyPair == nil {
			return nil
		}
		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	if keyPair == nil {
		d.SetId("")
		tflog.Info(ctx, "key pair not exist", map[string]interface{}{
			"keyId": keyId,
		})
		return nil
	}

	// key pair info
	_ = d.Set("key_name", keyPair.KeyName)
	_ = d.Set("public_key", keyPair.PublicKey)
	_ = d.Set("key_description", keyPair.KeyDescription)

	return diags
}
