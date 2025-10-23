package oauth

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yaoapp/kun/log"
	"github.com/yaoapp/yao/openapi/oauth/acl"
	"github.com/yaoapp/yao/openapi/oauth/authorized"
	"github.com/yaoapp/yao/openapi/oauth/types"
	"github.com/yaoapp/yao/openapi/response"
)

// Guard is the OAuth guard middleware
func (s *Service) Guard(c *gin.Context) {
	// Get the token from the request
	token := s.getAccessToken(c)

	// Validate the token
	if token == "" {
		response.RespondWithError(c, http.StatusUnauthorized, types.ErrTokenMissing)
		c.Abort()
		return
	}

	// Validate the token
	claims, err := s.VerifyToken(token)
	if err != nil {
		response.RespondWithError(c, http.StatusUnauthorized, types.ErrInvalidToken)
		c.Abort()
		return
	}

	// Auto refresh the token
	if claims.ExpiresAt.Before(time.Now()) {
		s.tryAutoRefreshToken(c, claims)
	}

	// Set Authorized Info in context
	sessionID := s.getSessionID(c)
	authorized.SetInfo(c, claims, sessionID, s.UserID)

	// Check if ACL is enabled
	if acl.Global == nil || !acl.Global.Enabled() {
		return
	}

	// Check permissions and enforce rate limits when ACL is configured
	ok, err := acl.Global.Enforce(c)
	if err != nil {
		log.Error("[OAuth] ACL enforcement failed: %v", err)
		s.handleACLError(c, err)
		return
	}

	// If permissions are not granted but no error returned, it's an unexpected state
	// This should not happen with the current implementation
	if !ok {
		response.RespondWithError(c, http.StatusForbidden, types.ErrForbidden)
		c.Abort()
		return
	}
}

// GetAuthorizedInfo gets authorized info from context
// Deprecated: Use authorized.GetInfo(c) instead
func GetAuthorizedInfo(c *gin.Context) *types.AuthorizedInfo {
	return authorized.GetInfo(c)
}

func (s *Service) tryAutoRefreshToken(c *gin.Context, _ *types.TokenClaims) {
	refreshToken := s.getRefreshToken(c)
	if refreshToken == "" {
		response.RespondWithError(c, http.StatusUnauthorized, types.ErrRefreshTokenMissing)
		c.Abort()
		return
	}

	// Verify the refresh token
	_, err := s.VerifyToken(refreshToken)
	if err != nil {
		response.RespondWithError(c, http.StatusUnauthorized, types.ErrInvalidRefreshToken)
		c.Abort()
		return
	}

	// @Todo: Auto refresh the token
}

func (s *Service) getAccessToken(c *gin.Context) string {
	token := c.GetHeader("Authorization")
	if token == "" {
		cookie, err := c.Cookie("__Host-access_token")
		if err != nil {
			return ""
		}
		token = cookie
	}
	return strings.TrimPrefix(token, "Bearer ")
}

// GetAccessToken gets the access token from the request (public method)
func (s *Service) GetAccessToken(c *gin.Context) string {
	return s.getAccessToken(c)
}

func (s *Service) getRefreshToken(c *gin.Context) string {
	token := c.GetHeader("Authorization")
	if token == "" {
		cookie, err := c.Cookie("__Host-refresh_token")
		if err != nil {
			return ""
		}
		token = cookie
	}
	return strings.TrimPrefix(token, "Bearer ")
}

// GetRefreshToken gets the refresh token from the request (public method)
func (s *Service) GetRefreshToken(c *gin.Context) string {
	return s.getRefreshToken(c)
}

