---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "bamboo_deployment Resource - bamboo"
subcategory: ""
description: |-
  This resource define deployment.
  In order for the execution to be successful, the user must have user access to all the specified repositories.
---

# bamboo_deployment (Resource)

This resource define deployment.

In order for the execution to be successful, the user must have user access to all the specified repositories.



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) Name of the deployment.
- `plan_key` (String) Plan key that will be the source of the deployment.

### Optional

- `assignment_version` (String) Assignment version, used to force update the permission.
- `assignments` (Block List) Assignment block (see [below for nested schema](#nestedblock--assignments))
- `description` (String) Description the deployment.
- `repositories` (List of String) This deployment will add this list of linked repositories into its permission.
- `retain_on_delete` (Boolean) Default value is `true`, and if the value set to `false` when the resource destroyed, the deployment will be removed.

### Read-Only

- `computed_groups` (Attributes List) Computed assignment. (see [below for nested schema](#nestedatt--computed_groups))
- `computed_users` (Attributes List) Computed assignment. (see [below for nested schema](#nestedatt--computed_users))
- `id` (String) Numeric id of the deployment.
- `repository_specs_managed` (Boolean) Computer value that defines the repository is managed by spec.

<a id="nestedblock--assignments"></a>
### Nested Schema for `assignments`

Required:

- `permissions` (List of String) List of permissions assignable to the users and groups (READ, CREATE, CREATEREPOSITORY, ADMINISTRATION, CLONE, WRITE, BUILD, VIEWCONFIGURATION)
- `priority` (Number) Priority of this block

Optional:

- `groups` (List of String) List of group names.
- `users` (List of String) List of usernames.


<a id="nestedatt--computed_groups"></a>
### Nested Schema for `computed_groups`

Read-Only:

- `name` (String) Name of the entity in the assignment.
- `permissions` (List of String) List of permission owned by the entity in the assignment.


<a id="nestedatt--computed_users"></a>
### Nested Schema for `computed_users`

Read-Only:

- `name` (String) Name of the entity in the assignment.
- `permissions` (List of String) List of permission owned by the entity in the assignment.