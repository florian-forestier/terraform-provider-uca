package provider

import (
	"context"
	"net/http"
	"os"
	"strings"

	tfdatasource "github.com/hashicorp/terraform-plugin-framework/datasource"
	tffunction "github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/path"
	tfprovider "github.com/hashicorp/terraform-plugin-framework/provider"
	tfschema "github.com/hashicorp/terraform-plugin-framework/provider/schema"
	tfresource "github.com/hashicorp/terraform-plugin-framework/resource"
	tftypes "github.com/hashicorp/terraform-plugin-framework/types"
)

type HasContext interface {
	Context() struct {
		UserToken  string
		Endpoint   string
		HttpClient *http.Client
	}
}

type ProviderModel struct {
	UserToken tftypes.String `tfsdk:"user_token"`
	Endpoint  tftypes.String `tfsdk:"endpoint"`
}

type Provider struct {
	userToken string
	endpoint  string
	client    *http.Client
}

var _ tfprovider.Provider = &Provider{}
var _ tfprovider.ProviderWithFunctions = &Provider{}

func New() func() tfprovider.Provider {
	return func() tfprovider.Provider {
		return &Provider{}
	}
}

func (p *Provider) Metadata(_ context.Context, _ tfprovider.MetadataRequest, resp *tfprovider.MetadataResponse) {
	resp.TypeName = "uca"
}

func (p *Provider) Schema(_ context.Context, _ tfprovider.SchemaRequest, resp *tfprovider.SchemaResponse) {
	// Only require user_token when UCA_USER_TOKEN is unset, so Terraform still prompts for it.
	tokenFromEnv := os.Getenv("UCA_USER_TOKEN") != ""

	resp.Schema = tfschema.Schema{
		Attributes: map[string]tfschema.Attribute{
			"user_token": tfschema.StringAttribute{
				MarkdownDescription: "Your auth token. Can also be set with the `UCA_USER_TOKEN` environment variable.",
				Required:            !tokenFromEnv,
				Optional:            tokenFromEnv,
				Sensitive:           true,
			},
			"endpoint": tfschema.StringAttribute{
				MarkdownDescription: "API Endpoint",
				Required:            true,
				Sensitive:           false,
			},
		},
	}
}

func (p *Provider) Configure(ctx context.Context, req tfprovider.ConfigureRequest, resp *tfprovider.ConfigureResponse) {
	var data ProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	p.userToken = data.UserToken.ValueString()
	if p.userToken == "" {
		p.userToken = os.Getenv("UCA_USER_TOKEN")
	}
	if p.userToken == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("user_token"),
			"Missing user token",
			"Set user_token in the provider configuration or the UCA_USER_TOKEN environment variable.",
		)
		return
	}

	p.endpoint = data.Endpoint.ValueString()

	if !strings.HasSuffix(p.endpoint, "/") {
		p.endpoint = p.endpoint + "/"
	}

	p.client = http.DefaultClient
	resp.DataSourceData = p // will be usable by DataSources
	resp.ResourceData = p   // will be usable by Resources
}

func (p *Provider) Resources(_ context.Context) []func() tfresource.Resource {
	return []func() tfresource.Resource{
		NewServerResource,
		NewSecurityGroupResource,
		NewSecurityRuleResource,
		NewSecurityGroupAttachmentResource,
	}
}

func (p *Provider) DataSources(_ context.Context) []func() tfdatasource.DataSource {
	return []func() tfdatasource.DataSource{}
}

func (p *Provider) Functions(_ context.Context) []func() tffunction.Function {
	return []func() tffunction.Function{}
}
