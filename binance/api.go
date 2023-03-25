package binance

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"trading/api"
	"trading/request"
	"trading/utils"
)

type ErrorJson struct {
	Code    string `json:"code"`
	Message string `json:"msg"`
}

const (
	prodBaseUrl string = "https://api.binance.com"
	devBaseUrl  string = "https://testnet.binance.vision"
)

type apiArg = api.ApiArg

func getSignature(message string) string {
	var secret = os.Getenv("API_SECRET")
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(message))
	return fmt.Sprintf("%x", mac.Sum(nil))
}

func getBaseApi() string {
	env := os.Getenv("ENV")
	if env == "production" {
		return prodBaseUrl
	}
	return devBaseUrl
}

func setHeaders(req *http.Request) {
	req.Header.Add("accept", "application/json")
	utils.LogInfo(req.URL.Hostname())
	var key = os.Getenv("API_KEY")

	// timestamp := fmt.Sprintf("timestamp=%d", time.Now().Unix() * 1000)
	// req.Header.Set("X-MBX-SIGNATURE", getSignature(timestamp))
	req.Header.Set("X-MBX-APIKEY", key)
	// req.Header.Set("X-MBX-TIMESTAMP", timestamp)
}

type binanceApi[R any] struct {
	url     string
	arg     api.ApiArg
	payload interface{}
}

// API documentation https://binance.github.io/binance-api-swagger/
var Endpoints = api.ListAPI{
	PriceAverage: api.Api{
		Method: http.MethodGet,
		Path:   "/api/v3/avgPrice",
	},
	ServerTime: api.Api{
		Method: http.MethodGet,
		Path:   "/api/v3/time",
	},
	PriceLatest: api.Api{
		Method: http.MethodGet,
		Path:   "/api/v3/ticker/price",
	},
	KLines: api.Api{
		Method: http.MethodGet,
		Path:   "/api/v3/klines",
	},
	OpenOrders: api.Api{
		Method: http.MethodGet,
		Path:   "/api/v3/openOrders",
	},
}

func New[R any](api api.ApiArg) binanceApi[R] {
	var url = strings.Join([]string{getBaseApi(), api.Path}, "")
	return binanceApi[R]{url, api, nil}
}

func (binance *binanceApi[R]) GetEndpoint() string {
	return binance.url
}

func (binance *binanceApi[R]) GetMethod() string {
	return binance.arg.Method
}

func (binance *binanceApi[R]) GetPayload() interface{} {
	return binance.payload
}

func (binance *binanceApi[R]) Request() api.RequestResponse[R] {
	return binance.RequestWithQuery(nil)
}

func (binance *binanceApi[R]) RequestWithQuery(params map[string]string) api.RequestResponse[R] {
	var data R
	binance.url = parseParams(binance.url, params)
	res, reqError := request.Request(binance, setHeaders)

	if reqError != nil {
		return api.RequestResponse[R]{Body: data, Ok: false, Error: reqError, Response: *res}
	}

	body, _ := ioutil.ReadAll(res.Body)
	unMarshalError := json.Unmarshal(body, &data)

	if unMarshalError != nil {
		return api.RequestResponse[R]{Body: data, Ok: false, Error: unMarshalError, Response: *res}
	}
	Ok := res.StatusCode <= 299
	if !Ok {
		utils.LogWarn(string(body))
	}
	defer res.Body.Close()
	return api.RequestResponse[R]{Body: data, Ok: Ok, Error: nil, Response: *res}

}

func parseParams(endpoint string, param map[string]string) string {
	u, _ := url.Parse(endpoint)
	query := u.Query()
	for k, v := range param {
		query.Add(k, v)
	}
	u.RawQuery = query.Encode()
	return u.String()
}
