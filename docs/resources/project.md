---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "bamboo_project Resource - bamboo"
subcategory: ""
description: |-
  This resource define project.
  The priority block has a priority that defines the final assigned permissions of the user or group.
---

# bamboo_project (Resource)

This resource define project.

The priority block has a priority that defines the final assigned permissions of the user or group.



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `key` (String) Project key.
- `name` (String) Project name.

### Optional

- `assignment_version` (String) Assignment version, used to force update the permission.
- `assignments` (Block List) Assignment block (see [below for nested schema](#nestedblock--assignments))
- `description` (String) Project description.
- `retain_on_delete` (Boolean) Default value is `true`, and if the value set to `false` when the resource destroyed, the project will be removed.

### Read-Only

- `computed_groups` (Attributes List) Computed assignment. (see [below for nested schema](#nestedatt--computed_groups))
- `computed_users` (Attributes List) Computed assignment. (see [below for nested schema](#nestedatt--computed_users))

<a id="nestedblock--assignments"></a>
### Nested Schema for `assignments`

Required:

- `permissions` (List of String) List of permissions assignable to the users and groups (READ, VIEWCONFIGURATION, WRITE, BUILD, CLONE, CREATE, CREATEREPOSITORY, ADMINISTRATION)
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
