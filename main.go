package main

import (
	"log"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

func main() {
	mygo.App.WhenReady(func() {
		dir, err := mygo.App.Path(mygo.PathUserData)
		if err != nil {
			log.Fatal(err)
		}

		app, err := NewApp(dir)
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
			// The page's own --bg, until the view draws its first frame.
			BackgroundColor: "light-dark(#f5f5f7, #1c1c1e)",
			Content:         ui.View(app.view),
		})
		app.win = win
		app.installMenu()
	})

	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
