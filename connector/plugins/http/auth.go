package http

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
)

const (
	AuthMethodBasic = "basic"
	AuthMethodBearer = "bearer"
	AuthMethodHeaderToken = "header_token"
	AuthMethodQueryToken = "query_token"
)


type AuthConfig struct {
	AuthMethod          string `json:"omitempty"` // none , bearer , basic ,header-token, query-token
	AuthToken           string `json:"omitempty"` // Bearer token
	AuthUsername        string `json:"omitempty"` // Username for Basic auth
	AuthPassword        string `json:"omitempty"` // Password for Basic auth
	AuthCustomParamName string `json:"omitempty"` // Name of custom header that stores token. Or name of query parameter that holds token.
}

func (conn *Connector) isRequestAllowed(w http.ResponseWriter,r *http.Request,streamAuth AuthConfig,flowId string ) bool {
	authMethod := conn.config.GlobalAuth.AuthMethod
	isGlobalAuth := true
	if streamAuth.AuthMethod != "" {
		// Node configs overrides global configurations
		authMethod = streamAuth.AuthMethod
		isGlobalAuth = false
	}
	log.Debugf("Auth method %s , flowId = %s",authMethod,flowId)
	if authMethod == "" {
		return true
	}
	switch authMethod {
	case AuthMethodBasic:
		user,pass,ok := r.BasicAuth()
		if ok {
			if isGlobalAuth {
				if conn.config.GlobalAuth.AuthUsername == user && conn.config.GlobalAuth.AuthPassword == pass {
					return true
				}
			}else {
				if streamAuth.AuthUsername == user && streamAuth.AuthPassword == pass {
					return true
				}
			}
		}
		realm := "tpflow"
		if !isGlobalAuth {
			realm = flowId
		}
		w.Header().Set("WWW-Authenticate", fmt.Sprintf(`Basic realm="%s"`,realm))
		w.WriteHeader(401)
		w.Write([]byte("Unauthorised.\n"))
		return false

	case AuthMethodBearer:
		authHeader := r.Header.Get("Authorization")
		if isGlobalAuth {
			if authHeader == fmt.Sprintf("Bearer %s",conn.config.GlobalAuth.AuthToken) {
				return true
			}
		}else{
			if authHeader == fmt.Sprintf("Bearer %s",streamAuth.AuthToken) {
				return true
			}
		}

	case AuthMethodHeaderToken:
		if isGlobalAuth {
			if r.Header.Get(conn.config.GlobalAuth.AuthCustomParamName) == fmt.Sprintf("Bearer %s",conn.config.GlobalAuth.AuthToken) {
				return true
			}
		}else{
			if r.Header.Get(streamAuth.AuthCustomParamName) == fmt.Sprintf("Bearer %s",streamAuth.AuthToken) {
				return true
			}
		}
	case AuthMethodQueryToken:
		if isGlobalAuth {
			if r.URL.Query().Get(conn.config.GlobalAuth.AuthCustomParamName) == conn.config.GlobalAuth.AuthToken {
				return true
			}
		}else {
			if r.URL.Query().Get(streamAuth.AuthCustomParamName) == streamAuth.AuthToken {
				return true
			}
		}

	}
	w.WriteHeader(401)
	return false
}