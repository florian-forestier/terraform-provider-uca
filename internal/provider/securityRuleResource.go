package provider

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	tfresource "github.com/hashicorp/terraform-plugin-framework/resource"
	tfschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	tftypes "github.com/hashicorp/terraform-plugin-framework/types"
)

var _ tfresource.Resource = &SecurityRuleResource{}

type SecurityRuleResource struct {
	client    *http.Client
	userToken string
	endpoint  string
}

type SecurityRuleModel struct {
	Id              tftypes.String `tfsdk:"id"`
	Name            tftypes.String `tfsdk:"name"`
	Description     tftypes.String `tfsdk:"description"`
	Port            tftypes.Int64  `tfsdk:"port"`
	Protocol        tftypes.String `tfsdk:"protocol"`  // "tcp", "udp"
	Direction       tftypes.String `tfsdk:"direction"` // "ingress" or "egress"
	IPRange         tftypes.String `tfsdk:"ip_range"`  // Could be both in IPv6 or IPv4 format
	SecurityGroupId tftypes.String `tfsdk:"security_group_id"`
}

func NewSecurityRuleResource() tfresource.Resource {
	return &SecurityRuleResource{}
}

func (r *SecurityRuleResource) Context() struct {
	UserToken  string
	Endpoint   string
	HttpClient *http.Client
} {
	return struct {
		UserToken  string
		Endpoint   string
		HttpClient *http.Client
	}{
		UserToken:  r.userToken,
		Endpoint:   r.endpoint,
		HttpClient: r.client,
	}
}

func (r *SecurityRuleResource) Metadata(_ context.Context, req tfresource.MetadataRequest, resp *tfresource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_security_rule"
}

func (r *SecurityRuleResource) Schema(_ context.Context, _ tfresource.SchemaRequest, resp *tfresource.SchemaResponse) {
	resp.Schema = tfschema.Schema{
		MarkdownDescription: "Manage a Security Rule",
		Attributes: map[string]tfschema.Attribute{
			"id": tfschema.StringAttribute{
				MarkdownDescription: "Security Rule ID (read-only)",
				Required:            false,
				Computed:            true,
			},
			"security_group_id": tfschema.StringAttribute{
				MarkdownDescription: "Security Group related to this rule",
				Required:            true,
				Computed:            false,
			},
			"name": tfschema.StringAttribute{
				MarkdownDescription: "Security Rule name",
				Required:            true,
				Computed:            false,
			},
			"description": tfschema.StringAttribute{
				MarkdownDescription: "Security Rule description",
				Required:            false,
				Computed:            false,
				Optional:            true,
			},
			"protocol": tfschema.StringAttribute{
				MarkdownDescription: "Security Rule protocol (\"TCP\", or \"UDP\")",
				Required:            true,
				Computed:            false,
			},
			"port": tfschema.Int64Attribute{
				MarkdownDescription: "Security Rule port",
				Required:            true,
				Computed:            false,
			},
			"direction": tfschema.StringAttribute{
				MarkdownDescription: "Security Rule direction (\"inbound\" or \"outbound\")",
				Required:            true,
				Computed:            false,
			},
			"ip_range": tfschema.StringAttribute{
				MarkdownDescription: "Security Rule IP Range. Can be an IPv4 or IPv6 CIDR notation (192.168.0.0/16 or bc8:bd04::/64 for example)",
				Required:            true,
				Computed:            false,
			},
		},
	}
}

func (r *SecurityRuleResource) Configure(_ context.Context, req tfresource.ConfigureRequest, resp *tfresource.ConfigureResponse) {
	if req.ProviderData == nil { // this means the provider.go Configure method hasn't been called yet, so wait longer
		return
	}
	provider, ok := req.ProviderData.(*Provider)
	if !ok {
		resp.Diagnostics.AddError("Could not create HTTP client", fmt.Sprintf("Expected *http.Client, got: %T", req.ProviderData))
		return
	}
	r.client = provider.client
	r.userToken = provider.userToken

	if !strings.HasSuffix(r.endpoint, "/") {
		r.endpoint = r.endpoint + "/"
	}

	r.endpoint = provider.endpoint
}

