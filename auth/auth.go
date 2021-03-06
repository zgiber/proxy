package auth

import (
	"fmt"
	"math/big"
	"time"

	"crypto/rand"

	jwt "github.com/dgrijalva/jwt-go"
)

const (
	randomCharSet = "1234567890qwertyuiopasdfghjklzxcvbnmQWERTYUIOPASDFGHJKLZXCVBNM"
)

var (
	JWTPublicKey  = "replace_me_with_secure_key"
	JWTPrivateKey = "replace_me_with_secure_key"
)

// TokenExchanger exchanges a token for another.
// It's primary purpose is to exchange opaque
// public tokens for short lived JWT for non-public usage
type TokenExchanger interface {
	Exchange(token string) (string, error)
}

// NewRandomToken returns a crypto safe
// random token of the given length
func NewRandomToken(length int) ([]byte, error) {
	l := big.NewInt(int64(len(randomCharSet)))

	result := make([]byte, length, length)
	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, l)
		if err != nil {
			return nil, err
		}
		result[i] = randomCharSet[n.Int64()]
	}
	return result, nil
}

// NewRandomToken returns a crypto safe
// random token of the given length as string
func NewRandomTokenString(length int) (string, error) {
	b, err := NewRandomToken(length)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// NewJWT returns a JWT signed with JWTPrivateKey
func NewJWT(claims map[string]interface{}, validFor time.Duration) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	for k, v := range claims {
		token.Claims[k] = v
	}

	token.Claims["exp"] = time.Now().UTC().Add(validFor).Unix()
	return token.SignedString([]byte(JWTPrivateKey))
}

// JWTClaims attempts to veify the signature on JWT and returns
// the claims in the JWT, and a bool indicating whether
// the signature is verified successfully.
func JWTClaims(tokenString string) (map[string]interface{}, bool) {
	token, err := jwt.Parse(tokenString, signingKeyJWT)
	if err != nil {
		return nil, false
	}

	return token.Claims, token.Valid
}

func signingKeyJWT(token *jwt.Token) (interface{}, error) {

	// only process HMAC signing for now
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
	}

	// TODO: add error checking to the type assertions above.
	return []byte(JWTPrivateKey), nil
}
