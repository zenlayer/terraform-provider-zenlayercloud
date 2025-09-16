Provides a resource to assign secondary private ipv4(s) from subnet to vNIC.

~> **NOTE:** The vNIC must contains IPv4 ip stack type

Example Usage

Prepare a vNIC
```hcl
variable "region" {
  default = "asia-east-1"
}

resource "zenlayercloud_zec_vpc" "foo" {
  name = "example"
  cidr_block = "10.0.0.0/16"
  enable_ipv6 = true
}

resource "zenlayercloud_zec_subnet" "ipv4" {
  vpc_id = zenlayercloud_zec_vpc.foo.id
  region_id	 = var.region
  name       = "test-subnet"
  cidr_block = "10.0.0.0/24"
}

resource "zenlayercloud_zec_vnic" "vnic" {
  subnet_id = zenlayercloud_zec_subnet.ipv4.id
  name = "example-vnic"
}

```

Assign secondary private ipv4s to vNIC by `secondary_private_ip_count`
```hcl
resource "zenlayercloud_zec_vnic_ipv4" "foo" {
  vnic_id 	 = zenlayercloud_zec_vnic.vnic.id
  secondary_private_ip_count  			 =  3
}
```

Assign secondary private ipv4 to vNIC by `secondary_private_ip_addresses`
```hcl
resource "zenlayercloud_zec_vnic_ipv4" "foo" {
  vnic_id 	 = ""
  secondary_private_ip_addresses  			 =  ["10.0.0.3", "10.0.0.4"]
}
```

Import

Disk instance can be imported, e.g.

```
$ terraform import zenlayercloud_zec_vnic_ipv4.test vnic-id
```
