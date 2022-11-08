# Terraform Provider For Zenlayer Cloud

## Requirements:
---

- [Terraform](https://www.terraform.io/downloads.html) 0.12.x
- [Go](https://golang.org/doc/install) 1.13

## Building The Provider

Clone repository to `$GOPATH/src/github.com/zenlayer/terraform-provider-zenlayercloud`

```shell
$ mkdir -p $GOPATH/src/github.com/zenlayer
$ cd $GOPATH/src/github.com/zenlayer
$ git clone https://github.com/zenlayer/terraform-provider-zenlayercloud.git
```

Enter the provider directory and build the provider

```bash
$ cd $GOPATH/src/github.com/zenlayer/terraform-provider-zenlayercloud
$ make build
```
