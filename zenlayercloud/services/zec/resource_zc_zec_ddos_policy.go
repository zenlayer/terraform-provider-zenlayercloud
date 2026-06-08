package zec

/*
Provides a ZEC DDoS protection policy resource.

~> **NOTE:** TCP and UDP cannot be blocked simultaneously in `block_protocol`.

~> **NOTE:** When moving an EIP from one policy to another within the same
Terraform config, the target policy must declare `depends_on` the source policy.
This ensures Terraform detaches the EIP from the source first before attaching
it to the target, avoiding API conflict errors from concurrent updates.

Example Usage

```hcl
resource "zenlayercloud_zec_ddos_policy" "example" {
  policy_name  = "my-ddos-policy"
  ipv4_id_list = [zenlayercloud_zec_eip.example.id]

  black_ip_list    = ["1.2.3.4"]
  white_ip_list    = ["5.6.7.8"]
  ip_black_timeout = 60

  block_protocol = ["ICMP"]
  block_regions  = ["CN"]

  port {
    protocol       = "UDP"
    src_port_start = 0
    src_port_end   = 65535
    dst_port_start = 0
    dst_port_end   = 65535
  }

  reflect_udp_port {
    port = 123
  }

  traffic_control {
    bps_enabled = true
    bps         = 100000000
    pps_enabled = true
    pps         = 100000
  }

  tags = {
    env = "prod"
  }
}
```

Import

DDoS policies can be imported using the policy ID:

```
$ terraform import zenlayercloud_zec_ddos_policy.example pol-xxxxxxxx
```
*/

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/common"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/services/zrm"
	common2 "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/common"
	user "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/user20240529"
	zec "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/zec20250901"
)

// validateIntBetween64 validates that an integer field falls within the given
// int64 range. It is used in place of validation.IntBetween when the upper
// bound exceeds the 32-bit int max (e.g. 2147483648), which would overflow the
// untyped constant when cross-compiling to 32-bit platforms (386, arm).
func validateIntBetween64(min, max int64) schema.SchemaValidateFunc {
	return func(i interface{}, k string) (warnings []string, errors []error) {
		v, ok := i.(int)
		if !ok {
			errors = append(errors, fmt.Errorf("expected type of %q to be int", k))
			return
		}
		if int64(v) < min || int64(v) > max {
			errors = append(errors, fmt.Errorf("expected %q to be in the range (%d - %d), got %d", k, min, max, v))
		}
		return
	}
}

