package zec

const (
	INVALID_VPC_NOT_FOUND  = "INVALID_VPC_NOT_FOUND"
	INVALID_DISK_NOT_FOUND = "INVALID_DISK_NOT_FOUND"
	INVALID_NIC_NOT_FOUND  = "INVALID_NIC_NOT_FOUND"
	INVALID_VPC_ROUTE_NOT_FOUND  = "INVALID_VPC_ROUTE_NOT_FOUND"

	// ZecDiskStatusRecycle Disk Status
	ZecDiskStatusRecycle   = "RECYCLED"
	ZecDiskStatusRecycling = "RECYCLING"
	ZecDiskStatusAttaching = "ATTACHING"
	ZecDiskStatusDetaching = "DETACHING"
	ZecDiskStatusCreating  = "CREATING"
	ZecDiskStatusDeleting  = "DELETING"
	ZecDiskStatusResizing  = "CHANGING"
	ZecDiskStatusInUse     = "IN_USE"
	ZecDiskStatusAvailable = "AVAILABLE"

	ZecEipStatusCreating     = "CREATING"
	ZecEipStatusCreateFailed = "CREATE_FAILED"
	ZecEipStatusBINDED       = "BINDED"
	ZecEipStatusAvailable    = "UNBIND"
	ZecEipStatusDeleting     = "DELETING"
	ZecEipStatusRecycle      = "RECYCLED"
	ZecEipStatusRecycling    = "RECYCLING"

	ZecInstanceStatusPending      = "PENDING"
	ZecInstanceStatusDeloying     = "DEPLOYING"
	ZecInstanceStatusCreateFailed = "CREATE_FAILED"
	ZecInstanceStatusRebuilding   = "REBUILDING"
	ZecInstanceStatusStopped      = "STOPPED"
	ZecInstanceStatusRunning      = "RUNNING"
	ZecInstanceStatusBooting      = "BOOTING"
	ZecInstanceStatusStopping     = "STOPPING"
	ZecInstanceStatusReleasing    = "RELEASING"
	ZecInstanceStatusRecycle      = "RECYCLE"
	ZecInstanceStatusRecycling    = "RECYCLING"
	ZecInstanceStatusResizing     = "RESIZING"

	ZecVnicStatusCreating     = "PENDING"
	ZecVnicStatusAvailable    = "AVAILABLE"
	ZecVnicStatusAttaching    = "ATTACHING"
	ZecVnicStatusDetaching    = "DETACHING"
	ZecVnicStatusDeleting     = "DELETING"
	ZecVnicStatusCreateFailed = "CREATE_FAILED"
	ZecVnicStatusUsed         = "USED"
)

var (
	InstanceOperatingStatus = []string{
		ZecInstanceStatusPending,
		ZecInstanceStatusDeloying,
		ZecInstanceStatusRebuilding,
		ZecInstanceStatusBooting,
		ZecInstanceStatusStopping,
		ZecInstanceStatusReleasing,
		ZecInstanceStatusRecycling,
		ZecInstanceStatusResizing,
	}

	ImageTypes = []string{
		"PUBLIC_IMAGE",
		"CUSTOM_IMAGE",
	}
)
