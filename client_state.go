package authboss

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	"net"
	"net/http"
	"reflect"
	"strings"
)

const (
	// SessionKey is the primarily used key by authboss.
	SessionKey = "uid"
	// SessionHalfAuthKey is used for sessions that have been authenticated by
	// the remember module. This serves as a way to force full authentication
	// by denying half-authed users access to sensitive areas.
	SessionHalfAuthKey = "halfauth"
	// SessionLastAction is the session key to retrieve the
	// last action of a user.
	SessionLastAction = "last_action"
	// Session2FA is set when a user has been authenticated with a second factor
	Session2FA = "twofactor"
	// Session2FAAuthToken is a random token set in the session to be verified
	// by e-mail.
	Session2FAAuthToken = "twofactor_auth_token"
	// Session2FAAuthed is in the session (and set to "true") when the user
	// has successfully verified the token sent via e-mail in the two factor
	// e-mail authentication process.
	Session2FAAuthed = "twofactor_authed"
	// SessionOAuth2State is the xsrf protection key for oauth.
	SessionOAuth2State = "oauth2_state"
	// SessionOAuth2Params is the additional settings for oauth
	// like redirection/remember.
	SessionOAuth2Params = "oauth2_params"

	// CookieRemember is used for cookies and form input names.
	CookieRemember = "rm"

	// FlashSuccessKey is used for storing success flash messages on the session
	FlashSuccessKey = "flash_success"
	// FlashErrorKey is used for storing success flash messages on the session
	FlashErrorKey = "flash_error"
)

// ClientStateEventKind is an enum.
type ClientStateEventKind int

// ClientStateEvent kinds
const (
	// ClientStateEventPut means you should put the key-value pair into the
	// client state.
	ClientStateEventPut ClientStateEventKind = iota
	// ClientStateEventDel means you should delete the key-value pair from the
	// client state.
	ClientStateEventDel
	// ClientStateEventDelAll means you should delete EVERY key-value pair from
	// the client state - though a whitelist of keys that should not be deleted
	// may be passed through as a comma separated list of keys in
	// the ClientStateEvent.Key field.
	ClientStateEventDelAll
)

// ClientStateEvent are the different events that can be recorded during
// a request.
type ClientStateEvent struct {
	Kind  ClientStateEventKind
	Key   string
	Value string
}

// ClientStateReadWriter is used to create a cookie storer from an http request.
// Keep in mind security considerations for your implementation, Secure,
// HTTP-Only, etc. flags.
//
// There's two major uses for this. To create session storage, and remember me
// cookies.
type ClientStateReadWriter interface {
	// ReadState should return a map like structure allowing it to look up
	// any values in the current session, or any cookie in the request
	ReadState(*http.Request) (ClientState, error)
	// WriteState can sometimes be called with a nil ClientState in the event
	// that no ClientState was read in from LoadClientState
	WriteState(http.ResponseWriter, ClientState, []ClientStateEvent) error
}

// UnderlyingResponseWriter retrieves the response
// writer underneath the current one. This allows us
// to wrap and later discover the particular one that we want.
// Keep in mind this should not be used to call the normal methods
// of a response writer, just additional ones particular to that type
// because it's possible to introduce subtle bugs otherwise.
type UnderlyingResponseWriter interface {
	UnderlyingResponseWriter() http.ResponseWriter
}

// ClientState represents the client's current state and can answer queries
// about it.
type ClientState interface {
	Get(key string) (string, bool)
}

// ClientStateResponseWriter is used to write out the client state at the last
// moment before the response code is written.
type ClientStateResponseWriter struct {
	http.ResponseWriter

	cookieStateRW  ClientStateReadWriter
	sessionStateRW ClientStateReadWriter

	cookieState  ClientState
	sessionState ClientState

	hasWritten         bool
	cookieStateEvents  []ClientStateEvent
	sessionStateEvents []ClientStateEvent
}

// LoadClientStateMiddleware wraps all requests with the
// ClientStateResponseWriter as well as loading the current client
// state into the context for use.
func (a *Authboss) LoadClientStateMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writer := a.NewResponse(w)
		request, err := a.LoadClientState(writer, r)
		if err != nil {
			logger := a.RequestLogger(r)
			logger.Errorf("failed to load client state %+v", err)

			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		h.ServeHTTP(writer, request)
	})
}

