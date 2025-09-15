/*
Provides a accelerator resource.

~> **NOTE:** Only L4 listener can be configured when domain is null.

~> **NOTE:** The Domain is not allowed to be the same as origin, otherwise a loop will be formed, making acceleration unusable.

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

		lifecycle {
			create_before_destroy = true
		}
	}

	resource "zenlayercloud_zga_accelerator" "default" {
	  accelerator_name = "accelerator_test"
	  charge_type = "ByTrafficPackage"
	  domain = "test.com"
	  relate_domains = ["a.test.com"]
	  origin_region_id = "DE"
	  origin = ["10.10.10.10"]
	  backup_origin = ["10.10.10.14"]
	  certificate_id = resource.zenlayercloud_zga_certificate.default.id
	  accelerate_regions {
	    accelerate_region_id = "KR"
	  }
	  accelerate_regions {
	    accelerate_region_id = "US"
	  }
	  l4_listeners {
	    protocol = "udp"
	    port_range = "53/54"
	    back_port_range = "53/54"
	  }
	  l4_listeners {
	    port = 80
	    back_port = 80
	    protocol = "tcp"
	  }
	  l7_listeners {
	    port = 443
	    back_port = 80
	    protocol = "https"
	    back_protocol = "http"
	  }
	  l7_listeners {
	    port_range = "8888/8890"
	    back_port_range = "8888/8890"
	    protocol = "http"
	    back_protocol = "http"
	  }
	  protocol_opts {
	    websocket = true
	    gzip = false
	  }
	  access_control {
	    enable = true
	    rules {
	      listener = "https:443"
	      directory = "/"
	      policy = "deny"
	      cidr_ip = ["10.10.10.10"]
	    }
	    rules {
	      listener = "udp:53/54"
	      directory = "/"
	      policy = "accept"
	      cidr_ip = ["10.10.10.11/8"]
	    }
	  }
	}

```
Import

Accelerator can be imported using the id, e.g.

```
terraform import zenlayercloud_zga_accelerator.default acceleratorId
```
*/
package zenlayercloud

import (
	"context"
	"errors"
	"fmt"
	common2 "github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/common"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	"github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/common"
	zga "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/zga20230706"
)

