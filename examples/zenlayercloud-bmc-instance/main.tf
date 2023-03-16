data "zenlayercloud_bmc_zones" "default" {

}

data "zenlayercloud_bmc_instance_types" "default" {
  availability_zone = data.zenlayercloud_bmc_zones.default.zones.0.id
}

# Get a centos image which also supported to install on given instance type
data "zenlayercloud_bmc_images" "default" {
  catalog          = "centos"
  instance_type_id = data.zenlayercloud_bmc_instance_types.default.instance_types.0.id
}

resource "zenlayercloud_bmc_subnet" "default" {
  availability_zone = data.zenlayercloud_bmc_zones.default.zones.0.id
  name              = "test-subnet"
  cidr_block        = "10.0.10.0/24"
}

# Create a web server
resource "zenlayercloud_bmc_instance" "web" {
  availability_zone    = data.zenlayercloud_bmc_zones.default.zones.0.id
  image_id             = data.zenlayercloud_bmc_images.default.images.0.image_id
  internet_charge_type = "ByBandwidth"
  instance_type_id     = data.zenlayercloud_bmc_instance_types.default.instance_types.0.id
  password             = "Example~123"
  instance_name        = "web"
  subnet_id            = zenlayercloud_bmc_subnet.default.id
}