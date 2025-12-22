Provide a resource to associate DHCP Options Set to subnet.

Example Usage

```hcl
# Create a DHCP Options Set
resource "zenlayercloud_zec_dhcp_options_set" "example_dhcp" {
  domain_name_servers = ["8.8.8.8", "8.8.4.4"]
  description = "example dhcp options set"
}

# Create a VPC and subnet
resource "zenlayercloud_zec_vpc" "foo" {
  name        = "example"
  cidr_block  = "10.0.0.0/24"
  enable_ipv6 = true
  mtu         = 1300
}

resource "zenlayercloud_zec_subnet" "example_subnet" {
  vpc_id     = zenlayercloud_zec_vpc.foo.id
  region_id  = "asia-east-1"
  cidr_block = "10.0.0.0/24"
  ipv6_type  = "Public"
}

# Create a DHCP Options Set Attachment
resource "zenlayercloud_zec_dhcp_options_set_attachment" "example" {
  subnet_id = zenlayercloud_zec_subnet.example_subnet.id
  dhcp_options_set_id = zenlayercloud_zec_dhcp_options_set.example_dhcp.id
}
```

Import

DHCP Options Set Attachment instance can be imported, e.g.

```
$ terraform import zenlayercloud_zec_dhcp_options_set_attachment.example subnet-id:dhcp-options-set-id
```