func resourceZenlayerCloudAccelerator() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceZenlayerCloudAcceleratorCreate,
		ReadContext:   resourceZenlayerCloudAcceleratorRead,
		UpdateContext: resourceZenlayerCloudAcceleratorUpdate,
		DeleteContext: resourceZenlayerCloudAcceleratorDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(common2.ZgaCreateTimeout),
			Update: schema.DefaultTimeout(common2.ZgaUpdateTimeout),
		},
		CustomizeDiff: customdiff.All(
			IPAcceleratorValidFunc(),
			DomainAcceleratorValidFunc(),
			PortCheckValidFunc,
			ProtocolOptsCheckValidFunc,
			AccessControlValidFunc,
		),
		Schema: map[string]*schema.Schema{
			"accelerator_name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringLenBetween(0, 64),
				Description:  "The name of accelerator. The max length of accelerator name is 64.",
			},
			"charge_type": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.StringInSlice([]string{
					ZgaAcceleratorChargeTypeBandwidth,
					ZgaAcceleratorChargeTypeTrafficPackage,
					ZgaAcceleratorChargeTypeBandwidth95,
					ZgaAcceleratorChargeTypeTraffic,
				}, false),
				Description: "The charge type of the accelerator. The default charge type of the account will be used. Modification is not supported. Valid values are `ByTrafficPackage`, `ByBandwidth95`, `ByBandwidth`, `ByTraffic`.",
			},
			"accelerator_status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Status of the accelerator. Values are `Accelerating`, `NotAccelerate`, `Deploying`, `StopAccelerate`, `AccelerateFailure`.",
			},
			"resource_group_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The resource group id the accelerator belongs to, default to Default Resource Group. Modification is not supported.",
			},
			"certificate_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The certificate of the accelerator. Required when exist https protocol accelerate.",
			},
			"cname": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Cname of the accelerator.",
			},
			"domain": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Main domain of the accelerator. Required when L7 http or https accelerate, globally unique and no duplication is allowed. Supports generic domain names, like: *.zenlayer.com.",
			},
			"relate_domains": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "Relate domains of the accelerator. Globally unique and no duplication is allowed. The max length of relate domains is 10.",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"origin_region_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of the orgin region. Modification is not supported.",
			},
			"origin": {
				Type:        schema.TypeSet,
				Required:    true,
				Description: "Endpoints of the origin. Only one endpoint is allowed to be configured, when the endpoint is CNAME.",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"backup_origin": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "Backup endpoint of the origin. Backup orgin only be configured when origin configured with IP. Only one back endpoint is allowed to be configured, when the back endpoint is CNAME.",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"accelerate_regions": {
				Type:        schema.TypeSet,
				Description: "Accelerate region of the accelerator.",
				Required:    true,
				Set:         AccelerateRegionsSchemaSetFuncfunc,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"accelerate_region_id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "ID of the accelerate region.",
						},
						"bandwidth": {
							Type:        schema.TypeInt,
							Optional:    true,
							Computed:    true,
							Description: "Bandwidth limit of the accelerate region. Exceeding the account speed limit is not allowed. Unit: Mbps.",
						},
						"vip": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.IsIPAddress,
							Description:  "Virtual IP the accelerate region. Modification is not supported.",
						},
					},
				},
			},
			"l4_listeners": {
				Type:        schema.TypeSet,
				Description: "L4 listeners of the accelerator.",
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"port": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IsPortNumber,
							Description:  "The port of the l4 listener. Only port or portRange can be configured, and duplicate ports are not allowed.",
						},
						"back_port": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IsPortNumber,
							Description:  "The Return-to-origin port of the l4 listener.",
						},
						"port_range": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: IsPortRange,
							Description:  "The port range of the l4 listener. Only port or portRange can be configured. Use a slash (/) to separate the starting and ending ports, like: 1/200. The max range: 300.",
						},
						"back_port_range": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: IsPortRange,
							Description:  "The Return-to-origin port range of the l4 listener. Use a slash (/) to separate the starting and ending ports, like: 1/200.",
						},
						"protocol": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{ZgaTCPL4Protocol, ZgaUDPL4Protocol}, false),
							Description:  "The protocol of the l4 listener. Valid values: `tcp`, `udp`.",
						},
					},
				},
			},
			"l7_listeners": {
				Type:        schema.TypeSet,
				Description: "L7 listeners of the accelerator.",
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"port": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IsPortNumberOrZero,
							Description:  "The port of the l7 listener. Only port or portRange can be configured, and duplicate ports are not allowed.",
						},
						"back_port": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IsPortNumberOrZero,
							Description:  "The Return-to-origin port of the l7 listener.",
						},
						"port_range": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: IsPortRange,
							Description:  "The port range of the l7 listener. Only port or portRange can be configured. Use a slash (/) to separate the starting and ending ports, like: 1/200. The max range: 300.",
						},
						"back_port_range": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: IsPortRange,
							Description:  "The Return-to-origin port range of the l7 listener. Use a slash (/) to separate the starting and ending ports, like: 1/200.",
						},
						"protocol": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{ZgaHTTPL7Protocol, ZgaHTTPSL7Protocol}, false),
							Description:  "The protocol of the l4 listener. Valid values: `http`, `https`.",
						},
						"back_protocol": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{ZgaHTTPL7Protocol, ZgaHTTPSL7Protocol}, false),
							Description:  "The Return-to-origin protocol of the l7 listener. Valid values: http and https. The default is equal to protocol.",
						},
						"host": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The Return-to-origin host of the l7 listener.",
						},
					},
				},
			},
			"protocol_opts": {
				Optional:    true,
				Computed:    true,
				Type:        schema.TypeList,
				MaxItems:    1,
				Description: "Protocol opts of the accelerator.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"toa": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Whether to enable TOA. Default is `false`.",
						},
						"toa_value": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     253,
							Description: "TOA verison. Default is `253`.",
						},
						"websocket": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Whether to enable websocket. Default is `false`.",
						},
						"proxy_protocol": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Whether to enable proxyProtocol. Default is `false`.",
						},
						"gzip": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Whether to enable gzip. Default is `false`.",
						},
					},
				},
			},
			"health_check": {
				Optional:    true,
				Computed:    true,
				Type:        schema.TypeList,
				MaxItems:    1,
				Description: "Health check of the accelerator.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enable": {
							Type:        schema.TypeBool,
							Required:    true,
							Description: "Whether to enable health check. If the enable is `false`, the alarm will be set to `false` and the port will be cleared.",
						},
						"alarm": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Whether to enable alarm. Default is `false`.",
						},
						"port": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      0,
							ValidateFunc: validation.IsPortNumberOrZero,
							Description:  "The port of health check.",
						},
					},
				},
			},
			"access_control": {
				Optional:    true,
				Computed:    true,
				Type:        schema.TypeList,
				MaxItems:    1,
				Description: "Access control of the accelerator.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enable": {
							Type:        schema.TypeBool,
							Required:    true,
							Description: "Whether to enable access control. Default is `true`.",
						},
						"rules": {
							Type:        schema.TypeSet,
							Optional:    true,
							Description: "Rules of the access control.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"listener": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: IsAcListener,
										Description:  "The listener of the rule. Valid values are `$protocol:$port`, `$protocol:$portRange`, `all`.",
									},
									"directory": {
										Type:        schema.TypeString,
										Optional:    true,
										Default:     "/",
										Description: "The directory of the rule. Not configurable with L4 listener. Default is `/`. Wildcards supported: *.",
									},
									"cidr_ip": {
										Type:        schema.TypeSet,
										Required:    true,
										Description: "The cidr ip of the rule.",
										Elem: &schema.Schema{
											Type: schema.TypeString,
											ValidateFunc: func(i interface{}, s string) ([]string, []error) {
												warnings, err := validation.IsIPAddress(i, s)
												if len(err) == 0 {
													return warnings, err
												}
												return validation.IsCIDR(i, s)
											},
										},
									},
									"policy": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice([]string{"accept", "deny"}, false),
										Description:  "The policy of the rule. Valid values are `accept`, `deny`.",
									},
									"note": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "The note of the rule.",
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func SharpenDomain(d *schema.ResourceData) *zga.Domain {
	domain, ok := d.Get("domain").(string)
	if !ok {
		return nil
	}
	result := zga.Domain{
		Domain: domain,
	}
	relateDomainsSet, ok := d.Get("relate_domains").(*schema.Set)
	if !ok {
		return &result
	}
	result.RelateDomains = InterfaceSliceToString(relateDomainsSet.List())
	return &result
}

