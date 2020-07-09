package puppetca

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
	"strings"

	"github.com/greennosedmule/go-puppetca/puppetca"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourcePuppetCACertificate() *schema.Resource {
	return &schema.Resource{
		Create: resourcePuppetCACertificateCreate,
		Read:   resourcePuppetCACertificateRead,
		Update: resourcePuppetCACertificateUpdate,
		Delete: resourcePuppetCACertificateDelete,

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"autosign": &schema.Schema{
				Type:     schema.TypeBool,
				Required: true,
				ForceNew: false,
			},
			"fingerprint": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"state": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourcePuppetCACertificateCreate(d *schema.ResourceData, meta interface{}) error {
	name := d.Get("name").(string)
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
	waitResult, waitErr := stateConf.WaitForState()
	if waitErr != nil {
		return fmt.Errorf(
			"Error waiting for certificate (%s) to be found: %s", name, waitErr)
	}

	jsonResult := waitResult.(string)
	var result map[string]interface{}
	json.Unmarshal([]byte(jsonResult), &result)

	if (result["state"] == "requested") && d.Get("autosign").(bool) {
		signErr := signCert(client, name)
		if signErr != nil {
			return fmt.Errorf(
				"Error signing certificate (%s): %s", name, signErr)
		}
	}

	d.SetId(name)
	d.Set("name", result["name"])
	d.Set("fingerprint", result["fingerprint"])
	d.Set("state", result["state"])

	return resourcePuppetCACertificateRead(d, meta)
}

func findCert(client puppetca.Client, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		cert, err := client.GetCertByName(name)
		if err != nil {
			log.Printf("[DEBUG][puppetca] Finding certificate: %s", err)
			if strings.Contains(err.Error(), "404 Not Found") {
				return nil, "not found", nil
			}
			return nil, "request error", err
		}

		return cert, "found", nil
	}
}

func signCert(client puppetca.Client, name string) error {
	err := client.SignCertByName(name)
	return err
}

func resourcePuppetCACertificateRead(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[INFO] Refreshing Certificate: %s", d.Id())

	client := meta.(puppetca.Client)
	name := d.Get("name").(string)

	jsonResult, err := client.GetCertByName(name)
	if err != nil {
		d.SetId("")
		return fmt.Errorf("Error retrieving certificate for %s: %v", name, err)
	}

	var result map[string]interface{}
	json.Unmarshal([]byte(jsonResult), &result)

	d.Set("name", result["name"])
	d.Set("fingerprint", result["fingerprint"])
	d.Set("state", result["state"])

	return nil
}

func resourcePuppetCACertificateUpdate(d *schema.ResourceData, meta interface{}) error {
	return resourcePuppetCACertificateRead(d, meta)
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
