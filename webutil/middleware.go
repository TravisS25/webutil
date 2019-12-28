package webutil

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"

	"github.com/TravisS25/httputil"
)

//////////////////////////////////////////////////////////////////
//---------------------- SERVER ERR MGS -----------------------
//////////////////////////////////////////////////////////////////

const (
	serverErrTxt     = "Server error"
	forbiddenURLTxt  = "Forbidden to access url"
	invalidCookieTxt = "Invalid cookie"
)

//////////////////////////////////////////////////////////////////
//------------------------ QUERY TYPES --------------------------
//////////////////////////////////////////////////////////////////

// Query types to be used against the Middleware#QueryDB function
const (
	UserQuery = iota
	GroupQuery
	RoutingQuery
	SessionQuery
)

//////////////////////////////////////////////////////////////////
//----------------------- STRINGS CONSTS -----------------------
//////////////////////////////////////////////////////////////////

const (
	// GroupKey is used as a key when pulling a user's groups out from cache
	GroupKey = "%s-groups"

	// URLKey is used as a key when pulling a user's allowed urls from cache
	URLKey = "%s-urls"
)

//////////////////////////////////////////////////////////////////
//------------------------ CTX KEYS ---------------------------
//////////////////////////////////////////////////////////////////

var (
	// UserCtxKey is variable used as context key in middleware functions
	// to get logged in user
	UserCtxKey = MiddlewareKey{KeyName: "user"}

	// GroupCtxKey is variable used as context key in middleware functions
	// to get all user groups current logged in user is in
	GroupCtxKey = MiddlewareKey{KeyName: "groupName"}

	// MiddlewareUserCtxKey is variable used as context key in middleware
	// functions to get a subset of user information of current
	// logged in user
	MiddlewareUserCtxKey = MiddlewareKey{KeyName: "middlewareUser"}
)

//////////////////////////////////////////////////////////////////
//----------------------- CONFIG STRUCTS -----------------------
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

// MiddlewareKey is config struct used to create context keys
// for middleware functions
type MiddlewareKey struct {
	KeyName string
}

// SessionConfig is config struct used to name a session
// with "SessionName" variable and then have arbitrary
// keys attached to to it with "Keys" variable
type SessionConfig struct {
	SessionName string
	Keys        SessionKeys
}

// SessionKeys is config struct used with an instance
// of SessionConfig to define key names for a certain session
type SessionKeys struct {
	UserKey string
}

// AuthHandlerConfig is used as config struct for AuthHandler
// These settings are not required but if user wants to use things
// like a different session store besides a database, these should
// be set
type AuthHandlerConfig struct {
	// SessionStore is used to implement a backend to store sessions
	// besides a database like file system or in-memory database
	// i.e. Redis
	SessionStore SessionStore

	// SessionKeys is just an arbitrary set of common key names to store
	// in a session values
	SessionConfig SessionConfig

	// QueryForSession is used for inserting a session value from a database
	// to the entity that implements SessionStore
	// This is used in the case where a person logs in while the entity that
	// implements SessionStore is down and must query session from database
	//
	// If this is set, the implementing function should return the session id
	// from a database which will then be set to SessionStore if/when it comes back up
	//
	// This is bascially a recovery method if implementing SessionStore ever
	// goes down or some how gets its values flushed
	QueryForSession func(w http.ResponseWriter, db httputil.Querier, userID string) (sessionID string, err error)

	// DecodeCookieErrResponse is config used to respond to user if decoding
	// a cookie is invalid
	// This usually happens when a user sends an invalid cookie on request
	//
	// Default status value is http.StatusBadRequest
	// Default response value is []byte("Invalid cookie")
	DecodeCookieErrResponse HTTPResponseConfig

	// ServerErrResponse is config used to respond to user if some type
	// of server error occurs
	//
	// Default status value is http.StatusInternalServerError
	// Default response value is []byte("Server error")
	ServerErrResponse HTTPResponseConfig

	// NoRowsErrResponse is config used to respond to user if the returned
	// error result of AuthHandler#queryForUser is sql.ErrNoRows
	// This should be returned if there are no results when trying to grab
	// a user from the database
	//
	// Default status value is http.StatusInternalServerError
	// Default response value is []byte("User Not Found")
	//NoRowsErrResponse HTTPResponseConfig
}

