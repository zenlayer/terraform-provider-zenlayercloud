Provide a resource to create a ZLB listener.

Example Usage

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

Create TCP Listener with health check enabled

```hcl
resource "zenlayercloud_zlb_listener" "listener" {
  zlb_id               = zenlayercloud_zlb_instance.zlb.id
  listener_name        = "tcp-listener"
  protocol             = "TCP"
  health_check_enabled = true
  port                 = 80
  scheduler            = "mh"
  kind                 = "FNAT"
  health_check_type    = "TCP"
}
```

Import

ZLB listener can be imported, e.g.

```
$ terraform import zenlayercloud_zlb_listener.listener zlb-id:listener-id
```
