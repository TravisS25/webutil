package webutil

//go:generate mockgen -source=api_util.go -destination=../webutilmock/api_util_mock.go -package=webutilmock
//go:generate mockgen -source=api_util.go -destination=api_util_mock_test.go -package=webutil

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/gorilla/csrf"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/pkg/errors"
)

//////////////////////////////////////////////////////////////////
//-------------------------- CONSTS ---------------------------
//////////////////////////////////////////////////////////////////

const (
	// TokenHeader is key string for "X-CSRF-TOKEN" header
	TokenHeader = "X-CSRF-TOKEN"

	// CookieHeader is key string for "Cookie" header
	CookieHeader = "Cookie"

	// SetCookieHeader is key string for "Set-Cookie" header
	SetCookieHeader = "Set-Cookie"
)

const (
	// ContentTypeBinary is key string for content type header "application/octet-stream"
	ContentTypeBinary = "application/octet-stream"

	// ContentTypeForm is key string for content type header "application/x-www-form-urlencoded"
	ContentTypeForm = "application/x-www-form-urlencoded"

	// ContentTypeJSON is key string for content type header "application/json"
	ContentTypeJSON = "application/json"

	// ContentTypePDF is key string for content type header "application/pdf"
	ContentTypePDF = "application/pdf"

	// ContentTypeHTML is key string for content type header "text/html; charset=utf-8"
	ContentTypeHTML = "text/html; charset=utf-8"

	// ContentTypeText is key string for content type header "text/plain; charset=utf-8"
	ContentTypeText = "text/plain; charset=utf-8"

	// ContenTypeJPG is key string for content type header "image/jpeg"
	ContenTypeJPG = "image/jpeg"

	// ContentTypePNG is key string for content type header "image/png"
	ContentTypePNG = "image/png"
)

//////////////////////////////////////////////////////////////////
//---------------------- CUSTOM ERRORS ------------------------
//////////////////////////////////////////////////////////////////

var (
	// ErrBodyRequired is used for when a post/put request does not contain a body in request
	ErrBodyRequired = errors.New("webutil: request must have body")

	// ErrInvalidJSON is used when there is an error unmarshalling a struct
	ErrInvalidJSON = errors.New("webutil: invalid json")

	// ErrServer is used when there is a server error
	ErrServer = errors.New("webutil: server error, please try again later")
)

//////////////////////////////////////////////////////////////////
//------------------------ INTERFACES --------------------------
//////////////////////////////////////////////////////////////////

// CookieError is wrapper interface for securecookie.Error
// to be able to generate mocks
type cookieError interface {
	securecookie.Error
}

//////////////////////////////////////////////////////////////////
//------------------------- FUNCTIONS --------------------------
//////////////////////////////////////////////////////////////////

// SetToken is wrapper function for setting csrf token header
func SetToken(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(TokenHeader, csrf.Token(r))
}

// // ServerError takes given err along with customMessage and writes back to client
// // then logs the error given the logFile
// func ServerError(w http.ResponseWriter, err error, customMessage string) {
// 	w.WriteHeader(http.StatusInternalServerError)

// 	if customMessage != "" {
// 		w.Write([]byte(customMessage))
// 	} else {
// 		w.Write([]byte(ErrServer.Error()))
// 	}
// }

// // HasServerError is wrapper for ServerError that returns if error passed
// // is nil or not.  Point of function is simply to reduce code lines by
// // a caller function
// func HasServerError(w http.ResponseWriter, err error, customMessage string) bool {
// 	if err != nil {
// 		ServerError(w, err, customMessage)
// 		return true
// 	}

// 	return false
// }