func (r *SecurityRuleResource) Create(ctx context.Context, req tfresource.CreateRequest, resp *tfresource.CreateResponse) {
	var plannedState SecurityRuleModel
	req.Plan.Get(ctx, &plannedState)

	data := struct {
		Name            string `json:"name"`
		Description     string `json:"description"`
		Protocol        string `json:"protocol"`
		Port            int64  `json:"port"`
		Direction       string `json:"direction"`
		IPRange         string `json:"ip_range"`
		SecurityGroupId string `json:"security_group_id"`
	}{
		Name:            plannedState.Name.ValueString(),
		Description:     plannedState.Description.ValueString(),
		Protocol:        plannedState.Protocol.ValueString(),
		Port:            plannedState.Port.ValueInt64(),
		Direction:       plannedState.Direction.ValueString(),
		IPRange:         plannedState.IPRange.ValueString(),
		SecurityGroupId: plannedState.SecurityGroupId.ValueString(),
	}

	res, _, err := MakeHTTPRequest(data, true, r, "POST", fmt.Sprintf("security-groups/%s/rules", data.SecurityGroupId), []int{http.StatusOK, http.StatusCreated})
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")
		return
	}

	plannedState.Id = tftypes.StringValue(res["id"])

	resp.State.Set(ctx, &plannedState)
}

func (r *SecurityRuleResource) Read(ctx context.Context, req tfresource.ReadRequest, resp *tfresource.ReadResponse) {
	var currentState SecurityRuleModel
	req.State.Get(ctx, &currentState)

	_, code, err := MakeHTTPRequest(nil, true, r, "GET", fmt.Sprintf("security-groups/%s/rules/%s", currentState.SecurityGroupId.ValueString(), currentState.Id.ValueString()), []int{http.StatusOK, http.StatusNotFound})
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")
		return
	}

	if code == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.State.Set(ctx, &currentState)
	return
}

func (r *SecurityRuleResource) Delete(ctx context.Context, req tfresource.DeleteRequest, resp *tfresource.DeleteResponse) {
	var currentState SecurityRuleModel
	req.State.Get(ctx, &currentState)

	_, _, err := MakeHTTPRequest(nil, false, r, "DELETE", fmt.Sprintf("security-groups/%s/rules/%s", currentState.SecurityGroupId.ValueString(), currentState.Id.ValueString()), []int{http.StatusOK, http.StatusNoContent, http.StatusNotFound})
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")
		return
	}
}

func (r *SecurityRuleResource) Update(ctx context.Context, req tfresource.UpdateRequest, resp *tfresource.UpdateResponse) {
	var currentState SecurityRuleModel
	var plannedState SecurityRuleModel

	req.State.Get(ctx, &currentState)
	req.Plan.Get(ctx, &plannedState)

	if plannedState.SecurityGroupId != currentState.SecurityGroupId {
		resp.Diagnostics.AddError("You cannot change the security group id of a rule. Permission denied.", "")
	}

	dataToSend := struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Protocol    string `json:"protocol"`
		Port        int64  `json:"port"`
		Direction   string `json:"direction"`
		IPRange     string `json:"ip_range"`
	}{
		Name:        plannedState.Name.ValueString(),
		Description: plannedState.Description.ValueString(),
		Protocol:    plannedState.Protocol.ValueString(),
		Port:        plannedState.Port.ValueInt64(),
		Direction:   plannedState.Direction.ValueString(),
		IPRange:     plannedState.IPRange.ValueString(),
	}

	_, _, err := MakeHTTPRequest(dataToSend, false, r, "PUT", fmt.Sprintf("security-groups/%s/rules/%s", currentState.SecurityGroupId.ValueString(), currentState.Id.ValueString()), []int{http.StatusOK, http.StatusNoContent, http.StatusNotFound})
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")
		return
	}

	plannedState.Id = currentState.Id

	resp.State.Set(ctx, &plannedState)
}