// GroupHandlerConfig is config struct used for GroupHandler
// The settings don't have to be set but if programmer wants to
// be able to store user group information in cache instead
// of database, this can be achieved by implementing CacheStore
type GroupHandlerConfig struct {
	// CacheStore is used for retrieving results from a in-memory
	// database like Redis
	CacheStore CacheStore

	// IgnoreCacheNil will query database for group information
	// even if cache returns nil
	// CacheStore must be initialized to use this
	IgnoreCacheNil bool

	// ServerErrResponse is config used to respond to user if some type
	// of server error occurs
	//
	// Default status value is http.StatusInternalServerError
	// Default response value is []byte("Server error")
	ServerErrResponse HTTPResponseConfig
}

// RoutingHandlerConfig is config struct for RoutingHandler
// These settings don't have to be set but if user wishes
// to use caching for routing paths
type RoutingHandlerConfig struct {
	// CacheStore is used for retrieving results from a in-memory
	// database like Redis
	CacheStore CacheStore

	// IgnoreCacheNil will query database for routing information
	// even if cache returns nil
	// CacheStore must be initialized for this to activate
	IgnoreCacheNil bool

	// ServerErrResponse is config used to respond to user if some type
	// of server error occurs
	//
	// Default status value is http.StatusInternalServerError
	// Default response value is []byte("Server Error")
	ServerErrResponse HTTPResponseConfig

	// UnauthorizedErrResponse is config used to respond to user if none
	// of the nonUserURLs keys or queried urls match the apis
	// a user is allowed to access
	//
	// Default status value is http.StatusForbidden
	// Default response value is []byte("Forbidden to access url")
	ForbiddenURLErrResponse HTTPResponseConfig
}

type middlewareUser struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

//////////////////////////////////////////////////////////////////
//--------------------------- TYPES ---------------------------
//////////////////////////////////////////////////////////////////

type QueryDB func(w http.ResponseWriter, res *http.Request, db httputil.Querier) ([]byte, error)

//////////////////////////////////////////////////////////////////
//-------------------------- STRUCTS --------------------------
//////////////////////////////////////////////////////////////////

type AuthHandler struct {
	db           httputil.DBInterfaceV2
	queryForUser QueryDB
	config       AuthHandlerConfig
}

func NewAuthHandler(
	db httputil.DBInterfaceV2,
	queryForUser QueryDB,
	config AuthHandlerConfig,
) *AuthHandler {
	return &AuthHandler{
		db:           db,
		queryForUser: queryForUser,
		config:       config,
	}
}

