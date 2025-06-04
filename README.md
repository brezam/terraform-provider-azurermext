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

I recommend using the environment variables option.


# Resources/Data Sources
## [Resource] azurermext_cosmosdb_ip_range_filter
This resource manages the IP rules of a CosmosDB account.

Unlike the official `azurerm_cosmosdb_account` it ignores additional IPs.
That's literally the only reason to use this, to take advantage of this ignoring so additional IPs aren't considered drifts.

You should add the ip_range_filter of the official resource in its `ignore_changes` block to avoid both resources
from conflicting.

# Examples
## azurermext_cosmosdb_ip_range_filter
This example simulates having a CosmosDB account and using this resource to take care of IP rules:
```terraform
resource "azurerm_cosmosdb_account" "example" {
  ...
  # fill in all fields you want EXCEPT ip_range_filter
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

Whatever IPs the account has that aren't listed are ignored. If you add an IP that's already present nothing happens in apply.
If you try to remove an IP that's in the state then the API will try to remove that IP.
