package user

type User struct {
	firstname  string
	lastname   string
	middlename string
	account    Account
}

func CreateUser() *User {
	var firstname, lastname, middlename string
	return &User{
		firstname:  firstname,
		lastname:   lastname,
		middlename: middlename,
		account:    *GetAccount(),
	}
}

func (user *User) GetAccount() *Account {
	return &user.account
}
