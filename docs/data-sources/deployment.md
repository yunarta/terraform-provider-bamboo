---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "bamboo_deployment Data Source - bamboo"
subcategory: ""
description: |-
  This data source used define a lookup of deployment project by name.
---

# bamboo_deployment (Data Source)

This data source used define a lookup of deployment project by name.



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) Deployment project name.

### Read-Only

- `groups` (Map of List of String) A map with the permission as the key and list of groups as the value.
- `id` (String) Computed deployment id.
- `users` (Map of List of String) A map with the permission as the key and list of users as the value.