func SharpenOrigin(d *schema.ResourceData) zga.Origin {
	var result = zga.Origin{
		OriginRegionId: d.Get("origin_region_id").(string),
		Origin:         InterfaceSliceToString(d.Get("origin").(*schema.Set).List()),
	}
	backOriginsSet, ok := d.Get("backup_origin").(*schema.Set)
	if !ok {
		return result
	}
	result.BackupOrigin = InterfaceSliceToString(backOriginsSet.List())
	return result
}

func SharpenAccelerateRegion(d *schema.ResourceData) []zga.AccelerateRegion {
	var (
		accRegionsV = d.Get("accelerate_regions").(*schema.Set).List()
		result      = make([]zga.AccelerateRegion, 0, len(accRegionsV))
	)
	for _, accRegionV := range accRegionsV {
		accRegion, ok := accRegionV.(map[string]interface{})
		if !ok {
			continue
		}
		region := zga.AccelerateRegion{
			AccelerateRegionId: accRegion["accelerate_region_id"].(string),
		}
		vip, ok := accRegion["vip"].(string)
		if ok {
			region.Vip = vip
		}
		bandwidth, ok := accRegion["bandwidth"].(int)
		if ok {
			region.Bandwidth = bandwidth
		}
		result = append(result, region)
	}
	return result
}

func SharpenL4Listeners(d *schema.ResourceData) []*zga.AccelerationRuleL4Listener {
	l4ListenersSet, ok := d.Get("l4_listeners").(*schema.Set)
	if !ok {
		return nil
	}
	var (
		l4ListenersV = l4ListenersSet.List()
		result       = make([]*zga.AccelerationRuleL4Listener, 0, len(l4ListenersV))
	)
	for _, l4ListenerV := range l4ListenersV {
		l4Listener, ok := l4ListenerV.(map[string]interface{})
		if !ok {
			continue
		}
		listener := zga.AccelerationRuleL4Listener{
			Protocol: l4Listener["protocol"].(string),
		}
		port, ok := l4Listener["port"].(int)
		if ok {
			listener.Port = port
			backPort, ok := l4Listener["back_port"].(int)
			if ok {
				listener.BackPort = backPort
			}
		}
		portRange, ok := l4Listener["port_range"].(string)
		if ok {
			listener.PortRange = portRange
			backPortRange, ok := l4Listener["back_port_range"].(string)
			if ok {
				listener.BackPortRange = backPortRange
			}
		}
		result = append(result, &listener)
	}
	return result
}

func SharpenL7Listeners(d *schema.ResourceData) []*zga.AccelerationRuleL7Listener {
	l7ListenersSet, ok := d.Get("l7_listeners").(*schema.Set)
	if !ok {
		return nil
	}
	var (
		l7ListenersV = l7ListenersSet.List()
		result       = make([]*zga.AccelerationRuleL7Listener, 0, len(l7ListenersV))
	)
	for _, l7ListenerV := range l7ListenersV {
		l7Listener, ok := l7ListenerV.(map[string]interface{})
		if !ok {
			continue
		}
		listener := zga.AccelerationRuleL7Listener{
			Protocol: l7Listener["protocol"].(string),
		}
		host, ok := l7Listener["host"].(string)
		if ok {
			listener.Host = host
		}
		backProtocol, ok := l7Listener["back_protocol"].(string)
		if ok {
			listener.BackProtocol = backProtocol
		}
		port, ok := l7Listener["port"].(int)
		if ok {
			listener.Port = port
			backPort, ok := l7Listener["back_port"].(int)
			if ok {
				listener.BackPort = backPort
			}
		}
		portRange, ok := l7Listener["port_range"].(string)
		if ok {
			listener.PortRange = portRange
			backPortRange, ok := l7Listener["back_port_range"].(string)
			if ok {
				listener.BackPortRange = backPortRange
			}
		}
		result = append(result, &listener)
	}
	return result
}

func SharpenProtocolOpts(d *schema.ResourceData) *zga.AccelerationRuleProtocolOpts {
	protocolOptsV, ok := d.Get("protocol_opts").([]interface{})
	if !ok || len(protocolOptsV) == 0 {
		return nil
	}
	protocolOpts, ok := protocolOptsV[0].(map[string]interface{})
	if !ok {
		return nil
	}
	var result zga.AccelerationRuleProtocolOpts
	toa, ok := protocolOpts["toa"].(bool)
	if ok {
		result.Toa = &toa
	}
	toaValue, ok := protocolOpts["toa_value"].(int)
	if ok {
		result.ToaValue = toaValue
	}
	websocket, ok := protocolOpts["websocket"].(bool)
	if ok {
		result.Websocket = &websocket
	}
	proxyProtocol, ok := protocolOpts["proxy_protocol"].(bool)
	if ok {
		result.ProxyProtocol = &proxyProtocol
	}
	gzip, ok := protocolOpts["gzip"].(bool)
	if ok {
		result.Gzip = &gzip
	}
	return &result
}

