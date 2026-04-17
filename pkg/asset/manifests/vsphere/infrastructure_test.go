package vsphere

import (
	"testing"

	"github.com/stretchr/testify/assert"

	configv1 "github.com/openshift/api/config/v1"
	"github.com/openshift/installer/pkg/asset/installconfig"
	"github.com/openshift/installer/pkg/ipnet"
	"github.com/openshift/installer/pkg/types"
	vsphereconfig "github.com/openshift/installer/pkg/types/vsphere"
)

func TestGetInfraPlatformSpecHostGroupAffinity(t *testing.T) {
	ic := installconfig.MakeAsset(&types.InstallConfig{
		Networking: &types.Networking{
			MachineNetwork: []types.MachineNetworkEntry{
				{CIDR: *ipnet.MustParseCIDR("10.0.0.0/16")},
			},
		},
		Platform: types.Platform{
			VSphere: &vsphereconfig.Platform{
				VCenters: []vsphereconfig.VCenter{
					{
						Server:      "test-vcenter",
						Port:        443,
						Datacenters: []string{"test-datacenter"},
					},
				},
				FailureDomains: []vsphereconfig.FailureDomain{
					{
						Name:       "fd-a",
						Region:     "region-a",
						Zone:       "zone-a",
						Server:     "test-vcenter",
						RegionType: vsphereconfig.ComputeClusterFailureDomain,
						ZoneType:   vsphereconfig.HostGroupFailureDomain,
						Topology: vsphereconfig.Topology{
							Datacenter:     "test-datacenter",
							ComputeCluster: "/test-datacenter/host/test-cluster",
							Datastore:      "/test-datacenter/datastore/test-datastore",
							Networks:       []string{"test-portgroup"},
							HostGroup:      "host-group-a",
						},
					},
				},
			},
		},
	})

	platformSpec := GetInfraPlatformSpec(ic, "cluster-id")
	if !assert.Len(t, platformSpec.FailureDomains, 1) {
		return
	}

	fd := platformSpec.FailureDomains[0]
	if assert.NotNil(t, fd.RegionAffinity) {
		assert.Equal(t, configv1.ComputeClusterFailureDomainRegion, fd.RegionAffinity.Type)
	}
	if assert.NotNil(t, fd.ZoneAffinity) {
		assert.Equal(t, configv1.HostGroupFailureDomainZone, fd.ZoneAffinity.Type)
		if assert.NotNil(t, fd.ZoneAffinity.HostGroup) {
			assert.Equal(t, "host-group-a", fd.ZoneAffinity.HostGroup.HostGroup)
			assert.Equal(t, "cluster-id-fd-a", fd.ZoneAffinity.HostGroup.VMGroup)
			assert.Equal(t, "cluster-id-fd-a", fd.ZoneAffinity.HostGroup.VMHostRule)
		}
	}
}
