package puppetca

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
)

// Client is a Puppet CA client
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient returns a new Client
func NewClient(baseURL, keyFile, certFile, caFile string) (c Client, err error) {
	// Load client cert
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		err = errors.Wrapf(err, "failed to load client cert at %s", certFile)
		return
	}

	// Load CA cert
	caCert, err := ioutil.ReadFile(caFile)
	if err != nil {
		err = errors.Wrapf(err, "failed to load CA cert at %s", caFile)
		return
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	// Setup HTTPS client
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
	}
	tr := &http.Transport{TLSClientConfig: tlsConfig}
	httpClient := &http.Client{Transport: tr}
	c = Client{baseURL, httpClient}

	return
}

// GetCertByName returns the certificate of a node by its name
func (c *Client) GetCertByName(nodename string) (string, error) {
	pem, err := c.Get(fmt.Sprintf("certificate/%s", nodename))
	if err != nil {
		return "", errors.Wrapf(err, "failed to retrieve certificate %s", nodename)
	}
	return pem, nil
}

// DeleteCertByName deletes the certificate of a given node
func (c *Client) DeleteCertByName(nodename string) error {
	_, err := c.Delete(fmt.Sprintf("certificate_status/%s", nodename))
	if err != nil {
		return errors.Wrapf(err, "failed to delete certificate %s", nodename)
	}
	return nil
}

// Get performs a GET request
func (c *Client) Get(path string) (string, error) {
	return c.Do("GET", path)
}

// Delete performs a DELETE request
func (c *Client) Delete(path string) (string, error) {
	return c.Do("DELETE", path)
}

// Do performs an HTTP request
func (c *Client) Do(method, path string) (string, error) {
	fullPath := fmt.Sprintf("%s/puppet-ca/v1/%s", c.baseURL, path)
	uri, err := url.Parse(fullPath)
	if err != nil {
		return "", errors.Wrapf(err, "failed to parse URL %s", fullPath)
	}
	req := http.Request{
		Method: method,
		URL:    uri,
	}
	resp, err := c.httpClient.Do(&req)
	if err != nil {
		return "", errors.Wrapf(err, "failed to %s URL %s", method, fullPath)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to %s URL %s, got: %s", method, fullPath, resp.Status)
	}
	defer resp.Body.Close()
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrapf(err, "failed to read body response from %s")
	}

	return string(content), nil
}
