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

var _ tfresource.Resource = &SecurityGroupResource{}

type SecurityGroupResource struct {
	client    *http.Client
	userToken string
	endpoint  string
}

type SecurityGroupModel struct {
	Id   tftypes.String `tfsdk:"id"`
	Name tftypes.String `tfsdk:"name"`
}

func NewSecurityGroupResource() tfresource.Resource {
	return &SecurityGroupResource{}
}

func (r *SecurityGroupResource) Context() struct {
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

func (r *SecurityGroupResource) Metadata(_ context.Context, req tfresource.MetadataRequest, resp *tfresource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_security_group"
}

func (r *SecurityGroupResource) Schema(_ context.Context, _ tfresource.SchemaRequest, resp *tfresource.SchemaResponse) {
	resp.Schema = tfschema.Schema{
		MarkdownDescription: "Manage a Security Group",
		Attributes: map[string]tfschema.Attribute{
			"id": tfschema.StringAttribute{
				MarkdownDescription: "Security Group ID (read-only)",
				Required:            false,
				Computed:            true,
			},
			"name": tfschema.StringAttribute{
				MarkdownDescription: "The name for the security group",
				Required:            true,
				Computed:            false,
			},
		},
	}
}

func (r *SecurityGroupResource) Configure(_ context.Context, req tfresource.ConfigureRequest, resp *tfresource.ConfigureResponse) {
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

func (r *SecurityGroupResource) Create(ctx context.Context, req tfresource.CreateRequest, resp *tfresource.CreateResponse) {
	var plannedState SecurityGroupModel
	req.Plan.Get(ctx, &plannedState)

	data := struct {
		Name string `json:"name"`
	}{
		Name: plannedState.Name.ValueString(),
	}

	res, _, err := MakeHTTPRequest(data, true, r, "POST", "security-groups", []int{http.StatusOK, http.StatusCreated})
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")
		return
	}

	plannedState.Id = tftypes.StringValue(res["id"])
	plannedState.Name = tftypes.StringValue(res["name"])

	resp.State.Set(ctx, &plannedState)
}

func (r *SecurityGroupResource) Read(ctx context.Context, req tfresource.ReadRequest, resp *tfresource.ReadResponse) {
	var currentState SecurityGroupModel
	req.State.Get(ctx, &currentState)

	res, code, err := MakeHTTPRequest(nil, true, r, "GET", fmt.Sprintf("security-groups/%s", currentState.Id.ValueString()), []int{http.StatusOK, http.StatusNotFound})
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")
		return
	}

	if code == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}

	currentState.Id = tftypes.StringValue(res["id"])
	currentState.Name = tftypes.StringValue(res["name"])

	resp.State.Set(ctx, &currentState)
	return
}

func (r *SecurityGroupResource) Delete(ctx context.Context, req tfresource.DeleteRequest, resp *tfresource.DeleteResponse) {
	var currentState SecurityGroupModel
	req.State.Get(ctx, &currentState)

	_, _, err := MakeHTTPRequest(nil, false, r, "DELETE", fmt.Sprintf("security-groups/%s", currentState.Id.ValueString()), []int{http.StatusOK, http.StatusNoContent, http.StatusNotFound})
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")
		return
	}
}

func (r *SecurityGroupResource) Update(ctx context.Context, req tfresource.UpdateRequest, resp *tfresource.UpdateResponse) {
	var currentState SecurityGroupModel
	var plannedState SecurityGroupModel

	req.State.Get(ctx, &currentState)
	req.Plan.Get(ctx, &plannedState)

	dataToSend := struct {
		Name string `json:"name"`
	}{
		Name: plannedState.Name.ValueString(),
	}

	_, _, err := MakeHTTPRequest(dataToSend, false, r, "PUT", fmt.Sprintf("security-groups/%s", currentState.Id.ValueString()), []int{http.StatusOK, http.StatusNoContent, http.StatusNotFound})
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")
		return
	}

	req.State.Set(ctx, &plannedState)
}
