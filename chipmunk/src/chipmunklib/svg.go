package chipmunklib

import (
	"encoding/xml"
	"fmt"

	"github.com/lucasb-eyer/go-colorful"
	"github.com/remogatto/mandala"
	"github.com/vova616/chipmunk"
	"github.com/vova616/chipmunk/vect"
)

type svgLine struct {
	X1 float32 `xml:"x1,attr"`
	Y1 float32 `xml:"y1,attr"`
	X2 float32 `xml:"x2,attr"`
	Y2 float32 `xml:"y2,attr"`
}

type svgRect struct {
	Width     float32 `xml:"width,attr"`
	Height    float32 `xml:"height,attr"`
	X         float32 `xml:"x,attr"`
	Y         float32 `xml:"y,attr"`
	Transform string  `xml:"transform,attr"`
}

type svgGroup struct {
	Transform string    `xml:"transform,attr"`
	Rects     []svgRect `xml:"rect"`
	Line      svgLine   `xml:"line"`
}

type svgFile struct {
	XMLName xml.Name   `xml:"svg"`
	Width   float32    `xml:"width,attr"`
	Height  float32    `xml:"height,attr"`
	Groups  []svgGroup `xml:"g"`
}

func (w *World) CreateFromSvg(filename string) {
	var svg svgFile

	responseCh := make(chan mandala.LoadResourceResponse)
	mandala.ReadResource(filename, responseCh)
	response := <-responseCh

	if response.Error != nil {
		mandala.Fatalf(response.Error.Error())
	}
	buf := response.Buffer

	err := xml.Unmarshal(buf, &svg)
	if err != nil {
		mandala.Fatalf(err.Error())
	}

	scaleX := float32(w.width) / svg.Width
	scaleY := float32(w.height) / svg.Height

	for _, group := range svg.Groups {
		for _, rect := range group.Rects {
			rX := (rect.X + rect.Width/2) * scaleX
			rY := float32(w.height) - (rect.Y+rect.Height/2)*scaleY
			rW := rect.Width * scaleX
			rH := rect.Height * scaleY
			box := newBox(w, rW, rH)
			pos := vect.Vect{
				vect.Float(rX),
				vect.Float(rY),
			}

			box.physicsBody.SetPosition(pos)

			if rect.Transform != "" {
				var a, b, c float32
				_, err = fmt.Sscanf(rect.Transform, "rotate(%f %f,%f)", &a, &b, &c)
				if err != nil {
					mandala.Fatalf(err.Error())
				}
				box.physicsBody.SetAngle(90 / chipmunk.DegreeConst)
			}

			box.openglShape.SetColor(colorful.HappyColor())
			w.addBox(box)
		}
	}
	line := svg.Groups[0].Line
	w.setGround(newGround(
		w,
		0,
		float32(w.height)-float32(line.Y1*scaleY),
		float32(w.width),
		float32(w.height)-float32(line.Y2*scaleY),
	))
}
