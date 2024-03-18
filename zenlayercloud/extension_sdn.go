package zenlayercloud

const (
	SdnStatusCreating   = "CREATING"
	SdnStatusDeleting   = "DELETING"
	SdnStatusUpdating   = "UPDATING"
	SdnStatusRunning    = "RUNNING"
	SdnStatusRecovering = "RECOVERING"
	SdnStatusRecycle    = "RECYCLED"
	SdnStatusReleasing  = "DESTROYING"
)

var (
	SdnOperatingStatus = []string{
		SdnStatusCreating,
		SdnStatusDeleting,
		SdnStatusUpdating,
		SdnStatusRecovering,
		SdnStatusReleasing,
	}
)

const (
	ROUTE_TYPE_BGP    = "BGP"
	ROUTE_TYPE_STATIC = "STATIC"
)

var ROUTE_TYPES = []string{ROUTE_TYPE_BGP, ROUTE_TYPE_STATIC}

const (
	POINT_TYPE_PORT    = "PORT"
	POINT_TYPE_TENCENT = "TENCENT"
	POINT_TYPE_AWS     = "AWS"
	POINT_TYPE_GOOGLE  = "GOOGLE"
	POINT_TYPE_VPC     = "VPC"
)

var ENDPOINT_TYPES = []string{POINT_TYPE_PORT, POINT_TYPE_TENCENT, POINT_TYPE_AWS, POINT_TYPE_GOOGLE}
var CLOUD_ENDPOINT_TYPES = []string{POINT_TYPE_TENCENT, POINT_TYPE_AWS, POINT_TYPE_GOOGLE}
var EDGE_POINT_TYPES = []string{POINT_TYPE_VPC, POINT_TYPE_PORT, POINT_TYPE_TENCENT, POINT_TYPE_AWS, POINT_TYPE_GOOGLE}

const (
	PRODUCT_PRIVATE_CONNECT = "PrivateConnect"
	PRODUCT_CLOUD_ROUTER    = "CloudRouter"
)

var PRODUCT_TYPES = []string{PRODUCT_PRIVATE_CONNECT, PRODUCT_CLOUD_ROUTER}

func IsOperating(status string) bool {
	return IsContains(SdnOperatingStatus, status)
}
