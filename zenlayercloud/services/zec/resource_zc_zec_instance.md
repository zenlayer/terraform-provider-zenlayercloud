Provides a instance resource.

~> **NOTE:** Currently it's not able to create default public IPv4 through instance resource. Please use `zenlayercloud_zec_eip` resource to create EIP and then call resource `zenlayer_zec_eip_association` to bind it to resource

Example Usage

```hcl

variable "availability_zone" {
	default = "asia-east-1a"
}

variable "region" {
	default = "asia-east-1"
}

resource "zenlayercloud_zec_security_group" "sg" {
  name = "Test-SecurityGroup1"
}

resource "zenlayercloud_zec_vpc_security_group_attachment" "foo" {
	vpc_id = zenlayercloud_zec_vpc.foo.id
	security_group_id = zenlayercloud_zec_security_group.sg.id
}

# Create subnet (IPv4 IP stack)
resource "zenlayercloud_zec_subnet" "ipv4" {
	vpc_id = zenlayercloud_zec_vpc.foo.id
	region_id	 = var.region
	name       = "test-subnet"
	cidr_block = "10.0.0.0/24"
}

# Create instance Using key pair
data "zenlayercloud_key_pairs" "all" {
}

data "zenlayercloud_zec_images" "ubuntu" {
  availability_zone = var.availability_zone
  category = "Ubuntu"
}

# Create a Instance
resource "zenlayercloud_zec_instance" "instance" {
  availability_zone = var.availability_zone
  instance_type = "z2a.cpu.1"
  image_id =data.zenlayercloud_zec_images.ubuntu.images.0.id
  instance_name = "Example-Instance"
  key_id = data.zenlayercloud_key_pairs.all.key_pairs.0.key_id
  subnet_id = zenlayercloud_zec_subnet.ipv4.id
  system_disk_size = 20
}

```

# Import

Instance can be imported using the id, e.g.

```
terraform import zenlayercloud_zec_instance.instance instance-id
```
