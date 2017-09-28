package puppetca

import (
	"fmt"

	"github.com/camptocamp/go-puppetca/puppetca"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

// Provider returns a terraform.ResourceProvider.
func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"base_url": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("PUPPETCA_URL", ""),
				Description: descriptions["base_url"],
			},
			"private_key": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("PUPPETCA_PRIVATE_KEY", ""),
				Description: descriptions["private_key"],
			},
			"certificate": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("PUPPETCA_CERTIFICATE", ""),
				Description: descriptions["certificate"],
			},
			"ca_cert": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("PUPPETCA_CA_CERT", ""),
				Description: descriptions["ca_cert"],
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"puppetca_certificate": resourcePuppetCACertificate(),
		},

		ConfigureFunc: providerConfigure,
	}
}

var descriptions map[string]string

func init() {
	descriptions = map[string]string{
		"base_url": "The base URL to the Puppet CA",

		"private_key": "A Puppet private key to authenticate on the CA",

		"certificate": "A Puppet certificate to authenticate on the CA",

		"ca_cert": "The Puppet CA certificate",
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	baseURL := d.Get("base_url").(string)
	privateKey := d.Get("private_key").(string)
	certificate := d.Get("certificate").(string)
	caCert := d.Get("ca_cert").(string)

	if baseURL == "" {
		return nil, fmt.Errorf("No base_url provided")
	}

	client, err := puppetca.NewClient(baseURL, privateKey, certificate, caCert)
	if err != nil {
		return nil, err
	}

	return client, nil
}
