package auth

import "github.com/alexedwards/argon2id"

func HashPassword(password string) (string, error){
	hashedPassword, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return "", err
	}
	return hashedPassword, nil
}

func CheckPasswordHash(password, hash string) (bool, error){
	passwordHashed, err := HashPassword(password)
	if err != nil {
		return false,err
	}
	if hash != passwordHashed {
		return false, nil
	} 

	return true, nil
}
