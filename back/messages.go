package main

type Msg struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}
