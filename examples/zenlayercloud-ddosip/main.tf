data "zenlayercloud_bmc_ddos_ips" "foo" {

}

output "eip" {
  value = data.zenlayercloud_bmc_ddos_ips.foo.ip_list
}

