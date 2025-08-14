/*
Use this data source to get all zga accelerator.

Example Usage
```hcl
data "zenlayercloud_zga_accelerators" "all" {
}
```
*/
package zenlayercloud

import (
	"context"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"

	zga "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/zga20230706"
)

func dataSourceZenlayerCloudZgaAccelerators() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceZenlayerCloudZgaAcceleratorsRead,
		Schema: map[string]*schema.Schema{
			"accelerator_ids": {
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Description: "IDs of the accelerator to be queried.",
			},
			"accelerator_name": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 64),
				Description:  "The name of accelerator. The max length of accelerator name is 64.",
			},
			"accelerator_status": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					ZgaAcceleratorStatusAccelerating,
					ZgaAcceleratorStatusNotAccelerate,
					ZgaAcceleratorStatusDeploying,
					ZgaAcceleratorStatusStopAccelerate,
					ZgaAcceleratorStatusAccelerateFailure,
				}, false),
				Description: "Status of the accelerator to be queried. Valid values are `Accelerating`, `NotAccelerate`, `Deploying`, `StopAccelerate`, `AccelerateFailure`.",
			},
			"accelerate_region_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Accelerate region of the accelerator to be queried.",
			},
			"vip": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.IsIPAddress,
				Description:  "Virtual IP of the accelerator to be queried.",
			},
			"domain": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Domain of the accelerator to be queried.",
			},
			"origin": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Origin of the accelerator to be queried.",
			},
			"origin_region_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Origin region of the accelerator to be queried.",
			},
			"cname": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Cname of the accelerator to be queried.",
			},
			"resource_group_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The ID of resource group that the accelerator grouped by.",
			},
			"result_output_file": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Used to save results.",
			},
			"accelerators": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "An information list of accelerator. Each element contains the following attributes:",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"accelerator_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of the accelerator.",
						},
						"accelerator_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Type of the accelerator.",
						},
						"accelerator_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Name of the accelerator.",
						},
						"charge_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The charge type of the accelerator.",
						},
						"domain": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Main domain of the accelerator.",
						},
						"relate_domains": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "Relate domains of the accelerator.",
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
						"accelerator_status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Status of the accelerator.",
						},
						"cname": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Cname of the accelerator.",
						},
						"origin_region_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of the orgin region.",
						},
						"origin_region_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Name of the orgin region.",
						},
						"origin": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "Endpoints of the origin.",
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
						"backup_origin": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "Backup endpoint of the origin.",
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
						"accelerate_regions": {
							Type:        schema.TypeSet,
							Description: "Accelerate region of the accelerator.",
							Computed:    true,
							Set:         AccelerateRegionsSchemaSetFuncfunc,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"accelerate_region_id": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "ID of the accelerate region.",
									},
									"accelerate_region_name": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Name of the accelerate region.",
									},
									"accelerate_region_status": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Configuration status of the accelerate region.",
									},
									"vip": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Virtual IP the accelerate region.",
									},
									"bandwidth": {
										Type:        schema.TypeInt,
										Computed:    true,
										Description: "Virtual IP the accelerate region.",
										// Description: "Bandwidth limit of the accelerate region. Unit: Mbps.",
									},
								},
							},
						},
						"l4_listeners": {
							Type:        schema.TypeList,
							Description: "L4 listeners of the accelerator.",
							Computed:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"port": {
										Type:        schema.TypeInt,
										Computed:    true,
										Description: "The port of the l4 listener.",
									},
									"back_port": {
										Type:        schema.TypeInt,
										Computed:    true,
										Description: "The Return-to-origin port of the l4 listener.",
									},
									"port_range": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The port range of the l4 listener.",
									},
									"back_port_range": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The Return-to-origin port range of the l4 listener.",
									},
									"protocol": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The protocol of the l4 listener.",
									},
								},
							},
						},
						"l7_listeners": {
							Type:        schema.TypeList,
							Description: "L7 listeners of the accelerator.",
							Computed:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"port": {
										Type:        schema.TypeInt,
										Computed:    true,
										Description: "The port of the l7 listener.",
									},
									"back_port": {
										Type:        schema.TypeInt,
										Computed:    true,
										Description: "The Return-to-origin port of the l7 listener.",
									},
									"port_range": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The port range of the l7 listener.",
									},
									"back_port_range": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The Return-to-origin port range of the l7 listener.",
									},
									"protocol": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The protocol of the l7 listener.",
									},
									"back_protocol": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The Return-to-origin protocol of the l7 listener.",
									},
									"host": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The Return-to-origin host of the l7 listener.",
									},
								},
							},
						},
						"protocol_opts": {
							Computed:    true,
							Type:        schema.TypeList,
							Description: "Protocol opts of the accelerator.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"toa": {
										Type:        schema.TypeBool,
										Computed:    true,
										Description: "Whether to enable TOA.",
									},
									"toa_value": {
										Type:        schema.TypeInt,
										Computed:    true,
										Description: "TOA verison.",
									},
									"websocket": {
										Type:        schema.TypeBool,
										Computed:    true,
										Description: "Whether to enable websocket.",
									},
									"proxy_protocol": {
										Type:        schema.TypeBool,
										Computed:    true,
										Description: "Whether to enable proxyProtocol.",
									},
									"gzip": {
										Type:        schema.TypeBool,
										Computed:    true,
										Description: "Whether to enable gzip.",
									},
								},
							},
						},
						"certificate": {
							Computed:    true,
							Type:        schema.TypeList,
							Description: "Certificate info of the accelerator.",
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
						"access_control": {
							Computed:    true,
							Type:        schema.TypeList,
							Description: "Access control of the accelerator.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"enable": {
										Type:        schema.TypeBool,
										Computed:    true,
										Description: "Whether to enable access control.",
									},
									"rules": {
										Type:        schema.TypeList,
										Description: "Rules of the access control.",
										Computed:    true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"listener": {
													Type:        schema.TypeString,
													Computed:    true,
													Description: "The listener of the rule.",
												},
												"directory": {
													Type:        schema.TypeString,
													Computed:    true,
													Description: "The directory of the rule.",
												},
												"cidr_ip": {
													Type:        schema.TypeList,
													Computed:    true,
													Description: "The cidr ip of the rule.",
													Elem:        &schema.Schema{Type: schema.TypeString},
												},
												"policy": {
													Type:        schema.TypeString,
													Computed:    true,
													Description: "The policy of the rule.",
												},
												"note": {
													Type:        schema.TypeString,
													Computed:    true,
													Description: "The note of the rule.",
												},
											},
										},
									},
								},
							},
						},
						"health_check": {
							Computed:    true,
							Type:        schema.TypeList,
							Description: "Health check of the accelerator.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"enable": {
										Type:        schema.TypeBool,
										Computed:    true,
										Description: "Whether to enable health check.",
									},
									"alarm": {
										Type:        schema.TypeBool,
										Computed:    true,
										Description: "Whether to enable alarm.",
									},
									"port": {
										Type:        schema.TypeInt,
										Computed:    true,
										Description: "The port of health check.",
									},
								},
							},
						},
						"create_time": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Create time of the accelerator.",
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

