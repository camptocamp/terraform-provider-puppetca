package provider

import (
	"context"
	"os"

	"github.com/camptocamp/go-puppetca/puppetca"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type Provider struct {
	name        string
	version     string
	dataSources []func() datasource.DataSource
	resources   []func() resource.Resource

	client puppetca.Client
}

type Model struct {
	Url           types.String `tfsdk:"url"`
	CACertificate types.String `tfsdk:"ca"`
	Certificate   types.String `tfsdk:"cert"`
	PrivateKey    types.String `tfsdk:"key"`
}

func (p *Provider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = p.name
	resp.Version = p.version
}

func (p *Provider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"url": schema.StringAttribute{
				Optional:    true,
				Description: "URL of the Puppet CA",
			},
			"ca": schema.StringAttribute{
				Optional:    true,
				Description: "Puppet CA certificate",
			},
			"cert": schema.StringAttribute{
				Optional:    true,
				Description: "Certificate to authenticate against Puppet CA",
			},
			"key": schema.StringAttribute{
				Sensitive:   true,
				Optional:    true,
				Description: "Private key to authenticate against Puppet CA",
			},
		},
	}
}

func (p *Provider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config Model

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	url := os.Getenv("PUPPETCA_URL")
	cacert := os.Getenv("PUPPETCA_CA")
	cert := os.Getenv("PUPPETCA_CERT")
	key := os.Getenv("PUPPETCA_KEY")

	if !config.Url.IsNull() {
		url = config.Url.ValueString()
	}

	if !config.CACertificate.IsNull() {
		cacert = config.CACertificate.ValueString()
	}

	if !config.Certificate.IsNull() {
		cert = config.Certificate.ValueString()
	}

	if !config.PrivateKey.IsNull() {
		key = config.PrivateKey.ValueString()
	}

	if url == "" {
		url = "https://puppet:8140"
	}

	var err error

	p.client, err = puppetca.NewClient(url, key, cert, cacert)

	if err != nil {
		resp.Diagnostics.AddError("Failed to create Puppet CA client", "Reason: "+err.Error())
		return
	}

	tflog.Info(ctx, "Successfully created Puppet CA client", map[string]any{
		"url": url,
	})
}

func (p *Provider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return p.dataSources
}

func (p *Provider) Resources(ctx context.Context) []func() resource.Resource {
	return p.resources
}

func (p *Provider) Client() *puppetca.Client {
	return &p.client
}

func NewFactory(name string, version string, ds []func(p *Provider) datasource.DataSource, rs []func(p *Provider) resource.Resource) func() provider.Provider {
	return func() provider.Provider {
		p := &Provider{
			name:    name,
			version: version,
		}

		p.dataSources = make([]func() datasource.DataSource, len(ds))

		for i, d := range ds {
			d := d

			p.dataSources[i] = func() datasource.DataSource {
				return d(p)
			}
		}

		p.resources = make([]func() resource.Resource, len(rs))

		for i, r := range rs {
			r := r

			p.resources[i] = func() resource.Resource {
				return r(p)
			}
		}

		var _ provider.Provider = p

		return p
	}
}