func ResourceZenlayerCloudZecDDoSPolicy() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceZenlayerCloudZecDDoSPolicyCreate,
		ReadContext:   resourceZenlayerCloudZecDDoSPolicyRead,
		UpdateContext: resourceZenlayerCloudZecDDoSPolicyUpdate,
		DeleteContext: resourceZenlayerCloudZecDDoSPolicyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		CustomizeDiff: customdiff.All(
			ddosPolicyBlockProtocolValidFunc(),
			ddosPolicyIpBlackTimeoutRequiredFunc(),
		),
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"policy_name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.All(validation.StringLenBetween(2, 63), validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9\-.]*[a-zA-Z0-9]$`), "must start and end with a letter or digit, and contain only letters, digits, hyphens, and dots")),
				Description:  "Name of the DDoS protection policy. 2-63 characters, only letters, digits, `-`, and `.` are allowed, must start and end with a letter or digit.",
			},
			"resource_group_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The resource group ID. If not specified, the default resource group is used.",
			},
			"ipv4_id_list": {
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "List of EIP IDs to attach to this policy. Each EIP can only be attached to one policy at a time. When moving an EIP from one policy to another, the target policy resource must declare `depends_on` the source policy to ensure detach completes before attach.",
			},
			"black_ip_list": {
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "List of blacklisted IP addresses.",
			},
			"white_ip_list": {
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "List of whitelisted IP addresses.",
			},
			"ip_black_timeout": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(1, 10080),
				Description:  "Blacklist timeout in minutes. Valid range: 1-10080. Required when black_ip_list is set.",
			},
			"block_protocol": {
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString, ValidateFunc: validation.StringInSlice([]string{"TCP", "UDP", "ICMP"}, false)},
				Description: "List of protocols to block. Valid values: `TCP`, `UDP`, `ICMP`. Note: `TCP` and `UDP` cannot be blocked simultaneously.",
			},
			"block_regions": {
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "List of region IDs to block. Use `DescribePolicyRegions` API to get available region IDs.",
			},
			"port": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Port blocking rules.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"protocol": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"TCP", "UDP"}, false),
							Description:  "Protocol type. Valid values: `TCP`, `UDP`.",
						},
						"src_port_start": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(0, 65535),
							Description:  "Source port range start value. Range: 0-65535.",
						},
						"src_port_end": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(0, 65535),
							Description:  "Source port range end value. Range: 0-65535.",
						},
						"dst_port_start": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(0, 65535),
							Description:  "Destination port range start value. Range: 0-65535.",
						},
						"dst_port_end": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(0, 65535),
							Description:  "Destination port range end value. Range: 0-65535.",
						},
						"action": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"Drop"}, false),
							Description:  "Action to take on match. Valid values: `Drop`.",
						},
					},
				},
			},
			"fingerprint_rule": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Fingerprint filtering rules.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"protocol": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"TCP", "UDP", "ICMP"}, false),
							Description:  "Protocol type. Valid values: `TCP`, `UDP`, `ICMP`.",
						},
						"src_port_start": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(0, 65535),
							Description:  "Source port range start value. Range: 0-65535.",
						},
						"src_port_end": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(0, 65535),
							Description:  "Source port range end value. Range: 0-65535.",
						},
						"dst_port_start": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(0, 65535),
							Description:  "Destination port range start value. Range: 0-65535.",
						},
						"dst_port_end": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(0, 65535),
							Description:  "Destination port range end value. Range: 0-65535.",
						},
						"min_pkt_length": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(1, 1500),
							Description:  "Minimum packet length to filter. Range: 1-1500.",
						},
						"max_pkt_length": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(1, 1500),
							Description:  "Maximum packet length to filter. Range: 1-1500.",
						},
						"offset": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(0, 1500),
							Description:  "Payload offset for fingerprint matching. Range: 0-1500.",
						},
						"match_bytes": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Bytes to match in the payload. Hexadecimal lowercase, zero-padded to 2 digits (e.g. `deadbeef`).",
						},
						"action": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"Drop"}, false),
							Description:  "Action to take on match. Valid values: `Drop`.",
						},
					},
				},
			},
			"reflect_udp_port": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Additional UDP reflection attack source ports to block, on top of the system built-in defaults. Use `DescribeReflectUdpPortOptions` to query the built-in default ports.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"port": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(0, 65535),
							Description:  "UDP reflection source port to block. Range: 0-65535.",
						},
					},
				},
			},
			"traffic_control": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "Source IP rate limiting configuration.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"bps_enabled": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Whether to enable bps rate limiting.",
						},
						"bps": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validateIntBetween64(8192, 2147483648),
							Description:  "Bps rate limit value. Valid range: [8192, 2147483648].",
						},
						"pps_enabled": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Whether to enable pps rate limiting.",
						},
						"pps": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(32, 50000),
							Description:  "Pps rate limit value. Valid range: [32, 50000].",
						},
						"syn_bps_enabled": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Whether to enable SYN bps rate limiting.",
						},
						"syn_bps": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validateIntBetween64(8192, 2147483648),
							Description:  "SYN bps rate limit value. Valid range: [8192, 2147483648].",
						},
						"syn_pps_enabled": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Whether to enable SYN pps rate limiting.",
						},
						"syn_pps": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(1, 100000),
							Description:  "SYN pps rate limit value. Valid range: [1, 100000].",
						},
					},
				},
			},
			"tags": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "Tags associated with the DDoS policy.",
			},
			"create_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The time when the DDoS policy was created.",
			},
		},
	}
}