// Get Session ID from cookies, headers, or query string
func (s *Service) getSessionID(c *gin.Context) string {

	// 0. If has __sid in context, return it
	sid, ok := c.Get("__sid")
	if ok {
		return sid.(string)
	}

	// 1. Try to get Session ID from cookies first
	if sid, err := c.Cookie("__Host-session_id"); err == nil && sid != "" {
		return sid
	}

	// 2. Try to get Session ID from X-Session-ID header
	if sessionHeader := c.GetHeader("X-Session-ID"); sessionHeader != "" {
		return sessionHeader
	}

	// 3. Try to get Session ID from query string
	if sessionQuery := c.Query("session_id"); sessionQuery != "" {
		return sessionQuery
	}

	// 4. Try alternative query parameter names
	if sessionQuery := c.Query("sid"); sessionQuery != "" {
		return sessionQuery
	}

	return ""
}

// handleACLError handles ACL errors and returns appropriate HTTP responses
func (s *Service) handleACLError(c *gin.Context, err error) {
	// Check if it's an ACL error with detailed information
	if aclErr, ok := err.(*acl.Error); ok {
		var statusCode int
		var errResponse *types.ErrorResponse

		switch aclErr.Type {
		case acl.ErrorTypeRateLimitExceeded:
			statusCode = http.StatusTooManyRequests
			errResponse = types.ErrRateLimitExceeded
			// Set Retry-After header if available
			if aclErr.RetryAfter > 0 {
				c.Header("Retry-After", fmt.Sprintf("%d", aclErr.RetryAfter))
			}

		case acl.ErrorTypeQuotaExceeded:
			statusCode = http.StatusTooManyRequests
			errResponse = &types.ErrorResponse{
				Code:             "quota_exceeded",
				ErrorDescription: aclErr.Message,
			}

		case acl.ErrorTypeInsufficientScope:
			statusCode = http.StatusForbidden
			// Include detailed scope information for insufficient scope errors
			requiredScopes, _ := aclErr.Details["required_scopes"].([]string)
			missingScopes, _ := aclErr.Details["missing_scopes"].([]string)

			errResponse = &types.ErrorResponse{
				Code:             "insufficient_scope",
				ErrorDescription: "The access token does not have the required scope",
				Reason:           aclErr.Message,
				RequiredScopes:   requiredScopes,
				MissingScopes:    missingScopes,
			}

		case acl.ErrorTypePermissionDenied:
			statusCode = http.StatusForbidden
			// Include detailed information for permission denied errors
			requiredScopes, _ := aclErr.Details["required_scopes"].([]string)
			missingScopes, _ := aclErr.Details["missing_scopes"].([]string)

			// Use standard ErrorResponse format with extended ACL fields
			errResponse = &types.ErrorResponse{
				Code:             "forbidden",
				ErrorDescription: "You do not have permission to access this resource",
				Reason:           aclErr.Message,
				RequiredScopes:   requiredScopes,
				MissingScopes:    missingScopes,
			}

		case acl.ErrorTypeResourceNotAllowed:
			statusCode = http.StatusForbidden
			errResponse = types.ErrAccessDenied

		case acl.ErrorTypeMethodNotAllowed:
			statusCode = http.StatusMethodNotAllowed
			errResponse = types.ErrMethodNotAllowed

		case acl.ErrorTypeIPBlocked, acl.ErrorTypeGeoRestricted, acl.ErrorTypeTimeRestricted:
			statusCode = http.StatusForbidden
			errResponse = types.ErrAccessDenied

		case acl.ErrorTypeInvalidRequest:
			statusCode = http.StatusBadRequest
			errResponse = &types.ErrorResponse{
				Code:             "invalid_request",
				ErrorDescription: aclErr.Message,
			}

		case acl.ErrorTypeInternal:
			statusCode = http.StatusInternalServerError
			errResponse = types.ErrACLInternalError

		default:
			statusCode = http.StatusInternalServerError
			errResponse = types.ErrACLInternalError
		}

		response.RespondWithError(c, statusCode, errResponse)
		c.Abort()
		return
	}

	// If it's not an ACL error, treat it as an internal error
	response.RespondWithError(c, http.StatusInternalServerError, types.ErrACLInternalError)
	c.Abort()
}
