# Notifications

OTF can send notifications for run state transitions. OTF implements the [TFC notifications API](https://developer.hashicorp.com/terraform/cloud-docs/api-docs/notification-configurations), which means you can use the same documented API endpoints to configure notifications. Alternatively you can use the [`tfe` terraform provider](https://registry.terraform.io/providers/hashicorp/tfe/latest/docs/resources/notification_configuration).

!!! note
	Currently you cannot configure notifications via the UI.

Support exists for the following destination types:

* `generic`: Generic HTTP POST notifications
* `slack`: Slack messages
* `gcppubsub`: GCP Pub/Sub topic messages (*OTF specific)

!!! note
	Currently there is no support for the `email` or `microsoft-teams`
	destination types (which TFC *does* support).

## GCP Pub Sub

OTF can send notifications to a [GCP Pub/Sub
topic](https://cloud.google.com/pubsub/docs/overview). To configure these
notifications see the [TFC notifications API
documentation](https://developer.hashicorp.com/terraform/cloud-docs/api-docs/notification-configurations#create-a-notification-configuration), in particular the endpoint for creating a notification configuration.

For the `destination-type` field, use `gcppubsub`.

For the `url` field, enter `gcppubsub://<project-id>/<topic>`, where `<project_id>` is the GCP project ID and `<topic>` is the Pub/Sub topic ID.

Ensure `otfd` has access to default credentials for a service account which has
necessary permissions to publish messages to the configured topic.

The payload of the messages is the same as that [documented for the `generic`
destination
type](https://developer.hashicorp.com/terraform/cloud-docs/api-docs/notification-configurations#run-notification-payload) (using the JSON format).

Additionally, attributes are added to each message:

|key|value|
|-|-|
|`otf.ninja/v1/workspace.name`|`<workspace_name>`|
|`otf.ninja/v1/workspace.id`|`<workspace_id>`|
|`otf.ninja/v1/tags/<tag_name>`|`true`|

Attributes permit you to [filter messages from a subscription](https://cloud.google.com/pubsub/docs/subscription-message-filter#filtering_syntax) in GCP
