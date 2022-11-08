package zenlayercloud

const (
	defaultPassword           = "Zenlayer+1"
	defaultHostname           = "tf-ci-test"
	defaultInstanceType       = "S8I"
	defaultZoneId             = "SEL-A"
	defaultInternetChargeType = BmcInternetChargeTypeBandwidth
)

const defaultVariable = `

variable "number" {
  default = "1"
}

variable "password" {
  default = "` + defaultPassword + `"
}

variable "availability_zone" {
  default = "` + defaultZoneId + `"
}

variable "instance_type_id" {
  default = "` + defaultInstanceType + `"
}

variable "hostname" {
  default = "` + defaultHostname + `"
}

variable "image_id" {
  default = null
}

variable "internet_charge_type" {
  default = "` + defaultInternetChargeType + `"
}

`
