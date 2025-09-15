
resource "zenlayercloud_zec_vpc" "foo" {
  name = "example"
  cidr_block = "10.0.0.0/16"
  enable_ipv6 = true
}

resource "zenlayercloud_zec_vpc_security_group_attachment" "foo" {
  vpc_id = zenlayercloud_zec_vpc.foo.id
  security_group_id = "1304682049596034008"
}

# Create subnet (IPv4 IP stack)
resource "zenlayercloud_zec_subnet" "ipv4" {
  vpc_id = zenlayercloud_zec_vpc.foo.id
  region_id	 = var.region
  name       = "test-subnet"
  cidr_block = "10.0.0.0/24"
}

# Create a vNIC
resource "zenlayercloud_zec_vnic" "vnic" {
  subnet_id = zenlayercloud_zec_subnet.ipv4.id
  name = "example"
}

resource "zenlayercloud_zec_vnic_ipv4" "vnic_ipv4" {
  vnic_id = zenlayercloud_zec_vnic.vnic.id
  secondary_private_ip_count = 1
}

data "zenlayercloud_zec_vnics" "foo" {
  ids =  [zenlayercloud_zec_vnic.vnic.id]
}

output "out" {
  value = data.zenlayercloud_zec_vnics.foo
}

#resource "zenlayercloud_zec_vnic_attachment" "example" {
#  vnic_id = "1475945014486896967"
#  instance_id = "1478001766933994003"
#  #  name = "example"
#  #  id = "1450531612340006968"
#}