func (a *AuthHandler) MiddlewareFunc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var userBytes []byte
		var middlewareUser middlewareUser
		var session *sessions.Session
		var err error

		// Setting up default values from passed configs if none are set
		SetHTTPResponseDefaults(&a.config.DecodeCookieErrResponse, http.StatusBadRequest, []byte(invalidCookieTxt))
		SetHTTPResponseDefaults(&a.config.ServerErrResponse, http.StatusInternalServerError, []byte(serverErrTxt))

		setUser := func() error {
			userBytes, err = a.queryForUser(w, r, a.db)

			if err != nil {
				isFatalErr := true

				switch err.(type) {
				case securecookie.Error:
					cookieErr := err.(securecookie.Error)

					if cookieErr.IsDecode() {
						isFatalErr = false
						w.WriteHeader(*a.config.DecodeCookieErrResponse.HTTPStatus)
						w.Write(a.config.DecodeCookieErrResponse.HTTPResponse)
					}

					w.WriteHeader(*a.config.ServerErrResponse.HTTPStatus)
					w.Write(a.config.ServerErrResponse.HTTPResponse)
				default:
					if err == sql.ErrNoRows {
						isFatalErr = false
						next.ServeHTTP(w, r)
						//return err
					} else {
						w.WriteHeader(*a.config.ServerErrResponse.HTTPStatus)
						w.Write(a.config.ServerErrResponse.HTTPResponse)
					}
				}

				if isFatalErr {
					httputil.Logger.Errorf("query for user err: %s", err.Error())
				}

				return err
			}

			err = json.Unmarshal(userBytes, &middlewareUser)

			if err != nil {
				w.WriteHeader(*a.config.ServerErrResponse.HTTPStatus)
				w.Write(a.config.ServerErrResponse.HTTPResponse)
				return err
			}

			return nil
		}

		// If user sets SessionStore, then we try retrieving session from implemented
		// SessionStore which usually is a file system or in-memory database i.e. Redis
		if a.config.SessionStore != nil {
			session, err = a.config.SessionStore.Get(r, a.config.SessionConfig.SessionName)

			if err != nil {
				w.WriteHeader(*a.config.ServerErrResponse.HTTPStatus)
				w.Write(a.config.ServerErrResponse.HTTPResponse)
				return
			}

			// If session is considered new, that means
			// either current user is truly not logged in or cache was/is down
			if session.IsNew {
				//fmt.Printf("new session\n")

				// First we determine if user is sending a cookie with our user cookie key
				// If they are, try retrieving from db if AuthHandler#queryForUser is set
				// Else, continue to next handler
				if _, err = r.Cookie(a.config.SessionConfig.SessionName); err == nil {
					//fmt.Printf("has cookie but not found in store\n")
					if err = setUser(); err != nil {
						fmt.Printf("within user\n")
						return
					}

					// Here we test to see if our session backend is responsive
					// If it is, that means current user logged in while cache was down
					// and was using the database to grab their sessions but since session
					// backend is back up, we can grab current user's session from
					// database and set it to session backend and use that instead of database
					// for future requests
					if err = a.config.SessionStore.Ping(); err == nil && a.config.QueryForSession != nil {
						//fmt.Printf("ping successful\n")
						sessionStr, err := a.config.QueryForSession(w, a.db, middlewareUser.ID)

						if err != nil {
							if err == sql.ErrNoRows {
								fmt.Printf("auth middleware db no row found\n")
								next.ServeHTTP(w, r)
								return
							}

							fmt.Printf("within query session\n")

							w.WriteHeader(*a.config.ServerErrResponse.HTTPStatus)
							w.Write(a.config.ServerErrResponse.HTTPResponse)
							return
						}

						fmt.Printf("session bytes: %s\n", sessionStr)

						session, err = a.config.SessionStore.New(r, a.config.SessionConfig.SessionName)

						if err != nil {
							fmt.Printf("within new session\n")
							w.WriteHeader(*a.config.ServerErrResponse.HTTPStatus)
							w.Write(a.config.ServerErrResponse.HTTPResponse)
							return
						}

						session.ID = sessionStr
						fmt.Printf("session id: %s\n", session.ID)
						session.Values[a.config.SessionConfig.Keys.UserKey] = userBytes
						session.Save(r, w)
					}

					//setCtxAndServe()
				} else {
					//fmt.Printf("new session, no cookie\n")
					next.ServeHTTP(w, r)
					return
				}
			} else {
				//fmt.Printf("not new session")
				if val, ok := session.Values[a.config.SessionConfig.Keys.UserKey]; ok {
					//fmt.Printf("found in session")
					userBytes = val.([]byte)
					err := json.Unmarshal(userBytes, &middlewareUser)

					if err != nil {
						httputil.Logger.Errorf("invalid json from session: %s", err.Error())
						w.WriteHeader(*a.config.ServerErrResponse.HTTPStatus)
						w.Write(a.config.ServerErrResponse.HTTPResponse)
						return
					}
				} else {
					next.ServeHTTP(w, r)
					return
				}
			}
		} else {
			if err = setUser(); err != nil {
				return
			}
		}

		ctx := context.WithValue(r.Context(), UserCtxKey, userBytes)
		ctxWithEmail := context.WithValue(ctx, MiddlewareUserCtxKey, middlewareUser)
		next.ServeHTTP(w, r.WithContext(ctxWithEmail))
	})
}

