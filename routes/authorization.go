package routes

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/drd-engineering/TwinCape/environments"
	"github.com/gin-gonic/gin"
)

// DRDApplicationIdentification is authorization for identify the request is from drd app
func DRDApplicationIdentification(auths ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("Drd-Identification")
		if len(apiKey) < 1 {
			c.AbortWithStatus(401)
			return
		}
		if apiKey != environments.Get("DRD_IDENTIFICATION") {
			c.AbortWithStatus(401)
			return
		}
		c.Next()
	}
}

// AuthorizationBearer is authorization middleware for identify the request client logged in or not
func AuthorizationBearer(auths ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		bearerToken := c.GetHeader("Authorization")
		strArr := strings.Split(bearerToken, " ")
		if len(strArr) < 2 {
			c.Abort()
			c.JSON(http.StatusUnauthorized,
				gin.H{"message": "Please provide authorization token"})
			return
		}
		claims := jwt.StandardClaims{}
		tokenString := strArr[1]
		_, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) {
			//Make sure that the token method conform to "SigningMethodHMAC"
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(environments.Get("ACCESS_SECRET_KEY")), nil
		})
		if err != nil {
			c.Abort()
			c.JSON(http.StatusUnauthorized,
				gin.H{"message": err.Error()})
			return
		}
		if claims.Issuer != "SSO_TWINCAPE" {
			c.Abort()
			c.JSON(http.StatusUnauthorized,
				gin.H{"message": "Invalid refresh token"})
			return
		}
		userID := claims.Audience
		c.Set("userID", userID)
		c.Next()
	}
}
