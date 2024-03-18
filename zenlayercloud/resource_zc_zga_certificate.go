/*
Provides a certificate resource.

~> **NOTE:** Modification of the certificate and key is not supported. If you want to change it, you need to create a new certificate.

~> **NOTE:** When the certificate and key are set to empty strings, the Update will not take effect.

Example Usage
```hcl

	resource "zenlayercloud_zga_certificate" "default" {
		certificate  = <<EOF

-----BEGIN CERTIFICATE-----
[......] # cert contents
-----END CERTIFICATE-----
EOF

	key = <<EOF

-----BEGIN RSA PRIVATE KEY-----
[......] # key contents
-----END RSA PRIVATE KEY-----
EOF

		label = "certificate"

		lifecycle {
			create_before_destroy = true
		}
	}

```
Import

Certificate can be imported using the id, e.g.

```
terraform import zenlayercloud_zga_certificate.default certificateId
```
*/
package zenlayercloud

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	zga "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/zga20230706"
)

func resourceZenlayerCloudCertificate() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceZenlayerCloudCertificateCreate,
		ReadContext:   resourceZenlayerCloudCertificateRead,
		UpdateContext: resourceZenlayerCloudCertificateUpdate,
		DeleteContext: resourceZenlayerCloudCertificateDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(zgaCreateTimeout),
		},
		Schema: map[string]*schema.Schema{
			"certificate": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				Description:      "The content of certificate.",
				DiffSuppressFunc: suppressEmptyString,
				StateFunc:        StateTrimSpace,
			},
			"key": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				Sensitive:        true,
				DiffSuppressFunc: suppressEmptyString,
				Description:      "The key of the certificate.",
				StateFunc:        StateTrimSpace,
			},
			"label": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The label of the certificate. Modification is not supported.",
			},
			"resource_group_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The resource group id the certificate belongs to, default to Default Resource Group. Modification is not supported.",
			},
			"common": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Common of the certificate.",
			},
			"create_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Uploaded time of the certificate.",
			},
			"end_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Expiration time of the certificate.",
			},
		},
	}
}

func resourceZenlayerCloudCertificateCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	certificateId := ""
	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		request := zga.NewCreateCertificateRequest()
		request.CertificateContent = d.Get("certificate").(string)
		request.CertificateKey = d.Get("key").(string)
		request.CertificateLabel = d.Get("label").(string)
		request.ResourceGroupId = d.Get("resource_group_id").(string)
		response, errRet := meta.(*connectivity.ZenlayerCloudClient).WithZgaClient().CreateCertificate(request)
		if errRet != nil {
			tflog.Error(ctx, "Fail to create certificate.", map[string]interface{}{
				"action":  request.GetAction(),
				"request": toJsonString(request),
				"err":     errRet.Error(),
			})
			return retryError(ctx, errRet, InternalServerError)
		}

		tflog.Info(ctx, "Create certificate success", map[string]interface{}{
			"action":   request.GetAction(),
			"request":  toJsonString(request),
			"response": toJsonString(response),
		})

		certificateId = response.Response.CertificateId
		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(certificateId)

	return resourceZenlayerCloudCertificateRead(ctx, d, meta)
}

func resourceZenlayerCloudCertificateDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	certificateId := d.Id()
	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
		errRet := NewZgaService(meta.(*connectivity.ZenlayerCloudClient)).DeleteCertificatesById(ctx, certificateId)
		if errRet != nil {
			switch {
			case isExpectError(errRet, []string{"INVALID_CERTIFICATE_NOT_FOUND"}):
				// DO NOTHING
			case isExpectError(errRet, []string{"CERTIFICATE_IS_USING"}):
				return resource.NonRetryableError(fmt.Errorf("certificate %s still in used", certificateId))
			default:
				return retryError(ctx, errRet, InternalServerError)
			}
		}
		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceZenlayerCloudCertificateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// TODO: support to update label and resource_group_id
	return resourceZenlayerCloudCertificateRead(ctx, d, meta)
}

func resourceZenlayerCloudCertificateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var (
		diags         diag.Diagnostics
		certificateId = d.Id()
		certInfo      *zga.CertificateInfo
	)
	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead)-time.Minute, func() *resource.RetryError {
		var errRet error
		certInfo, errRet = NewZgaService(meta.(*connectivity.ZenlayerCloudClient)).DescribeCertificateById(ctx, certificateId)
		if errRet != nil {
			return retryError(ctx, errRet, InternalServerError)
		}
		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	if certInfo == nil {
		d.SetId("")
		tflog.Info(ctx, "certificate not exist or created failed or recycled", map[string]interface{}{
			"certificateId": certificateId,
		})
		return nil
	}

	_ = d.Set("common", certInfo.Common)
	_ = d.Set("create_time", certInfo.CreateTime)
	_ = d.Set("end_time", certInfo.EndTime)
	_ = d.Set("resource_group_id", certInfo.ResourceGroupId)
	_ = d.Set("label", certInfo.CertificateLabel)

	return diags
}

func StateTrimSpace(v interface{}) string {
	s, ok := v.(string)

	if !ok {
		return ""
	}

	return strings.TrimSpace(s)
}

func suppressEmptyString(_, oldValue, newValue string, d *schema.ResourceData) bool {
	if !d.IsNewResource() && newValue == "" && oldValue != "" {
		return true
	}
	return false
}
