Provides a resource to bind a ZEC instance to a high-availability virtual IP (HaVip).

The bound instance's network interface must reside in the same subnet as the HaVip.

Example Usage

```hcl
variable "region" {
  default = "asia-east-1"
}

resource "zenlayercloud_zec_vpc" "foo" {
  name       = "example"
  cidr_block = "10.0.0.0/16"
}

resource "zenlayercloud_zec_subnet" "subnet" {
  vpc_id     = zenlayercloud_zec_vpc.foo.id
  region_id  = var.region
  name       = "example-subnet"
  cidr_block = "10.0.0.0/24"
}

resource "zenlayercloud_zec_havip" "havip" {
  subnet_id = zenlayercloud_zec_subnet.subnet.id
  name      = "example-havip"
}

resource "zenlayercloud_zec_instance" "instance" {
  # ... omit instance configuration
}

resource "zenlayercloud_zec_havip_association" "binding" {
  ha_vip_id   = zenlayercloud_zec_havip.havip.id
  instance_id = zenlayercloud_zec_instance.instance.id
}
```

Import

HaVip association can be imported using the id (ha_vip_id:instance_id), e.g.

```
$ terraform import zenlayercloud_zec_havip_association.binding havip-id:instance-id
```
