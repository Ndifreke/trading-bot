package kline

var MockKLineData = []KlineData{
    {
      Open: 100,
      High: 101,
      Low: 99,
      Close:  100,
      
    },
	{
		Open: 100,
		High: 101,
		Low: 99,
		Close:  100,
		
	  },
	  
}


var MockKLinePullUpData = []KlineData{
	{Open: 15.0, Close: 10.0, High: 40.0, Low: 5.0},
}

type kLineMock struct {
	data []KlineData
}

func (k kLineMock) RefreshData() []KlineData {
	return k.data
}

func (kline kLineMock) KLineData() []KlineData {
	return kline.data
}

func (kline kLineMock) GetPointLimit() int {
	return len(kline.data)
}

func GetKLineMock(data []KlineData) KLineInterface {
	return kLineMock{
		data: data,
	}
}
