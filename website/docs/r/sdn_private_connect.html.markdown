---
subcategory: "Cloud Networking(SDN)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_sdn_private_connect"
sidebar_current: "docs-zenlayercloud-resource-sdn_private_connect"
description: |-
  Provides a resource to manage layer 2 private connect.
---

# zenlayercloud_sdn_private_connect

Provides a resource to manage layer 2 private connect.

## Example Usage

```hcl
resource "zenlayercloud_sdn_private_connect" "aws-port-test" {
  connect_name      = "Test"
  connect_bandwidth = 20
  endpoints {
    port_id       = "xxxxxxxxx"
    endpoint_type = "TENCENT"
    vlan_id       = "1019"
  }
  endpoints {
    datacenter    = "SOF1"
    cloud_region  = "eu-west-1"
    cloud_account = "123412341234"
    endpoint_type = "AWS"
    vlan_id       = "1457"
  }

  resource "zenlayercloud_sdn_private_connect" "aws-tencent-test" {
    connect_name      = "Test"
    connect_bandwidth = 20
    endpoints {
      datacenter    = "HKG2"
      cloud_region  = "ap-hongkong-a-kc"
      cloud_account = "123412341234"
      endpoint_type = "TENCENT"
      vlan_id       = "1019"
    }
    endpoints {
      datacenter    = "SOF1"
      cloud_region  = "eu-west-1"
      cloud_account = "123412341234"
      endpoint_type = "AWS"
      vlan_id       = "1457"
    }
```

## Argument Reference

The following arguments are supported:

* `endpoints` - (Required, List) Access points of private connect. Length must be equal to 2.
* `connect_bandwidth` - (Optional, Int) The bandwidth of private connect. Valid range: [1,500]. Unit: Mbps.
* `connect_name` - (Optional, String) The private connect name. Up to 255 characters in length are allowed.
* `force_delete` - (Optional, Bool) Indicate whether to force delete the private connect. Default is `false`. If set true, the private connect will be permanently deleted instead of being moved into the recycle bin.

The `endpoints` object supports the following:

* `endpoint_type` - (Required, String, ForceNew) The type of the access point, Valid values: PORT,AWS,TENCENT and GOOGLE.
* `cloud_account` - (Optional, String, ForceNew) The account of public cloud access point. If cloud type is GOOGLE, the value is google pairing key. This value is available only when `endpoint_type` within cloud type (AWS, GOOGLE and TENCENT).
* `cloud_region` - (Optional, String, ForceNew) Region of cloud access point. This value is available only when `endpoint_type` within cloud type (AWS, GOOGLE and TENCENT).
* `datacenter` - (Optional, String, ForceNew) The ID of data center.
* `port_id` - (Optional, String, ForceNew) The ID of the port. This value is required when `endpoint_type` is `PORT`.
* `vlan_id` - (Optional, Int, ForceNew) VLAN ID of the access point. Value range: from 1 to 4096.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the resource.
* `connectivity_status` - Network connectivity state. ACTIVE means the network is connected. DOWN which means not connected.
* `create_time` - Create time of the private connect.
* `expired_time` - Expired time of the private connect.
* `resource_group_id` - The resource group ID.
* `resource_group_name` - The Name of resource group.
* `status` - The business state of private connect.


## Import

Private Connect can be imported, e.g.

```
$ terraform import zenlayercloud_sdn_private_connect.foo xxxxxx
```

