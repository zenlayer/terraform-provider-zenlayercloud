package zenlayercloud

const (
	VmChargeTypePostpaid                    = "POSTPAID"
	VmChargeTypePrepaid                     = "PREPAID"
	VmInternetChargeTypeBandwidth           = "ByBandwidth"
	VmInternetChargeTypeTrafficPackage      = "ByTrafficPackage"
	VmInternetChargeTypeInstanceBandwidth95 = "ByInstanceBandwidth95"
	VmInternetChargeTypeClusterBandwidth95  = "ByClusterBandwidth95"

	VmInstanceStatusPending      = "PENDING"
	VmInstanceStatusDeloying     = "DEPLOYING"
	VmInstanceStatusCreateFailed = "CREATE_FAILED"
	VmInstanceStatusRebuilding   = "REBUILDING"
	VmInstanceStatusStopped      = "STOPPED"
	VmInstanceStatusRunning      = "RUNNING"
	VmInstanceStatusBooting      = "BOOTING"
	VmInstanceStatusStopping     = "STOPPING"
	VmInstanceStatusReleasing    = "RELEASING"
	VmInstanceStatusRecycle      = "RECYCLE"

	VmImageTypePublic = "PUBLIC_IMAGE"
	VmImageTypeCustom = "CUSTOM_IMAGE"

	ImageStatusAvailable = "AVAILABLE"

	VmImageStatusCreating = "CREATING"

	VmImageStatusUnavailable = "UNAVAILABLE"

	VmSubnetStatusCreating  = "Creating"
	VmSubnetStatusAvailable = "Available"
	VmSubnetStatusFailed    = "Failed"

	VmDiskStatusInUse     = "IN_USE"
	VmDiskStatusAvailable = "AVAILABLE"
	VmDiskStatusAttaching = "ATTACHING"
	VmDiskStatusDetaching = "DETACHING"
	VmDiskStatusCreating  = "CREATING"
	VmDiskStatusDeleting  = "DELETING"
	VmDiskStatusRecycle   = "RECYCLED"
)

var (
	SecurityGroupRuleDirection = []string{
		"ingress", "egress",
	}

	SecurityGroupRuleIpProtocol = []string{
		"tcp", "udp", "icmp", "all",
	}

	SecurityGroupRulePolicy = []string{
		"accept",
	}
	VmImageTypes = []string{
		VmImageTypePublic,
		VmImageTypeCustom,
	}
	ImageCategories = []string{
		"CentOS", "Windows", "Ubuntu", "Debian",
	}
	VmOsTypes = []string{
		"windows", "linux",
	}

	VmChargeTypes = []string{
		VmChargeTypePostpaid,
		VmChargeTypePrepaid,
	}

	VmInternetChargeTypes = []string{
		VmInternetChargeTypeBandwidth,
		VmInternetChargeTypeTrafficPackage,
		VmInternetChargeTypeInstanceBandwidth95,
		VmInternetChargeTypeClusterBandwidth95,
	}

	VmInstanceOperatingStatus = []string{
		VmInstanceStatusPending,
		VmInstanceStatusDeloying,
		VmInstanceStatusStopping,
		VmInstanceStatusBooting,
		VmInstanceStatusReleasing,
		VmInstanceStatusRebuilding,
	}
)

func vmInstanceIsOperating(instanceStatus string) bool {
	return IsContains(VmInstanceOperatingStatus, instanceStatus)
}