// ddosPolicyIpBlackTimeoutRequiredFunc validates that ip_black_timeout is set when black_ip_list or white_ip_list is non-empty.
// The API requires ipBlackTimeout whenever a black/white IP list is provided.
func ddosPolicyIpBlackTimeoutRequiredFunc() schema.CustomizeDiffFunc {
	return func(ctx context.Context, d *schema.ResourceDiff, meta interface{}) error {
		blackList := d.Get("black_ip_list").(*schema.Set).List()
		if len(blackList) > 0 && d.Get("ip_black_timeout").(int) == 0 {
			return fmt.Errorf("ip_black_timeout is required when black_ip_list is set (valid range: 1-10080 minutes)")
		}
		return nil
	}
}

// ddosPolicyBlockProtocolValidFunc validates that TCP and UDP are not both present in block_protocol
func ddosPolicyBlockProtocolValidFunc() schema.CustomizeDiffFunc {
	return customdiff.ValidateValue("block_protocol", func(ctx context.Context, value, meta interface{}) error {
		protocols := value.(*schema.Set).List()
		hasTCP := false
		hasUDP := false
		for _, p := range protocols {
			switch p.(string) {
			case "TCP":
				hasTCP = true
			case "UDP":
				hasUDP = true
			}
		}
		if hasTCP && hasUDP {
			return fmt.Errorf("TCP and UDP cannot be blocked simultaneously in block_protocol")
		}
		return nil
	})
}

func resourceZenlayerCloudZecDDoSPolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common.LogElapsed(ctx, "resource.zenlayercloud_zec_ddos_policy.create")()

	zecService := ZecService{client: meta.(*connectivity.ZenlayerCloudClient)}

	request := zec.NewCreatePolicyRequest()
	request.PolicyName = common2.String(d.Get("policy_name").(string))

	if v, ok := d.GetOk("resource_group_id"); ok {
		request.ResourceGroupId = common2.String(v.(string))
	}
	if v, ok := d.GetOk("black_ip_list"); ok {
		request.BlackIpList = common.ToStringList(v.(*schema.Set).List())
	}
	if v, ok := d.GetOk("white_ip_list"); ok {
		request.WhiteIpList = common.ToStringList(v.(*schema.Set).List())
	}
	if v, ok := d.GetOk("ip_black_timeout"); ok {
		request.IpBlackTimeout = common2.Integer(v.(int))
	}
	if v, ok := d.GetOk("block_protocol"); ok {
		request.BlockProtocol = common.ToStringList(v.(*schema.Set).List())
	}
	if v, ok := d.GetOk("block_regions"); ok {
		request.BlockRegions = common.ToStringList(v.(*schema.Set).List())
	}
	if v, ok := d.GetOk("port"); ok {
		request.Ports = expandDDoSPolicyPorts(v.([]interface{}))
	}
	if v, ok := d.GetOk("fingerprint_rule"); ok {
		request.Finger = expandDDoSFingerprintRules(v.([]interface{}))
	}
	if v, ok := d.GetOk("reflect_udp_port"); ok {
		request.ReflectUdpPort = expandDDoSReflectUdpPorts(v.([]interface{}))
	}
	if v, ok := d.GetOk("traffic_control"); ok {
		request.TrafficControl = expandDDoSTrafficControl(v.([]interface{}))
	}
	if tags := common.GetTags(d, "tags"); len(tags) > 0 {
		request.Tags = buildTagAssociation(tags)
	}

	var policyId string
	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		response, err := zecService.client.WithZec2Client().CreatePolicy(request)
		if err != nil {
			return common.RetryError(ctx, err, common.InternalServerError, common2.NetworkError)
		}
		if response == nil || response.Response == nil || response.Response.PolicyId == nil {
			return resource.NonRetryableError(fmt.Errorf("failed to get policy ID from response"))
		}
		policyId = *response.Response.PolicyId
		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(policyId)

	// attach EIPs if specified
	if v, ok := d.GetOk("ipv4_id_list"); ok {
		ipv4Ids := common.ToStringList(v.(*schema.Set).List())
		if len(ipv4Ids) > 0 {
			if err := zecService.AttachDDoSPolicy(ctx, policyId, ipv4Ids); err != nil {
				return diag.FromErr(fmt.Errorf("error attaching EIPs to DDoS policy %s: %w", policyId, err))
			}
		}
	}

	return resourceZenlayerCloudZecDDoSPolicyRead(ctx, d, meta)
}

func resourceZenlayerCloudZecDDoSPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common.LogElapsed(ctx, "resource.zenlayercloud_zec_ddos_policy.read")()

	zecService := ZecService{client: meta.(*connectivity.ZenlayerCloudClient)}
	policyId := d.Id()

	policy, err := zecService.DescribeDDoSPolicyById(ctx, policyId)
	if err != nil {
		return diag.FromErr(err)
	}
	if policy == nil {
		d.SetId("")
		tflog.Info(ctx, "zec ddos policy not found, removing from state", map[string]interface{}{
			"policyId": policyId,
		})
		return nil
	}

	_ = d.Set("policy_name", policy.PolicyName)
	_ = d.Set("create_time", policy.CreateTime)

	// black/white ip list - API returns BlackIps/WhiteIps
	if len(policy.BlackIps) > 0 {
		_ = d.Set("black_ip_list", policy.BlackIps)
	} else {
		_ = d.Set("black_ip_list", []string{})
	}
	if len(policy.WhiteIps) > 0 {
		_ = d.Set("white_ip_list", policy.WhiteIps)
	} else {
		_ = d.Set("white_ip_list", []string{})
	}
	if policy.BlackIpListExpireAt != nil {
		_ = d.Set("ip_black_timeout", policy.BlackIpListExpireAt)
	}

	// block_protocol - API returns BlockProtocols
	if len(policy.BlockProtocols) > 0 {
		_ = d.Set("block_protocol", policy.BlockProtocols)
	} else {
		_ = d.Set("block_protocol", []string{})
	}

	// block_regions
	if len(policy.BlockRegions) > 0 {
		_ = d.Set("block_regions", policy.BlockRegions)
	} else {
		_ = d.Set("block_regions", []string{})
	}

	// port rules
	_ = d.Set("port", flattenDDoSPolicyPorts(policy.Ports))

	// fingerprint rules - API returns FingerPrintRules
	_ = d.Set("fingerprint_rule", flattenDDoSFingerprintRules(policy.FingerPrintRules))

	// reflect udp ports
	_ = d.Set("reflect_udp_port", flattenDDoSReflectUdpPorts(policy.ReflectUdpPort))

	// traffic control
	_ = d.Set("traffic_control", flattenDDoSTrafficControl(policy.TrafficControl))

	// DescribePolicyDetail returns AttachmentIps as public IPs; resolve to EIP IDs for state.
	ipv4IdList, err := zecService.ResolveIpv4IdsFromAttachmentIps(ctx, policy.AttachmentIps)
	if err != nil {
		return diag.FromErr(err)
	}
	_ = d.Set("ipv4_id_list", ipv4IdList)

	return nil
}

func resourceZenlayerCloudZecDDoSPolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common.LogElapsed(ctx, "resource.zenlayercloud_zec_ddos_policy.update")()

	zecService := ZecService{client: meta.(*connectivity.ZenlayerCloudClient)}
	policyId := d.Id()
	d.Partial(true)

	// Update policy name (no configType required)
	if d.HasChange("policy_name") {
		req := zec.NewModifyPolicyRequest()
		req.PolicyId = &policyId
		req.PolicyName = common2.String(d.Get("policy_name").(string))
		if err := retryModifyPolicy(ctx, d, zecService, req); err != nil {
			return diag.FromErr(err)
		}
	}

	// Update IP black/white list
	if d.HasChanges("black_ip_list", "white_ip_list", "ip_black_timeout") {
		req := zec.NewModifyPolicyRequest()
		req.PolicyId = &policyId
		req.ConfigType = common2.String("IpList")
		if v, ok := d.GetOk("black_ip_list"); ok {
			req.BlackIpList = common.ToStringList(v.(*schema.Set).List())
		} else {
			req.BlackIpList = []string{}
		}
		if v, ok := d.GetOk("white_ip_list"); ok {
			req.WhiteIpList = common.ToStringList(v.(*schema.Set).List())
		} else {
			req.WhiteIpList = []string{}
		}
		if v, ok := d.GetOk("ip_black_timeout"); ok {
			req.IpBlackTimeout = common2.Integer(v.(int))
		}
		if err := retryModifyPolicy(ctx, d, zecService, req); err != nil {
			return diag.FromErr(err)
		}
	}

	// Update block protocol
	if d.HasChange("block_protocol") {
		req := zec.NewModifyPolicyRequest()
		req.PolicyId = &policyId
		req.ConfigType = common2.String("BlockProtocol")
		if v, ok := d.GetOk("block_protocol"); ok {
			req.BlockProtocol = common.ToStringList(v.(*schema.Set).List())
		} else {
			req.BlockProtocol = []string{}
		}
		if err := retryModifyPolicy(ctx, d, zecService, req); err != nil {
			return diag.FromErr(err)
		}
	}

	// Update block regions
	if d.HasChange("block_regions") {
		req := zec.NewModifyPolicyRequest()
		req.PolicyId = &policyId
		req.ConfigType = common2.String("BlockRegion")
		if v, ok := d.GetOk("block_regions"); ok {
			req.BlockRegions = common.ToStringList(v.(*schema.Set).List())
		} else {
			req.BlockRegions = []string{}
		}
		if err := retryModifyPolicy(ctx, d, zecService, req); err != nil {
			return diag.FromErr(err)
		}
	}

	// Update port blocking rules
	if d.HasChange("port") {
		req := zec.NewModifyPolicyRequest()
		req.PolicyId = &policyId
		req.ConfigType = common2.String("Port")
		if v, ok := d.GetOk("port"); ok {
			req.Ports = expandDDoSPolicyPorts(v.([]interface{}))
		} else {
			req.Ports = []*zec.DdosPolicyPort{}
		}
		if err := retryModifyPolicy(ctx, d, zecService, req); err != nil {
			return diag.FromErr(err)
		}
	}

	// Update fingerprint rules
	if d.HasChange("fingerprint_rule") {
		req := zec.NewModifyPolicyRequest()
		req.PolicyId = &policyId
		req.ConfigType = common2.String("Fingerprint")
		if v, ok := d.GetOk("fingerprint_rule"); ok {
			req.Finger = expandDDoSFingerprintRules(v.([]interface{}))
		} else {
			req.Finger = []*zec.DdosFingerprintRule{}
		}
		if err := retryModifyPolicy(ctx, d, zecService, req); err != nil {
			return diag.FromErr(err)
		}
	}

	// Update reflect UDP ports
	if d.HasChange("reflect_udp_port") {
		req := zec.NewModifyPolicyRequest()
		req.PolicyId = &policyId
		req.ConfigType = common2.String("UdpReflect")
		if v, ok := d.GetOk("reflect_udp_port"); ok {
			req.ReflectUdpPort = expandDDoSReflectUdpPorts(v.([]interface{}))
		} else {
			req.ReflectUdpPort = []*zec.DdosReflectUdpPort{}
		}
		if err := retryModifyPolicy(ctx, d, zecService, req); err != nil {
			return diag.FromErr(err)
		}
	}

	// Update traffic control
	if d.HasChange("traffic_control") {
		req := zec.NewModifyPolicyRequest()
		req.PolicyId = &policyId
		req.ConfigType = common2.String("TrafficControl")
		if v, ok := d.GetOk("traffic_control"); ok {
			req.TrafficControl = expandDDoSTrafficControl(v.([]interface{}))
		}
		if err := retryModifyPolicy(ctx, d, zecService, req); err != nil {
			return diag.FromErr(err)
		}
	}

	// Update EIP bindings: detach removed EIPs first, then attach new EIPs
	if d.HasChange("ipv4_id_list") {
		old, new := d.GetChange("ipv4_id_list")
		oldSet := old.(*schema.Set)
		newSet := new.(*schema.Set)

		removed := common.ToStringList(oldSet.Difference(newSet).List())
		added := common.ToStringList(newSet.Difference(oldSet).List())

		if len(removed) > 0 {
			if err := zecService.DetachDDoSPolicy(ctx, policyId, removed); err != nil {
				return diag.FromErr(fmt.Errorf("error detaching EIPs from DDoS policy %s: %w", policyId, err))
			}
		}
		if len(added) > 0 {
			if err := zecService.AttachDDoSPolicy(ctx, policyId, added); err != nil {
				return diag.FromErr(fmt.Errorf("error attaching EIPs to DDoS policy %s: %w", policyId, err))
			}
		}
	}

	// Update resource group
	if d.HasChange("resource_group_id") {
		err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate)-time.Minute, func() *resource.RetryError {
			request := user.NewAddResourceResourceGroupRequest()
			request.ResourceGroupId = common2.String(d.Get("resource_group_id").(string))
			request.Resources = []string{policyId}
			_, err := meta.(*connectivity.ZenlayerCloudClient).WithUsrClient().AddResourceResourceGroup(request)
			if err != nil {
				return common.RetryError(ctx, err, common.InternalServerError, common2.NetworkError)
			}
			return nil
		})
		if err != nil {
			return diag.FromErr(err)
		}
	}

	// Update tags via ZRM service
	if d.HasChange("tags") {
		zrmService := zrm.NewZrmService(meta.(*connectivity.ZenlayerCloudClient))
		if err := zrmService.ModifyResourceTags(ctx, d, policyId); err != nil {
			return diag.FromErr(err)
		}
	}

	d.Partial(false)
	return resourceZenlayerCloudZecDDoSPolicyRead(ctx, d, meta)
}

func resourceZenlayerCloudZecDDoSPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common.LogElapsed(ctx, "resource.zenlayercloud_zec_ddos_policy.delete")()

	zecService := ZecService{client: meta.(*connectivity.ZenlayerCloudClient)}
	policyId := d.Id()

	// Detach all EIPs before deleting the policy
	if v, ok := d.GetOk("ipv4_id_list"); ok {
		ipv4Ids := common.ToStringList(v.(*schema.Set).List())
		if len(ipv4Ids) > 0 {
			if err := zecService.DetachDDoSPolicy(ctx, policyId, ipv4Ids); err != nil {
				return diag.FromErr(fmt.Errorf("error detaching EIPs before deleting DDoS policy %s: %w", policyId, err))
			}
		}
	}

	request := zec.NewDeletePolicyRequest()
	request.PolicyId = &policyId

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		_, err := zecService.client.WithZec2Client().DeletePolicy(request)
		if err != nil {
			if sdkErr, ok := err.(*common2.ZenlayerCloudSdkError); ok {
				if sdkErr.Code == common.ResourceNotFound {
					return nil
				}
			}
			return common.RetryError(ctx, err, common.InternalServerError, common2.NetworkError)
		}
		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

// retryModifyPolicy executes a ModifyPolicy request with retry logic
func retryModifyPolicy(ctx context.Context, d *schema.ResourceData, svc ZecService, req *zec.ModifyPolicyRequest) error {
	return resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
		resp, err := svc.client.WithZec2Client().ModifyPolicy(req)
		defer common.LogApiRequest(ctx, "ModifyPolicy", req, resp, err)
		if err != nil {
			return common.RetryError(ctx, err, common.InternalServerError, common2.NetworkError)
		}
		return nil
	})
}

