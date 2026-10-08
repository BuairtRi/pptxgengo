package wmdesign

type NativeGeometry struct {
	SourceGeometrySHA256 string  `json:"source_geometry_sha256"`
	Kind                 string  `json:"kind"`
	Parent               string  `json:"parent,omitempty"`
	X                    float64 `json:"x_pt"`
	Y                    float64 `json:"y_pt"`
	W                    float64 `json:"width_pt"`
	H                    float64 `json:"height_pt"`
	Rotation             float64 `json:"rotation_deg,omitempty"`
	FlipH                bool    `json:"flip_h,omitempty"`
	FlipV                bool    `json:"flip_v,omitempty"`
	Child                *Rect   `json:"child_space,omitempty"`
}
