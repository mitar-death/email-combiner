package droparea

import (
	"image/color"
	"website-copier/cmd/utils"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

type DropAreaWidget struct {
	widget.BaseWidget
	FilePath  *string
	FileEntry *widget.Entry
	Window    fyne.Window
	Label     string
}

var _ fyne.Widget = (*DropAreaWidget)(nil)
var _ fyne.Tappable = (*DropAreaWidget)(nil)

func NewDropAreaWidget(filePath *string, fileEntry *widget.Entry, win fyne.Window, label string) *DropAreaWidget {
	da := &DropAreaWidget{
		FilePath:  filePath,
		FileEntry: fileEntry,
		Window:    win,
		Label:     label,
	}
	da.ExtendBaseWidget(da)
	return da
}

func (d *DropAreaWidget) CreateRenderer() fyne.WidgetRenderer {
	rect := canvas.NewRectangle(color.NRGBA{R: 220, G: 220, B: 220, A: 255})
	rect.SetMinSize(fyne.NewSize(400, 100))
	label := widget.NewLabel(d.Label)
	objects := []fyne.CanvasObject{rect, label}

	return &dropAreaRenderer{
		widget:  d,
		rect:    rect,
		label:   label,
		objects: objects,
	}
}

type dropAreaRenderer struct {
	widget  *DropAreaWidget
	rect    *canvas.Rectangle
	label   *widget.Label
	objects []fyne.CanvasObject
}

func (r *dropAreaRenderer) Layout(size fyne.Size) {
	r.rect.Resize(size)
	labelSize := r.label.MinSize()
	r.label.Move(fyne.NewPos(
		(size.Width-labelSize.Width)/2,
		(size.Height-labelSize.Height)/2,
	))
}

func (r *dropAreaRenderer) MinSize() fyne.Size {
	return fyne.NewSize(400, 100)
}

func (r *dropAreaRenderer) Refresh() {
	canvas.Refresh(r.widget)
}

func (r *dropAreaRenderer) Destroy() {}

func (r *dropAreaRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

// Implement Tappable interface
func (d *DropAreaWidget) Tapped(event *fyne.PointEvent) {
	utils.ShowFileOpenDialog(d.FilePath, d.FileEntry, d.Window)
}

func (d *DropAreaWidget) TappedSecondary(event *fyne.PointEvent) {}
