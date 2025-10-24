# terraform-provider-uca

A provider to create ephemeral instances for your labs at Université Clermont-Auvergne (Technologies Cloud ZZ3 FISA, DevOps M1).

## Install the provider

```tf
terraform {
  required_providers {
    uca = {
      source = "registry.terraform.io/florian-forestier/uca"
      version = "1.0.1"
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

#### `uca_server`

This resource allows you to create a server for 4hrs, extendable up to 12hrs using the UI.

The following arguments are required:
* `name` : the displayed name of your server. Must comply with RFC 1123 (only alphanums and dashes).
* `username` : the default username created on your server.
* `ssh_key` : the public ssh key to use for your default username.

The following data is returned as output:

* `id` : the unique id of the generated server (UUIDv4);
* `ipv4` : the server's IPv4 address ;
* `expiration_date` : the server's expiration date (can be extented up to 12 hours via UI).

**This resource does not support in-place update. Please destroy and apply again when needed.**