// NewResponse wraps the ResponseWriter with a ClientStateResponseWriter
func (a *Authboss) NewResponse(w http.ResponseWriter) *ClientStateResponseWriter {
	return &ClientStateResponseWriter{
		ResponseWriter: w,
		cookieStateRW:  a.Config.Storage.CookieState,
		sessionStateRW: a.Config.Storage.SessionState,
	}
}

// LoadClientState loads the state from sessions and cookies
// into the ResponseWriter for later use.
func (a *Authboss) LoadClientState(w http.ResponseWriter, r *http.Request) (*http.Request, error) {
	if a.Storage.SessionState != nil {
		state, err := a.Storage.SessionState.ReadState(r)
		if err != nil {
			return nil, err
		} else if state != nil {
			c := MustClientStateResponseWriter(w)
			c.sessionState = state
			r = r.WithContext(context.WithValue(r.Context(), CTXKeySessionState, state))
		}
	}
	if a.Storage.CookieState != nil {
		state, err := a.Storage.CookieState.ReadState(r)
		if err != nil {
			return nil, err
		} else if state != nil {
			c := MustClientStateResponseWriter(w)
			c.cookieState = state
			r = r.WithContext(context.WithValue(r.Context(), CTXKeyCookieState, state))
		}
	}

	return r, nil
}

// MustClientStateResponseWriter tries to find a csrw inside the response
// writer by using the UnderlyingResponseWriter interface.
func MustClientStateResponseWriter(w http.ResponseWriter) *ClientStateResponseWriter {
	for {
		fmt.Printf("##TYPE-OF-[%+v]##", reflect.TypeOf(w))
		if e, ok := w.(*echo.Response); ok {
			// fmt.Printf("%+v", w)
			w = e.Writer
			continue
		}

		if c, ok := w.(*ClientStateResponseWriter); ok {
			// fmt.Printf("%+v", w)
			return c
		}

		if u, ok := w.(UnderlyingResponseWriter); ok {
			w = u.UnderlyingResponseWriter()
			continue
		}

		panic(fmt.Sprintf("ResponseWriter must be a ClientStateResponseWriter or UnderlyingResponseWriter in (see: authboss.LoadClientStateMiddleware): %T", w))
	}
}

// WriteHeader writes the header, but in order to handle errors from the
// underlying ClientStateReadWriter, it has to panic.
func (c *ClientStateResponseWriter) WriteHeader(code int) {
	if !c.hasWritten {
		if err := c.putClientState(); err != nil {
			panic(err)
		}
	}
	c.ResponseWriter.WriteHeader(code)
}

// Header retrieves the underlying headers
func (c ClientStateResponseWriter) Header() http.Header {
	return c.ResponseWriter.Header()
}

// Hijack implements the http.Hijacker interface by calling the
// underlying implementation if available.
func (c ClientStateResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := c.ResponseWriter.(http.Hijacker)
	if ok {
		return h.Hijack()
	}
	return nil, nil, errors.New("authboss: underlying ResponseWriter does not support hijacking")
}

// Write ensures that the client state is written before any writes
// to the body occur (before header flush to http client)
func (c *ClientStateResponseWriter) Write(b []byte) (int, error) {
	if !c.hasWritten {
		if err := c.putClientState(); err != nil {
			return 0, err
		}
	}
	return c.ResponseWriter.Write(b)
}

// UnderlyingResponseWriter for this instance
func (c *ClientStateResponseWriter) UnderlyingResponseWriter() http.ResponseWriter {
	return c.ResponseWriter
}

// EchoResponseWriter for this instance
func (c *ClientStateResponseWriter) EchoResponseWriter(e echo.Context) http.ResponseWriter {
	return e.Response().Writer
}

func (c *ClientStateResponseWriter) putClientState() error {
	if c.hasWritten {
		panic("should not call putClientState twice")
	}
	c.hasWritten = true

	if len(c.cookieStateEvents) == 0 && len(c.sessionStateEvents) == 0 {
		return nil
	}

	if c.sessionStateRW != nil && len(c.sessionStateEvents) > 0 {
		err := c.sessionStateRW.WriteState(c, c.sessionState, c.sessionStateEvents)
		if err != nil {
			return err
		}
	}
	if c.cookieStateRW != nil && len(c.cookieStateEvents) > 0 {
		err := c.cookieStateRW.WriteState(c, c.cookieState, c.cookieStateEvents)
		if err != nil {
			return err
		}
	}

	return nil
}

