package tabula

import (
	"errors"
	"github.com/pdbogen/mapbot/model/context"
	"github.com/pdbogen/mapbot/model/mark"
	"image"
	"image/draw"
)

func (t *Tabula) addMarks(in image.Image, ctx context.Context) error {
	dirMarkSlice := []mark.Mark{}
	for _, dirMarks := range ctx.GetMarks(*t.Id) {
		for _, mark := range dirMarks {
			dirMarkSlice = append(dirMarkSlice, mark)
		}
	}

	if err := t.addMarkSlice(in, dirMarkSlice); err != nil {
		return err
	}

	return t.addMarkSlice(in, t.Marks)
}

func (t *Tabula) addMarkSlice(in image.Image, marks []mark.Mark) error {
	drawable, ok := in.(draw.Image)
	if !ok {
		return errors.New("image provided could not be used as a draw.Image")
	}

	for _, mark := range marks {
		switch mark.Direction {
		case "n":
			t.squareAtFloat(drawable, float32(mark.Point.X), float32(mark.Point.Y)-.1, float32(mark.Point.X)+1, float32(mark.Point.Y)+.1, 0, mark.Color)
		case "s":
			t.squareAtFloat(drawable, float32(mark.Point.X), float32(mark.Point.Y)+.9, float32(mark.Point.X)+1, float32(mark.Point.Y)+1.1, 0, mark.Color)
		case "e":
			t.squareAtFloat(drawable, float32(mark.Point.X)+.9, float32(mark.Point.Y), float32(mark.Point.X)+1.1, float32(mark.Point.Y)+1, 0, mark.Color)
		case "w":
			t.squareAtFloat(drawable, float32(mark.Point.X)-.1, float32(mark.Point.Y), float32(mark.Point.X)+.1, float32(mark.Point.Y)+1, 0, mark.Color)
		case "ne":
			t.squareAtFloat(drawable, float32(mark.Point.X)+.9, float32(mark.Point.Y)-.1, float32(mark.Point.X)+1.1, float32(mark.Point.Y)+.1, 0, mark.Color)
		case "se":
			t.squareAtFloat(drawable, float32(mark.Point.X)+.9, float32(mark.Point.Y)+.9, float32(mark.Point.X)+1.1, float32(mark.Point.Y)+1.1, 0, mark.Color)
		case "nw":
			t.squareAtFloat(drawable, float32(mark.Point.X)-.1, float32(mark.Point.Y)-.1, float32(mark.Point.X)+.1, float32(mark.Point.Y)+.1, 0, mark.Color)
		case "sw":
			t.squareAtFloat(drawable, float32(mark.Point.X)-.1, float32(mark.Point.Y)+.9, float32(mark.Point.X)+.1, float32(mark.Point.Y)+1.1, 0, mark.Color)
		default:
			t.squareAt(drawable, image.Rect(mark.Point.X, mark.Point.Y, mark.Point.X+1, mark.Point.Y+1), 1, mark.Color)
		}
	}

	return nil
}

func (t *Tabula) WithMarks(marks []mark.Mark) *Tabula {
	t.Marks = make([]mark.Mark, len(marks))
	for i, m := range marks {
		t.Marks[i] = m
	}
	return t
}