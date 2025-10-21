package types

type Point struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type Shape struct {
	LayerNumber int       `json:"layer_number"`
	Type        string    `json:"type"`
	X1          int       `json:"x1,omitempty"`
	Y1          int       `json:"y1,omitempty"`
	X2          int       `json:"x2,omitempty"`
	Y2          int       `json:"y2,omitempty"`
	Radius      int       `json:"radius,omitempty"`
	Color       string    `json:"color"`
	StrokeWidth float64   `json:"stroke_width"`
	Points      []float64 `json:"points,omitempty"`
}

type SaveCanvasRequest struct {
	CanvasID *string `json:"canvasId,omitempty"`
	UserID   string  `json:"userId"`
	Layers   []Layer `json:"layers"`
	Shapes   []Shape `json:"shapes"`
}

type Layer struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Visible bool   `json:"visible"`
	Locked  bool   `json:"locked"`
}
