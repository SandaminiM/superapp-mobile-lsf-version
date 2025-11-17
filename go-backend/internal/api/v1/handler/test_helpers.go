package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"

	"go-backend/internal/auth"
)

// TestUserInfo represents a test user for mocking JWT authentication
type TestUserInfo struct {
	Email  string
	Groups []string
}

// createRequestWithAuth creates an HTTP request with mocked user info in context
// This bypasses JWT middleware for testing by directly setting user info in context
func createRequestWithAuth(method, url string, body []byte, userInfo *TestUserInfo) *http.Request {
	var req *http.Request
	if body != nil {
		req = httptest.NewRequest(method, url, bytes.NewReader(body))
	} else {
		req = httptest.NewRequest(method, url, nil)
	}

	if userInfo != nil {
		// Mock user info in context (bypassing JWT middleware)
		req = auth.SetUserInfo(req, &auth.CustomJwtPayload{
			Email:  userInfo.Email,
			Groups: userInfo.Groups,
		})
	}

	return req
}
