# terraform-provider-uca

A provider to create ephemeral instances for your labs at Université Clermont-Auvergne (Technologies Cloud ZZ3 FISA, DevOps M1).

## Install the provider

```tf
terraform {
  required_providers {
    uca = {
      source = "registry.terraform.io/florian-forestier/uca"
      version = "1.0.2"
    }
  }
}
```

## Use the provider

### Provider configuration

```tf
provider "uca" {
  endpoint = "https://cloud-ui.edu.forestier.re/api/"
  user_token = "YOUR_API_KEY_HERE"
}
```

Specify the endpoint to the API server of the infrastructure (not the frontend!).

The user token can be generated on the UI. **NEVER SHARE OR SEND YOUR TOKEN ONLINE**.

You can set `UCA_USER_TOKEN` in your environment to avoid typing it down in your files.

### Resources

