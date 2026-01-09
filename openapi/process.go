package openapi

import (
	"strings"

	"github.com/yaoapp/gou/process"
	"github.com/yaoapp/kun/exception"
	"github.com/yaoapp/kun/log"
	"github.com/yaoapp/yao/openapi/oauth"
	"github.com/yaoapp/yao/openapi/oauth/types"
	"github.com/yaoapp/yao/openapi/user"
)

func init() {
	process.Register("OpenApi.AccessToken.Make", ProcessOauthTokenMake)
	process.Register("OpenApi.AccessToken.Verify", ProcessOauthTokenVerify)
	process.Register("OpenApi.AccessToken.Enabled", ProcessOauthTokenEnabled)
	process.Register("OpenApi.Features", ProcessOpenApiFeatures)
}
func CheckAuthIsInit() {
	if oauth.OAuth == nil {
		exception.New("OAuth service not initialized", 500).Throw()
	}
}
func ProcessOauthTokenEnabled(process *process.Process) interface{} {
	if Server == nil || oauth.OAuth == nil {
		return false
	}
	return true
}
func ProcessOpenApiFeatures(process *process.Process) interface{} { 
	features := map[string]interface{}{
		"openapi_enable": true,
	}
	if Server == nil || oauth.OAuth == nil {
		features["openapi_enable"] = false
	}
	return features
}
func ProcessOauthTokenMake(process *process.Process) interface{} {
	CheckAuthIsInit()
	process.ValidateArgNums(1)
	input := process.ArgsString(0)
	expiresIn := 3600
	if process.NumOfArgs() > 1 {
		expiresIn = process.ArgsInt(1)
	}
	scopes := []string{}
	if process.NumOfArgs() > 2 {
		scopes = process.ArgsStrings(2)
	}
	claims := make(map[string]interface{})
	if process.NumOfArgs() > 3 {
		claims = process.ArgsMap(3)
	}
	claims["__input"] = input
	token, err := makeAccessToken(input, expiresIn, scopes, claims)
	if err != nil {
		exception.New("%s error: %s", 400, err).Throw()
	}
	return token
}

func ProcessOauthTokenVerify(process *process.Process) interface{} {
	CheckAuthIsInit()
	process.ValidateArgNums(1)
	token := process.ArgsString(0)
	claims, err := verifyAccessToken(token)
	if err != nil {
		exception.New("%s error: %s", 400, err).Throw()
	}
	return claims
}

func verifyAccessToken(token string) (*types.TokenClaims, error) {
	CheckAuthIsInit()
	claims, err := oauth.OAuth.VerifyToken(token)
	if err != nil {
		return nil, err
	}
	userId, err := oauth.OAuth.UserID(claims.ClientID, claims.Subject)
	if err == nil {
		claims.Extra["__user_id"] = userId
	}else {
		log.Error("Failed to get user ID: %v", err)
	}

	return claims, nil
}


// makeAccessToken 生成访问令牌，scopes为空时，校验时会使用用户关联的scope
func makeAccessToken(input string, expiresIn int, scopes []string, claims map[string]interface{}) (string, error) {
	CheckAuthIsInit()
	yaoClientConfig := user.GetYaoClientConfig()
	//scopes为空时会使用用户关联的scope
	//scopes为用户acl权限对象列表，会作为访问权限控制
	subject, err := oauth.OAuth.Subject(yaoClientConfig.ClientID, input)
	accessToken, err := oauth.OAuth.MakeAccessToken(yaoClientConfig.ClientID, strings.Join(scopes, " "), subject, expiresIn, claims)
	if err != nil {
		return "", err
	}
	return accessToken, nil
}
