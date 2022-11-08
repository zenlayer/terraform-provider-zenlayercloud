data "zenlayercloud_bmc_vpc_regions" "default_region" {

}

resource "zenlayercloud_bmc_vpc" "foo" {
  region     = data.zenlayercloud_bmc_vpc_regions.default_region.regions.0.id
  cidr_block = "10.0.0.0/26"
}