func dataSourceZenlayerCloudZgaAcceleratorsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer logElapsed(ctx, "data_source.zenlayercloud_zga_accelerators.read")()

	var af AcceleratorsFilter
	if v, ok := d.GetOk("accelerator_ids"); ok {
		acceleratorIds := v.(*schema.Set).List()
		if len(acceleratorIds) > 0 {
			af.AcceleratorIds = toStringList(acceleratorIds)
		}
	}
	if v, ok := d.GetOk("accelerator_name"); ok {
		af.AcceleratorName = v.(string)
	}
	if v, ok := d.GetOk("accelerator_status"); ok {
		af.AcceleratorStatus = v.(string)
	}
	if v, ok := d.GetOk("accelerate_region_id"); ok {
		af.AccelerateRegionId = v.(string)
	}
	if v, ok := d.GetOk("vip"); ok {
		af.Vip = v.(string)
	}
	if v, ok := d.GetOk("domain"); ok {
		af.Domain = v.(string)
	}
	if v, ok := d.GetOk("origin"); ok {
		af.Origin = v.(string)
	}
	if v, ok := d.GetOk("origin_region_id"); ok {
		af.OriginRegionId = v.(string)
	}
	if v, ok := d.GetOk("cname"); ok {
		af.Cname = v.(string)
	}
	if v, ok := d.GetOk("resource_group_id"); ok {
		af.ResourceGroupId = v.(string)
	}

	var accelerators []*zga.AcceleratorInfo
	err := resource.RetryContext(ctx, readRetryTimeout, func() *resource.RetryError {
		var errRet error
		accelerators, errRet = NewZgaService(meta.(*connectivity.ZenlayerCloudClient)).DescribeAcceleratorsByFilter(ctx, &af)
		if errRet != nil {
			return retryError(ctx, errRet, InternalServerError, ReadTimedOut)
		}
		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	var (
		length          = len(accelerators)
		acceleratorList = make([]map[string]interface{}, 0, length)
		ids             = make([]string, 0, length)
	)
	for _, accelerator := range accelerators {
		acceleratorList = append(acceleratorList, flattenAccelerator(accelerator))
		ids = append(ids, accelerator.AcceleratorId)
	}

	sort.StringSlice(ids).Sort()

	d.SetId(dataResourceIdHash(ids))

	err = d.Set("accelerators", acceleratorList)
	if err != nil {
		return diag.FromErr(err)
	}

	output, ok := d.GetOk("result_output_file")
	if ok && output.(string) != "" {
		if err := writeToFile(output.(string), acceleratorList); err != nil {
			return diag.FromErr(err)
		}
	}
	return nil
}

func flattenAccelerateRegions(acceelrateRegions []*zga.AccelerateRegionInfo) []map[string]interface{} {
	if acceelrateRegions == nil {
		return nil
	}
	var result = make([]map[string]interface{}, 0, len(acceelrateRegions))
	for _, region := range acceelrateRegions {
		result = append(result, map[string]interface{}{
			"accelerate_region_id":     region.AccelerateRegionId,
			"accelerate_region_name":   region.AccelerateRegionName,
			"accelerate_region_status": region.AccelerateRegionStatus,
			"vip":                      region.Vip,
			"bandwidth":                region.Bandwidth,
		})
	}
	return result
}

func flattenL4Listeners(listeners []*zga.AccelerationRuleL4Listener) []map[string]interface{} {
	if listeners == nil {
		return nil
	}
	var result = make([]map[string]interface{}, 0, len(listeners))
	for _, listener := range listeners {
		m := map[string]interface{}{
			"protocol": listener.Protocol,
		}
		if listener.Port != 0 {
			m["port"] = listener.Port
		}
		if listener.BackPort != 0 {
			m["back_port"] = listener.BackPort
		}
		if listener.PortRange != "" {
			m["port_range"] = listener.PortRange
		}
		if listener.BackPortRange != "" {
			m["back_port_range"] = listener.BackPortRange
		}
		result = append(result, m)
	}
	return result
}

func flattenL7Listeners(listeners []*zga.AccelerationRuleL7Listener) []map[string]interface{} {
	if listeners == nil {
		return nil
	}
	var result = make([]map[string]interface{}, 0, len(listeners))
	for _, listener := range listeners {
		m := map[string]interface{}{
			"protocol":      listener.Protocol,
			"back_protocol": listener.BackProtocol,
		}
		if listener.Port != 0 {
			m["port"] = listener.Port
		}
		if listener.BackPort != 0 {
			m["back_port"] = listener.BackPort
		}
		if listener.PortRange != "" {
			m["port_range"] = listener.PortRange
		}
		if listener.BackPortRange != "" {
			m["back_port_range"] = listener.BackPortRange
		}
		if listener.Host != "" {
			m["host"] = listener.Host
		}
		result = append(result, m)
	}
	return result
}

func flattenProtocolOpts(protocolOpts *zga.AccelerationRuleProtocolOpts) []interface{} {
	if protocolOpts == nil {
		return nil
	}
	m := map[string]interface{}{
		"toa_value": protocolOpts.ToaValue,
	}
	if protocolOpts.Toa != nil {
		m["toa"] = *protocolOpts.Toa
	}
	if protocolOpts.Websocket != nil {
		m["websocket"] = *protocolOpts.Websocket
	}
	if protocolOpts.ProxyProtocol != nil {
		m["proxy_protocol"] = *protocolOpts.ProxyProtocol
	}
	if protocolOpts.Gzip != nil {
		m["gzip"] = *protocolOpts.Gzip
	}
	return []interface{}{m}
}

func flattenAccessControl(accessControl *zga.AccessControl) []interface{} {
	if accessControl == nil {
		return nil
	}

	m := map[string]interface{}{
		"enable": accessControl.Enable,
	}

	rules := make([]map[string]interface{}, 0, len(accessControl.Rules))
	for _, rule := range accessControl.Rules {
		if rule == nil {
			continue
		}
		rules = append(rules, map[string]interface{}{
			"listener":  rule.Listener,
			"directory": rule.Directory,
			"cidr_ip":   splitStringByCommaOrSemicolon(rule.CidrIp),
			"policy":    rule.Policy,
			"note":      rule.Note,
		})
	}
	m["rules"] = rules

	return []interface{}{m}
}

func flattenHealthCheck(healthCheck *zga.HealthCheck) []interface{} {
	if healthCheck == nil {
		return nil
	}
	m := map[string]interface{}{
		"enable": healthCheck.Enable,
		"alarm":  healthCheck.Alarm,
		"port":   healthCheck.Port,
	}
	return []interface{}{m}
}

func flattenAccelerator(accelerator *zga.AcceleratorInfo) map[string]interface{} {
	if accelerator == nil {
		return nil
	}

	m := map[string]interface{}{
		"accelerator_id":     accelerator.AcceleratorId,
		"accelerator_type":   accelerator.AcceleratorType,
		"accelerator_name":   accelerator.AcceleratorName,
		"charge_type":        accelerator.ChargeType,
		"accelerator_status": accelerator.AcceleratorStatus,
		"cname":              accelerator.Cname,
		"create_time":        accelerator.CreateTime,
		"resource_group_id":  accelerator.ResourceGroupId,
		"origin_region_id":   accelerator.Origin.OriginRegionId,
		"origin_region_name": accelerator.Origin.OriginRegionName,
		"origin":             splitStringByCommaOrSemicolon(accelerator.Origin.Origin),
		"backup_origin":      splitStringByCommaOrSemicolon(accelerator.Origin.BackupOrigin),
		"accelerate_regions": flattenAccelerateRegions(accelerator.AccelerateRegions),
		"l4_listeners":       flattenL4Listeners(accelerator.L4Listeners),
		"l7_listeners":       flattenL7Listeners(accelerator.L7Listeners),
		"protocol_opts":      flattenProtocolOpts(accelerator.ProtocolOpts),
		"access_control":     flattenAccessControl(accelerator.AccessControl),
		"health_check":       flattenHealthCheck(accelerator.HealthCheck),
	}

	if accelerator.Domain != nil {
		m["domain"] = accelerator.Domain.Domain
		m["relate_domains"] = splitStringByCommaOrSemicolon(accelerator.Domain.RelateDomains)
	}
	if certInfo := flattenCertificate(accelerator.Certificate); certInfo != nil {
		m["certificate"] = []interface{}{certInfo}
	}
	return m
}

type AcceleratorsFilter struct {
	AcceleratorIds     []string
	AcceleratorName    string
	AcceleratorStatus  string
	AccelerateRegionId string
	Vip                string
	Domain             string
	Origin             string
	OriginRegionId     string
	Cname              string
	ResourceGroupId    string
}

func splitStringByCommaOrSemicolon(value string) (result []string) {
	if value == "" {
		return
	}
	comma := ","
	if strings.Contains(value, comma) {
		result = strings.Split(value, comma)
	} else {
		result = strings.Split(value, ";")
	}
	return
}

func joinArrayByComma(values []string) string {
	return strings.Join(values, ",")
}

func AccelerateRegionsSchemaSetFuncfunc(val interface{}) int {
	m := val.(map[string]interface{})
	regionID := m["accelerate_region_id"].(string)
	return schema.HashString(regionID)
}