// --- expand helpers ---

func expandDDoSPolicyPorts(raw []interface{}) []*zec.DdosPolicyPort {
	ports := make([]*zec.DdosPolicyPort, 0, len(raw))
	for _, v := range raw {
		m := v.(map[string]interface{})
		port := &zec.DdosPolicyPort{
			Protocol:     common2.String(m["protocol"].(string)),
			SrcPortStart: common2.Integer(m["src_port_start"].(int)),
			SrcPortEnd:   common2.Integer(m["src_port_end"].(int)),
			DstPortStart: common2.Integer(m["dst_port_start"].(int)),
			DstPortEnd:   common2.Integer(m["dst_port_end"].(int)),
			Action:       common2.String(m["action"].(string)),
		}
		ports = append(ports, port)
	}
	return ports
}

func expandDDoSFingerprintRules(raw []interface{}) []*zec.DdosFingerprintRule {
	rules := make([]*zec.DdosFingerprintRule, 0, len(raw))
	for _, v := range raw {
		m := v.(map[string]interface{})
		rule := &zec.DdosFingerprintRule{
			Protocol:     common2.String(m["protocol"].(string)),
			SrcPortStart: common2.Integer(m["src_port_start"].(int)),
			SrcPortEnd:   common2.Integer(m["src_port_end"].(int)),
			DstPortStart: common2.Integer(m["dst_port_start"].(int)),
			DstPortEnd:   common2.Integer(m["dst_port_end"].(int)),
			MinPktLength: common2.Integer(m["min_pkt_length"].(int)),
			MaxPktLength: common2.Integer(m["max_pkt_length"].(int)),
		}
		if val, ok := m["offset"].(int); ok {
			rule.Offset = common2.Integer(val)
		}
		if val, ok := m["match_bytes"].(string); ok && val != "" {
			rule.MatchBytes = common2.String(val)
		}
		if val, ok := m["action"].(string); ok && val != "" {
			rule.Action = common2.String(val)
		}
		rules = append(rules, rule)
	}
	return rules
}

func expandDDoSReflectUdpPorts(raw []interface{}) []*zec.DdosReflectUdpPort {
	ports := make([]*zec.DdosReflectUdpPort, 0, len(raw))
	for _, v := range raw {
		m := v.(map[string]interface{})
		ports = append(ports, &zec.DdosReflectUdpPort{
			Port: common2.Integer(m["port"].(int)),
		})
	}
	return ports
}

func expandDDoSTrafficControl(raw []interface{}) *zec.DdosTrafficControl {
	if len(raw) == 0 || raw[0] == nil {
		return nil
	}
	m := raw[0].(map[string]interface{})
	tc := &zec.DdosTrafficControl{}
	if v, ok := m["bps_enabled"].(bool); ok {
		tc.BpsEnabled = common2.Bool(v)
	}
	if v, ok := m["bps"].(int); ok && v != 0 {
		bps := int64(v)
		tc.Bps = &bps
	}
	if v, ok := m["pps_enabled"].(bool); ok {
		tc.PpsEnabled = common2.Bool(v)
	}
	if v, ok := m["pps"].(int); ok && v != 0 {
		pps := int64(v)
		tc.Pps = &pps
	}
	if v, ok := m["syn_bps_enabled"].(bool); ok {
		tc.SynBPSEnabled = common2.Bool(v)
	}
	if v, ok := m["syn_bps"].(int); ok && v != 0 {
		synBps := int64(v)
		tc.SynBPS = &synBps
	}
	if v, ok := m["syn_pps_enabled"].(bool); ok {
		tc.SynPPSEnabled = common2.Bool(v)
	}
	if v, ok := m["syn_pps"].(int); ok && v != 0 {
		synPps := int64(v)
		tc.SynPPS = &synPps
	}
	return tc
}