func SharpenHealthCheck(d *schema.ResourceData) *zga.HealthCheck {
	healthCheckV, ok := d.Get("health_check").([]interface{})
	if !ok || len(healthCheckV) == 0 {
		return nil
	}
	healthCheck, ok := healthCheckV[0].(map[string]interface{})
	if !ok {
		return nil
	}
	result := zga.HealthCheck{
		Enable: healthCheck["enable"].(bool),
	}
	port, ok := healthCheck["port"].(int)
	if ok {
		result.Port = port
	}
	alarm, ok := healthCheck["alarm"].(bool)
	if ok {
		result.Alarm = alarm
	}
	return &result
}

func SharpenAccessControl(d *schema.ResourceData) (exist bool, enable bool, result []zga.AccessControlRule) {
	accessControlV, ok := d.Get("access_control").([]interface{})
	if !ok || len(accessControlV) == 0 {
		return false, false, nil
	}
	accessControl, ok := accessControlV[0].(map[string]interface{})
	if !ok {
		return false, false, nil
	}
	enable = accessControl["enable"].(bool)
	rulesSet, ok := accessControl["rules"].(*schema.Set)
	if !ok {
		return true, enable, nil
	}
	rulesV := rulesSet.List()
	result = make([]zga.AccessControlRule, 0, len(rulesV))
	for _, ruleV := range rulesV {
		rule, ok := ruleV.(map[string]interface{})
		if !ok {
			continue
		}
		r := zga.AccessControlRule{
			Listener: rule["listener"].(string),
			CidrIp:   InterfaceSliceToString(rule["cidr_ip"].(*schema.Set).List()),
			Policy:   rule["policy"].(string),
		}
		directory, ok := rule["directory"].(string)
		if ok {
			r.Directory = directory
		}
		note, ok := rule["note"].(string)
		if ok {
			r.Note = note
		}
		result = append(result, r)
	}
	return true, enable, result
}

func resourceZenlayerCloudAcceleratorCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var (
		acceleratorId string
		zgaService    = NewZgaService(meta.(*connectivity.ZenlayerCloudClient))
	)
	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		request := zga.NewCreateAcceleratorRequest()
		request.AcceleratorName = d.Get("accelerator_name").(string)
		request.ChargeType = d.Get("charge_type").(string)
		request.CertificateId = d.Get("certificate_id").(string)
		request.ResourceGroupId = d.Get("resource_group_id").(string)
		request.Domain = SharpenDomain(d)
		request.Origin = SharpenOrigin(d)
		request.AccelerateRegions = SharpenAccelerateRegion(d)
		request.L4Listeners = SharpenL4Listeners(d)
		request.L7Listeners = SharpenL7Listeners(d)
		request.ProtocolOpts = SharpenProtocolOpts(d)
		request.HealthCheck = SharpenHealthCheck(d)
		response, errRet := zgaService.client.WithZgaClient().CreateAccelerator(request)
		if errRet != nil {
			tflog.Error(ctx, "Fail to create accelerator.", map[string]interface{}{
				"action":  request.GetAction(),
				"request": common2.ToJsonString(request),
				"err":     errRet.Error(),
			})
			return common2.RetryError(ctx, errRet, common2.InternalServerError)
		}

		tflog.Info(ctx, "Create accelerator success", map[string]interface{}{
			"action":   request.GetAction(),
			"request":  common2.ToJsonString(request),
			"response": common2.ToJsonString(response),
		})

		acceleratorId = response.Response.AcceleratorId
		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(acceleratorId)

	err = waitAcceleratorDeploySuccess(ctx, zgaService, d, acceleratorId)
	if err != nil {
		return diag.FromErr(err)
	}

	exist, acEnable, rules := SharpenAccessControl(d)
	if !exist {
		return resourceZenlayerCloudAcceleratorRead(ctx, d, meta)
	}

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		if len(rules) != 0 {
			errRet := zgaService.ModifyAcceleratorAccessControl(ctx, acceleratorId, rules)
			if errRet != nil {
				return common2.RetryError(ctx, errRet, common2.InternalServerError)
			}
		}
		if acEnable {
			errRet := zgaService.OpenAcceleratorAccessControl(ctx, acceleratorId)
			if errRet != nil {
				return common2.RetryError(ctx, errRet, common2.InternalServerError)
			}
		}
		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	err = waitAcceleratorDeploySuccess(ctx, zgaService, d, acceleratorId)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceZenlayerCloudAcceleratorRead(ctx, d, meta)
}

func resourceZenlayerCloudAcceleratorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var (
		diags           diag.Diagnostics
		acceleratorId   = d.Id()
		acceleratorInfo *zga.AcceleratorInfo
	)
	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead)-time.Minute, func() *resource.RetryError {
		var errRet error
		acceleratorInfo, errRet = NewZgaService(meta.(*connectivity.ZenlayerCloudClient)).DescribeAcceleratorById(ctx, acceleratorId)
		if errRet != nil {
			return common2.RetryError(ctx, errRet, common2.InternalServerError)
		}
		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	if acceleratorInfo == nil {
		d.SetId("")
		tflog.Info(ctx, "accelerator not exist or created failed or recycled", map[string]interface{}{
			"acceleratorId": acceleratorId,
		})
		return nil
	}

	_ = d.Set("accelerator_name", acceleratorInfo.AcceleratorName)
	_ = d.Set("charge_type", acceleratorInfo.ChargeType)
	_ = d.Set("resource_group_id", acceleratorInfo.ResourceGroupId)
	_ = d.Set("cname", acceleratorInfo.Cname)
	_ = d.Set("accelerator_status", acceleratorInfo.AcceleratorStatus)
	_ = d.Set("origin_region_id", acceleratorInfo.Origin.OriginRegionId)
	_ = d.Set("origin", splitStringByCommaOrSemicolon(acceleratorInfo.Origin.Origin))
	_ = d.Set("backup_origin", splitStringByCommaOrSemicolon(acceleratorInfo.Origin.BackupOrigin))
	_ = d.Set("accelerate_regions", flattenResourceAccelerateRegions(acceleratorInfo.AccelerateRegions))
	_ = d.Set("l4_listeners", flattenL4Listeners(acceleratorInfo.L4Listeners))
	_ = d.Set("l7_listeners", flattenL7Listeners(acceleratorInfo.L7Listeners))
	_ = d.Set("protocol_opts", flattenProtocolOpts(acceleratorInfo.ProtocolOpts))
	_ = d.Set("health_check", flattenHealthCheck(acceleratorInfo.HealthCheck))
	_ = d.Set("access_control", flattenAccessControl(acceleratorInfo.AccessControl))

	if acceleratorInfo.Domain != nil {
		_ = d.Set("domain", acceleratorInfo.Domain.Domain)
		_ = d.Set("relate_domains", splitStringByCommaOrSemicolon(acceleratorInfo.Domain.RelateDomains))
	} else {
		_ = d.Set("domain", "")
		_ = d.Set("relate_domains", nil)
	}

	if acceleratorInfo.Certificate != nil {
		_ = d.Set("certificate_id", acceleratorInfo.Certificate.CertificateId)
	} else {
		_ = d.Set("certificate_id", "")
	}
	return diags
}

func flattenResourceAccelerateRegions(acceelrateRegions []*zga.AccelerateRegionInfo) []map[string]interface{} {
	if acceelrateRegions == nil {
		return nil
	}
	var result = make([]map[string]interface{}, 0, len(acceelrateRegions))
	for _, region := range acceelrateRegions {
		result = append(result, map[string]interface{}{
			"accelerate_region_id": region.AccelerateRegionId,
			"vip":                  region.Vip,
			"bandwidth":            region.Bandwidth,
		})
	}
	return result
}

func resourceZenlayerCloudAcceleratorUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var (
		acceleratorId = d.Id()
		zgaService    = NewZgaService(meta.(*connectivity.ZenlayerCloudClient))
	)
	if d.HasChange("accelerator_name") {
		err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate)-time.Minute, func() *resource.RetryError {
			errRet := zgaService.ModifyAcceleratorName(ctx, acceleratorId, d.Get("accelerator_name").(string))
			if errRet != nil {
				return common2.RetryError(ctx, errRet, common2.InternalServerError, common.NetworkError)
			}
			return nil
		})
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("domain") || d.HasChange("relate_domains") {
		domain := SharpenDomain(d)
		if domain != nil {
			err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate)-time.Minute, func() *resource.RetryError {
				errRet := zgaService.ModifyAcceleratorDomain(ctx, acceleratorId, *domain)
				if errRet != nil {
					return common2.RetryError(ctx, errRet, common2.InternalServerError, common.NetworkError)
				}
				return nil
			})
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	if d.HasChange("certificate_id") {
		err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate)-time.Minute, func() *resource.RetryError {
			errRet := zgaService.ModifyAcceleratorCertificateId(ctx, acceleratorId, d.Get("certificate_id").(string))
			if errRet != nil {
				return common2.RetryError(ctx, errRet, common2.InternalServerError, common.NetworkError)
			}
			return nil
		})
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("origin") || d.HasChange("backup_origin") {
		err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate)-time.Minute, func() *resource.RetryError {
			errRet := zgaService.ModifyAcceleratorOrigin(ctx, acceleratorId, SharpenOrigin(d))
			if errRet != nil {
				return common2.RetryError(ctx, errRet, common2.InternalServerError, common.NetworkError)
			}
			return nil
		})
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("accelerate_regions") {
		err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate)-time.Minute, func() *resource.RetryError {
			errRet := zgaService.ModifyAcceleratorAccRegions(ctx, acceleratorId, SharpenAccelerateRegion(d))
			if errRet != nil {
				return common2.RetryError(ctx, errRet, common2.InternalServerError, common.NetworkError)
			}
			return nil
		})
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("l4_listeners") || d.HasChange("l7_listeners") {
		err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate)-time.Minute, func() *resource.RetryError {
			errRet := zgaService.ModifyAcceleratorListener(ctx, acceleratorId, SharpenL4Listeners(d), SharpenL7Listeners(d))
			if errRet != nil {
				return common2.RetryError(ctx, errRet, common2.InternalServerError, common.NetworkError)
			}
			return nil
		})
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("protocol_opts") {
		protocolOpts := SharpenProtocolOpts(d)
		if protocolOpts != nil {
			err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate)-time.Minute, func() *resource.RetryError {
				errRet := zgaService.ModifyAcceleratorProtocolOpts(ctx, acceleratorId, *protocolOpts)
				if errRet != nil {
					return common2.RetryError(ctx, errRet, common2.InternalServerError, common.NetworkError)
				}
				return nil
			})
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	if d.HasChange("health_check") {
		healthCheck := SharpenHealthCheck(d)
		if healthCheck != nil {
			err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate)-time.Minute, func() *resource.RetryError {
				errRet := zgaService.ModifyAcceleratorHealthCheck(ctx, acceleratorId, *healthCheck)
				if errRet != nil {
					return common2.RetryError(ctx, errRet, common2.InternalServerError, common.NetworkError)
				}
				return nil
			})
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	exist, enable, rules := SharpenAccessControl(d)
	if exist && d.HasChange("access_control.0.rules") {
		err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate)-time.Minute, func() *resource.RetryError {
			errRet := zgaService.ModifyAcceleratorAccessControl(ctx, acceleratorId, rules)
			if errRet != nil {
				return common2.RetryError(ctx, errRet, common2.InternalServerError, common.NetworkError)
			}
			return nil
		})
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if exist && d.HasChange("access_control.0.enable") {
		err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate)-time.Minute, func() *resource.RetryError {
			var errRet error
			if enable {
				errRet = zgaService.OpenAcceleratorAccessControl(ctx, acceleratorId)
			} else {
				errRet = zgaService.CloseAcceleratorAccessControl(ctx, acceleratorId)
			}
			if errRet != nil {
				return common2.RetryError(ctx, errRet, common2.InternalServerError, common.NetworkError)
			}
			return nil
		})
		if err != nil {
			return diag.FromErr(err)
		}
	}

	err := waitAcceleratorDeploySuccess(ctx, zgaService, d, acceleratorId)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceZenlayerCloudAcceleratorRead(ctx, d, meta)
}

func resourceZenlayerCloudAcceleratorDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	acceleratorId := d.Id()
	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
		errRet := NewZgaService(meta.(*connectivity.ZenlayerCloudClient)).DeleteAcceleratorById(ctx, acceleratorId)
		if errRet != nil {
			switch {
			case common2.IsExpectError(errRet, []string{"INVALID_ACCELERATOR_NOT_FOUND"}):
				// DO NOTHING
			default:
				return common2.RetryError(ctx, errRet, common2.InternalServerError)
			}
		}
		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func waitAcceleratorDeploySuccess(ctx context.Context, zgaService *ZgaService, d *schema.ResourceData, acceleratorId string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			ZgaAcceleratorStatusDeploying,
		},
		Target: []string{
			ZgaAcceleratorStatusAccelerating,
		},
		Refresh:        zgaService.AcceleratorStateRefreshFunc(ctx, acceleratorId, []string{ZgaAcceleratorStatusAccelerateFailure}),
		Timeout:        d.Timeout(schema.TimeoutCreate) - time.Minute,
		Delay:          10 * time.Second,
		MinTimeout:     5 * time.Second,
		NotFoundChecks: 3,
	}
	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return fmt.Errorf("error waiting for zga accelerator (%s) to be created: %v", acceleratorId, err)
	}
	return nil
}

func AccessControlValidFunc(_ context.Context, diff *schema.ResourceDiff, _ interface{}) error {
	accessControlV, ok := diff.Get("access_control").([]interface{})
	if !ok || len(accessControlV) == 0 {
		return nil
	}
	accessControl, ok := accessControlV[0].(map[string]interface{})
	if !ok {
		return nil
	}
	rulesSet, ok := accessControl["rules"].(*schema.Set)
	if !ok {
		return nil
	}
	rulesV := rulesSet.List()
	for _, ruleV := range rulesV {
		rule, ok := ruleV.(map[string]interface{})
		if ok {
			listener := rule["listener"].(string)
			directory, _ := rule["directory"].(string)
			if (listener == ZgaAccessControlAllListener ||
				strings.Contains(listener, ZgaUDPL4Protocol) ||
				strings.Contains(listener, ZgaTCPL4Protocol)) && directory != "/" {
				return errors.New("directory cannot be configured when listener is `all` or `tcp` or `udp`")
			}
		}
	}
	return nil
}

func ProtocolOptsCheckValidFunc(_ context.Context, diff *schema.ResourceDiff, _ interface{}) error {
	protocolOptsV, ok := diff.Get("protocol_opts").([]interface{})
	if !ok || len(protocolOptsV) == 0 {
		return nil
	}
	protocolOpts, ok := protocolOptsV[0].(map[string]interface{})
	if !ok {
		return nil
	}
	toa, _ := protocolOpts["toa"].(bool)
	proxyProtocol, _ := protocolOpts["proxy_protocol"].(bool)
	if toa && proxyProtocol {
		return errors.New("proxy_protocol and toa are not allowed to be configured at the same time")
	}
	return nil
}

