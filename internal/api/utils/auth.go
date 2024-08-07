// auth.go

package utils

import (
	handlers "c2c/internal/api/handlers"
	models "c2c/internal/models"
	"c2c/internal/tools"
	"fmt"
	"net/http"
	"os"
	"strings"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

func extractToken(c *gin.Context) string {
	jwtToken := c.GetHeader("Authorization")
	strArr := strings.Split(jwtToken, " ")
	if len(strArr) == 2 {
		return strArr[1]
	}
	return ""
}

func verifyToken(c *gin.Context, e *handlers.Env) (bool, error) {
	tokenStr := extractToken(c)

	key_filename := fmt.Sprintf("%s/key", e.Config.Keypath)
	if strings.Contains(c.Request.URL.Path, "assconfig") {
		key_filename = fmt.Sprintf("%s/assconfigkey", e.Config.Keypath)
	}

	key, err := os.ReadFile(key_filename)
	if err != nil {
		if strings.Contains(c.Request.URL.Path, "assconfig") {
			key = []byte(JWT_CONFIG_KEY)
		} else {
			key = []byte(JWT_KEY)
		}
	}

	// Solution 1 but doesn't check expiration
	// parts := strings.Split(tokenStr, ".")
	// err = jwt.SigningMethodHS256.Verify(strings.Join(parts[0:2], "."), parts[2], key)
	// if err != nil {
	// 	tools.Logger.Errorf("%s", err)
	// 	return false, nil
	// }

	// Solution 2 and can get claims
	_, err = jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		// check token signing method etc
		return key, nil
	})
	if err != nil {
		tools.Logger.Errorf("%s", err)
		return false, err
	}

	// if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
	// 	return claims, true
	// } else {
	// 	tools.Logger.Printf("Invalid JWT Token")
	// 	return nil, false
	// }

	return true, nil
}

func IsAuthorized(endpoint func(c *gin.Context, e *handlers.Env), e *handlers.Env, protected_method []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if e.Config.Authentication.Enable && Contains(protected_method, c.Request.Method) {
			if ok, err := verifyToken(c, e); err != nil || !ok {
				var resp models.GasResponse
				err := models.C2CError{ErrorCode: "NOT_AUHTORIZED"}
				resp.SetError(&err)
				c.JSON(http.StatusOK, resp)
				// c.String(http.StatusForbidden, "Forbidden Acces")
			} else {
				endpoint(c, e)
			}
		} else {
			endpoint(c, e)
		}
	}
}
