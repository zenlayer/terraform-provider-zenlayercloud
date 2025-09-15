Provides an eip resource associated with resource including vNIC, ZLB.

Example Usage

Bind EIP to vNIC

```hcl
variable "region" {
  default = "asia-east-1"
}

resource "zenlayercloud_zec_vpc" "foo" {
  name        = "example"
  cidr_block  = "10.0.0.0/16"
  enable_ipv6 = true
}

# Create subnet (IPv4 IP stack)
resource "zenlayercloud_zec_subnet" "ipv4" {
  vpc_id     = zenlayercloud_zec_vpc.foo.id
  region_id  = var.region
  name       = "test-subnet"
  cidr_block = "10.0.0.0/24"
}

resource "zenlayercloud_zec_vnic" "vnic" {
  subnet_id = zenlayercloud_zec_subnet.ipv4.id
  name      = "example-vnic"
}

resource "zenlayercloud_zec_eip" "eip" {
  region_id            = var.region
  name                 = "example"
  ip_network_type      = "BGPLine"
  internet_charge_type = "ByBandwidth"
  bandwidth            = 10
}

resource "zenlayercloud_zec_eip_association" "eip_association" {
  eip_id             = zenlayercloud_zec_eip.eip.id
  associated_id      = zenlayercloud_zec_vnic.vnic.id
  associated_type    = "NIC"
  private_ip_address = zenlayercloud_zec_vnic.vnic.primary_ipv4
}
```

Bind EIP to ZLB
```hcl
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


resource "zenlayercloud_zlb_listener" "tcp_listener" {
  zlb_id               = zenlayercloud_zlb_instance.zlb.id
  listener_name        = "tcp-listener"
  protocol             = "TCP"
  health_check_enabled = true
  port                 = 8080
  scheduler            = "mh"
  kind                 = "FNAT"
  health_check_type    = "TCP"
}

resource "zenlayercloud_zec_eip" "eip" {
  region_id = var.region
  name = "example"
  ip_network_type = "BGPLine"
  internet_charge_type = "ByBandwidth"
  bandwidth = 10
}

#Associate EIP with LB
resource "zenlayercloud_zec_eip_association" "eip_association" {
  eip_id             = zenlayercloud_zec_eip.eip.id
  associated_id      = zenlayercloud_zlb_instance.zlb.id
  associated_type    = "LB"
}

```

# Import

EIP association can be imported, e.g.

```
$ terraform import zenlayercloud_zec_eip_association.eip_association eip-id:associated-id:associated-type
```
