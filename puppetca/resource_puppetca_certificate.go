package puppetca

import (
	"fmt"
	"log"

	"github.com/camptocamp/go-puppetca/puppetca"
	"github.com/hashicorp/terraform/helper/schema"
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
			"cert": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourcePuppetCACertificateCreate(d *schema.ResourceData, meta interface{}) error {
	name := d.Get("name").(string)
	log.Printf("[INFO][puppetca] Creating Certificate: %s", name)
	client := meta.(*puppetca.Client)

	cert, err := client.GetCertByName(name)
	if err != nil {
		return fmt.Errorf("Error retrieving certificate for %s: %v", name, err)
	}
	d.SetId(name)
	d.Set("cert", cert)
	return resourcePuppetCACertificateRead(d, meta)
}

func resourcePuppetCACertificateRead(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[INFO] Refreshing Certificate: %s", d.Id())

	client := meta.(*puppetca.Client)
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

	client := meta.(*puppetca.Client)
	name := d.Get("name").(string)

	err := client.DeleteCertByName(name)
	if err != nil {
		return fmt.Errorf("Error deleting certificate for %s: %v", name, err)
	}

	d.SetId("")
	return nil
}
