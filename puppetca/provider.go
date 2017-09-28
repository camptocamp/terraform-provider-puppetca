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
			"url": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("PUPPETCA_URL", "https://puppet:8140"),
				Description: descriptions["url"],
			},
			"cert": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("PUPPETCA_CERT", ""),
				Description: descriptions["cert"],
			},
			"key": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("PUPPETCA_KEY", ""),
				Description: descriptions["key"],
			},
			"ca": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("PUPPETCA_CA", ""),
				Description: descriptions["ca"],
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
		"url": "The URL of the Puppet CA",

		"cert": "A Puppet certificate to authenticate on the CA",

		"key": "A Puppet private key to authenticate on the CA",

		"ca": "The Puppet CA certificate",
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	baseURL := d.Get("url").(string)
	certificate := d.Get("cert").(string)
	privateKey := d.Get("key").(string)
	caCert := d.Get("ca").(string)

	if baseURL == "" {
		return nil, fmt.Errorf("No url provided")
	}

	client, err := puppetca.NewClient(baseURL, privateKey, certificate, caCert)
	if err != nil {
		return nil, err
	}

	return client, nil
}
