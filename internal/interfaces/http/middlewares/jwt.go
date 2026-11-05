package middlewares

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/FrostBitzX/smart-task-ai/internal/interfaces/http/responses"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// cache JWKS ไว้ไม่ต้อง fetch ทุก request
var (
	jwksCache     map[string]interface{}
	jwksCacheMu   sync.RWMutex
	jwksCacheTime time.Time
	jwksCacheTTL  = 1 * time.Hour
)

func getJWKS() (map[string]interface{}, error) {
	jwksCacheMu.RLock()
	if jwksCache != nil && time.Since(jwksCacheTime) < jwksCacheTTL {
		defer jwksCacheMu.RUnlock()
		return jwksCache, nil
	}
	jwksCacheMu.RUnlock()

	baseURL := os.Getenv("SUPABASE_URL")
	if baseURL == "" {
		baseURL = os.Getenv("SUPABASE_PUBLIC_URL")
	}
	url := baseURL + "/auth/v1/.well-known/jwks.json"
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	jwksCacheMu.Lock()
	jwksCache = result
	jwksCacheTime = time.Now()
	jwksCacheMu.Unlock()

	return result, nil
}

func getPublicKey(kid string) (*ecdsa.PublicKey, error) {
	jwks, err := getJWKS()
	if err != nil {
		return nil, err
	}

	keys, ok := jwks["keys"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid JWKS format")
	}

	for _, k := range keys {
		keyMap, ok := k.(map[string]interface{})
		if !ok {
			continue
		}
		if keyMap["kid"] != kid {
			continue
		}

		crv, _ := keyMap["crv"].(string)
		xStr, _ := keyMap["x"].(string)
		yStr, _ := keyMap["y"].(string)
		if xStr == "" || yStr == "" {
			return nil, fmt.Errorf("invalid EC key material for kid: %s", kid)
		}

		var curve elliptic.Curve
		switch crv {
		case "P-256":
			curve = elliptic.P256()
		case "P-384":
			curve = elliptic.P384()
		case "P-521":
			curve = elliptic.P521()
		default:
			return nil, fmt.Errorf("unsupported EC curve: %s", crv)
		}

		xBytes, err := base64.RawURLEncoding.DecodeString(xStr)
		if err != nil {
			return nil, fmt.Errorf("failed to decode x coordinate: %w", err)
		}
		yBytes, err := base64.RawURLEncoding.DecodeString(yStr)
		if err != nil {
			return nil, fmt.Errorf("failed to decode y coordinate: %w", err)
		}

		return &ecdsa.PublicKey{
			Curve: curve,
			X:     new(big.Int).SetBytes(xBytes),
			Y:     new(big.Int).SetBytes(yBytes),
		}, nil
	}
	return nil, fmt.Errorf("key not found: %s", kid)
}

func JWTMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		auth := c.Get("Authorization")
		if auth == "" {
			return unauthorized(c, "missing token", "MISSING_TOKEN")
		}

		const bearerPrefix = "Bearer "
		if !strings.HasPrefix(auth, bearerPrefix) {
			return unauthorized(c, "invalid token format", "INVALID_TOKEN_FORMAT")
		}

		tokenStr := strings.TrimPrefix(auth, bearerPrefix)

		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			switch token.Method.(type) {
			case *jwt.SigningMethodECDSA:
				// ES256 — ใช้ public key จาก JWKS
				kid, _ := token.Header["kid"].(string)
				return getPublicKey(kid)

			case *jwt.SigningMethodHMAC:
				// HS256 legacy — ใช้ JWT_SECRET เดิม
				return []byte(os.Getenv("JWT_SECRET")), nil

			default:
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
		})

		if err != nil || !token.Valid {
			return unauthorized(c, "invalid token", "INVALID_TOKEN")
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return unauthorized(c, "invalid token claims", "INVALID_TOKEN_CLAIMS")
		}

		accountID, _ := claims["AccountId"].(string)
		nodeID, _ := claims["NodeId"].(string)
		email, _ := claims["Email"].(string)
		username, _ := claims["Username"].(string)

		if accountID == "" {
			return unauthorized(c, "missing account id in token", "MISSING_ACCOUNT_ID")
		}
		if nodeID == "" {
			return unauthorized(c, "missing node id in token", "MISSING_NODE_ID")
		}

		c.Locals("jwt_claims", map[string]interface{}{
			"AccountId": accountID,
			"NodeId":    nodeID,
			"Email":     email,
			"Username":  username,
		})

		return c.Next()
	}
}

func unauthorized(c *fiber.Ctx, msg, code string) error {
	return c.Status(fiber.StatusUnauthorized).JSON(responses.ErrorResponse{
		Success: false,
		Message: msg,
		Data:    nil,
		Error: responses.ErrorDetail{
			Code:    fiber.StatusUnauthorized,
			Message: code,
		},
	})
}
