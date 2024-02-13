package webutil

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/gorilla/csrf"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/pkg/errors"
)

const (
	bodyRequiredTxt = "Request must have body"
	invalidJSONTxt  = "Invalid json"
	serverErrTxt    = "Server error, please try again later"

	// IntBase is default base to use for converting string to int64
	IntBase = 10

	// IntBitSize is default bit size to use for converting string to int64
	IntBitSize = 64
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
//------------------------- STRUCTS ----------------------------
//////////////////////////////////////////////////////////////////

// HTTPResponseConfig is used to give default header and response
// values of an http request
// This will mainly be used for middleware config structs
// to allow user of middleware more control on what gets
// send back to the user
type HTTPResponseConfig struct {
	HTTPStatus   *int
	HTTPResponse []byte
}

//////////////////////////////////////////////////////////////////
//------------------------- FUNCTIONS --------------------------
//////////////////////////////////////////////////////////////////

// SetToken is wrapper function for setting csrf token header
func SetToken(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(TOKEN_HEADER, csrf.Token(r))
}

// SendPayload is a wrapper for converting the payload map parameter into json and
// sending to the client
func SendPayload(w http.ResponseWriter, payload interface{}, errResp HTTPResponseConfig) error {
	w.Header().Set("Content-Type", JSON_CONTENT_HEADER)
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
