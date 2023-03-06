package trends

const (
	Percentage ExecutionPrinceType = "Percent"
	Price ExecutionPrinceType = "Price"
)

type TrendConfig struct {
	Support float32
	Resistance float32
	SpikeCount int
}


type LowerHighTrendConfig struct{

}

type LowerLowTrendConfig struct {

}



func DetectRange(){
	
}


func ScanAndBugLongDrop(){}

type F struct{
N int
}

func do(t struct{N int}){}

func init(){
	var g = struct{
		N : 9
	}
do(g);
}

/***
Sell automatic buy and sell currency such that anyone that reaches that
limit gets executed first
E.G BUSD/APT buy=1 sell 2,
BUSD/
*/


/***
If it goes out of the trend parameter, dont buy only sell when it is on the HIGH SIDE ONLY
**/

/**
Wait for a buying opportunity on multiple assets
**/
