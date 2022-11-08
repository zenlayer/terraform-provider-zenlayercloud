#data "zenlayercloud_bmc_eip_zones" "eip_zone" {
#
#}
variable "availability_zone" {
  default = "SEL-A"
}

data "zenlayercloud_bmc_eips" "foo" {
  availability_zone = var.availability_zone
}