func portAllocated(allocated *big.Int, offset int) bool {
	if allocated.Bit(offset) == 1 {
		return false
	}
	allocated.SetBit(allocated, offset, 1)
	return true
}

func ParsePort(data map[string]interface{}) (bport, eport int, err error) {
	port, ok := data["port"].(int)
	if ok && port != 0 {
		bport, eport = port, port
		backport, ok := data["back_port"].(int)
		if !ok || backport == 0 {
			err = errors.New("required back_port when port exist")
			return
		}
	}
	portRange, ok := data["port_range"].(string)
	if ok && portRange != "" {
		if port != 0 {
			err = errors.New("port exclude with port range")
			return
		}
		ports := strings.Split(portRange, "/")
		if len(ports) != 2 {
			err = errors.New("expected port_range separate with (/)")
			return
		}
		// ignore error, already check in IsPortRange
		bport, _ = strconv.Atoi(ports[0])
		eport, _ = strconv.Atoi(ports[1])
		backPortRange, ok := data["back_port_range"].(string)
		if !ok || backPortRange == "" {
			err = errors.New("required back_port_range when port_range exist")
			return
		}
		ports = strings.Split(backPortRange, "/")
		if len(ports) != 2 {
			err = errors.New("expected back_port_range separate with (/)")
			return
		}
		backbport, _ := strconv.Atoi(ports[0])
		backeport, _ := strconv.Atoi(ports[1])
		if eport-bport != backeport-backbport {
			err = errors.New("port_range length not equal back_port_range")
			return
		}
	}
	if eport == 0 || bport == 0 {
		err = errors.New("required port or port_range")
	}
	return
}

func PortCheckValidFunc(_ context.Context, diff *schema.ResourceDiff, _ interface{}) error {
	var (
		allowAccessControlListener = make(map[string]struct{})
		allowSingleTcpPort         = make(map[int]struct{})
		udpAllocated               = big.NewInt(0)
		tcpAllocated               = big.NewInt(0)
	)
	l7ListenersSet, ok := diff.Get("l7_listeners").(*schema.Set)
	if ok {
		l7ListenersV := l7ListenersSet.List()
		for _, l7ListenerV := range l7ListenersV {
			l7Listener, ok := l7ListenerV.(map[string]interface{})
			if ok {
				protocol := l7Listener["protocol"].(string)
				bport, eport, err := ParsePort(l7Listener)
				if err != nil {
					return err
				}
				if bport == eport {
					acListener := fmt.Sprintf("%s:%d", protocol, bport)
					allowAccessControlListener[acListener] = struct{}{}
					allowSingleTcpPort[bport] = struct{}{}
				} else {
					acListener := fmt.Sprintf("%s:%d/%d", protocol, bport, eport)
					allowAccessControlListener[acListener] = struct{}{}
				}
				for ; bport <= eport; bport++ {
					if !portAllocated(tcpAllocated, bport) {
						return fmt.Errorf("tcp port conflict in %d", bport)
					}
				}
			}
		}
	}

	l4ListenersSet, ok := diff.Get("l4_listeners").(*schema.Set)
	if ok {
		l4ListenersV := l4ListenersSet.List()
		for _, l4ListenerV := range l4ListenersV {
			l4Listener, ok := l4ListenerV.(map[string]interface{})
			if ok {
				protocol := l4Listener["protocol"].(string)
				bport, eport, err := ParsePort(l4Listener)
				if err != nil {
					return err
				}
				if bport == eport {
					if protocol == ZgaTCPL4Protocol {
						allowSingleTcpPort[bport] = struct{}{}
					}
					acListener := fmt.Sprintf("%s:%d", protocol, bport)
					allowAccessControlListener[acListener] = struct{}{}
				} else {
					acListener := fmt.Sprintf("%s:%d/%d", protocol, bport, eport)
					allowAccessControlListener[acListener] = struct{}{}
				}
				for ; bport <= eport; bport++ {
					switch protocol {
					case ZgaTCPL4Protocol:
						if !portAllocated(tcpAllocated, bport) {
							return fmt.Errorf("tcp port conflict in %d", bport)
						}
					default:
						if !portAllocated(udpAllocated, bport) {
							return fmt.Errorf("udp port conflict in %d", bport)
						}
					}
				}
			}
		}
	}

	if len(allowAccessControlListener) == 0 {
		return errors.New("required listeners")
	}

	accessControlV, ok := diff.Get("access_control").([]interface{})
	if ok && len(accessControlV) == 1 {
		accessControl, ok := accessControlV[0].(map[string]interface{})
		if ok {
			rulesSet, ok := accessControl["rules"].(*schema.Set)
			if ok {
				rulesV := rulesSet.List()
				for _, ruleV := range rulesV {
					rule, ok := ruleV.(map[string]interface{})
					if ok {
						listener := rule["listener"].(string)
						_, allow := allowAccessControlListener[listener]
						if !allow && listener != ZgaAccessControlAllListener {
							return fmt.Errorf("%s access control listeners should be included in acceleration listeners", listener)
						}
					}
				}
			}
		}
	}

	healthCheckV, ok := diff.Get("health_check").([]interface{})
	if ok && len(healthCheckV) == 1 {
		hc, ok := healthCheckV[0].(map[string]interface{})
		if ok {
			hcPort, ok := hc["port"].(int)
			if ok && hcPort != 0 {
				_, allow := allowSingleTcpPort[hcPort]
				if !allow {
					return fmt.Errorf("%d health check port should be included in tcp single port", hcPort)
				}
			}
		}
	}
	return nil
}

