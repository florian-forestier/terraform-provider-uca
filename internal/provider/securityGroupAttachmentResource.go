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

var _ tfresource.Resource = &SecurityGroupAttachmentResource{}

type SecurityGroupAttachmentResource struct {
	client    *http.Client
	userToken string
	endpoint  string
}

type SecurityGroupAttachmentModel struct {
	SecurityGroupId tftypes.String `tfsdk:"security_group_id"`
	ServerId        tftypes.String `tfsdk:"server_id"`
}

func NewSecurityGroupAttachmentResource() tfresource.Resource {
	return &SecurityGroupAttachmentResource{}
}

func (r *SecurityGroupAttachmentResource) Context() struct {
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

func (r *SecurityGroupAttachmentResource) Metadata(_ context.Context, req tfresource.MetadataRequest, resp *tfresource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_security_group_attachment"
}

func (r *SecurityGroupAttachmentResource) Schema(_ context.Context, _ tfresource.SchemaRequest, resp *tfresource.SchemaResponse) {
	resp.Schema = tfschema.Schema{
		MarkdownDescription: "Manage a Security Rule",
		Attributes: map[string]tfschema.Attribute{
			"security_group_id": tfschema.StringAttribute{
				MarkdownDescription: "Security Group related to this rule",
				Required:            true,
				Computed:            false,
			},
			"server_id": tfschema.StringAttribute{
				MarkdownDescription: "Server ID related to this rule",
				Required:            true,
				Computed:            false,
			},
		},
	}
}

func (r *SecurityGroupAttachmentResource) Configure(_ context.Context, req tfresource.ConfigureRequest, resp *tfresource.ConfigureResponse) {
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

func (r *SecurityGroupAttachmentResource) Create(ctx context.Context, req tfresource.CreateRequest, resp *tfresource.CreateResponse) {
	var plannedState SecurityGroupAttachmentModel
	req.Plan.Get(ctx, &plannedState)

	data := struct {
		SecurityGroupId string `json:"security_group_id"`
		ServerId        string `json:"server_id"`
	}{
		ServerId:        plannedState.ServerId.ValueString(),
		SecurityGroupId: plannedState.SecurityGroupId.ValueString(),
	}

	_, _, err := MakeHTTPRequest(data, true, r, "POST", fmt.Sprintf("servers/%s/security-groups", data.ServerId), []int{http.StatusOK, http.StatusCreated})
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")
		return
	}

	resp.State.Set(ctx, &plannedState)
}

func (r *SecurityGroupAttachmentResource) Read(ctx context.Context, req tfresource.ReadRequest, resp *tfresource.ReadResponse) {
	var currentState SecurityGroupAttachmentModel
	req.State.Get(ctx, &currentState)

	_, code, err := MakeHTTPRequest(nil, true, r, "GET", fmt.Sprintf("servers/%s/security-groups/%s", currentState.ServerId.ValueString(), currentState.SecurityGroupId.ValueString()), []int{http.StatusOK, http.StatusNotFound})
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

func (r *SecurityGroupAttachmentResource) Delete(ctx context.Context, req tfresource.DeleteRequest, resp *tfresource.DeleteResponse) {
	var currentState SecurityGroupAttachmentModel
	req.State.Get(ctx, &currentState)

	_, _, err := MakeHTTPRequest(nil, false, r, "DELETE", fmt.Sprintf("servers/%s/security-groups/%s", currentState.ServerId.ValueString(), currentState.SecurityGroupId.ValueString()), []int{http.StatusOK, http.StatusNoContent, http.StatusNotFound})
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")
		return
	}
}

func (r *SecurityGroupAttachmentResource) Update(ctx context.Context, req tfresource.UpdateRequest, resp *tfresource.UpdateResponse) {
	var currentState SecurityGroupAttachmentModel
	var plannedState SecurityGroupAttachmentModel

	req.State.Get(ctx, &currentState)
	req.Plan.Get(ctx, &plannedState)

	_, _, err := MakeHTTPRequest(nil, false, r, "DELETE", fmt.Sprintf("servers/%s/security-groups/%s", currentState.ServerId.ValueString(), currentState.SecurityGroupId.ValueString()), []int{http.StatusOK, http.StatusNoContent, http.StatusNotFound})
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")
		return
	}

	data := struct {
		SecurityGroupId string `json:"security_group_id"`
		ServerId        string `json:"server_id"`
	}{
		ServerId:        plannedState.ServerId.ValueString(),
		SecurityGroupId: plannedState.SecurityGroupId.ValueString(),
	}

	_, _, err = MakeHTTPRequest(data, true, r, "POST", fmt.Sprintf("servers/%s/security-groups", data.ServerId), []int{http.StatusOK, http.StatusCreated})
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")
		return
	}

	resp.State.Set(ctx, &plannedState)

}
