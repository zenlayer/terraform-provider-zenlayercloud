/*
Provides a resource to manage datacenter port.

Example Usage

```hcl
resource "zenlayercloud_sdn_port" "foo" {
  name       			= "my_name"
  datacenter    		= "xxxxx-xxxxx-xxxxx"
  remarks				= "Test"
  port_type				= "1G"
  business_entity_name  = "John"
}
```

Import

Port can be imported, e.g.

```
$ terraform import zenlayercloud_sdn_port.foo xxxxxx
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

func resourceZenlayerCloudDcPorts() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceZenlayerCloudDcPortsCreate,
		ReadContext:   resourceZenlayerCloudDcPortsRead,
		UpdateContext: resourceZenlayerCloudDcPortsUpdate,
		DeleteContext: resourceZenlayerCloudDcPortsDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "Terraform-Port",
				ValidateFunc: validation.StringLenBetween(1, 255),
				Description:  "Port name. Up to 255 characters in length are allowed.",
			},
			"datacenter": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of data center.",
			},
			"remarks": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(2, 255),
				Description:  "Description of port.",
			},
			"port_type": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Type of port. eg. 1G/10G/40G.",
			},
			"port_charge_type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The charge type of port. Valid values: `PREPAID`, `POSTPAID`.",
			},
			"business_entity_name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(2, 255),
				Description:  "Your business entity name. The entity name to be used on the Letter of Authorization (LOA).",
			},
			"force_delete": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Indicate whether to force delete the port. Default is `false`. If set true, the port will be permanently deleted instead of being moved into the recycle bin.",
			},
			"datacenter_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The name of datacenter.",
			},
			"loa_status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The LOA state.",
			},
			"loa_url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The LOA URL address.",
			},
			"connect_status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The network connectivity state of port.",
			},
			"port_status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The business status of port.",
			},
			"create_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Create time of the port.",
			},
			"expired_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Expired time of the port.",
			},
		},
	}
}

func resourceZenlayerCloudDcPortsDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common2.LogElapsed(ctx, "resource.zenlayercloud_sdn_port.delete")()

	// force_delete: terminate and then delete
	forceDelete := d.Get("force_delete").(bool)

	sdnService := SdnService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	portId := d.Id()
	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
		errRet := sdnService.DeletePortById(ctx, portId)
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
		port, errRet := sdnService.DescribePortById(ctx, portId)
		if errRet != nil {
			return common2.RetryError(ctx, errRet, common2.InternalServerError)
		}
		if port == nil {
			notExist = true
			return nil
		}

		if port.PortStatus == SdnStatusRecycle {
			//in recycling
			return nil
		}

		if IsOperating(port.PortStatus) {
			return resource.RetryableError(fmt.Errorf("waiting for port %s recycling, current status: %s", port.PortId, port.PortStatus))
		}

		return resource.NonRetryableError(fmt.Errorf("port status is not recycle, current status %s", port.PortStatus))
	})

	if err != nil {
		return diag.FromErr(err)
	}

	if notExist || !forceDelete {
		return nil
	}

	tflog.Debug(ctx, "Releasing Port ...", map[string]interface{}{
		"portId": portId,
	})

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
		errRet := sdnService.DestroyPort(ctx, portId)
		if errRet != nil {

			ee, ok := errRet.(*common.ZenlayerCloudSdkError)
			if !ok {
				return common2.RetryError(ctx, errRet)
			}
			if ee.Code == "INVALID_PORT_NOT_FOUND" || ee.Code == common2.ResourceNotFound {
				// port doesn't exist
				return nil
			}
			return common2.RetryError(ctx, errRet, common2.InternalServerError)
		}
		return nil
	})

	return nil
}

func resourceZenlayerCloudDcPortsUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdnService := SdnService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}
	portId := d.Id()
	if d.HasChanges("name", "remarks", "business_entity_name") {
		err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate)-time.Minute, func() *resource.RetryError {
			portName := d.Get("name").(string)
			remarks := d.Get("remarks").(string)
			businessEntityName := d.Get("business_entity_name").(string)
			err := sdnService.ModifyPort(ctx, portId, portName, remarks, businessEntityName)
			if err != nil {
				return common2.RetryError(ctx, err, common2.InternalServerError, common.NetworkError)
			}
			return nil
		})

		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceZenlayerCloudDcPortsRead(ctx, d, meta)
}

func resourceZenlayerCloudDcPortsCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common2.LogElapsed(ctx, "resource.zenlayercloud_sdn_port.create")()

	sdnService := SdnService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	request := sdn.NewCreatePortRequest()
	request.PortName = d.Get("name").(string)
	request.DcId = d.Get("datacenter").(string)
	request.BusinessEntityName = d.Get("business_entity_name").(string)
	request.PortType = d.Get("port_type").(string)

	if v, ok := d.GetOk("remarks"); ok {
		request.PortRemarks = v.(string)
	}
	portId := ""

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		response, err := meta.(*connectivity.ZenlayerCloudClient).WithSdnClient().CreatePort(request)
		if err != nil {
			tflog.Info(ctx, "Fail to create port.", map[string]interface{}{
				"action":  request.GetAction(),
				"request": common2.ToJsonString(request),
				"err":     err.Error(),
			})
			return common2.RetryError(ctx, err)
		}

		tflog.Info(ctx, "Create port success", map[string]interface{}{
			"action":   request.GetAction(),
			"request":  common2.ToJsonString(request),
			"response": common2.ToJsonString(response),
		})

		if response.Response.PortId == "" {
			err = fmt.Errorf("portId is nil")
			return resource.NonRetryableError(err)
		}
		portId = response.Response.PortId

		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(portId)

	stateConf := &resource.StateChangeConf{
		Pending: []string{
			SdnStatusCreating,
		},
		Target: []string{
			SdnStatusRunning,
		},
		Refresh:        sdnService.PortStateRefreshFunc(ctx, portId, []string{}),
		Timeout:        d.Timeout(schema.TimeoutCreate) - time.Minute,
		Delay:          10 * time.Second,
		MinTimeout:     5 * time.Second,
		NotFoundChecks: 3,
	}

	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return diag.FromErr(fmt.Errorf("error waiting for port (%s) to be created: %v", d.Id(), err))
	}

	return resourceZenlayerCloudDcPortsRead(ctx, d, meta)
}

func resourceZenlayerCloudDcPortsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common2.LogElapsed(ctx, "resource.zenlayercloud_sdn_port.read")()

	var diags diag.Diagnostics

	portId := d.Id()

	sdnService := SdnService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	var portInfo *sdn.PortInfo
	var errRet error

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead)-time.Minute, func() *resource.RetryError {
		portInfo, errRet = sdnService.DescribePortById(ctx, portId)
		if errRet != nil {
			return common2.RetryError(ctx, errRet)
		}
		if portInfo != nil && IsOperating(portInfo.PortStatus) {
			return resource.RetryableError(fmt.Errorf("waiting for port %s operation, current status: %s", portInfo.PortId, portInfo.PortStatus))
		}

		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	if portInfo == nil {
		d.SetId("")
		tflog.Info(ctx, "port not exist", map[string]interface{}{
			"portId": portId,
		})
		return nil
	}

	// port info

	_ = d.Set("name", portInfo.PortName)
	_ = d.Set("port_type", portInfo.PortType)
	_ = d.Set("remarks", portInfo.PortRemarks)
	_ = d.Set("datacenter", portInfo.DcId)
	_ = d.Set("datacenter_name", portInfo.DcName)
	_ = d.Set("loa_status", portInfo.LoaStatus)
	_ = d.Set("loa_url", portInfo.LoaDownloadUrl)
	_ = d.Set("port_charge_type", portInfo.PortChargeType)
	_ = d.Set("business_entity_name", portInfo.BusinessEntityName)
	_ = d.Set("connect_status", portInfo.ConnectionStatus)
	_ = d.Set("port_status", portInfo.PortStatus)
	_ = d.Set("create_time", portInfo.CreatedTime)
	_ = d.Set("expired_time", portInfo.ExpiredTime)

	return diags
}
