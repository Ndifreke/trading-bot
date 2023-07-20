package graph

import "trading/kline"

var MockDowntrendPullupData = []kline.KlineData{
	//the high is in the body of the bar
	{Open: 10.0, Close: 10.0, High: 12.0, Low: 5.0},
	{Open: 10.0, Close: 10.0, High: 12.0, Low: 5.0},
	{Open: 9.0, Close: 5.0, High: 1.0, Low: 5.0},
}
