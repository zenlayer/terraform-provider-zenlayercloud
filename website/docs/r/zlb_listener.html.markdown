---
subcategory: "Zenlayer Load Balancing(ZLB)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_zlb_listener"
sidebar_current: "docs-zenlayercloud-resource-zlb_listener"
description: |-
  Provide a resource to create a ZLB listener.
---

# zenlayercloud_zlb_listener

Provide a resource to create a ZLB listener.

## Example Usage

Prepare a ZLB instance

```hcl
variable "region" {
  default = "asia-east-1"
}

resource "zenlayercloud_zec_vpc" "foo" {
  name        = "example"
  cidr_block  = "10.0.0.0/16"
  enable_ipv6 = true
}

resource "zenlayercloud_zlb_instance" "zlb" {
  region_id = var.region
  vpc_id    = zenlayercloud_zec_vpc.foo.id
  zlb_name  = "example-5"
}
```

Create TCP Listener with health check enabled (single port)

```hcl
resource "zenlayercloud_zlb_listener" "listener" {
  zlb_id               = zenlayercloud_zlb_instance.zlb.id
  listener_name        = "tcp-listener"
  protocol             = "TCP"
  health_check_enabled = true
  port                 = "80"
  scheduler            = "mh"
  kind                 = "FNAT"
  health_check_type    = "TCP"
  persistent           = 300
}
```

is required and only one port is allowed.

```hcl
resource "zenlayercloud_zlb_listener" "all_ports_listener" {
  zlb_id               = zenlayercloud_zlb_instance.zlb.id
  listener_name        = "all-ports-listener"
  protocol             = "TCP"
  health_check_enabled = true
  port                 = "0" # All ports
  scheduler            = "mh"
  kind                 = "FNAT"
  health_check_type    = "TCP"
  health_check_port    = 8080 # Required when using all ports with TCP/HTTP_GET health check
}
```

## Argument Reference

The following arguments are supported:

* `listener_name` - (Required, String) The name of the load balancer listener.
* `port` - (Required, String) The port of listener. Multiple ports are separated by commas. When the port is a range, connect with -, for example: 10000-10005. Use '0' to represent all ports. The value range of the port is 0 to 65535. Please note that the port cannot overlap with other ports of the listener.
* `protocol` - (Required, String, ForceNew) The protocol of listener. Valid values: `TCP`, `UDP`.
* `zlb_id` - (Required, String, ForceNew) The ID of load balancer that the listener belongs to.
* `health_check_conn_timeout` - (Optional, Int) Connection timeout for health check. Valid values: `1` to `15`. `health_check_conn_timeout` takes effect only if `health_check_enabled` is set to true. Default is `2`.
* `health_check_delay_loop` - (Optional, Int) Interval between health checks. Measured in second. Valid values: `3` to `30`. `health_check_delay_loop` takes effect only if `health_check_enabled` is set to true. Default is `3`.
* `health_check_delay_try` - (Optional, Int) Health check delay try time.Valid values: `1` to `15`. `health_check_delay_try` takes effect only if `health_check_enabled` is set to true. Default is `2`.
* `health_check_enabled` - (Optional, Bool) Indicates whether health check is enabled. Default is `true`. When health check is disabled, other health check parameter can't be set.
* `health_check_fail_open` - (Optional, Bool) Indicates whether to enable fail-open for health check. When enabled, traffic will be sent to backend servers even if health checks fail. Default is `false`.
* `health_check_http_get_url` - (Optional, String) HTTP request URL for health check. The value should start with '/', Default is `/`.
* `health_check_http_status_code` - (Optional, Int) HTTP status code for health check. Required when `check_type` is `HTTP_GET`. Valid values: `100` to `599`.
* `health_check_port` - (Optional, Int) Health check port. Valid values: `1` to `65535`. **When the listener is configured with all ports (port = '0') and `health_check_type` is `TCP` or `HTTP_GET`, this parameter is required and only one port is allowed.** For other cases, defaults to the listener's port.
* `health_check_retry` - (Optional, Int) Number of retry attempts for health check. Valid values: `1` to `5`. `health_check_retry` takes effect only if `health_check_enabled` is set to true. Default is `2`.
* `health_check_type` - (Optional, String) Health check protocols. Valid values: `PING_CHECK`, `TCP`, `HTTP_GET`.
* `idle_timeout` - (Optional, Int) Idle timeout for data connections in seconds. If no request is received within this timeout, the LB temporarily disconnects until the next request re-establishes a new connection. Must be greater than 0. Default is `300`.
* `kind` - (Optional, String) Forwarding mode of the listener. Valid values: `DR`(stands for Direct Routing), `FNAT`(stands for Full NAT), `DNAT`(stands for Destination NAT). Default is `FNAT`.
* `persistent` - (Optional, Int) Session persistence duration in seconds. Set to 0 or omit this field to disable session affinity. When set to a positive value, the load balancer will maintain session affinity for the specified duration.
* `scheduler` - (Optional, String) Scheduling algorithm of the listener. Valid values: `mh`, `rr`, `wrr`, `lc`, `wlc`, `sh`, `dh`. Default value: `mh`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the resource.
* `create_time` - Create time of the listener.


## Import

ZLB listener can be imported, e.g.

```
$ terraform import zenlayercloud_zlb_listener.listener zlb-id:listener-id
```