// SendPayload is a wrapper for converting the payload map parameter into json and
// sending to the client
func SendPayload(w http.ResponseWriter, payload interface{}, errResp HTTPResponseConfig) error {
	SetHTTPResponseDefaults(&errResp, http.StatusInternalServerError, []byte(ErrInvalidJSON.Error()))
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
func GetMiddlewareUser(r *http.Request) *middlewareUser {
	if r.Context().Value(MiddlewareUserCtxKey) == nil {
		return nil
	}

	return r.Context().Value(MiddlewareUserCtxKey).(*middlewareUser)
}

// // HasBodyError checks if the "Body" field of the request parameter is nil or not
// // If nil, we write to client with error message, 406 status and return true
// // Else return false
// func HasBodyError(w http.ResponseWriter, r *http.Request, bodyRespConfig HTTPResponseConfig) bool {
// 	SetHTTPResponseDefaults(&bodyRespConfig, http.StatusNotAcceptable, []byte(ErrBodyRequired.Error()))

// 	if r.Body == nil || r.Body == http.NoBody {
// 		w.WriteHeader(*bodyRespConfig.HTTPStatus)
// 		w.Write(bodyRespConfig.HTTPResponse)
// 		return true
// 	}

// 	return false
// }

// LogoutUser deletes user session based on session object passed along with userSession parameter
// If userSession is empty string, then string "user" will be used to delete from session object
func LogoutUser(w http.ResponseWriter, r *http.Request, sessionStore sessions.Store, userSession string) error {
	if r.Context().Value(UserCtxKey) != nil {
		var session *sessions.Session
		var err error

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
		session.Save(r, w)
	}

	return nil
}

// GetUserGroups is wrapper for to returning group map from context of request
// where the keys are the groups the current user is in
// If there is no groupctx, returns nil
func GetUserGroups(r *http.Request) map[string]bool {
	if r.Context().Value(GroupCtxKey) != nil {
		return r.Context().Value(GroupCtxKey).(map[string]bool)
	}

	return nil
}

// HasGroup is a wrapper for finding if given groups names is in
// group context of given request
// If a group name is found, return true; else returns false
// The search is based on OR logic so if any one of the given strings
// is found, function will return true
func HasGroup(r *http.Request, searchGroups ...string) bool {
	groupMap := r.Context().Value(GroupCtxKey).(map[string]bool)

	for _, searchGroup := range searchGroups {
		if _, ok := groupMap[searchGroup]; ok {
			return true
		}
	}

	return false
}

// PanicHandlerFunc is wrapper util function for using
// against negroni#Recovery#PanicHandlerFunc function
//
// This function gives functionality of emailing a panic
// error message to desired parties along with slight
// formatting abilities of the sent message
//
// emailConfig:
//		Config struct for emailing error message.  If email
//		can't be sent, function will panic with error message
// subSearchStrings:
// 		Substring list of a library(s) path you wish to search for
// 		which will be taken from full stack trace and narrowed down
// 		to only display that library(s) in the message.  This is just
//		to help reduce the clutter of a stacktrace that you don't
//		care about
// func PanicHandlerFunc(to []string, from, subject string, subSearchStrings []string, mail SendMessage) func(*negroni.PanicInformation) {
// 	return func(info *negroni.PanicInformation) {
// 		var stack string
// 		ss := strings.Fields(info.StackAsString())

// 		if subSearchStrings == nil {
// 			for _, v := range ss {
// 				stack += v + "<br />"
// 			}
// 		} else {
// 			if len(subSearchStrings) == 0 {
// 				for _, v := range ss {
// 					stack += v + "<br />"
// 				}
// 			} else {
// 				for _, v := range ss {
// 					for _, t := range subSearchStrings {
// 						if strings.Contains(v, t) {
// 							stack += v + "<br />"
// 						}
// 					}
// 				}
// 			}
// 		}

// 		html := info.RequestDescription() + "<br /><br />" + stack
// 		err := SendEmail(
// 			to,
// 			from,
// 			subject,
// 			nil,
// 			[]byte(html),
// 			mail,
// 		)

// 		if err != nil {
// 			panic("sending mail error: " + err.Error())
// 		}
// 	}
// }

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

// GetJSONBuffer takes interface and json encodes it into a buffer and returns buffer
func GetJSONBuffer(item interface{}) (bytes.Buffer, error) {
	var buffer bytes.Buffer

	encoder := json.NewEncoder(&buffer)

	if err := encoder.Encode(&item); err != nil {
		return bytes.Buffer{}, err
	}

	return buffer, nil
}

// SetSecureCookie is used to set a cookie from a session
// The code used is copied pasted from the RedisStore#Save function from the redis store library
func SetSecureCookie(w http.ResponseWriter, session *sessions.Session, keyPairs ...[]byte) error {
	encoded, err := securecookie.EncodeMulti(session.Name(), session.ID, securecookie.CodecsFromPairs(keyPairs...)...)
	if err != nil {
		return err
	}
	http.SetCookie(w, sessions.NewCookie(session.Name(), encoded, session.Options))
	return nil
}

// func WriteError(w http.ResponseWriter, res HTTPResponseConfig) {
// 	http.Error(
// 		w,
// 		string(res.HTTPResponse),
// 		*res.HTTPStatus,
// 	)
// }
