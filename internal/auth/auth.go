package auth

import (
	// "os/user"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

// makerefreshtoken
func MakeRefreshToken() (string, error) {
	//generate 32bits randomly
	key := make([]byte, 32) //make a slice of bytes of base32
	_, err := rand.Read(key)
	if err != nil {
		fmt.Println("error creating randomKey")
		return "", err
	}
	return hex.EncodeToString(key), nil //convert it to hex string
}

func GetAPIKey(headers http.Header) (string, error) {
	authorizationString := headers.Get("Authorization")

	//solve if its empty
	if authorizationString == "" {
		return "", errors.New("missing Authorization header")
	}

	//strip the apikey prefiv
	_, after, found := strings.Cut(authorizationString, "ApiKey ")
	//if apikey prefix not found
	if !found {
		return "", errors.New(`authorization header must start with "ApiKey"`)
	}
	after = strings.TrimSpace(after)
	if after == "" {
		return "", errors.New("missing token")
	}

	return after, nil

}
func GetBearerToken(headers http.Header) (string, error) {
	authString := headers.Get("Authorization")
	//solve if its empty
	if authString == "" {
		return "", errors.New("missing Authorization header")

	}
	//strip the bearer prefix
	_, after, found := strings.Cut(authString, "Bearer ")

	//if no bearer prefix is found
	if !found {
		return "", errors.New("Authorization header must start with\"Bearer\"")
	}
	// if there is no token string
	after = strings.TrimSpace(after)
	if after == "" {
		return "", errors.New("missing bearer token")
	}
	return after, nil
}
func HashPassword(password string) (string, error) {
	hashedPassword, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return "", err
	}
	return hashedPassword, nil
}

func CheckPasswordHash(password, hash string) (bool, error) {
	match, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		return false, err
	}
	return match, nil
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	//create jwt payload
	claims := jwt.RegisteredClaims{
		Issuer:    "chirpy-access",
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
		Subject:   userID.String(),
	}
	//builds unsigned token -header + claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	//signs jwt token
	signedToken, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	//empty struct to parse into
	claims := &jwt.RegisteredClaims{}

	//1.splits the token into header.payload.signature
	//2.rederives the signature using the secret returned by your keyfunc
	//3. compares it against the signature embedded in the token
	//4. also automatically checks expiresat against the current time
	_, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			return []byte(tokenSecret), nil
		})

	//checks wrong secret and timestamp check at once
	if err != nil {
		return uuid.Nil, err
	}
	//pulls out the subject
	userIDStr := claims.Subject

	//converts string to uuid
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, err
	}

	return userID, nil
}
