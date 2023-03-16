data "zenlayercloud_zones" "default" {

}

data "zenlayercloud_instance_types" "default" {
  availability_zone = data.zenlayercloud_zones.default.zones.0.id
}

# Get a centos image which also supported to install on given instance type
data "zenlayercloud_images" "default" {
  availability_zone = data.zenlayercloud_zones.default.zones.0.id
  category          = "CentOS"
}

resource "zenlayercloud_subnet" "default" {
  name              = "test-subnet"
  cidr_block        = "10.0.10.0/24"
}

# Create a web server
resource "zenlayercloud_instance" "web" {
  availability_zone    = data.zenlayercloud_zones.default.zones.0.id
  image_id             = data.zenlayercloud_images.default.images.0.image_id
  internet_charge_type = "ByBandwidth"
  instance_type        = data.zenlayercloud_instance_types.default.instance_types.0.id
  password             = "Example~123"
  instance_name        = "web"
  subnet_id            = zenlayercloud_subnet.default.id
  system_disk_size     = 100
}