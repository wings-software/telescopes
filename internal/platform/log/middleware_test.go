// Copyright © 2019 Banzai Cloud
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package log

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/goph/logur"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestMiddlewareCorrelationId_GeneratesWhenMissing(t *testing.T) {
	router := gin.New()
	router.Use(MiddlewareCorrelationId())
	router.GET("/ping", func(c *gin.Context) {
		assert.NotEmpty(t, c.GetString(ContextKey))
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestMiddlewareCorrelationId_UsesIncomingHeader(t *testing.T) {
	const cid = "corr-123"
	router := gin.New()
	router.Use(MiddlewareCorrelationId())
	router.GET("/ping", func(c *gin.Context) {
		assert.Equal(t, cid, c.GetString(ContextKey))
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set(defaultHeader, cid)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestMiddlewareCorrelationId_CustomHeaderOption(t *testing.T) {
	const cid = "custom-cid"
	router := gin.New()
	router.Use(MiddlewareCorrelationId(Header("X-Request-ID")))
	router.GET("/ping", func(c *gin.Context) {
		assert.Equal(t, cid, c.GetString(ContextKey))
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set("X-Request-ID", cid)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGinMiddlewareLogger_LogLevelsByStatus(t *testing.T) {
	tests := []struct {
		name          string
		status        int
		wantLevel     logur.Level
		addGinError   bool
		addPrivateErr bool
		query         string
		pipelineUUID  string
		correlationID string
	}{
		{
			name:      "2xx info",
			status:    http.StatusOK,
			wantLevel: logur.Info,
		},
		{
			name:      "4xx warn",
			status:    http.StatusBadRequest,
			wantLevel: logur.Warn,
			query:     "accountId=abc",
		},
		{
			name:      "5xx error",
			status:    http.StatusInternalServerError,
			wantLevel: logur.Error,
		},
		{
			name:        "4xx with gin bind error logs warn",
			status:      http.StatusBadRequest,
			wantLevel:   logur.Warn,
			addGinError: true,
		},
		{
			name:          "4xx with private error uses private message",
			status:        http.StatusBadRequest,
			wantLevel:     logur.Warn,
			addPrivateErr: true,
		},
		{
			name:          "includes pipeline and correlation fields",
			status:        http.StatusOK,
			wantLevel:     logur.Info,
			pipelineUUID:  "pipeline-1",
			correlationID: "cid-1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := logur.NewTestLogger()
			router := gin.New()
			router.Use(GinMiddlewareLogger(logger))
			router.GET("/test", func(c *gin.Context) {
				if tt.correlationID != "" {
					c.Set(ContextKey, tt.correlationID)
				}
				if tt.addGinError {
					_ = c.Error(errors.New("bind failed"))
				}
				if tt.addPrivateErr {
					_ = c.Error(errors.New("private failure")).SetType(gin.ErrorTypePrivate)
				}
				c.Status(tt.status)
			})

			path := "/test"
			if tt.query != "" {
				path = path + "?" + tt.query
			}
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, path, nil)
			if tt.pipelineUUID != "" {
				req.Header.Set("Banzai-Cloud-Pipeline-UUID", tt.pipelineUUID)
			}
			router.ServeHTTP(w, req)

			require.GreaterOrEqual(t, logger.Count(), 1)
			event := logger.LastEvent()
			require.NotNil(t, event)
			assert.Equal(t, tt.wantLevel, event.Level)
			assert.Equal(t, tt.status, event.Fields["statusCode"])

			if tt.addPrivateErr {
				assert.Contains(t, event.Line, "private failure")
			} else if tt.addGinError {
				assert.Contains(t, event.Line, "bind failed")
			} else {
				assert.Equal(t, "ginLogger", event.Line)
			}

			if tt.query != "" {
				assert.Contains(t, event.Fields["path"], "?"+tt.query)
			}
			if tt.pipelineUUID != "" {
				assert.Equal(t, tt.pipelineUUID, event.Fields["pipeline-instance"])
			}
			if tt.correlationID != "" {
				assert.Equal(t, tt.correlationID, event.Fields["correlation-id"])
			}
		})
	}
}

func TestGinMiddlewareLogger_SkipPath(t *testing.T) {
	logger := logur.NewTestLogger()
	router := gin.New()
	router.Use(GinMiddlewareLogger(logger, "/health"))
	router.GET("/health", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	router.GET("/other", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/health", nil))
	assert.Equal(t, 0, logger.Count())

	router.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/other", nil))
	assert.Equal(t, 1, logger.Count())
	assert.Equal(t, logur.Info, logger.LastEvent().Level)
}
