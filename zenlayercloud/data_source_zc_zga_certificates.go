/*
Use this data source to get all zga certificates.

Example Usage
```hcl
data "zenlayercloud_zga_certificates" "all" {
}
```
*/
package zenlayercloud

import (
	"context"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/common"
	"sort"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	zga "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/zga20230706"
)

func dataSourceZenlayerCloudZgaCertificates() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceZenlayerCloudZgaCertificatesRead,
		Schema: map[string]*schema.Schema{
			"certificate_ids": {
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Description: "IDs of the certificates to be queried.",
			},
			"certificate_label": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Label of the certificate to be queried.",
			},
			"dns_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "DNS Name of the certificate to be queried.",
			},
			"resource_group_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The ID of resource group that the certificate grouped by.",
			},
			"expired": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Whether the certificate has expired.",
			},
			"result_output_file": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Used to save results.",
			},
			"certificates": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "An information list of certificate. Each element contains the following attributes:",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"certificate_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of the certificate.",
						},
						"certificate_label": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Label of the certificate.",
						},
						"common": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Common of the certificate.",
						},
						"fingerprint": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Md5 fingerprint of the certificate.",
						},
						"issuer": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Issuer of the certificate.",
						},
						"dns_names": {
							Type:        schema.TypeSet,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Computed:    true,
							Description: "DNS Names of the certificate.",
						},
						"algorithm": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Algorithm of the certificate.",
						},
						"create_time": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Upload time of the certificate.",
						},
						"start_time": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Start time of the certificate.",
						},
						"end_time": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Expiration time of the certificate.",
						},
						"expired": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether the certificate has expired.",
						},
						"resource_group_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of resource group that the instance belongs to.",
						},
					},
				},
			},
		},
	}
}

func dataSourceZenlayerCloudZgaCertificatesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common.LogElapsed(ctx, "data_source.zenlayercloud_zga_certificates.read")()

	var cf CertificatesFilter
	if v, ok := d.GetOk("certificate_ids"); ok {
		certificateIds := v.(*schema.Set).List()
		if len(certificateIds) > 0 {
			cf.CertificateIds = common.ToStringList(certificateIds)
		}
	}
	if v, ok := d.GetOk("certificate_label"); ok {
		cf.CertificateLabel = v.(string)
	}
	if v, ok := d.GetOk("dns_name"); ok {
		cf.DnsName = v.(string)
	}
	if v, ok := d.GetOk("resource_group_id"); ok {
		cf.ResourceGroupId = v.(string)
	}
	if v, ok := d.GetOk("expired"); ok {
		expired := v.(bool)
		cf.Expired = &expired
	}

	var certs []*zga.CertificateInfo
	err := resource.RetryContext(ctx, common.ReadRetryTimeout, func() *resource.RetryError {
		var errRet error
		certs, errRet = NewZgaService(meta.(*connectivity.ZenlayerCloudClient)).
			DescribeCertificatesByFilter(ctx, &cf)
		if errRet != nil {
			return common.RetryError(ctx, errRet, common.InternalServerError, common.ReadTimedOut)
		}
		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	var (
		length   = len(certs)
		certList = make([]map[string]interface{}, 0, length)
		ids      = make([]string, 0, length)
	)
	for _, certificate := range certs {
		certList = append(certList, flattenCertificate(certificate))
		ids = append(ids, certificate.CertificateId)
	}

	sort.StringSlice(ids).Sort()

	d.SetId(common.DataResourceIdHash(ids))

	err = d.Set("certificates", certList)
	if err != nil {
		return diag.FromErr(err)
	}

	output, ok := d.GetOk("result_output_file")
	if ok && output.(string) != "" {
		if err := common.WriteToFile(output.(string), certList); err != nil {
			return diag.FromErr(err)
		}
	}
	return nil
}

func flattenCertificate(certificate *zga.CertificateInfo) map[string]interface{} {
	if certificate == nil {
		return nil
	}

	m := map[string]interface{}{
		"certificate_id":    certificate.CertificateId,
		"certificate_label": certificate.CertificateLabel,
		"common":            certificate.Common,
		"fingerprint":       certificate.Fingerprint,
		"issuer":            certificate.Issuer,
		"dns_names":         certificate.Sans,
		"algorithm":         certificate.Algorithm,
		"create_time":       certificate.CreateTime,
		"start_time":        certificate.StartTime,
		"end_time":          certificate.EndTime,
		"expired":           certificate.Expired,
		"resource_group_id": certificate.ResourceGroupId,
	}

	return m
}

type CertificatesFilter struct {
	CertificateIds   []string
	CertificateLabel string
	DnsName          string
	ResourceGroupId  string
	Expired          *bool
}
