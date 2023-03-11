package webutil

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/TravisS25/webutil/webutilcfg"
	"github.com/gorilla/csrf"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/pkg/errors"
)

const (
	bodyRequiredTxt = "Request must have body"
	invalidJSONTxt  = "Invalid json"
	serverErrTxt    = "Server error, please try again later"

	IDRegexStr   = "[0-9]+"
	UUIDRegexStr = "[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$"

	IDParam   = "{id:" + IDRegexStr + "}"
	UUIDParam = "{id}"
)

//////////////////////////////////////////////////////////////////
//---------------------- CUSTOM ERRORS ------------------------
//////////////////////////////////////////////////////////////////

var (
	// ErrBodyRequired is used for when a post/put request does not contain a body in request
	ErrBodyRequired = errors.New("webutil: " + bodyRequiredTxt)

	// ErrInvalidJSON is used when there is an error unmarshalling a struct
	ErrInvalidJSON = errors.New("webutil: " + invalidJSONTxt)
)

//////////////////////////////////////////////////////////////////
//------------------------- FUNCTIONS --------------------------
//////////////////////////////////////////////////////////////////

// SetToken is wrapper function for setting csrf token header
func SetToken(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(webutilcfg.TokenHeader, csrf.Token(r))
}

// SendPayload is a wrapper for converting the payload map parameter into json and
// sending to the client
func SendPayload(w http.ResponseWriter, payload interface{}, errResp HTTPResponseConfig) error {
	w.Header().Set("Content-Type", webutilcfg.ContentTypeJSON)
	SetHTTPResponseDefaults(&errResp, http.StatusInternalServerError, []byte(invalidJSONTxt))
	jsonString, err := json.Marshal(payload)

	if err != nil {
		w.WriteHeader(*errResp.HTTPStatus)
		w.Write(errResp.HTTPResponse)
		return err
	}

	w.Write(jsonString)
	return nil
}

// GetUser returns a user if set in userctx, else returns nil
func GetUser(r *http.Request) []byte {
	if r.Context().Value(UserCtxKey) == nil {
		return nil
	}

	return r.Context().Value(UserCtxKey).([]byte)
}

// GetMiddlewareUser returns a user's email if set in userctx, else returns nil
func GetMiddlewareUser(r *http.Request) *MiddlewareUser {
	if r.Context().Value(MiddlewareUserCtxKey) == nil {
		return nil
	}

	return r.Context().Value(MiddlewareUserCtxKey).(*MiddlewareUser)
}

// LogoutUser deletes user session based on session object passed along with userSession parameter
// If userSession is empty string, then string "user" will be used to delete from session object
func LogoutUser(w http.ResponseWriter, r *http.Request, sessionStore sessions.Store, userSession string) error {
	var err error

	if r.Context().Value(UserCtxKey) != nil {
		var session *sessions.Session

		if userSession == "" {
			session, err = sessionStore.Get(r, "user")
		} else {
			session, err = sessionStore.Get(r, userSession)
		}

		if err != nil {
			return err
		}

		session.Options = &sessions.Options{
			MaxAge: -1,
		}
		err = session.Save(r, w)
	}

	return err
}

// GetRouting is wrapper for returning routing map where
// keys in map should be url templates of all routes
// current user is allowed to access
func GetRouting(r *http.Request) map[string]bool {
	if r.Context().Value(RoutingCtxKey) != nil {
		return r.Context().Value(RoutingCtxKey).(map[string]bool)
	}

	return nil
}

// GetUserGroups is wrapper for returning group map from context of request
// where the keys are the groups the current user is in
// If there is no groupctx, returns nil
func GetUserGroups(r *http.Request) map[string]bool {
	if r.Context().Value(GroupCtxKey) != nil {
		return r.Context().Value(GroupCtxKey).(map[string]bool)
	}

	return nil
}

func GetGroupArray(r *http.Request) []string {
	if r.Context().Value(GroupCtxKey) != nil {
		gMap := r.Context().Value(GroupCtxKey).(map[string]bool)
		gl := make([]string, 0, len(gMap))

		for k := range gMap {
			gl = append(gl, k)
		}

		return gl
	}

	return nil
}

// HasGroup is a wrapper for finding if given groups names is in
// group context of given request
// If a group name is found, return true; else returns false
// The search is based on OR logic so if any one of the given strings
// is found, function will return true
func HasGroup(r *http.Request, searchGroups ...string) bool {
	if groupMap := GetUserGroups(r); groupMap != nil {
		for _, searchGroup := range searchGroups {
			if _, ok := groupMap[searchGroup]; ok {
				return true
			}
		}
	}

	return false
}

// DecodeCookie takes in a cookie name which value should be encoded and then takes the
// authKey and encryptKey variables passed to decode the value of the cookie
func DecodeCookie(r *http.Request, cookieName string, authKey, encryptKey []byte) (string, error) {
	var cookieVal string
	sc := securecookie.New(authKey, encryptKey)
	ec, err := r.Cookie(cookieName)

	if err != nil {
		return "", err
	}

	err = sc.Decode(cookieName, ec.Value, &cookieVal)

	if err != nil {
		return "", err
	}

	return cookieVal, nil
}

func DecodeCookieNoRequest(cookieName, val string, authKey, encryptKey []byte) (string, error) {
	var cookieVal string

	sc := securecookie.New(authKey, encryptKey)
	err := sc.Decode(cookieName, val, &cookieVal)

	if err != nil {
		return "", err
	}

	return cookieVal, nil
}

// GetJSONBuffer takes interface and json encodes it into a buffer and returns buffer
func GetJSONBuffer(item interface{}) (bytes.Buffer, error) {
	var buffer bytes.Buffer

	encoder := json.NewEncoder(&buffer)

	if err := encoder.Encode(&item); err != nil {
		return bytes.Buffer{}, err
	}

	return buffer, nil
}

// SetSecureCookie is used to set a cookie from a session to header and returns encoded cookie
// The code used is copied pasted from the RedisStore#Save function from the redis store library
func SetSecureCookie(w http.ResponseWriter, session *sessions.Session, keyPairs ...[]byte) (string, error) {
	encoded, err := securecookie.EncodeMulti(session.Name(), session.ID, securecookie.CodecsFromPairs(keyPairs...)...)
	if err != nil {
		return "", err
	}
	http.SetCookie(w, sessions.NewCookie(session.Name(), encoded, session.Options))
	return encoded, nil
}
