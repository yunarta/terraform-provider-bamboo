---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "bamboo_linked_repository_dependency Resource - bamboo"
subcategory: ""
description: |-
  This resource define relationship where repository specified by id will requires access to list of specified required repositories.
  In order for the execution to be successful, the user must have admin access to all the required repositories.
---

# bamboo_linked_repository_dependency (Resource)

This resource define relationship where repository specified by id will requires access to list of specified required repositories.

In order for the execution to be successful, the user must have admin access to all the required repositories.



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `id` (String) Numeric id of the linked repository.
- `requires` (List of String) This repository will be added into to this list of linked repositories permissions.

### Optional

- `retain_on_delete` (Boolean) Default value is `true`, and if the value set to `false` when the resource destroyed, the permission will be removed.