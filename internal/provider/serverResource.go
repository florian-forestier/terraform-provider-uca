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

var _ tfresource.Resource = &ServerResource{}

type ServerResource struct {
	client    *http.Client
	userToken string
	endpoint  string
}

type ServerModel struct {
	Id       tftypes.String `tfsdk:"id"`
	Name     tftypes.String `tfsdk:"name"`
	IPv4     tftypes.String `tfsdk:"ipv4"`
	SSHKey   tftypes.String `tfsdk:"ssh_key"`
	Username tftypes.String `tfsdk:"username"`
}

func (r *ServerResource) Context() struct {
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

func NewServerResource() tfresource.Resource {
	return &ServerResource{}
}

func (r *ServerResource) Metadata(_ context.Context, req tfresource.MetadataRequest, resp *tfresource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server"
}

func (r *ServerResource) Schema(_ context.Context, _ tfresource.SchemaRequest, resp *tfresource.SchemaResponse) {
	resp.Schema = tfschema.Schema{
		MarkdownDescription: "Manage server",
		Attributes: map[string]tfschema.Attribute{
			"id": tfschema.StringAttribute{
				MarkdownDescription: "Server ID (read-only)",
				Required:            false,
				Computed:            true,
			},
			"name": tfschema.StringAttribute{
				MarkdownDescription: "The name for the server",
				Required:            true,
				Computed:            false,
			},
			"ssh_key": tfschema.StringAttribute{
				MarkdownDescription: "The public key for the server",
				Required:            true,
				Computed:            false,
			},
			"username": tfschema.StringAttribute{
				MarkdownDescription: "The server's configured username",
				Required:            true,
				Computed:            false,
			},
			"ipv4": tfschema.StringAttribute{
				MarkdownDescription: "The server's IPv4",
				Required:            false,
				Computed:            true,
			},
		},
	}
}

func (r *ServerResource) Configure(_ context.Context, req tfresource.ConfigureRequest, resp *tfresource.ConfigureResponse) {
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

func (r *ServerResource) Create(ctx context.Context, req tfresource.CreateRequest, resp *tfresource.CreateResponse) {
	var plannedState ServerModel
	req.Plan.Get(ctx, &plannedState)

	data := struct {
		Username string `json:"user"`
		SshKey   string `json:"ssh_key"`
		Name     string `json:"instance_name"`
	}{
		Username: plannedState.Username.ValueString(),
		SshKey:   plannedState.SSHKey.ValueString(),
		Name:     plannedState.Name.ValueString(),
	}

	res, _, err := MakeHTTPRequest(data, true, r, "POST", "servers", []int{http.StatusCreated})

	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")
		return
	}

	plannedState.Id = tftypes.StringValue(res["id"])
	plannedState.Name = tftypes.StringValue(res["instance_name"])
	plannedState.Username = tftypes.StringValue(res["user"])
	plannedState.IPv4 = tftypes.StringValue(res["ipv4"])
	plannedState.SSHKey = tftypes.StringValue(res["ssh_key"])

	resp.State.Set(ctx, &plannedState)
}

func (r *ServerResource) Read(ctx context.Context, req tfresource.ReadRequest, resp *tfresource.ReadResponse) {
	var currentState ServerModel
	req.State.Get(ctx, &currentState)

	res, code, err := MakeHTTPRequest(nil, true, r, "GET", fmt.Sprintf("servers/%s", currentState.Id.ValueString()), []int{http.StatusOK, http.StatusNotFound})

	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")
		return
	}

	if code == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}

	currentState.Id = tftypes.StringValue(res["id"])
	currentState.Name = tftypes.StringValue(res["instance_name"])
	currentState.Username = tftypes.StringValue(res["user"])
	currentState.IPv4 = tftypes.StringValue(res["ipv4"])
	currentState.SSHKey = tftypes.StringValue(res["ssh_key"])

	resp.State.Set(ctx, &currentState)
}

func (r *ServerResource) Delete(ctx context.Context, req tfresource.DeleteRequest, resp *tfresource.DeleteResponse) {
	var currentState ServerModel
	req.State.Get(ctx, &currentState)

	_, _, err := MakeHTTPRequest(nil, false, r, "DELETE", fmt.Sprintf("servers/%s", currentState.Id.ValueString()), []int{http.StatusOK, http.StatusNoContent, http.StatusNotFound})

	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")
		return
	}
}

func (r *ServerResource) Update(ctx context.Context, req tfresource.UpdateRequest, resp *tfresource.UpdateResponse) {
	var currentState ServerModel
	var plannedState ServerModel
	req.State.Get(ctx, &currentState)
	req.Plan.Get(ctx, &plannedState)

	if currentState.Id != plannedState.Id || currentState.IPv4 != plannedState.IPv4 || currentState.SSHKey != plannedState.SSHKey || currentState.Username != plannedState.Username {
		resp.Diagnostics.AddError("You are trying to change a read-only attribute. Request denied.", "Due to the nature of an instance, you cannot change the id, ipv4, ssh_key and username fields. Request denied.")
		return
	}

	dataToSend := struct {
		Name string `json:"name"`
	}{
		Name: plannedState.Name.ValueString(),
	}

	_, _, err := MakeHTTPRequest(dataToSend, true, r, "PUT", fmt.Sprintf("servers/%s", currentState.Id.ValueString()), []int{http.StatusOK, http.StatusNotFound, http.StatusNoContent})

	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")
		return
	}

	resp.State.Set(ctx, &plannedState)
}
