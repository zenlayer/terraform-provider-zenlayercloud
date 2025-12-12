Provide a resource to create a Load balancer instance.

~> **NOTE:** When creating a load balancer instance, ensure that there is at least a subnet is IPv4 stack type under the VPC and that there are at least 2 available IP addresses for allocation within the subnet.

Example Usage

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

Import

ZLB instance can be imported, e.g.

```
$ terraform import zenlayercloud_zlb_instance.zlb zlb-id
```
