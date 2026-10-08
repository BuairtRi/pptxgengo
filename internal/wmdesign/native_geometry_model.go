package wmdesign

import "github.com/buairtri/pptxgengo/pptx"

type NativeConnectorRoute = pptx.ConnectorRoute

type NativeGeometry struct {
	Route                *NativeConnectorRoute `json:"route,omitempty"`
	SourceGeometrySHA256 string                `json:"source_geometry_sha256"`
	Kind                 string                `json:"kind"`
	Parent               string                `json:"parent,omitempty"`
	X                    float64               `json:"x_pt"`
	Y                    float64               `json:"y_pt"`
	W                    float64               `json:"width_pt"`
	H                    float64               `json:"height_pt"`
	Rotation             float64               `json:"rotation_deg,omitempty"`
	FlipH                bool                  `json:"flip_h,omitempty"`
	FlipV                bool                  `json:"flip_v,omitempty"`
	Child                *Rect                 `json:"child_space,omitempty"`
}

// Padding is expressed in the container's native placement axes, before its
// rotation and parent scaling. This is an allocation contract, not an ink bound.
type DiagramPadding struct {
	Top    float64 `json:"top_pt"`
	Right  float64 `json:"right_pt"`
	Bottom float64 `json:"bottom_pt"`
	Left   float64 `json:"left_pt"`
}
type DiagramContainment struct {
	Container string         `json:"container"`
	Padding   DiagramPadding `json:"padding"`
}
type DiagramContainmentObservation struct {
	Member        string         `json:"member"`
	Container     string         `json:"container"`
	Padding       DiagramPadding `json:"padding"`
	Clearance     DiagramPadding `json:"clearance"`
	ContainerRect Rect           `json:"container_rect"`
	MemberBounds  Rect           `json:"member_bounds_in_container_axes"`
}
