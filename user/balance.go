package user

type Balance struct{
	Free float64
	Locked float64
	Asset string
}

func NewBalance() {}
