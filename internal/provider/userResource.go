package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
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

func NewServerResource() tfresource.Resource {
	return &ServerResource{}
}

func (r *ServerResource) Metadata(ctx context.Context, req tfresource.MetadataRequest, resp *tfresource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server"
}

func (r *ServerResource) Schema(ctx context.Context, req tfresource.SchemaRequest, resp *tfresource.SchemaResponse) {
	resp.Schema = tfschema.Schema{
		MarkdownDescription: "Manage server",
		Attributes: map[string]tfschema.Attribute{
			"id": tfschema.StringAttribute{
				MarkdownDescription: "Server ID",
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

func (r *ServerResource) Configure(ctx context.Context, req tfresource.ConfigureRequest, resp *tfresource.ConfigureResponse) {
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
	var state ServerModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	data := struct {
		Username string `json:"user"`
		SshKey   string `json:"ssh_key"`
		Name     string `json:"instance_name"`
	}{
		Username: state.Username.ValueString(),
		SshKey:   state.SSHKey.ValueString(),
		Name:     state.Name.ValueString(),
	}

	d, _ := json.Marshal(data)

	r2, _ := http.NewRequest("POST", r.endpoint+"servers", bytes.NewBuffer(d))
	r2.Header.Set("Content-Type", "application/json")
	r2.Header.Set("X-Auth-Token", r.userToken)

	response, err := r.client.Do(r2)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Error sending request: %s", err))
		return
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError("HTTP Error", fmt.Sprintf("Received non-OK HTTP status: %s", response.Status))
		return
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		resp.Diagnostics.AddError("Failed to Read Response Body", fmt.Sprintf("Could not read response body: %s", err))
		return
	}

	var data2 map[string]string
	_ = json.Unmarshal(body, &data2)

	state.Id = tftypes.StringValue(data2["id"])
	state.Name = tftypes.StringValue(data2["instance_name"])
	state.Username = tftypes.StringValue(data2["user"])
	state.IPv4 = tftypes.StringValue(data2["ipv4"])
	state.SSHKey = tftypes.StringValue(data2["ssh_key"])

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ServerResource) Read(ctx context.Context, req tfresource.ReadRequest, resp *tfresource.ReadResponse) {
	var state ServerModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r2, _ := http.NewRequest("GET", r.endpoint+"servers", nil)
	r2.Header.Set("Content-Type", "application/json")
	r2.Header.Set("X-Auth-Token", r.userToken)

	response, err := r.client.Do(r2)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read user, got error: %s", err))
		return
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}
	if response.StatusCode == http.StatusOK {
		body, err := io.ReadAll(response.Body)
		if err != nil {
			resp.Diagnostics.AddError("Failed to Read Response Body", fmt.Sprintf("Could not read response body: %s", err))
			return
		}

		var data2 []map[string]string
		_ = json.Unmarshal(body, &data2)

		for _, k := range data2 {
			if k["id"] == state.Id.ValueString() {
				state.Id = tftypes.StringValue(k["id"])
				state.Name = tftypes.StringValue(k["instance_name"])
				state.Username = tftypes.StringValue(k["user"])
				state.IPv4 = tftypes.StringValue(k["ipv4"])
				state.SSHKey = tftypes.StringValue(k["ssh_key"])
				resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
				break
			}
		}
	} else {
		resp.Diagnostics.AddError("HTTP Error", fmt.Sprintf("Received bad HTTP status: %s", response.Status))
	}
}

func (r *ServerResource) Delete(ctx context.Context, req tfresource.DeleteRequest, resp *tfresource.DeleteResponse) {
	var data ServerModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r2, _ := http.NewRequest("DELETE", r.endpoint+"servers/"+data.Id.ValueString(), nil)
	r2.Header.Set("Content-Type", "application/json")
	r2.Header.Set("X-Auth-Token", r.userToken)

	response, err := r.client.Do(r2)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete user, got error: %s", err))
		return
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK && response.StatusCode != http.StatusNoContent && response.StatusCode != http.StatusNotFound {
		resp.Diagnostics.AddError("HTTP Error", fmt.Sprintf("Received non-OK HTTP status: %s", response.Status))
		return
	}
	data.Id = tftypes.StringValue("")
	data.Name = tftypes.StringValue("")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ServerResource) Update(ctx context.Context, req tfresource.UpdateRequest, resp *tfresource.UpdateResponse) {
	resp.Diagnostics.AddError("Cannot update resource. Please run destroy then apply again.", "Cannot update resource. Please run destroy then apply again.")
	return
}
