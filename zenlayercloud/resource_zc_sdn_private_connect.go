/*
Provides a resource to manage layer 2 private connect.

Example Usage

```hcl

resource "zenlayercloud_sdn_private_connect" "aws-port-test" {
  connect_name       	= "Test"
  connect_bandwidth    	= 20
  endpoints	{
    port_id = "xxxxxxxxx"
    endpoint_type = "TENCENT"
    vlan_id = "1019"
	}
  endpoints {
    datacenter = "SOF1"
    cloud_region = "eu-west-1"
    cloud_account = "123412341234"
    endpoint_type = "AWS"
    vlan_id = "1457"
  }

resource "zenlayercloud_sdn_private_connect" "aws-tencent-test" {
  connect_name       	= "Test"
  connect_bandwidth    	= 20
  endpoints	{
    datacenter = "HKG2"
    cloud_region = "ap-hongkong-a-kc"
    cloud_account = "123412341234"
    endpoint_type = "TENCENT"
    vlan_id = "1019"
	}
  endpoints {
    datacenter = "SOF1"
    cloud_region = "eu-west-1"
    cloud_account = "123412341234"
    endpoint_type = "AWS"
    vlan_id = "1457"
  }
```

Import

Private Connect can be imported, e.g.

```
$ terraform import zenlayercloud_sdn_private_connect.foo xxxxxx
```
*/
package zenlayercloud

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
	"github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/common"
	sdn "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/sdn20230830"
	"time"
)

func resourceZenlayerCloudPrivateConnect() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceZenlayerCloudPrivateConnectCreate,
		ReadContext:   resourceZenlayerCloudPrivateConnectRead,
		UpdateContext: resourceZenlayerCloudPrivateConnectUpdate,
		DeleteContext: resourceZenlayerCloudPrivateConnectDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"connect_name": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "Terraform-Private-Connect",
				ValidateFunc: validation.StringLenBetween(1, 255),
				Description:  "The private connect name. Up to 255 characters in length are allowed.",
			},
			"connect_bandwidth": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      1,
				ValidateFunc: validation.IntBetween(1, 500),
				Description:  "The bandwidth of private connect. Valid range: [1,500]. Unit: Mbps.",
			},
			"endpoints": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "Access points of private connect. Length must be equal to 2.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"port_id": {
							Type:        schema.TypeString,
							Optional:    true,
							ForceNew:    true,
							Description: "The ID of the port. This value is required when `endpoint_type` is `PORT`.",
						},
						"cloud_region": {
							Type:     schema.TypeString,
							Optional: true,
							//ConflictsWith: []string{"endpoints.port_id"},
							ForceNew:    true,
							Description: "Region of cloud access point. This value is available only when `endpoint_type` within cloud type (AWS, GOOGLE and TENCENT).",
						},
						"cloud_account": {
							Type:     schema.TypeString,
							Optional: true,
							//ConflictsWith: []string{"endpoints.port_id"},
							ForceNew:    true,
							Description: "The account of public cloud access point. If cloud type is GOOGLE, the value is google pairing key. This value is available only when `endpoint_type` within cloud type (AWS, GOOGLE and TENCENT).",
						},
						"endpoint_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the access point.",
						},
						"vlan_id": {
							Type:        schema.TypeInt,
							Optional:    true,
							ForceNew:    true,
							Description: "VLAN ID of the access point. Value range: from 1 to 4096.",
						},
						"endpoint_type": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringInSlice(ENDPOINT_TYPES, false),
							Description:  "The type of the access point, Valid values: PORT,AWS,TENCENT and GOOGLE.",
						},
						"datacenter": {
							Type:        schema.TypeString,
							Optional:    true,
							ForceNew:    true,
							Description: "The ID of data center.",
						},
						"connectivity_status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Network connectivity state. ACTIVE means the network is connected. DOWN which means not connected.",
						},
					},
				},
			},
			"force_delete": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Indicate whether to force delete the private connect. Default is `false`. If set true, the private connect will be permanently deleted instead of being moved into the recycle bin.",
			},
			"connectivity_status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Network connectivity state. ACTIVE means the network is connected. DOWN which means not connected.",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The business state of private connect.",
			},
			"resource_group_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The resource group ID.",
			},
			"resource_group_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The Name of resource group.",
			},
			"create_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Create time of the private connect.",
			},
			"expired_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Expired time of the private connect.",
			},
		},
	}
}

func resourceZenlayerCloudPrivateConnectDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common2.LogElapsed(ctx, "resource.zenlayercloud_sdn_private_connect.delete")()

	// force_delete: terminate and then delete
	forceDelete := d.Get("force_delete").(bool)

	sdnService := SdnService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	connectId := d.Id()
	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
		errRet := sdnService.DeletePrivateConnectById(ctx, connectId)
		if errRet != nil {
			return common2.RetryError(ctx, errRet)
		}
		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	notExist := false

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
		connect, errRet := sdnService.DescribePrivateConnectById(ctx, connectId)
		if errRet != nil {
			return common2.RetryError(ctx, errRet, common2.InternalServerError)
		}
		if connect == nil {
			notExist = true
			return nil
		}

		if connect.PrivateConnectStatus == SdnStatusRecycle {
			//in recycling
			return nil
		}

		if IsOperating(connect.PrivateConnectStatus) {
			return resource.RetryableError(fmt.Errorf("waiting for connect %s recycling, current status: %s", connect.PrivateConnectId, connect.PrivateConnectStatus))
		}

		return resource.NonRetryableError(fmt.Errorf("connect status is not recycle, current status %s", connect.PrivateConnectStatus))
	})

	if err != nil {
		return diag.FromErr(err)
	}

	if notExist || !forceDelete {
		return nil
	}

	tflog.Debug(ctx, "Releasing private connect ...", map[string]interface{}{
		"connectId": connectId,
	})

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
		errRet := sdnService.DestroyPrivateConnect(ctx, connectId)
		if errRet != nil {

			ee, ok := errRet.(*common.ZenlayerCloudSdkError)
			if !ok {
				return common2.RetryError(ctx, errRet)
			}
			if ee.Code == "INVALID_PRIVATE_CONNECT_NOT_FOUND" || ee.Code == common2.ResourceNotFound {
				// connect doesn't exist
				return nil
			}
			return common2.RetryError(ctx, errRet, common2.InternalServerError)
		}
		return nil
	})

	return nil
}

func resourceZenlayerCloudPrivateConnectUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common2.LogElapsed(ctx, "resource.zenlayercloud_sdn_private_connect.update")()

	sdnService := SdnService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}
	connectId := d.Id()
	if d.HasChanges("connect_name") {
		err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate)-time.Minute, func() *resource.RetryError {
			connectName := d.Get("connect_name").(string)
			err := sdnService.ModifyPrivateConnectName(ctx, connectId, connectName)
			if err != nil {
				return common2.RetryError(ctx, err, common2.InternalServerError, common.NetworkError)
			}
			return nil
		})

		if err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChanges("connect_bandwidth") {
		err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate)-time.Minute, func() *resource.RetryError {
			bandwidth := d.Get("connect_bandwidth").(int)
			err := sdnService.ModifyPrivateConnectBandwidth(ctx, connectId, bandwidth)
			if err != nil {
				return common2.RetryError(ctx, err, common2.InternalServerError, common.NetworkError)
			}
			return nil
		})

		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceZenlayerCloudPrivateConnectRead(ctx, d, meta)
}

func resourceZenlayerCloudPrivateConnectCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common2.LogElapsed(ctx, "resource.zenlayercloud_sdn_private_connect.create")()

	sdnService := SdnService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}
	endpoints := d.Get("endpoints").([]interface{})
	if len(endpoints) != 2 {
		return diag.Errorf("The size of endpoint must equal to 2.")
	}

	request := sdn.NewCreatePrivateConnectRequest()
	request.PrivateConnectName = d.Get("connect_name").(string)
	request.BandwidthMbps = d.Get("connect_bandwidth").(int)
	request.EndpointA = parseEndpoint(endpoints[0])
	request.EndpointZ = parseEndpoint(endpoints[1])
	privateConnectId := ""

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		response, err := meta.(*connectivity.ZenlayerCloudClient).WithSdnClient().CreatePrivateConnect(request)
		if err != nil {
			tflog.Info(ctx, "Fail to create private connect.", map[string]interface{}{
				"action":  request.GetAction(),
				"request": common2.ToJsonString(request),
				"err":     err.Error(),
			})
			return common2.RetryError(ctx, err)
		}

		tflog.Info(ctx, "Create private connect success", map[string]interface{}{
			"action":   request.GetAction(),
			"request":  common2.ToJsonString(request),
			"response": common2.ToJsonString(response),
		})

		if response.Response.PrivateConnectId == "" {
			err = fmt.Errorf("connectId is nil")
			return resource.NonRetryableError(err)
		}
		privateConnectId = response.Response.PrivateConnectId

		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(privateConnectId)

	stateConf := &resource.StateChangeConf{
		Pending: []string{
			SdnStatusCreating,
		},
		Target: []string{
			SdnStatusRunning,
		},
		Refresh:        sdnService.PrivateConnectStateRefreshFunc(ctx, privateConnectId, []string{}),
		Timeout:        d.Timeout(schema.TimeoutCreate) - time.Minute,
		Delay:          10 * time.Second,
		MinTimeout:     5 * time.Second,
		NotFoundChecks: 3,
	}

	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return diag.FromErr(fmt.Errorf("error waiting for connect (%s) to be created: %v", d.Id(), err))
	}

	return resourceZenlayerCloudPrivateConnectRead(ctx, d, meta)
}

func parseEndpoint(endpointParam interface{}) sdn.CreateEndpointParam {
	c := sdn.CreateEndpointParam{}

	v := endpointParam.(map[string]interface{})
	endpointType := v["endpoint_type"].(string)
	if endpointType == POINT_TYPE_PORT {
		portId := v["port_id"].(string)
		c.PortId = portId
	} else {
		cloudRegionId := v["cloud_region"].(string)
		cloudAccountId := v["cloud_account"].(string)
		c.CloudRegionId = cloudRegionId
		c.CloudAccountId = cloudAccountId
		c.CloudType = endpointType
	}
	dcId := v["datacenter"].(string)
	c.DcId = dcId
	c.VlanId = v["vlan_id"].(int)
	return c
}

func resourceZenlayerCloudPrivateConnectRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common2.LogElapsed(ctx, "resource.zenlayercloud_sdn_private_connect.read")()

	var diags diag.Diagnostics

	connectId := d.Id()

	sdnService := SdnService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	var connect *sdn.PrivateConnect
	var errRet error

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead)-time.Minute, func() *resource.RetryError {
		connect, errRet = sdnService.DescribePrivateConnectById(ctx, connectId)
		if errRet != nil {
			return common2.RetryError(ctx, errRet)
		}
		if connect != nil && IsOperating(connect.PrivateConnectStatus) {
			return resource.RetryableError(fmt.Errorf("waiting for private connect %s operation, current status: %s", connect.PrivateConnectId, connect.PrivateConnectStatus))
		}

		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	if connect == nil {
		d.SetId("")
		tflog.Info(ctx, "private connect not exist", map[string]interface{}{
			"connectId": connectId,
		})
		return nil
	}

	// connect info

	_ = d.Set("connect_name", connect.PrivateConnectName)
	_ = d.Set("connect_bandwidth", connect.BandwidthMbps)
	_ = d.Set("connectivity_status", connect.ConnectivityStatus)
	_ = d.Set("connect_status", connect.PrivateConnectStatus)
	_ = d.Set("resource_group_id", connect.ResourceGroupId)
	_ = d.Set("resource_group_name", connect.ResourceGroupName)
	_ = d.Set("create_time", connect.CreateTime)
	_ = d.Set("expired_time", connect.ExpiredTime)

	var res = make([]interface{}, 0, 2)
	res = append(res, mappingConnectEndpoint(&connect.EndpointA))
	res = append(res, mappingConnectEndpoint(&connect.EndpointZ))
	_ = d.Set("endpoints", res)

	return diags
}
