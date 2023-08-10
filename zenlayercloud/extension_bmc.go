package zenlayercloud

const (
	ResourceTypeInstance = "instance"
	ResourceTypeEip      = "eip"
	ResourceTypeDdosIp   = "ddos_ip"
	ResourceTypeSubnet   = "subnet"
	ResourceTypeCidrIPv4 = "cidr_ipv4"
	ResourceTypeCidrIPv6 = "cidr_ipv6"

	BmcChargeTypePostpaid                    = "POSTPAID"
	BmcChargeTypePrepaid                     = "PREPAID"
	BmcInternetChargeTypeBandwidth           = "ByBandwidth"
	BmcInternetChargeTypeTrafficPackage      = "ByTrafficPackage"
	BmcInternetChargeTypeInstanceBandwidth95 = "ByInstanceBandwidth95"
	BmcInternetChargeTypeClusterBandwidth95  = "ByClusterBandwidth95"

	BmcInstanceStatusPending       = "PENDING"
	BmcInstanceStatusCreating      = "CREATING"
	BmcInstanceStatusCreateFailed  = "CREATE_FAILED"
	BmcInstanceStatusInstalling    = "INSTALLING"
	BmcInstanceStatusInstallFailed = "INSTALL_FAILED"
	BmcInstanceStatusRunning       = "RUNNING"
	BMC_INSTANCE_STATUS_STOPPED    = "STOPPED"
	BmcInstanceStatusBooting       = "BOOTING"
	BmcInstanceStatusStopping      = "STOPPING"
	BmcInstanceStatusRecycle       = "RECYCLE"
	BmcInstanceStatusRecycling     = "RECYCLING"

	BmcSubnetStatusAvailable         = "AVAILABLE"
	BmcSubnetStatusCreating          = "CREATING"
	BmcSubnetStatusPending           = "PENDING"
	BmcSubnetStatusDeleting          = "DELETING"
	BmcSubnetStatusCreateFailed      = "CREATE_FAILED"
	BmcSubnetStatusAssociate         = "ASSOCIATING"
	BmcSubnetInstanceStatusBinding   = "BINDING"
	BmcSubnetInstanceStatusUnbinding = "UNBINDING"
	BmcSubnetInstanceStatusBound     = "BOUND"

	BmcVpcStatusCreating        = "CREATING"
	BmcVpcStatusCreateFailed    = "CREATE_FAILED"
	BmcVpcStatusCreateAvailable = "AVAILABLE"
	BmcVpcStatusDeleting        = "DELETING"

	BmcEipStatusCreating      = "CREATING"
	BmcEipStatusCreateFailed  = "CREATE_FAILED"
	BmcEipStatusAssociating   = "ASSOCIATING"
	BmcEipStatusUnAssociating = "UNASSOCIATING"
	BmcEipStatusAssociated    = "ASSOCIATED"
	BmcEipStatusAvailable     = "AVAILABLE"
	BmcEipStatusReleasing     = "RELEASING"
	BmcEipStatusRecycle       = "RECYCLE"
	BmcEipStatusRecycling     = "RECYCLING"

	ImageTypePublic = "PUBLIC_IMAGE"
	ImageTypeCustom = "CUSTOM_IMAGE"
)

var (
	ImageTypes = []string{
		ImageTypePublic,
		ImageTypeCustom,
	}

	ImageCatalogs = []string{
		"centos", "windows", "ubuntu", "debian", "esxi",
	}

	OsTypes = []string{
		"windows", "linux",
	}

	BmcChargeTypes = []string{
		BmcChargeTypePostpaid,
		BmcChargeTypePrepaid,
	}
	BmcInternetChargeTypes = []string{
		BmcInternetChargeTypeBandwidth,
		BmcInternetChargeTypeTrafficPackage,
		BmcInternetChargeTypeInstanceBandwidth95,
		BmcInternetChargeTypeClusterBandwidth95,
	}
	InstanceOperatingStatus = []string{
		BmcInstanceStatusPending,
		BmcInstanceStatusStopping,
		BmcInstanceStatusBooting,
		BmcInstanceStatusInstalling,
		BmcInstanceStatusRecycling,
	}
	SubnetOperatingStatus = []string{
		BmcSubnetStatusCreating,
		BmcSubnetStatusPending,
		BmcSubnetStatusAssociate,
	}

	VpcOperatingStatus = []string{
		BmcVpcStatusDeleting,
		BmcVpcStatusCreating,
	}

	EipOperatingStatus = []string{
		BmcEipStatusCreating,
		BmcEipStatusAssociating,
		BmcEipStatusUnAssociating,
		BmcEipStatusReleasing,
	}
)

func instanceIsOperating(instanceStatus string) bool {
	return IsContains(InstanceOperatingStatus, instanceStatus)
}

func subnetIsOperating(subnetStatus string) bool {
	return IsContains(SubnetOperatingStatus, subnetStatus)
}
