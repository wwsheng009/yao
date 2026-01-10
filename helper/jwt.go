package helper

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/yaoapp/gou/process"
	"github.com/yaoapp/gou/session"
	"github.com/yaoapp/kun/any"
	"github.com/yaoapp/kun/exception"
	"github.com/yaoapp/kun/log"
	"github.com/yaoapp/yao/config"
)

const (
	// MaxTokenLength is the maximum allowed length for a JWT token
	MaxTokenLength = 4096
	// MaxTokenParts is the maximum allowed number of parts in a JWT token (header.payload.signature)
	MaxTokenParts = 3
)

// JwtClaims 用户Token
type JwtClaims struct {
	ID   string                 `json:"id"`
	SID  string                 `json:"sid"`
	Data map[string]interface{} `json:"data"`
	jwt.RegisteredClaims
}

// JwtToken JWT令牌
type JwtToken struct {
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expires_at"`
}

// JwtValidate JWT 校验
func JwtValidate(tokenString string, secret ...[]byte) *JwtClaims {
	// Check token length
	if len(tokenString) > MaxTokenLength {
		exception.New("Token too long", 401).Throw()
		return nil
	}

	// Check number of parts
	parts := strings.Split(tokenString, ".")
	if len(parts) > MaxTokenParts {
		exception.New("Invalid token format", 401).Throw()
		return nil
	}

	jwtSecret := []byte(config.Conf.JWTSecret)
	if len(secret) > 0 {
		jwtSecret = secret[0]
	}

	token, err := jwt.ParseWithClaims(tokenString, &JwtClaims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		log.Error("JWT ParseWithClaims Error: %s", err)
		exception.New("Invalid token", 401).Ctx(err.Error()).Throw()
		return nil
	}

	if claims, ok := token.Claims.(*JwtClaims); ok && token.Valid {
		return claims
	}

	exception.New("Invalid token", 401).Ctx(token.Claims).Throw()
	return nil
}

// JwtMake  生成 JWT
// option: {"subject":"<主题>", "audience": "<接收人>", "issuer":"<签发人>", "timeout": "<有效期,单位秒>", "sid":"<会话ID>"}
func JwtMake(id string, data map[string]interface{}, option map[string]interface{}, secret ...[]byte) JwtToken {

	jwtSecret := []byte(config.Conf.JWTSecret)
	if len(secret) > 0 {
		jwtSecret = secret[0]
	}

	now := time.Now()
	sid := ""
	timeout := time.Hour
	uid := id
	subject := "User Token"
	audience := []string{"Yao Process utils.jwt.Make"}
	issuer := fmt.Sprintf("xiaong:%s", id)

	if v, has := option["subject"]; has {
		subject = fmt.Sprintf("%v", v)
	}

	if v, has := option["audience"]; has {
		audience = []string{fmt.Sprintf("%v", v)}
	}

	if v, has := option["issuer"]; has {
		issuer = fmt.Sprintf("%v", v)
	}

	if v, has := option["sid"]; has {
		sid = fmt.Sprintf("%v", v)
	}

	if v, has := option["timeout"]; has {
		timeout = time.Duration(any.Of(v).CInt()) * time.Second
	}

	expiresAt := now.Add(timeout)
	if v, has := option["expires_at"]; has {
		expiresAt = time.Unix(int64(any.Of(v).CInt()), 0)
	}

	if sid == "" {
		sid = session.ID()
	}

	claims := &JwtClaims{
		ID:   uid,
		SID:  sid, // 会话ID
		Data: data,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uid,                           // 唯一ID
			Subject:   subject,                       // 主题
			Audience:  audience,                      // 接收人
			ExpiresAt: jwt.NewNumericDate(expiresAt), // 过期时间
			NotBefore: jwt.NewNumericDate(now),       // 生效时间
			IssuedAt:  jwt.NewNumericDate(now),       // 签发时间
			Issuer:    issuer,                        // 签发人
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		exception.New("JWT Make Error: %s", 500, err.Error()).Throw()
	}

	return JwtToken{
		Token:     tokenString,
		ExpiresAt: expiresAt.Unix(),
	}
}

// ProcessJwtMake xiang.helper.JwtMake 生成JWT
func ProcessJwtMake(process *process.Process) interface{} {
	process.ValidateArgNums(2)
	uid, err := InterfaceToString(process.Args[0])
	if err != nil || uid == "" {
		exception.New("Invalid id format:%s", 401, err.Error()).Throw()
	}

	data := process.ArgsMap(1)
	option := map[string]interface{}{}
	if process.NumOfArgs() > 2 {
		option = process.ArgsMap(2)
	}
	secret := ""
	if process.NumOfArgs() > 3 {
		secret = process.ArgsString(3)
	}
	if secret != "" {
		return JwtMake(uid, data, option, []byte(secret))
	}
	return JwtMake(uid, data, option)
}

// ProcessJwtValidate xiang.helper.JwtValidate 校验JWT
func ProcessJwtValidate(process *process.Process) interface{} {
	process.ValidateArgNums(1)
	tokenString := process.ArgsString(0)
	return JwtValidate(tokenString)
}

func InterfaceToString(v interface{}) (string, error) {
	if v == nil {
		return "", errors.New("value is nil")
	}

	switch val := v.(type) {

	case string:
		return val, nil

	case []byte:
		return string(val), nil

	// signed integers
	case int:
		return strconv.Itoa(val), nil
	case int8:
		return strconv.FormatInt(int64(val), 10), nil
	case int16:
		return strconv.FormatInt(int64(val), 10), nil
	case int32:
		return strconv.FormatInt(int64(val), 10), nil
	case int64:
		return strconv.FormatInt(val, 10), nil

	// unsigned integers
	case uint:
		return strconv.FormatUint(uint64(val), 10), nil
	case uint8:
		return strconv.FormatUint(uint64(val), 10), nil
	case uint16:
		return strconv.FormatUint(uint64(val), 10), nil
	case uint32:
		return strconv.FormatUint(uint64(val), 10), nil
	case uint64:
		return strconv.FormatUint(val, 10), nil

	// floats
	case float32:
		return strconv.FormatFloat(float64(val), 'f', -1, 32), nil
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64), nil

	// JSON number
	case json.Number:
		return val.String(), nil

	// implements Stringer
	case fmt.Stringer:
		return val.String(), nil

	default:
		return "", fmt.Errorf("unsupported type %T", v)
	}
}
