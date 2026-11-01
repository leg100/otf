package ui

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseApplyView(t *testing.T) {
	// Realistic apply log: ANSI colors, progress lines, a create, an
	// update, a replace (destroy + create on the same address) and an
	// operation still in progress.
	log := "\x1b[0m\x1b[1mmodule.router.uapi_dhcp_host.lease[\"alpine-bd-psql\"]: Creating...\x1b[0m\n" +
		"module.router.uapi_dhcp_host.lease[\"alpine-bd-psql\"]: Still creating... [10s elapsed]\n" +
		"module.router.uapi_dhcp_host.lease[\"alpine-bd-psql\"]: Creation complete after 12s [id=h_01ktr0]\n" +
		"module.router.uapi_network_interface.wan: Modifying... [id=wan]\n" +
		"module.router.uapi_network_interface.wan: Modifications complete after 2s [id=wan]\n" +
		"module.router.uapi_firewall_redirect.r[\"seedbox\"]: Destroying... [id=r_01m68v]\n" +
		"module.router.uapi_firewall_redirect.r[\"seedbox\"]: Destruction complete after 0s\n" +
		"module.router.uapi_firewall_redirect.r[\"seedbox\"]: Creating...\n" +
		"module.router.uapi_firewall_redirect.r[\"seedbox\"]: Creation complete after 1s [id=r_01m68v]\n" +
		"module.router.uapi_sqm_queue.q[\"vlan32\"]: Creating...\n" +
		"OpenTofu used the selected providers to generate the following execution\n" +
		"Apply complete! Resources: 3 added, 1 changed, 1 destroyed.\n"

	changes := parseApplyView(log)
	require.Len(t, changes, 5)

	assert.Equal(t, applyChangeView{
		Address:  `module.router.uapi_dhcp_host.lease["alpine-bd-psql"]`,
		Action:  "create",
		Duration: "12s",
		ID:       "h_01ktr0",
		Complete: true,
	}, changes[0])
	assert.Equal(t, applyChangeView{
		Address:  "module.router.uapi_network_interface.wan",
		Action:  "update",
		Duration: "2s",
		ID:       "wan",
		Complete: true,
	}, changes[1])
	assert.Equal(t, applyChangeView{
		Address:  `module.router.uapi_firewall_redirect.r["seedbox"]`,
		Action:  "destroy",
		Duration: "0s",
		ID:       "r_01m68v", // carried over from the Destroying line
		Complete: true,
	}, changes[2])
	assert.Equal(t, applyChangeView{
		Address:  `module.router.uapi_firewall_redirect.r["seedbox"]`,
		Action:  "create",
		Duration: "1s",
		ID:       "r_01m68v",
		Complete: true,
	}, changes[3])
	assert.Equal(t, applyChangeView{
		Address:  `module.router.uapi_sqm_queue.q["vlan32"]`,
		Action:  "create",
		Complete: false, // no completion line yet
	}, changes[4])

	assert.Equal(t, 3, countApplyChanges(changes, "create"))
	assert.Equal(t, 1, countApplyChanges(changes, "update"))
	assert.Equal(t, 1, countApplyChanges(changes, "destroy"))
}

func TestParseApplyViewEmpty(t *testing.T) {
	assert.Empty(t, parseApplyView("no operations here\n"))
	assert.Empty(t, parseApplyView(""))
}
