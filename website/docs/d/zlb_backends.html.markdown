---
subcategory: "Zenlayer Load Balancing(ZLB)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_zlb_backends"
sidebar_current: "docs-zenlayercloud-datasource-zlb_backends"
description: |-
  Use this data source to query backends for ZLB instance.
---

# zenlayercloud_zlb_backends

Use this data source to query backends for ZLB instance.

## Example Usage

Query all backend instances for ZLB

```hcl
data "zenlayercloud_zlb_backends" "all" {
  zlb_id = "<zlbId>"
}
```

Query backend instances by listener id

```hcl
data "zenlayercloud_zlb_instances" "foo" {
  zlb_id      = "<zlbId>"
  listener_id = "<listenerId>"
}
```

## Argument Reference

The following arguments are supported:

* `zlb_id` - (Required, String) The ID of load balancer that the backends belong to.
* `listener_id` - (Optional, String) The ID of the listener that the backends belong to.
* `result_output_file` - (Optional, String) Used to save results.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `backends` - An information list of backend servers. Each element contains the following attributes:
   * `backend_port` - Target port for request forwarding and health checks. If left empty, it will follow the listener's port configuration.
   * `instance_id` - ID of the backend server.
   * `listener_id` - The ID of the listener that the backend server belongs to.
   * `listener_name` - The name of the listener that the backend server belongs to.
   * `listener_port` - Listening port. Use commas (,) to separate multiple ports.Use a hyphen (-) to define a port range, e.g., 10000-10005.
   * `private_ip` - Private IP address of the network interface attached to the instance.
   * `protocol` - Protocol of the backend server.
   * `weight` - Forwarding weight of the backend server.


