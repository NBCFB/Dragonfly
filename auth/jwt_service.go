package auth

import (
	"github.com/dgrijalva/jwt-go"
	"log"
	"time"
)

var secKey = []byte("dragonfly@NBCFB_2018")


type CustomClaims struct {
	UserId string `json:"userId"`
	jwt.StandardClaims
}

// Create a new JWT token use customized claims
func CreateNewToken(userId string) (string, error) {
	claims := CustomClaims {
		userId,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 8).Unix(),
			Issuer:    "Cheeta",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secKey)

	if err != nil {
		return "", err
	} else {
		return tokenString, nil
	}

}

// Validate JWT token
func ValidateToken(tokenString string) (bool, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return mykey, nil
	})

	if _, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return true, nil
	} else {
		log.Println(err.Error())
		return false, err
	}
}