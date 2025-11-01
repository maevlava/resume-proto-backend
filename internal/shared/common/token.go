package common

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/maevlava/resume-backend/internal/shared/db"
	"github.com/rs/zerolog/log"
)

type CustomClaims struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}
type Token struct {
	Value    string           `json:"token"`
	Duration *jwt.NumericDate `json:"duration"`
}

func MakeJWT(user db.User, tokenSecret string, expiresIn time.Duration) (*Token, error) {
	// signing method
	signingMethod := jwt.SigningMethodHS256
	// claims
	claims := CustomClaims{
		Email:    user.Email,
		Username: user.Name,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "resume-proto",
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(expiresIn)),
			Subject:   user.ID.String(),
		},
	}
	// token
	tokenClaims := jwt.NewWithClaims(signingMethod, claims)

	// sign token
	signedToken, err := tokenClaims.SignedString([]byte(tokenSecret))
	if err != nil {
		log.Error().Err(err).Msg("Error signing token")
		return nil, err
	}

	// return token
	token := &Token{
		Value:    signedToken,
		Duration: claims.RegisteredClaims.ExpiresAt,
	}

	return token, nil
}

func ValidateJWT(tokenString string, tokenSecret string) (*CustomClaims, error) {
	claims := &CustomClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(tokenSecret), nil
	})
	if err != nil {
		log.Error().Err(err).Msg("Error validating token")
		return nil, err
	}
	if !token.Valid {
		return nil, fmt.Errorf("token is not valid")
	}
	subject, err := claims.GetSubject()
	if err != nil {
		log.Error().Err(err).Msg("Error getting subject from token")
		return nil, err
	}
	log.Info().Msgf("Subject: %s", subject)

	return claims, nil
}
