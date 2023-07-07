package request

import (
	"encoding/json"
	// "fmt"
	_ "fmt"

	// "fmt"
	"net/http"
	// "net/http/httputil"
	// "net/http/httputil"
	// "net/http/httputil"
)

type requestApi interface {
	GetMethod() string
	GetPayload() interface{}
	GetEndpoint() string
}

func Request(api requestApi, decorateFunc func(r *http.Request)) (*http.Response, error) {
	var method, payload, endpoint = api.GetMethod(), api.GetPayload(), api.GetEndpoint()
	client := http.Client{}
	json.Marshal(payload)
	req, reqError := http.NewRequest(method, endpoint, nil)

	if reqError != nil{
		return nil, reqError
	}

	if decorateFunc != nil{
		decorateFunc(req)
	}
	
	resp, err := client.Do(req)
	// fmt.Println(resp.Body)
	// dump,_ := httputil.DumpResponse(resp, true)
	// fmt.Printf("%s", dump)
	return resp, err
}