func DomainAcceleratorValidFunc() schema.CustomizeDiffFunc {
	return customdiff.IfValue("domain", func(ctx context.Context, value, meta interface{}) bool {
		return value != ""
	}, func(ctx context.Context, diff *schema.ResourceDiff, i interface{}) error {
		domain, ok := diff.Get("domain").(string)
		if !ok {
			return nil
		}
		relateDomainsSet, ok := diff.Get("relate_domains").(*schema.Set)
		if !ok {
			return nil
		}
		if relateDomainsSet.Contains(domain) {
			return fmt.Errorf("domain conflict: %s", domain)
		}
		return nil
	})
}

func IPAcceleratorValidFunc() schema.CustomizeDiffFunc {
	return customdiff.IfValue("domain", func(ctx context.Context, value, meta interface{}) bool {
		return value == ""
	}, func(ctx context.Context, diff *schema.ResourceDiff, i interface{}) error {
		if _, ok := diff.GetOk("relate_domains"); ok {
			return errors.New("ip accelerator cannot configure relate_domains")
		}
		if _, ok := diff.GetOk("certificate_id"); ok {
			return errors.New("ip accelerator cannot configure certificate_id")
		}
		if _, ok := diff.GetOk("l7_listeners"); ok {
			return errors.New("ip accelerator cannot configure l7_listeners")
		}
		protocolOptsV, ok := diff.Get("protocol_opts").([]interface{})
		if ok && len(protocolOptsV) == 1 {
			protocolOpts, ok := protocolOptsV[0].(map[string]interface{})
			if ok {
				websocket, ok := protocolOpts["websocket"].(bool)
				if ok && websocket {
					return errors.New("ip accelerator cannot configure websocket in protocol opts")
				}
				gzip, ok := protocolOpts["gzip"].(bool)
				if ok && gzip {
					return errors.New("ip accelerator cannot configure gzip in protocol opts")
				}
			}
		}
		return nil
	})
}

func IsPortRange(i interface{}, k string) (warnings []string, errors []error) {
	v, ok := i.(string)
	if !ok {
		errors = append(errors, fmt.Errorf("expected type of %s to be string", k))
		return
	}
	slice := strings.Split(v, "/")
	if len(slice) != 2 {
		errors = append(errors, fmt.Errorf("expected %s separate with (/)", k))
		return
	}
	bport, err := strconv.Atoi(slice[0])
	if err != nil {
		errors = append(errors, fmt.Errorf("expected %q to be a valid port number", k))
		return
	} else if bport < 1 {
		errors = append(errors, fmt.Errorf("expected %q port range (1, 65535]", k))
		return
	}
	eport, err := strconv.Atoi(slice[1])
	if err != nil {
		errors = append(errors, fmt.Errorf("expected %q to be a valid port number", k))
		return
	} else if eport > 65535 {
		errors = append(errors, fmt.Errorf("expected %q port range (1, 65535]", k))
		return
	}
	if bport > eport {
		errors = append(errors, fmt.Errorf("expected %q bport less than eport", k))
	}
	return
}

func IsAcListener(i interface{}, k string) (warnings []string, errors []error) {
	v, ok := i.(string)
	if !ok {
		errors = append(errors, fmt.Errorf("expected type of %s to be string", k))
		return
	}
	if v == ZgaAccessControlAllListener {
		return
	}

	splits := strings.Split(v, ":")
	if len(splits) != 2 {
		errors = append(errors, fmt.Errorf("expected %q like $protocol:$port", k))
		return
	}

	protocol := splits[0]
	switch protocol {
	case ZgaUDPL4Protocol, ZgaTCPL4Protocol, ZgaHTTPL7Protocol, ZgaHTTPSL7Protocol:
		portStr := splits[1]
		if !strings.Contains(portStr, "/") {
			port, err := strconv.Atoi(portStr)
			if err != nil {
				errors = append(errors, fmt.Errorf("expected %q like $protocol:$port", k))
			} else if port < 1 || port > 65535 {
				errors = append(errors, fmt.Errorf("expected %q port range (1, 65535]", k))
			}
			return
		}
		warnings, errors = IsPortRange(portStr, k)
	default:
		errors = append(errors, fmt.Errorf("expected %q protocol: `udp`, `tcp`, `http`, `https`", k))
	}
	return
}

func InterfaceSliceToString(slice []interface{}) string {
	var result = make([]string, 0, len(slice))
	for _, v := range slice {
		str, ok := v.(string)
		if !ok {
			continue
		}
		result = append(result, str)
	}
	return joinArrayByComma(result)
}
