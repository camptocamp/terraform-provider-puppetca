package resources

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/camptocamp/go-puppetca/puppetca"
	"github.com/camptocamp/terraform-provider-puppetca/internal/log"
	"github.com/camptocamp/terraform-provider-puppetca/internal/provider"
	"github.com/cenkalti/backoff/v4"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type Certificate struct {
	provider *provider.Provider
}

type CertificateModel struct {
	NodeName    types.String `tfsdk:"name"`
	Environment types.String `tfsdk:"env"`
	Sign        types.Bool   `tfsdk:"sign"`
	UsedBy      types.String `tfsdk:"usedby"`
	Content     types.String `tfsdk:"cert"`
}

func (r *Certificate) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_certificate"
}

func (r *Certificate) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version: 1,
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"env": schema.StringAttribute{
				Optional: true,
			},
			"sign": schema.BoolAttribute{
				Optional: true,
			},
			"usedby": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"cert": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (r *Certificate) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan, state CertificateModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	nodeName := plan.NodeName.ValueString()
	environment := plan.Environment.ValueString()
	sign := plan.Sign.ValueBool()

	certificate, err := retryGetOrSignCert(ctx, r.provider.Client(), nodeName, environment, sign)

	if err != nil {
		resp.Diagnostics.AddError("Failed to create certificate", "Reason: "+err.Error())

		return
	}

	state = plan
	state.Content = types.StringValue(certificate)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *Certificate) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state CertificateModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	nodeName := state.NodeName.ValueString()
	environment := state.Environment.ValueString()

	certificate, err := getCert(ctx, r.provider.Client(), nodeName, environment)

	if err != nil {
		if isErrNotFound(err) {
			resp.State.RemoveResource(ctx)

			return
		}

		resp.Diagnostics.AddError("Failed to read certificate", "Reason: "+err.Error())

		return
	}

	state.Content = types.StringValue(certificate)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *Certificate) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state, plan CertificateModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	nodeName := plan.NodeName.ValueString()
	environment := plan.Environment.ValueString()
	sign := plan.Sign.ValueBool()

	certificate, err := retryGetOrSignCert(ctx, r.provider.Client(), nodeName, environment, sign)

	if err != nil {
		resp.Diagnostics.AddError("Failed to update certificate", "Reason: "+err.Error())

		return
	}

	state = plan
	state.Content = types.StringValue(certificate)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *Certificate) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state CertificateModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	nodeName := state.NodeName.ValueString()
	environment := state.Environment.ValueString()

	err := deleteCert(ctx, r.provider.Client(), nodeName, environment)

	if err != nil && !isErrNotFound(err) {
		resp.Diagnostics.AddError("Failed to delete certificate", "Reason: "+err.Error())

		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *Certificate) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id := strings.Split(req.ID, ",")

	if len(id) != 2 {
		resp.Diagnostics.AddError("Invalid ID format", "Expected ID format is “<node name>,<environment>”")

		return
	}

	state := CertificateModel{
		NodeName:    types.StringValue(id[0]),
		Environment: types.StringValue(id[1]),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *Certificate) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		0: {
			PriorSchema: &schema.Schema{
				Attributes: map[string]schema.Attribute{
					"id":     schema.StringAttribute{},
					"name":   schema.StringAttribute{},
					"env":    schema.StringAttribute{},
					"sign":   schema.BoolAttribute{},
					"usedby": schema.StringAttribute{},
					"cert":   schema.StringAttribute{},
				},
			},
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
				var oldState struct {
					ID          types.String `tfsdk:"id"`
					NodeName    types.String `tfsdk:"name"`
					Environment types.String `tfsdk:"env"`
					Sign        types.Bool   `tfsdk:"sign"`
					UsedBy      types.String `tfsdk:"usedby"`
					Content     types.String `tfsdk:"cert"`
				}

				resp.Diagnostics.Append(req.State.Get(ctx, &oldState)...)

				if resp.Diagnostics.HasError() {
					return
				}

				newState := CertificateModel{
					NodeName:    oldState.NodeName,
					Environment: oldState.Environment,
					Sign:        oldState.Sign,
					UsedBy:      oldState.UsedBy,
					Content:     oldState.Content,
				}

				resp.Diagnostics.Append(resp.State.Set(ctx, newState)...)
			},
		},
	}
}

func NewCertificate(p *provider.Provider) resource.Resource {
	r := &Certificate{
		provider: p,
	}

	var _ resource.Resource = r
	var _ resource.ResourceWithImportState = r
	var _ resource.ResourceWithUpgradeState = r

	return r
}

func init() {
	resources = append(resources, NewCertificate)
}

func isErrNotFound(err error) bool {
	return strings.Contains(err.Error(), http.StatusText(http.StatusNotFound))
}

func getCert(ctx context.Context, client *puppetca.Client, nodeName string, environment string) (string, error) {
	logFields := log.CertificateFields(nodeName, environment)

	tflog.Trace(ctx, "Requesting certificate", logFields)

	certificate, err := client.GetCertByName(nodeName, environment)

	tflog.Trace(ctx, "Requested certificate", log.MergeFields(logFields, log.ErrorField(err), map[string]any{
		"certificate": certificate,
	}))

	return certificate, err
}

func signCert(ctx context.Context, client *puppetca.Client, nodeName string, environment string) error {
	logFields := log.CertificateFields(nodeName, environment)

	tflog.Trace(ctx, "Requesting certificate signing request", logFields)

	certificateSigningRequest, err := client.GetRequest(nodeName, environment)

	tflog.Trace(ctx, "Requested certificate signing request", log.MergeFields(logFields, log.ErrorField(err), map[string]any{
		"certificate_signing_request": certificateSigningRequest,
	}))

	if err != nil {
		return err
	}

	tflog.Trace(ctx, "Requesting certificate signing", logFields)

	err = client.SignRequest(nodeName, environment)

	tflog.Trace(ctx, "Requested certificate signing", log.MergeFields(logFields, log.ErrorField(err)))

	return err
}

func retryGetOrSignCert(ctx context.Context, client *puppetca.Client, nodeName string, environment string, sign bool) (string, error) {
	logFields := log.CertificateFields(nodeName, environment)

	getOrSignCert := func() (certificate string, err error) {
		defer func() {
			if err != nil && !isErrNotFound(err) {
				err = backoff.Permanent(err)
			}
		}()

		certificate, err = getCert(ctx, client, nodeName, environment)

		if err != nil && isErrNotFound(err) && sign {
			err = signCert(ctx, client, nodeName, environment)

			if err != nil {
				return "", err
			}

			certificate, err = getCert(ctx, client, nodeName, environment)
		}

		return certificate, err
	}

	return backoff.RetryNotifyWithData(getOrSignCert, backoff.WithContext(backoff.NewExponentialBackOff(), ctx),
		func(err error, delay time.Duration) {
			tflog.Trace(ctx, "Will retry requesting certificate after backoff delay", log.MergeFields(logFields, log.ErrorField(err), map[string]any{
				"delay": delay,
			}))
		},
	)
}

func deleteCert(ctx context.Context, client *puppetca.Client, nodeName string, environment string) error {
	logFields := log.CertificateFields(nodeName, environment)

	tflog.Trace(ctx, "Requesting certificate deletion", logFields)

	err := client.DeleteCertByName(nodeName, environment)

	tflog.Trace(ctx, "Requested certificate deletion", log.MergeFields(logFields, log.ErrorField(err)))

	return err
}
