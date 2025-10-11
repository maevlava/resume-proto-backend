package auth

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"time"
)

func MakeJWT(userId uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	// signing method
	signingMethod := jwt.SigningMethodHS256
	// claims
	claims := jwt.RegisteredClaims{
		Issuer:    "resume-proto",
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(expiresIn)),
		Subject:   userId.String(),
	}
	// token
	token := jwt.NewWithClaims(signingMethod, claims)
	// sign token
	signedToken, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		log.Error().Err(err).Msg("Error signing token")
		return "", err
	}
	// return token
	return signedToken, nil
}

func ValidateJWT(tokenString string, tokenSecret string) (*jwt.RegisteredClaims, error) {
	claims := &jwt.RegisteredClaims{}

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