func buildTagAssociation(tags map[string]string) *zec.TagAssociation {
	ta := &zec.TagAssociation{}
	for k, v := range tags {
		tmpKey := k
		tmpValue := v
		ta.Tags = append(ta.Tags, &zec.Tag{
			Key:   &tmpKey,
			Value: &tmpValue,
		})
	}
	return ta
}

// --- flatten helpers ---

func flattenDDoSPolicyPorts(ports []*zec.DdosPolicyPort) []interface{} {
	result := make([]interface{}, 0, len(ports))
	for _, p := range ports {
		m := map[string]interface{}{}
		if p.Protocol != nil {
			m["protocol"] = *p.Protocol
		}
		if p.SrcPortStart != nil {
			m["src_port_start"] = *p.SrcPortStart
		}
		if p.SrcPortEnd != nil {
			m["src_port_end"] = *p.SrcPortEnd
		}
		if p.DstPortStart != nil {
			m["dst_port_start"] = *p.DstPortStart
		}
		if p.DstPortEnd != nil {
			m["dst_port_end"] = *p.DstPortEnd
		}
		if p.Action != nil {
			m["action"] = *p.Action
		}
		result = append(result, m)
	}
	return result
}

func flattenDDoSFingerprintRules(rules []*zec.DdosFingerprintRule) []interface{} {
	result := make([]interface{}, 0, len(rules))
	for _, r := range rules {
		m := map[string]interface{}{}
		if r.Protocol != nil {
			m["protocol"] = *r.Protocol
		}
		if r.SrcPortStart != nil {
			m["src_port_start"] = *r.SrcPortStart
		}
		if r.SrcPortEnd != nil {
			m["src_port_end"] = *r.SrcPortEnd
		}
		if r.DstPortStart != nil {
			m["dst_port_start"] = *r.DstPortStart
		}
		if r.DstPortEnd != nil {
			m["dst_port_end"] = *r.DstPortEnd
		}
		if r.MinPktLength != nil {
			m["min_pkt_length"] = *r.MinPktLength
		}
		if r.MaxPktLength != nil {
			m["max_pkt_length"] = *r.MaxPktLength
		}
		if r.Offset != nil {
			m["offset"] = *r.Offset
		}
		if r.MatchBytes != nil {
			m["match_bytes"] = *r.MatchBytes
		}
		if r.Action != nil {
			m["action"] = *r.Action
		}
		result = append(result, m)
	}
	return result
}

func flattenDDoSReflectUdpPorts(ports []*zec.DdosReflectUdpPort) []interface{} {
	result := make([]interface{}, 0, len(ports))
	for _, p := range ports {
		m := map[string]interface{}{}
		if p.Port != nil {
			m["port"] = *p.Port
		}
		result = append(result, m)
	}
	return result
}

func flattenDDoSTrafficControl(tc *zec.DdosTrafficControl) []interface{} {
	if tc == nil {
		return []interface{}{}
	}
	m := map[string]interface{}{}
	if tc.BpsEnabled != nil {
		m["bps_enabled"] = *tc.BpsEnabled
	}
	if tc.Bps != nil {
		m["bps"] = int(*tc.Bps)
	}
	if tc.PpsEnabled != nil {
		m["pps_enabled"] = *tc.PpsEnabled
	}
	if tc.Pps != nil {
		m["pps"] = int(*tc.Pps)
	}
	if tc.SynBPSEnabled != nil {
		m["syn_bps_enabled"] = *tc.SynBPSEnabled
	}
	if tc.SynBPS != nil {
		m["syn_bps"] = int(*tc.SynBPS)
	}
	if tc.SynPPSEnabled != nil {
		m["syn_pps_enabled"] = *tc.SynPPSEnabled
	}
	if tc.SynPPS != nil {
		m["syn_pps"] = int(*tc.SynPPS)
	}
	return []interface{}{m}
}
