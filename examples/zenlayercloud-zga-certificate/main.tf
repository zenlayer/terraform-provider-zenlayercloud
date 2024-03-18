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

  label = "certificate_test"
}
