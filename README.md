# terraform-provider-uca

A provider to create ephemeral instances for your labs at Université Clermont-Auvergne (Technologies Cloud ZZ3 FISA, DevOps M1).

## Install the provider

```tf
terraform {
  required_providers {
    uca = {
      source = "registry.terraform.io/florian-forestier/uca"
      version = "1.2.0"
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

Manage an ephemeral server.

```tf
resource "uca_server" "my_server" {
  name     = "my-instance"
  username = "ubuntu"
  ssh_key  = "ssh-ed25519 AAAAC3Nza..."
}
```

- `id` (String, Computed): Server ID.
- `name` (String, Required): The name for the server.
- `username` (String, Required): The server's configured username.
- `ssh_key` (String, Required): The public key for the server.
- `ipv4` (String, Computed): The server's IPv4 address.

#### `uca_security_group`

Manage a Security Group.


```tf
resource "uca_security_group" "my_sg" {
  name = "my-security-group"
}
```

- `id` (String, Computed): Security Group ID.
- `name` (String, Required): The name for the security group.

#### `uca_security_rule`

Manage a Security Rule within a Security Group.

```tf
resource "uca_security_rule" "allow_ssh" {
  security_group_id = uca_security_group.my_sg.id
  name              = "allow-ssh"
  description       = "Allow SSH from anywhere"
  protocol          = "TCP" # Must be "TCP" or "UDP"
  port              = 22
  direction         = "inbound" # Must be "inbound" or "outbound"
  ip_range          = "0.0.0.0/0" # Must be a valid CIDR IPv4 or IPv6 notation
}
```

- `id` (String, Computed): Security Rule ID.
- `security_group_id` (String, Required): Security Group related to this rule.
- `name` (String, Required): Security Rule name.
- `description` (String, Optional): Security Rule description.
- `protocol` (String, Required): Security Rule protocol ("tcp" or "udp").
- `port` (Number, Required): Security Rule port.
- `direction` (String, Required): Security Rule direction ("ingress" or "egress").
- `ip_range` (String, Required): Security Rule IP Range (CIDR notation).
- `action` (String, Required): Security Rule action ("allow" or "deny").

#### `uca_security_group_attachment`

Attach a Security Group to a Server.

```tf
resource "uca_security_group_attachment" "my_attachment" {
  security_group_id = uca_security_group.my_sg.id
  server_id         = uca_server.my_server.id
}
```

- `security_group_id` (String, Required): Security Group ID to attach.
- `server_id` (String, Required): Server ID to attach the security group to.

