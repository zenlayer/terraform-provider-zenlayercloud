Provides a resource to manage key pair.

~> **NOTE:** This request is to import an SSH key pair to be used for later instance login..

~> **NOTE:** A key pair name and several public SSH keys are required.

Example Usage

```hcl
resource "zenlayercloud_key_pair" "foo" {
  key_name       	= "my_key"
  public_key    	= "ssh-rsa XXXXXXXXXXXX key"
  key_description	= "create a key pair"
}
```

Import

Key pair can be imported, e.g.

```
$ terraform import zenlayercloud_key_pair.foo key-xxxxxxx
```
