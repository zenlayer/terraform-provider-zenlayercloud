resource "zenlayercloud_zga_certificate" "default" {
  certificate  = <<EOF
-----BEGIN CERTIFICATE-----
[......] # cert contents
-----END CERTIFICATE-----
EOF

  key = <<EOF
-----BEGIN RSA PRIVATE KEY-----
[......] # key contents
-----END RSA PRIVATE KEY-----
EOF

  lifecycle {
    create_before_destroy = true
  }
}

resource "zenlayercloud_zga_accelerator" "default" {
  accelerator_name = "accelerator_test"
  charge_type = "ByTrafficPackage"
  domain = "test.com"
  relate_domains = ["a.test.com"]
  origin_region_id = "DE"
  origin = ["10.10.10.10"]
  backup_origin = ["10.10.10.14"]
  certificate_id = resource.zenlayercloud_zga_certificate.default.id
  accelerate_regions {
    accelerate_region_id = "KR"
  }
  accelerate_regions {
    accelerate_region_id = "US"
  }
  l4_listeners {
    protocol = "udp"
    port_range = "53/54"
    back_port_range = "53/54"
  } 
  l4_listeners {
    port = 80
    back_port = 80
    protocol = "tcp"
  } 
  l7_listeners {
    port = 443
    back_port = 80
    protocol = "https"
    back_protocol = "http"
  }
  l7_listeners {
    port_range = "8888/8890"
    back_port_range = "8888/8890"
    protocol = "http"
    back_protocol = "http"
  }
  protocol_opts {
    websocket = true
    gzip = false
  }
  access_control {
    enable = true
    rules {
      listener = "https:443"
      directory = "/"
      policy = "deny"
      cidr_ip = ["10.10.10.10"]
    }
    rules {
      listener = "udp:53/54"
      directory = "/"
      policy = "accept"
      cidr_ip = ["10.10.10.11/8"]
    }
  }
}