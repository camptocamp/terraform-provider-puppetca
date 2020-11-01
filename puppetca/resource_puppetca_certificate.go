package puppetca

import (
	"fmt"
	"log"
	"time"

	"github.com/camptocamp/go-puppetca/puppetca"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourcePuppetCACertificate() *schema.Resource {
	return &schema.Resource{
		Create: resourcePuppetCACertificateCreate,
		Read:   resourcePuppetCACertificateRead,
		Delete: resourcePuppetCACertificateDelete,

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"usedby": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"cert": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourcePuppetCACertificateCreate(d *schema.ResourceData, meta interface{}) error {
	name := d.Get("name").(string)
	usedby := d.Get("usedby").(string)
	log.Printf("[INFO][puppetca] Creating Certificate: %s", name)
	client := meta.(puppetca.Client)

	stateConf := &resource.StateChangeConf{
		Pending:        []string{"found", "not found"},
		Target:         []string{"found"},
		Refresh:        findCert(client, name),
		Timeout:        10 * time.Minute,
		Delay:          1 * time.Second,
		MinTimeout:     3 * time.Second,
		NotFoundChecks: 50,
	}
	cert, waitErr := stateConf.WaitForState()
	if waitErr != nil {
		return fmt.Errorf(
			"Error waiting for certificate (%s) to be found: %s", name, waitErr)
	}

	d.SetId(name)
	d.Set("usedby", usedby)
	d.Set("cert", cert)
	return resourcePuppetCACertificateRead(d, meta)
}

func findCert(client puppetca.Client, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		cert, err := client.GetCertByName(name)
		if err != nil {
			return nil, "not found", nil
		}

		return cert, "found", nil
	}
}

func resourcePuppetCACertificateRead(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[INFO] Refreshing Certificate: %s", d.Id())

	client := meta.(puppetca.Client)
	name := d.Get("name").(string)

	cert, err := client.GetCertByName(name)
	if err != nil {
		return fmt.Errorf("Error retrieving certificate for %s: %v", name, err)
	}
	d.Set("cert", cert)

	return nil
}

func resourcePuppetCACertificateDelete(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[INFO] Deleting Certificate: %s", d.Id())

	client := meta.(puppetca.Client)
	name := d.Get("name").(string)

	err := client.DeleteCertByName(name)
	if err != nil {
		return fmt.Errorf("Error deleting certificate for %s: %v", name, err)
	}

	d.SetId("")
	return nil
}
