Provide a resource to Elastic IP.

Example Usage

Create en EIP billing by flat rate
```hcl
variable "region" {
	default = "asia-southeast-1"
}

resource "zenlayercloud_zec_eip" "eip" {
	region_id = var.region
	name = "example"
	ip_network_type = "BGPLine"
	internet_charge_type = "ByBandwidth"
	bandwidth = 10
}
```

Import

EIP instance can be imported, e.g.

```
$ terraform import zenlayercloud_zec_eip.eip eip-id
```
