package main

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.CORS())

	e.GET("/ws", wsHandler)

	go handleMessages()
	go generateDummyData()

	e.Logger.Fatal(e.Start(":8080"))
}
