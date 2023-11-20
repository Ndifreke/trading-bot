package user

type User struct {
	firstname  string
	lastname   string
	middlename string
	account    AccountInterface
}

func CreateUser() *User {
	var firstname, lastname, middlename string
	return &User{
		firstname:  firstname,
		lastname:   lastname,
		middlename: middlename,
		account:    GetAccount(),
	}
}

func (user *User) GetAccount() AccountInterface {
	return user.account
}
