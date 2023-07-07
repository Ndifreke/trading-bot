package kline

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
