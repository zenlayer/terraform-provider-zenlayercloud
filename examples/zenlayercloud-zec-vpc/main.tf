
resource "zenlayercloud_zec_vpc" "foo" {
  name = "example"
  cidr_block = "10.0.0.0/16"
  enable_ipv6 = true
}

resource "zenlayercloud_zec_vpc_security_group_attachment" "foo" {
  vpc_id = zenlayercloud_zec_vpc.foo.id
  security_group_id = "1304682049596034008"
}

#Create subnet (IPv4 IP stack)
resource "zenlayercloud_zec_subnet" "ipv4" {
  vpc_id = zenlayercloud_zec_vpc.foo.id
  region_id	 = var.region
  name       = "test-subnet"
  cidr_block = "10.0.0.0/24"
}

#Create subnet (IPv4 & IPv6 IP stack)
resource "zenlayercloud_zec_subnet" "ipv4_ipv6" {
  vpc_id = zenlayercloud_zec_vpc.foo.id
  region_id	 = var.region
  name       = "IPv4-IPv6-Subnet"
  cidr_block = "10.0.3.0/24"
  ipv6_type = "Private"
}

#Create subnet (IPv6 IP stack) with Private IPv6
resource "zenlayercloud_zec_subnet" "ipv6_private" {
  vpc_id = zenlayercloud_zec_vpc.foo.id
  region_id	 = var.region
  name       = "IPv6-Private-Subnet"
  ipv6_type = "Private"
}

#Create subnet (IPv6 IP stack) with Public IPv6
resource "zenlayercloud_zec_subnet" "ipv6_public" {
  vpc_id = zenlayercloud_zec_vpc.foo.id
  region_id	 = var.region
  name       = "IPv6-Public-Subnet-1"
  ipv6_type = "Public"
}
#
