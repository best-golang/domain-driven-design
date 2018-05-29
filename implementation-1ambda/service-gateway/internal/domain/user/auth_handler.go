package user

import (
	"strings"

	e "github.com/a-trium/domain-driven-design/implementation-1ambda/service-gateway/internal/exception"
	"github.com/pkg/errors"
	"github.com/a-trium/domain-driven-design/implementation-1ambda/service-gateway/pkg/generated/swagger/swagserver/swagapi"
	"github.com/a-trium/domain-driven-design/implementation-1ambda/service-gateway/pkg/generated/swagger/swagserver/swagapi/auth"
	"github.com/go-openapi/runtime/middleware"
	dto "github.com/a-trium/domain-driven-design/implementation-1ambda/service-gateway/pkg/generated/swagger/swagmodel"
	"github.com/gorilla/sessions"
	"net/http"
	"fmt"
	"encoding/json"
	"github.com/go-openapi/runtime"
)

type AuthHandler interface {
	Configure(handlerRegistry *swagapi.GatewayAPI)
	Register(uid string, email string, password string) (*AuthClaim, e.Exception)
	Login(uid string, password string) (*AuthClaim, e.Exception)
}

type authHandlerImpl struct {
	userRepository Repository
	encryptor      Encryptor
	sessionStore   *sessions.CookieStore
}

const SessionCookieName = "SESSION"
const SessionFieldUID = "uid"
const SessionFieldAuthenticated = "authenticated"

func NewAuthHandler(repo Repository, encryptor Encryptor, sessionStore *sessions.CookieStore) AuthHandler {
	return &authHandlerImpl{
		userRepository: repo,
		encryptor:      encryptor,
		sessionStore:   sessionStore,
	}
}

func (c *authHandlerImpl) Configure(registry *swagapi.GatewayAPI) () {
	registry.AuthRegisterHandler = auth.RegisterHandlerFunc(
		func(params auth.RegisterParams) middleware.Responder {
			if params.Body == nil {
				err := errors.New("Empty Body")
				ex := e.NewBadRequestException(err)
				return auth.NewLoginDefault(ex.StatusCode()).WithPayload(ex.ToSwaggerError())
			}

			uid := params.Body.UID
			email := params.Body.Email
			password := params.Body.Password

			claim, ex := c.Register(uid, email, password)
			if ex != nil {
				return auth.NewRegisterDefault(ex.StatusCode()).WithPayload(ex.ToSwaggerError())
			}

			response := dto.AuthResponse{
				UID: claim.UID,
				// UserID: strconv.FormatUint(uint64(claim.UserID), 10),
			}
			return auth.NewRegisterOK().WithPayload(&response)
		})

	registry.AuthLoginHandler = auth.LoginHandlerFunc(
		func(params auth.LoginParams) middleware.Responder {
			if params.Body == nil {
				err := errors.New("Empty Body")
				ex := e.NewBadRequestException(err)
				return auth.NewLoginDefault(ex.StatusCode()).WithPayload(ex.ToSwaggerError())
			}

			uid := params.Body.UID
			password := params.Body.Password

			claim, ex := c.Login(uid, password)
			if ex != nil {
				return auth.NewLoginDefault(ex.StatusCode()).WithPayload(ex.ToSwaggerError())
			}

			response := &dto.AuthResponse{UID: claim.UID,}

			// set session value to mark user is logged in
			session, _ := c.sessionStore.Get(params.HTTPRequest, SessionCookieName)
			SetLoginSessionCookie(session, claim.UID)

			responder := auth.NewLoginOK().WithPayload(response)
			return NewLoginSessionResponder(responder, params.HTTPRequest, session)
		})

	registry.AuthLogoutHandler = auth.LogoutHandlerFunc(
		func(params auth.LogoutParams) middleware.Responder {
			session, _ := c.sessionStore.Get(params.HTTPRequest, SessionCookieName)
			CleanLoginSessionCookie(session)

			responder := auth.NewLogoutOK()
			return NewLogoutSessionResponder(responder, params.HTTPRequest, session)
		})

	registry.AuthWhoamiHandler = auth.WhoamiHandlerFunc(
		func(params auth.WhoamiParams) middleware.Responder {
			session, _ := c.sessionStore.Get(params.HTTPRequest, SessionCookieName)
			authenticated, uid := IsAuthenticated(session)

			// if not authenticated, then return empty uid
			if !authenticated {
				uid = ""
			}

			response := &dto.AuthResponse{UID: uid,}
			return auth.NewLoginOK().WithPayload(response)
		})

}

