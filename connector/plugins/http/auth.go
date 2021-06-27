package http

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strings"
)

const (
	AuthMethodBasic = "basic"
	AuthMethodNone = "none"
	AuthMethodBearer = "bearer"
	AuthMethodHeaderToken = "header_token"
	AuthMethodQueryToken = "query_token"

	AuthCodeAuthorized   = 1
	AuthCodeBasicFailed  = 2
	AuthCodeFailed       = 3
)


type AuthConfig struct {
	AuthMethod          string // none , bearer , basic ,header-token, query-token
	AuthToken           string // Bearer token
	AuthUsername        string // Username for Basic auth
	AuthPassword        string // Password for Basic auth
	AuthCustomParamName string // Name of custom header that stores token. Or name of query parameter that holds token.
}

func (conn *Connector) SendHttpAuthFailureResponse(authCode int , w http.ResponseWriter,flowId string) {
	switch authCode {
	case AuthCodeBasicFailed :
		t := strings.Split(flowId,"_")
		w.Header().Set("WWW-Authenticate", fmt.Sprintf(`Basic realm="%s"`,t[0]))
		w.WriteHeader(401)
		w.Write([]byte("Unauthorised.\n"))
	case AuthCodeFailed:
		w.WriteHeader(401)
	}
}

func (conn *Connector) isRequestAllowed(r *http.Request,streamAuth AuthConfig,flowId string ) int {
	authMethod := conn.config.GlobalAuth.AuthMethod
	isGlobalAuth := true
	if streamAuth.AuthMethod != "" {
		// Node configs overrides global configurations
		authMethod = streamAuth.AuthMethod
		isGlobalAuth = false
	}
	log.Debugf("Auth method %s , flowId = %s",authMethod,flowId)
	if authMethod == "" {
		return AuthCodeAuthorized
	}
	switch authMethod {
	case AuthMethodBasic:
		user,pass,ok := r.BasicAuth()
		if ok {
			if isGlobalAuth {
				if conn.config.GlobalAuth.AuthUsername == user && conn.config.GlobalAuth.AuthPassword == pass {
					return AuthCodeAuthorized
				}
			}else {
				if streamAuth.AuthUsername == user && streamAuth.AuthPassword == pass {
					return AuthCodeAuthorized
				}
			}
		}
		return AuthCodeBasicFailed

	case AuthMethodBearer:
		authHeader := r.Header.Get("Authorization")
		if isGlobalAuth {
			if authHeader == fmt.Sprintf("Bearer %s",conn.config.GlobalAuth.AuthToken) {
				return AuthCodeAuthorized
			}
		}else{
			if authHeader == fmt.Sprintf("Bearer %s",streamAuth.AuthToken) {
				return AuthCodeAuthorized
			}
		}

	case AuthMethodHeaderToken:
		if isGlobalAuth {
			if r.Header.Get(conn.config.GlobalAuth.AuthCustomParamName) == fmt.Sprintf("Bearer %s",conn.config.GlobalAuth.AuthToken) {
				return AuthCodeAuthorized
			}
		}else{
			if r.Header.Get(streamAuth.AuthCustomParamName) == fmt.Sprintf("Bearer %s",streamAuth.AuthToken) {
				return AuthCodeAuthorized
			}
		}
	case AuthMethodQueryToken:
		if isGlobalAuth {
			if r.URL.Query().Get(conn.config.GlobalAuth.AuthCustomParamName) == conn.config.GlobalAuth.AuthToken {
				return AuthCodeAuthorized
			}
		}else {
			if r.URL.Query().Get(streamAuth.AuthCustomParamName) == streamAuth.AuthToken {
				return AuthCodeAuthorized
			}
		}
	case AuthMethodNone:
		return AuthCodeAuthorized
	case "":
		return AuthCodeAuthorized

	}
	return AuthCodeFailed
}