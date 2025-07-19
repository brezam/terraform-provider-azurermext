# terraform-provider-azurermext
Terraform provider For Extra Azure Resource Management resources/datasources .

# Configuration
In order for the provider to be configured, you do the same thing as [azurerm's configuration using service principal client id and client secret](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/guides/service_principal_client_secret).

You can configure it like this:
```terraform
provider "azurermext" {
  client_id     = "xxxx-xxxx-xxxx"        # Also available as environment variable ARM_CLIENT_ID
  client_secret = "xxxx-xxxx-xxxx"        # Also available as environment variable ARM_CLIENT_SECRET
  tenant_id     = "xxxx-xxxx-xxxx"        # Also available as environment variable ARM_TENANT_ID
}
```

It's recommended to use the environment variables option, especially for the cient secret.


# Resources/Data Sources
## [Resource] azurermext_cosmosdb_ip_range_filter
This resource manages the IP rules for a CosmosDB account.

Unlike the official `azurerm_cosmosdb_account` resource, this version intentionally ignores additional IPs. This behavior is its primary advantage; by not considering extra IPs, it avoids detecting drift caused by external IP modifications.

To prevent conflicts between the two resources, include an `ignore_changes` for the ip_range_filter property in the official resource.

# Examples
## azurermext_cosmosdb_ip_range_filter
This example showcases having a CosmosDB account and using this resource to take care of its IP rules:
```terraform
resource "azurerm_cosmosdb_account" "example" {
  ...
  # fill in all fields with the exception of ip_range_filter
  ...

  lifecycle {
    ignore_changes = [ip_range_filter] # this is necessary to avoid conflicts in later applies
  }
}

resource "azurermext_cosmosdb_ip_range_filter" "example" {
  cosmosdb_account_id = azurerm_cosmosdb_account.example.id
  ip_rules = ["4.210.172.107", "13.88.56.148", "13.91.105.0/24", ...] # list of ip and ip ranges to add as firewall rules
}
```

Important considerations:
- Any IPs not explicitly listed in the configuration are ignored.
- Adding an IP that already exists will have no effect during the apply phase.
However, if you attempt to remove an IP that exists in the current state, the API will be called to remove that IP.
- Destroying the resource doesn't change anything. If you want to remove all managed IPs, simply apply an empty list instead.
