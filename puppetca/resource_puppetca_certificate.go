package puppetca

import (
	"fmt"
	"log"
	"net/http"
	"strings"
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
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		// Necessary to allow terraform to know if the certificate needs to be
		// updated or not when switching sign from false <=> true
		Update: resourcePuppetCACertificateCreate,

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"env": &schema.Schema{
				Type:     schema.TypeString,
				ForceNew: false,
				Optional: true,
			},
			"sign": &schema.Schema{
				Type: schema.TypeBool,
				// sign must be either ForceNew true or we need to implement
				// Update for this resource which does not really mean something
				// puppet wise. So instead we fall back to making sure the
				// certificate is present by calling resourcePuppetCACertificateRead
				ForceNew: false,
				Optional: true,
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
	env := d.Get("env").(string)
	sign := d.Get("sign").(bool)
	log.Printf("[INFO][puppetca] Creating Certificate: %s", name)
	client := meta.(puppetca.Client)

	stateConf := &resource.StateChangeConf{
		Pending:        []string{"found", "not found"},
		Target:         []string{"found"},
		Refresh:        signCert(client, name, env, sign),
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
	d.Set("cert", cert)
	return resourcePuppetCACertificateRead(d, meta)
}

func findCert(client puppetca.Client, name, env string) (string, string, error) {
	cert, err := client.GetCertByName(name, env)

	if err != nil || cert == "" {
		if strings.Contains(err.Error(), http.StatusText(http.StatusNotFound)) {
			// Expected error: puppetserver returns a 404 when the certificate
			// requested on a GET /puppet-ca/v1/certificate/ is not found. This
			// is translated to an error in golang. We can set the error to nil
			// in this case otherwise terraform will stop and report a bug in
			// the provider which is not the case
			return "", "not found", nil
		}
		return "", "not found", err
	}
	return cert, "found", nil
}

func signCert(client puppetca.Client, name, env string, sign bool) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		cert, status, err := findCert(client, name, env)
		if err != nil {
			return cert, status, err
		} else if cert == "" && sign == true {
			// do we have a CSR request to sign
			csrReq, err := client.GetRequest(name, env)
			if err != nil {
				return nil, "not found", err
			}
			if csrReq != "" {
				err = client.SignRequest(name, env)
				if err != nil {
					return nil, "not found", err
				}
				cert, status, err = findCert(client, name, env)
				if err != nil {
					return nil, status, err
				}
			}
		}
		return cert, status, nil
	}
}

func resourcePuppetCACertificateRead(d *schema.ResourceData, meta interface{}) error {
	name := d.Id()
	log.Printf("[INFO] Refreshing Certificate: %s", name)

	client := meta.(puppetca.Client)
	env := d.Get("env").(string)

	cert, err := client.GetCertByName(name, env)
	if err != nil {
		return fmt.Errorf("Error retrieving certificate for %s: %v", name, err)
	}
	d.Set("name", name)
	d.Set("cert", cert)

	return nil
}

func resourcePuppetCACertificateDelete(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[INFO] Deleting Certificate: %s", d.Id())

	client := meta.(puppetca.Client)
	name := d.Get("name").(string)
	env := d.Get("env").(string)

	err := client.DeleteCertByName(name, env)
	if err != nil {
		return fmt.Errorf("Error deleting certificate for %s: %v", name, err)
	}

	d.SetId("")
	return nil
}