// IsFullyAuthed returns false if the user has a SessionHalfAuth
// in his session.
func IsFullyAuthed(r *http.Request) bool {
	_, hasHalfAuth := GetSession(r, SessionHalfAuthKey)
	return !hasHalfAuth
}

// IsTwoFactored returns false if the user doesn't have a Session2FA
// in his session.
func IsTwoFactored(r *http.Request) bool {
	_, has2fa := GetSession(r, Session2FA)
	return has2fa
}

// DelAllSession deletes all variables in the session except for those on
// the whitelist.
//
// The whitelist is typically provided directly from the authboss config.
//
// This is the best way to ensure the session is cleaned up after use for
// a given user. An example is when a user is expired or logged out this method
// is called.
func DelAllSession(w http.ResponseWriter, whitelist []string) {
	delAllState(w, CTXKeySessionState, whitelist)
}

// DelKnownSession is deprecated. See DelAllSession for an alternative.
// DelKnownSession deletes all known session variables,
// effectively logging a user out.
func DelKnownSession(w http.ResponseWriter) {
	DelSession(w, SessionKey)
	DelSession(w, SessionHalfAuthKey)
	DelSession(w, SessionLastAction)
}

// DelKnownCookie deletes all known cookie variables, which can be used
// to delete remember me pieces.
func DelKnownCookie(w http.ResponseWriter) {
	DelCookie(w, CookieRemember)
}

// PutSession puts a value into the session
func PutSession(w http.ResponseWriter, key, val string) {
	putState(w, CTXKeySessionState, key, val)
}

// DelSession deletes a key-value from the session.
func DelSession(w http.ResponseWriter, key string) {
	delState(w, CTXKeySessionState, key)
}

// GetSession fetches a value from the session
func GetSession(r *http.Request, key string) (string, bool) {
	return getState(r, CTXKeySessionState, key)
}

// PutCookie puts a value into the session
func PutCookie(w http.ResponseWriter, key, val string) {
	putState(w, CTXKeyCookieState, key, val)
}

// DelCookie deletes a key-value from the session.
func DelCookie(w http.ResponseWriter, key string) {
	delState(w, CTXKeyCookieState, key)
}

// GetCookie fetches a value from the session
func GetCookie(r *http.Request, key string) (string, bool) {
	return getState(r, CTXKeyCookieState, key)
}

func putState(w http.ResponseWriter, CTXKey contextKey, key, val string) {
	setState(w, CTXKey, ClientStateEventPut, key, val)
}

func delState(w http.ResponseWriter, CTXKey contextKey, key string) {
	setState(w, CTXKey, ClientStateEventDel, key, "")
}

func delAllState(w http.ResponseWriter, CTXKey contextKey, whitelist []string) {
	setState(w, CTXKey, ClientStateEventDelAll, strings.Join(whitelist, ","), "")
}

func setState(w http.ResponseWriter, ctxKey contextKey, op ClientStateEventKind, key, val string) {
	fmt.Printf("%+v", w)
	csrw := MustClientStateResponseWriter(w)
	ev := ClientStateEvent{
		Kind: op,
		Key:  key,
	}

	if op == ClientStateEventPut {
		ev.Value = val
	}

	switch ctxKey {
	case CTXKeySessionState:
		csrw.sessionStateEvents = append(csrw.sessionStateEvents, ev)
	case CTXKeyCookieState:
		csrw.cookieStateEvents = append(csrw.cookieStateEvents, ev)
	}
}

func getState(r *http.Request, ctxKey contextKey, key string) (string, bool) {
	val := r.Context().Value(ctxKey)
	if val == nil {
		return "", false
	}

	state := val.(ClientState)
	return state.Get(key)
}

// FlashSuccess returns FlashSuccessKey from the session and removes it.
func FlashSuccess(w http.ResponseWriter, r *http.Request) string {
	str, ok := GetSession(r, FlashSuccessKey)
	if !ok {
		return ""
	}

	DelSession(w, FlashSuccessKey)
	return str
}

// FlashError returns FlashError from the session and removes it.
func FlashError(w http.ResponseWriter, r *http.Request) string {
	str, ok := GetSession(r, FlashErrorKey)
	if !ok {
		return ""
	}

	DelSession(w, FlashErrorKey)
	return str
}
