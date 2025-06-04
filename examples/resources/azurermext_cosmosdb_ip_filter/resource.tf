resource "azurermext_cosmosdb_ip_range_filter" "example" {
  cosmosdb_account_id = "xxx" # attribute 'id' of an azurerm_cosmosdb_account

  ip_rules = ["4.210.172.107", "13.88.56.148", "13.91.105.0/24"]
}
