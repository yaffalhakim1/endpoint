package main

import (
	"log"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"github.com/yaffalhakim1/endpoint/internal/app"
)

func main() {
	mygo.App.WhenReady(func() {
		dir, err := mygo.App.Path(mygo.PathUserData)
		if err != nil {
			log.Fatal(err)
		}

		window, err := app.New(dir)
		if err != nil {
			log.Fatal(err)
		}

		win := mygo.NewWindow(mygo.WindowOptions{
			Title:          "Endpoint",
			Width:          1180,
			Height:         780,
			MinWidth:       880,
			MinHeight:      560,
			StateKey:       "main",
			TitleBarStyle:  mygo.TitleBarHidden,
			TitleBarHeight: 44,
			// The buttons are drawn by the view, so the system draws none.
			HideWindowButtons: true,
			Vibrancy:          mygo.VibrancySidebar,
			// The window's own colour, until the view draws its first frame.
			BackgroundColor: "light-dark(#f5f5f7, #1c1c1e)",
			Content:         ui.View(window.View),
		})
		window.SetWindow(win)
		window.InstallMenu()
	})

	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