func (c *authHandlerImpl) Register(uid string, email string, password string) (*AuthClaim, e.Exception) {
	if strings.TrimSpace(uid) == "" || strings.TrimSpace(password) == "" {
		err := errors.New("Empty uid or password")
		return nil, e.NewBadRequestException(err)
	}

	encrypted, err := c.encryptor.Digest(password)
	if err != nil {
		wrap := errors.Wrap(err, "Failed to digest password")
		return nil, e.NewInternalServerException(wrap)
	}

	aid, ex := c.userRepository.CreateAuthIdentity(uid, email, encrypted)
	if ex != nil {
		return nil, ex
	}

	return aid.ToClaims(), nil
}

func (c *authHandlerImpl) Login(uid string, password string) (*AuthClaim, e.Exception) {
	if strings.TrimSpace(uid) == "" || strings.TrimSpace(password) == "" {
		err := errors.New("Empty uid or password")
		return nil, e.NewUnauthorizedException(err)
	}

	aid, ex := c.userRepository.FindAuthIdentityByUID(uid)
	if ex != nil {
		return nil, ex
	}

	if err := c.encryptor.Compare(aid.EncryptedPassword, password); err != nil {
		wrap := errors.Wrap(err, "Incorrect password")
		return nil, e.NewUnauthorizedException(wrap)
	}

	return aid.ToClaims(), nil
}

func SetLoginSessionCookie(session *sessions.Session, uid string) {
	session.Values[SessionFieldAuthenticated] = true
	session.Values[SessionFieldUID] = uid
}

func CleanLoginSessionCookie(session *sessions.Session) {
	session.Values[SessionFieldAuthenticated] = false
}

func IsAuthenticated(session *sessions.Session) (bool, string) {
	authenticated, ok := session.Values[SessionFieldAuthenticated].(bool)
	if !ok || !authenticated {
		return false, ""
	}

	uid, ok := session.Values[SessionFieldUID].(string)
	if !ok || uid == "" {
		return false, ""
	}

	return true, uid
}

type LoginSessionResponder struct {
	auth.LoginOK
	request *http.Request
	session *sessions.Session
}

func NewLoginSessionResponder(responder *auth.LoginOK, r *http.Request, session *sessions.Session) *LoginSessionResponder {
	return &LoginSessionResponder{
		*responder, r, session,
	}
}

func (responder *LoginSessionResponder) WriteResponse(w http.ResponseWriter, p runtime.Producer) {
	r := responder.request
	responder.session.Save(r, w)
	responder.LoginOK.WriteResponse(w, p)
}

type LogoutSessionResponder struct {
	auth.LogoutOK
	request *http.Request
	session *sessions.Session
}

func NewLogoutSessionResponder(responder *auth.LogoutOK, r *http.Request, session *sessions.Session) *LogoutSessionResponder {
	return &LogoutSessionResponder{
		*responder, r, session,
	}
}

func (responder *LogoutSessionResponder) WriteResponse(w http.ResponseWriter, p runtime.Producer) {
	r := responder.request
	responder.session.Save(r, w)
	responder.LogoutOK.WriteResponse(w, p)
}

func ConfigureSessionMiddleware(sessionStore sessions.Store, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// if CORS
		if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
			h.ServeHTTP(w, r)
			return
		}

		// if auth request
		if r.Method == http.MethodPost && strings.HasPrefix(r.URL.Path, "/api/auth/") {
			h.ServeHTTP(w, r)
			return
		}

		session, _ := sessionStore.Get(r, SessionCookieName)
		if authenticated, _ := IsAuthenticated(session); !authenticated {
			message := fmt.Sprintf("Not Authenticated: (%s) %s", r.Method, r.URL.Path)
			err := errors.New(message)
			ex := e.NewUnauthorizedException(err)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(ex.StatusCode())
			json.NewEncoder(w).Encode(ex)
			return
		}

		h.ServeHTTP(w, r)
	})
}
