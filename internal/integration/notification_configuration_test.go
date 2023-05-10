package integration

import "testing"

func TestIntegration_NotificationConfigurationService(t *testing.T) {
	t.Parallel()

	svc := setup(t, nil)
	org := svc.createOrganization(t, ctx)
}