// setConfig is really only here for testing purposes
func (a *AuthHandler) setConfig(config AuthHandlerConfig) {
	a.config = config
}

type GroupHandler struct {
	config         GroupHandlerConfig
	db             httputil.DBInterfaceV2
	queryForGroups QueryDB
}

func NewGroupHandler(
	db httputil.DBInterfaceV2,
	queryForGroups QueryDB,
	config GroupHandlerConfig,
) *GroupHandler {
	return &GroupHandler{
		config:         config,
		db:             db,
		queryForGroups: queryForGroups,
	}
}

func (g *GroupHandler) MiddlewareFunc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value(MiddlewareUserCtxKey)

		if user != nil {
			var groupMap map[string]bool
			var err error
			var groupBytes []byte

			// Setting up default values from passed configs if none are set
			SetHTTPResponseDefaults(&g.config.ServerErrResponse, http.StatusInternalServerError, []byte(serverErrTxt))
			user := user.(middlewareUser)
			groups := fmt.Sprintf(GroupKey, user.Email)

			setGroupFromDB := func() error {
				fmt.Printf("group middlware query db\n")
				groupBytes, err = g.queryForGroups(w, r, g.db)

				if err != nil {
					if err == sql.ErrNoRows {
						next.ServeHTTP(w, r)
						return err
					}

					w.WriteHeader(*g.config.ServerErrResponse.HTTPStatus)
					w.Write(g.config.ServerErrResponse.HTTPResponse)
					//w.Write([]byte("err from db"))
					return err
				}

				err = json.Unmarshal(groupBytes, &groupMap)

				if err != nil {
					w.WriteHeader(*g.config.ServerErrResponse.HTTPStatus)
					w.Write(g.config.ServerErrResponse.HTTPResponse)
					//w.Write([]byte("json err from set group"))
					return err
				}

				return nil
			}

			// If cache is set, try to get group info from cache
			// Else query from db
			if g.config.CacheStore != nil {
				groupBytes, err = g.config.CacheStore.Get(groups)

				if err != nil {
					// If err occurs and is not a nil err,
					// query from database
					if err != ErrCacheNil {
						if err = setGroupFromDB(); err != nil {
							return
						}
					} else {
						// If GroupHandlerConfig#IgnoreCacheNil is set,
						// then we ignore that the cache result came back
						// nil and query the database anyways
						if g.config.IgnoreCacheNil {
							if err = setGroupFromDB(); err != nil {
								return
							}
						} else {
							next.ServeHTTP(w, r)
							return
						}
					}
				} else {
					err = json.Unmarshal(groupBytes, &groupMap)

					if err != nil {
						w.WriteHeader(*g.config.ServerErrResponse.HTTPStatus)
						w.Write(g.config.ServerErrResponse.HTTPResponse)
						return
					}
				}
			} else {
				if err = setGroupFromDB(); err != nil {
					return
				}
			}

			ctx := context.WithValue(r.Context(), GroupCtxKey, groupMap)
			next.ServeHTTP(w, r.WithContext(ctx))
		} else {
			next.ServeHTTP(w, r)
		}
	})
}

type RoutingHandler struct {
	db          httputil.DBInterfaceV2
	queryDB     QueryDB
	pathRegex   httputil.PathRegex
	nonUserURLs map[string]bool
	config      RoutingHandlerConfig
}

func NewRoutingHandler(
	db httputil.DBInterfaceV2,
	queryDB QueryDB,
	pathRegex httputil.PathRegex,
	nonUserURLs map[string]bool,
	config RoutingHandlerConfig,
) *RoutingHandler {
	return &RoutingHandler{
		db:          db,
		queryDB:     queryDB,
		pathRegex:   pathRegex,
		nonUserURLs: nonUserURLs,
		config:      config,
	}
}

