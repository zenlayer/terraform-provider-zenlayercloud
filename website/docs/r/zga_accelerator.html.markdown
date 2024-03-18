---
subcategory: "Zenlayer Global Accelerator(ZGA)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_zga_accelerator"
sidebar_current: "docs-zenlayercloud-resource-zga_accelerator"
description: |-
  Provides a accelerator resource.
---

# zenlayercloud_zga_accelerator

Provides a accelerator resource.

~> **NOTE:** Only L4 listener can be configured when domain is null.

~> **NOTE:** The Domain is not allowed to be the same as origin, otherwise a loop will be formed, making acceleration unusable.

## Example Usage

```hcl
resource "zenlayercloud_zga_certificate" "default" {
  certificate = <<EOF

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
  charge_type      = "ByTrafficPackage"
  domain           = "test.com"
  relate_domains   = ["a.test.com"]
  origin_region_id = "DE"
  origin           = ["10.10.10.10"]
  backup_origin    = ["10.10.10.14"]
  certificate_id   = resource.zenlayercloud_zga_certificate.default.id
  accelerate_regions {
    accelerate_region_id = "KR"
  }
  accelerate_regions {
    accelerate_region_id = "US"
  }
  l4_listeners {
    protocol        = "udp"
    port_range      = "53/54"
    back_port_range = "53/54"
  }
  l4_listeners {
    port      = 80
    back_port = 80
    protocol  = "tcp"
  }
  l7_listeners {
    port          = 443
    back_port     = 80
    protocol      = "https"
    back_protocol = "http"
  }
  l7_listeners {
    port_range      = "8888/8890"
    back_port_range = "8888/8890"
    protocol        = "http"
    back_protocol   = "http"
  }
  protocol_opts {
    websocket = true
    gzip      = false
  }
  access_control {
    enable = true
    rules {
      listener  = "https:443"
      directory = "/"
      policy    = "deny"
      cidr_ip   = ["10.10.10.10"]
    }
    rules {
      listener  = "udp:53/54"
      directory = "/"
      policy    = "accept"
      cidr_ip   = ["10.10.10.11/8"]
    }
  }
}
```

## Argument Reference

The following arguments are supported:

* `accelerate_regions` - (Required, List) Accelerate region of the accelerator.
* `origin_region_id` - (Required, String) ID of the orgin region. Modification is not supported.
* `origin` - (Required, Set: [`String`]) Endpoints of the origin. Only one endpoint is allowed to be configured, when the endpoint is CNAME.
* `accelerator_name` - (Optional, String) The name of accelerator. The max length of accelerator name is 64.
* `access_control` - (Optional, List) Access control of the accelerator.
* `backup_origin` - (Optional, Set: [`String`]) Backup endpoint of the origin. Backup orgin only be configured when origin configured with IP. Only one back endpoint is allowed to be configured, when the back endpoint is CNAME.
* `certificate_id` - (Optional, String) The certificate of the accelerator. Required when exist https protocol accelerate.
* `charge_type` - (Optional, String) The charge type of the accelerator. The default charge type of the account will be used. Modification is not supported. Valid values are `ByTrafficPackage`, `ByBandwidth95`, `ByBandwidth`, `ByTraffic`.
* `domain` - (Optional, String) Main domain of the accelerator. Required when L7 http or https accelerate, globally unique and no duplication is allowed. Supports generic domain names, like: *.zenlayer.com.
* `health_check` - (Optional, List) Health check of the accelerator.
* `l4_listeners` - (Optional, Set) L4 listeners of the accelerator.
* `l7_listeners` - (Optional, Set) L7 listeners of the accelerator.
* `protocol_opts` - (Optional, List) Protocol opts of the accelerator.
* `relate_domains` - (Optional, Set: [`String`]) Relate domains of the accelerator. Globally unique and no duplication is allowed. The max length of relate domains is 10.
* `resource_group_id` - (Optional, String) The resource group id the accelerator belongs to, default to Default Resource Group. Modification is not supported.

The `accelerate_regions` object supports the following:

* `accelerate_region_id` - (Required, String) ID of the accelerate region.
* `bandwidth` - (Optional, Int) Bandwidth limit of the accelerate region. Exceeding the account speed limit is not allowed. Unit: Mbps.
* `vip` - (Optional, String) Virtual IP the accelerate region. Modification is not supported.

The `access_control` object supports the following:

* `enable` - (Required, Bool) Whether to enable access control. Default is `false`.
* `rules` - (Optional, Set) Rules of the access control.

The `health_check` object supports the following:

* `enable` - (Required, Bool) Whether to enable health check. If the enable is `false`, the alarm will be set to `false` and the port will be cleared.
* `alarm` - (Optional, Bool) Whether to enable alarm. Default is `false`.
* `port` - (Optional, Int) The port of health check.

The `l4_listeners` object supports the following:

* `protocol` - (Required, String) The protocol of the l4 listener. Valid values: `tcp`, `udp`.
* `back_port_range` - (Optional, String) The Return-to-origin port range of the l4 listener. Use a slash (/) to separate the starting and ending ports, like: 1/200.
* `back_port` - (Optional, Int) The Return-to-origin port of the l4 listener.
* `port_range` - (Optional, String) The port range of the l4 listener. Only port or portRange can be configured. Use a slash (/) to separate the starting and ending ports, like: 1/200. The max range: 300.
* `port` - (Optional, Int) The port of the l4 listener. Only port or portRange can be configured, and duplicate ports are not allowed.

The `l7_listeners` object supports the following:

* `back_protocol` - (Required, String) The Return-to-origin protocol of the l7 listener. Valid values: http and https. The default is equal to protocol.
* `protocol` - (Required, String) The protocol of the l4 listener. Valid values: `http`, `https`.
* `back_port_range` - (Optional, String) The Return-to-origin port range of the l7 listener. Use a slash (/) to separate the starting and ending ports, like: 1/200.
* `back_port` - (Optional, Int) The Return-to-origin port of the l7 listener.
* `host` - (Optional, String) The Return-to-origin host of the l7 listener.
* `port_range` - (Optional, String) The port range of the l7 listener. Only port or portRange can be configured. Use a slash (/) to separate the starting and ending ports, like: 1/200. The max range: 300.
* `port` - (Optional, Int) The port of the l7 listener. Only port or portRange can be configured, and duplicate ports are not allowed.

The `protocol_opts` object supports the following:

* `gzip` - (Optional, Bool) Whether to enable gzip. Default is `false`.
* `proxy_protocol` - (Optional, Bool) Whether to enable proxyProtocol. Default is `false`.
* `toa_value` - (Optional, Int) TOA verison. Default is `253`.
* `toa` - (Optional, Bool) Whether to enable TOA. Default is `false`.
* `websocket` - (Optional, Bool) Whether to enable websocket. Default is `false`.

The `rules` object supports the following:

* `cidr_ip` - (Required, Set) The cidr ip of the rule.
* `listener` - (Required, String) The listener of the rule. Valid values are `$protocol:$port`, `$protocol:$portRange`, `all`.
* `policy` - (Required, String) The policy of the rule. Valid values are `accept`, `deny`.
* `directory` - (Optional, String) The directory of the rule. Not configurable with L4 listener. Default is `/`. Wildcards supported: *.
* `note` - (Optional, String) The note of the rule.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the resource.
* `accelerator_status` - Status of the accelerator. Values are `Accelerating`, `NotAccelerate`, `Deploying`, `StopAccelerate`, `AccelerateFailure`.
* `cname` - Cname of the accelerator.


## Import

Accelerator can be imported using the id, e.g.

```
terraform import zenlayercloud_zga_accelerator.default acceleratorId
```

