Puppet CA Terraform Provider
=============================

This Terraform provider allows to connect to a Puppet Certificate Authority to verify that node certificates were signed, and clean them upon decommissioning the node.


Requirements
------------

-	[Terraform](https://www.terraform.io/downloads.html) 0.10.x
-	[Go](https://golang.org/doc/install) 1.8 (to build the provider plugin)

Building The Provider
---------------------

Clone repository to: `$GOPATH/src/github.com/camptocamp/terraform-provider-puppetca`

```sh
$ mkdir -p $GOPATH/src/github.com/camptocamp; cd $GOPATH/src/github.com/camptocamp
$ git clone git@github.com:camptocamp/terraform-provider-puppetca
```

Enter the provider directory and build the provider

```sh
$ cd $GOPATH/src/github.com/camptocamp/terraform-provider-puppetca
$ make build
```

Using the provider
----------------------

```hcl
provider puppetca {
  url = "https://puppetca.example.com:8140"
  cert = "certs/puppet.crt"
  key = "certs/puppet.key"
  ca = "certs/ca.pem"

}

resource "puppetca_certificate" "test" {
  name = "0a7842c26ad0.foo.com"
}
```



Developing the Provider
---------------------------

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (version 1.8+ is *required*). You'll also need to correctly setup a [GOPATH](http://golang.org/doc/code.html#GOPATH), as well as adding `$GOPATH/bin` to your `$PATH`.

To compile the provider, run `make build`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

```sh
$ make bin
...
$ $GOPATH/bin/terraform-provider-puppetca
...
```

In order to test the provider, you can simply run `make test`.

```sh
$ make test
```

In order to run the full suite of Acceptance tests, run `make testacc`.

*Note:* Acceptance tests create real resources, and often cost money to run.

```sh
$ make testacc
```
