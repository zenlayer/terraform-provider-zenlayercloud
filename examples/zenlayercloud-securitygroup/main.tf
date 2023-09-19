resource "zenlayercloud_security_group" "group1" {
  name        = "terraform-group1"
  description = "for terraform test 1 purpose"
}

resource "zenlayercloud_security_group" "group2" {
  name        = "terraform-group2"
  description = "for terraform test 2 purpose"
}

resource "zenlayercloud_security_group_rule" "rule1" {
  security_group_id = zenlayercloud_security_group.group2.id
  direction         = "egress"
  policy            = "accept"
  cidr_ip           = "10.0.0.0/16"
  ip_protocol       = "tcp"
  port_range        = "81"
}

resource "zenlayercloud_security_group_attachment" "attachment1" {
  security_group_id = zenlayercloud_security_group.group2.id
  instance_id       = "802452345504925912"
}