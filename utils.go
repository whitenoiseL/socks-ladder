package socks_ladder

import (
	"time"
	"math/rand"
)

const PasswordLength = 256
type Password [PasswordLength]byte

func init(){
	rand.Seed(time.Now().Unix())
}

func RandPassword() *Password {
	intArr := rand.Perm(PasswordLength)
	var password Password
	for i,v := range intArr {
		password[i] = byte(v)
		if i == v {
			return RandPassword()
		}
	}
	return &password
}