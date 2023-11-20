package user

type Balance struct {
	Locked float64
	Free   float64
	Asset  string
}

func NewBalance() {}
