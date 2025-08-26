
resource "zenlayercloud_zec_global_vpc" "foo" {
  name = "example"
  cidr_block = "10.0.0.0/24"
  enable_ipv6 = true
}
