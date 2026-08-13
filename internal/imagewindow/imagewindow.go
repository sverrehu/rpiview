package imagewindow

import (
	"image"
	"log"
	"sync"

	"github.com/gogpu/gg"
	"github.com/gogpu/gg/integration/ggcanvas"
	"github.com/gogpu/gogpu"
	"github.com/sverrehu/goutils/image/scaling"
)

type ImageWindow struct {
	app   *gogpu.App
	img   *image.Image
	mutex *sync.Mutex
}

func NewImageWindow() *ImageWindow {
	mutex := sync.Mutex{}
	return &ImageWindow{app: nil, img: nil, mutex: &mutex}
}

func (w *ImageWindow) SetImage(img *image.Image) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	w.img = img
	w.openOrRedraw()
}

func (w *ImageWindow) openOrRedraw() {
	if w.app == nil {
		w.Open()
	}
	w.app.RequestRedraw()
}

func (w *ImageWindow) Open() {
	w.mutex.Lock()
	w.app = gogpu.NewApp(gogpu.DefaultConfig().
		WithTitle("Image Viewer").
		WithSize(800, 600))
	w.mutex.Unlock()
	var canvas *ggcanvas.Canvas
	w.app.OnDraw(func(dc *gogpu.Context) {
		w.mutex.Lock()
		defer w.mutex.Unlock()
		width, height := dc.Width(), dc.Height()
		if width <= 0 || height <= 0 {
			return
		}
		if canvas == nil {
			provider := w.app.GPUContextProvider()
			if provider == nil {
				return
			}
			var err error
			canvas, err = ggcanvas.New(provider, width, height)
			if err != nil {
				log.Fatalf("Failed to create canvas: %v", err)
			}
		}
		cw, ch := canvas.Size()
		if cw != width || ch != height {
			if err := canvas.Resize(width, height); err != nil {
				log.Printf("Resize error: %v", err)
			}
			cw, ch = width, height
		}
		err := canvas.Draw(func(cc *gg.Context) {
			w.renderFrame(cc, width, height)
		})
		if err != nil {
			log.Printf("Draw error: %v", err)
		}
		err = canvas.Render(dc.RenderTarget())
		if err != nil {
			log.Printf("Render error: %v", err)
		}
		w.app.RequestRedraw()
	})
	err := w.app.Run()
	if err != nil {
		log.Fatalf("Failed to run app: %v", err)
	}
}

func (w *ImageWindow) renderFrame(cc *gg.Context, width, height int) {
	cc.ClearWithColor(gg.Black)
	if w.img != nil {
		fitImg := scaling.Fit(*w.img, width, height)
		imgBuf := gg.ImageBufFromImage(fitImg)
		cc.DrawImage(imgBuf, float64(width-imgBuf.Width())/2.0, float64(height-imgBuf.Height())/2.0)
	}
}
