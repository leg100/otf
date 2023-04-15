# Authorization

The authorization model largely follows that of TFC/E. An organization comprises a number of teams. A user is a member of one or more teams.

### Owners team

Every organization has an `owners` team. The user that creates an organization becomes its owner. The owners team must have at least one member and it cannot be deleted.

Members of the owners team enjoy broad privileges across an organization. "Owners" are the only users permitted to alter organization-level permissions. They are also automatically assigned all the organization-level permissions; these permissions cannot be unassigned.

### Synchronisation

Upon signing in, a user's organizations and teams are synchronised or "mapped" to those of their SSO provider. If an organization or team does not exist it is created.

The mapping varies according to the SSO provider. If the provider doesn't have the concept of an organization or team then equivalent units of authorization are used. Special rules apply to the mapping of the Owners team too. The exact mappings for each provider are listed here:

|provider|organization|team|owners|
|-|-|-|-|
|Github|organization|team|admin role or "owners" team|
|Gitlab|top-level group|access level|owners access level|

### Personal organization

A user is assigned a personal organization matching their username. They are automatically an owner of this organization. The organization is created the first time the user logs in.

### Permissions

Permissions are assigned to teams on two levels: organizations and workspaces. Organization permissions confer privileges across the organization:

* Manage Workspaces: Allows members to create and administrate all workspaces within the organization.
* Manage VCS Settings: Allows members to manage the set of VCS providers available within the organization.
* Manage Registry: Allows members to publish and delete modules within the organization.

Workspace permissions confer privileges on the workspace alone, and are based on the [fixed permission sets of TFC/TFE](https://developer.hashicorp.com/terraform/cloud-docs/users-teams-organizations/permissions#fixed-permission-sets):

* Read
* Plan
* Write
* Admin

See the [TFC/TFE documentation](https://developer.hashicorp.com/terraform/cloud-docs/users-teams-organizations/permissions#fixed-permission-sets) for more information on the privileges each permission set confers.
