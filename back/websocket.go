package main

import (
	"log"
	"math"
	"math/rand"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

var (
	upgrader  = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	clients   = make(map[*websocket.Conn]bool)
	broadcast = make(chan Msg)
)

func wsHandler(c echo.Context) error {
	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	clients[conn] = true
	defer func() {
		delete(clients, conn)
		conn.Close()
	}()

	for {
		var msg Msg
		if err := conn.ReadJSON(&msg); err != nil {
			log.Println("Read err:", err)
			break
		}
		broadcast <- msg
	}
	return nil
}

func handleMessages() {
	for msg := range broadcast {
		for client := range clients {
			if err := client.WriteJSON(msg); err != nil {
				log.Println("Write err:", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}

func generateDummyData() {
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	shift := 0.0
	for range ticker.C {

		shift += rand.Float64()*0.2 - 0.05

		min, max := 1000.0, -1000.0
		val := 0.0

		// Graph data: 50 random points
		graphData := make([]float64, 100)
		for i := range graphData {

			val = rand.Float64()*3 +
				15*math.Sin(float64(i)/10+shift) +
				15*math.Cos(float64(i)/12+shift*1.5) + 30

			if val < min {
				min = val
			}
			if val > max {
				max = val
			}

			graphData[i] = val

		}
		broadcast <- Msg{Type: "graph", Data: graphData}

		//fmt.Printf(" -[ %.3f <--> %.3f ]- \n", min, max)

		/*
			// Spectrum data: 20 random amplitudes
			spectrumData := make([]float64, 20)
			for i := range spectrumData {
				spectrumData[i] = rand.Float64() * 50
			}
			broadcast <- Msg{Type: "spectrum", Data: spectrumData}

			// Map data: random GPS point
			mapData := map[string]float64{
				"lat": 37.7749 + (rand.Float64()-0.5)*0.1, // SF area ±0.05
				"lng": -122.4194 + (rand.Float64()-0.5)*0.1,
			}
			broadcast <- Msg{Type: "map", Data: mapData}
		*/
	}
}
