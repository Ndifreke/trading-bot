package api

import "net/http"

type Api struct {
	Method string
	Path   string
}

type ListAPI struct {
	PriceAverage Api
	ServerTime   Api
	PriceLatest  Api
	KLines       Api
	OpenOrders   Api
	OpenOrder    Api
}

type ApiArg struct {
	Api
}

type RequestArg struct {
	Params  map[string]string
	Payload interface{}
}

type RequestResponse[B any] struct {
	Body     B
	Response http.Response
	Ok       bool
	Error    error
}