func (routing *RoutingHandler) MiddlewareFunc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//fmt.Printf("routing middleware\n")
		if r.Method != http.MethodOptions {
			var urlBytes []byte
			var urls map[string]bool
			var err error

			SetHTTPResponseDefaults(&routing.config.ForbiddenURLErrResponse, http.StatusForbidden, []byte(forbiddenURLTxt))
			SetHTTPResponseDefaults(&routing.config.ServerErrResponse, http.StatusInternalServerError, []byte(serverErrTxt))

			// Queries from db and sets the bytes returned to url map
			setURLsFromDB := func() error {
				urlBytes, err = routing.queryDB(w, r, routing.db)

				if err != nil {
					if err == sql.ErrNoRows {
						w.WriteHeader(*routing.config.ForbiddenURLErrResponse.HTTPStatus)
						w.Write(routing.config.ForbiddenURLErrResponse.HTTPResponse)
						return err
					}

					w.WriteHeader(*routing.config.ServerErrResponse.HTTPStatus)
					w.Write(routing.config.ServerErrResponse.HTTPResponse)
					return err
				}

				err = json.Unmarshal(urlBytes, &urls)

				if err != nil {
					w.WriteHeader(*routing.config.ServerErrResponse.HTTPStatus)
					w.Write(routing.config.ServerErrResponse.HTTPResponse)
					return err
				}

				return nil
			}

			pathExp, err := routing.pathRegex(r)

			if err != nil {
				w.WriteHeader(*routing.config.ServerErrResponse.HTTPStatus)
				w.Write(routing.config.ServerErrResponse.HTTPResponse)
				return
			}

			allowedPath := false
			user := r.Context().Value(MiddlewareUserCtxKey)

			if user != nil {
				//fmt.Printf("routing user\n")
				user := user.(middlewareUser)
				key := fmt.Sprintf(URLKey, user.Email)

				if routing.config.CacheStore != nil {
					urlBytes, err = routing.config.CacheStore.Get(key)

					if err != nil {
						if err != ErrCacheNil {
							if err = setURLsFromDB(); err != nil {
								return
							}
						} else {
							// If RoutingHandlerConfig#IgnoreCacheNil is set,
							// then we ignore that the cache result came back
							// nil and query the database anyways
							//
							// Else we return forbidden error
							if routing.config.IgnoreCacheNil {
								if err = setURLsFromDB(); err != nil {
									return
								}
							} else {
								w.WriteHeader(*routing.config.ForbiddenURLErrResponse.HTTPStatus)
								w.Write(routing.config.ForbiddenURLErrResponse.HTTPResponse)
								return
							}
						}
					} else {
						err = json.Unmarshal(urlBytes, &urls)

						if err != nil {
							w.WriteHeader(*routing.config.ServerErrResponse.HTTPStatus)
							w.Write(routing.config.ServerErrResponse.HTTPResponse)
							return
						}
					}

					if _, ok := urls[pathExp]; ok {
						allowedPath = true
					}
				} else {
					if err = setURLsFromDB(); err != nil {
						return
					}

					if _, ok := urls[pathExp]; ok {
						allowedPath = true
					}
				}
			} else {
				//fmt.Printf("non user\n")
				//fmt.Printf("non user urls: %v\n", routing.nonUserURLs)
				if _, ok := routing.nonUserURLs[pathExp]; ok {
					allowedPath = true
				}
			}

			// If returned urls do not match any urls user is allowed to
			// access, return with error response
			if !allowedPath {
				w.WriteHeader(*routing.config.ForbiddenURLErrResponse.HTTPStatus)
				w.Write(routing.config.ForbiddenURLErrResponse.HTTPResponse)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

//////////////////////////////////////////////////////////////////
//------------------------ FUNCTIONS --------------------------
//////////////////////////////////////////////////////////////////

// SetHTTPResponseDefaults is util function to set default values for passed
// config if values for nil
func SetHTTPResponseDefaults(config *HTTPResponseConfig, defaultStatus int, defaultResponse []byte) {
	if config.HTTPStatus == nil {
		config.HTTPStatus = &defaultStatus
	}
	if config.HTTPResponse == nil {
		config.HTTPResponse = defaultResponse
	}
}
