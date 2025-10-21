package utils

import (
	"bytes"
	"canvas_service/types"
	"fmt"

	svg "github.com/ajstarks/svgo"
)

func GenerateSVGFromShapes(shapes []types.Shape, width, height int) string {
	var buf bytes.Buffer
	canvas := svg.New(&buf)
	canvas.Start(width, height)

	for _, s := range shapes {
		switch s.Type {
		case "line":
			style := getStyle(s, false)
			canvas.Line(s.X1, s.Y1, s.X2, s.Y2, style)
		case "pixel":
			pixelStyle := fmt.Sprintf("fill:%s", s.Color)
			canvas.Circle(s.X1, s.Y1, 1, pixelStyle)
		case "circle":
			style := getStyle(s, true)
			canvas.Circle(s.X1, s.Y1, s.Radius, style)
		case "pen":
		case "stroke":
			style := getStyle(s, true)
			if len(s.Points) < 2 {
				continue
			}
			xCoords := make([]int, 0, len(s.Points)/2)
			yCoords := make([]int, 0, len(s.Points)/2)
			for i := 0; i < len(s.Points); i += 2 {
				if i+1 < len(s.Points) {
					xCoords = append(xCoords, int(s.Points[i]))
					yCoords = append(yCoords, int(s.Points[i+1]))
				}
			}
			canvas.Polyline(xCoords, yCoords, style)
		case "polygon":
			style := getStyle(s, false)
			if len(s.Points) < 2 {
				continue
			}
			xCoords := make([]int, 0, len(s.Points)/2)
			yCoords := make([]int, 0, len(s.Points)/2)
			for i := 0; i < len(s.Points); i += 2 {
				if i+1 < len(s.Points) {
					xCoords = append(xCoords, int(s.Points[i]))
					yCoords = append(yCoords, int(s.Points[i+1]))
				}
			}
			canvas.Polygon(xCoords, yCoords, style)
		}
	}
	canvas.End()
	return buf.String()
}

func getStyle(s types.Shape, noFill bool) string {
	if noFill {
		return fmt.Sprintf("fill:none;stroke:%s;stroke-width:%.f", s.Color, s.StrokeWidth)
	}
	return fmt.Sprintf("stroke:%s;stroke-width:%.f", s.Color, s.StrokeWidth)
}
