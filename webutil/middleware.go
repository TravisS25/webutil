package webutil

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/jmoiron/sqlx"
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

	// RoutingCtxKey is variable used as context key in middleware functions
	// to extract what urls a logged in user is allowed to access
	RoutingCtxKey = MiddlewareKey{KeyName: "routing"}

	// MiddlewareUserCtxKey is variable used as context key in middleware
	// functions to get a subset of user information of current
	// logged in user
	MiddlewareUserCtxKey = MiddlewareKey{KeyName: "MiddlewareUser"}
)

//////////////////////////////////////////////////////////////////
//------------------------ INTERFACES ---------------------------
//////////////////////////////////////////////////////////////////

// type MiddlewareAuth interface {
// 	GetID() string
// 	GetEmail() string
// }

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
	ServerErrorConfig

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
	QueryForSession func(db Querier, userID string) (sessionID string, err error)
}

// MiddlewareUser is config struct used to get the base
// authentication of user in middleware
type MiddlewareUser struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

//////////////////////////////////////////////////////////////////
//--------------------------- TYPES ---------------------------
//////////////////////////////////////////////////////////////////

// QueryDB should implement querying a database and returning
// results in bytes
type QueryDB func(req *http.Request, db Querier) ([]byte, error)

//////////////////////////////////////////////////////////////////
//-------------------------- STRUCTS --------------------------
//////////////////////////////////////////////////////////////////

type AuthHandler struct {
	db           Querier
	queryForUser QueryDB
	config       AuthHandlerConfig
}

func NewAuthHandler(
	db Querier,
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
		var middlewareUser MiddlewareUser
		var session *sessions.Session
		var err error
		var db *sqlx.DB

		// Setting up default values from passed configs if none are set
		SetHTTPResponseDefaults(
			&a.config.ClientErrorResponse,
			http.StatusBadRequest,
			[]byte(invalidCookieTxt),
		)
		SetHTTPResponseDefaults(
			&a.config.ServerErrorResponse,
			http.StatusInternalServerError,
			[]byte(serverErrTxt),
		)

		setUser := func() error {
			userBytes, err = a.queryForUser(r, a.db)

			if err != nil {
				canRecover := false

				switch err.(type) {
				case securecookie.Error:
					cookieErr := err.(securecookie.Error)

					if cookieErr.IsDecode() {
						http.Error(
							w,
							string(a.config.ClientErrorResponse.HTTPResponse),
							*a.config.ClientErrorResponse.HTTPStatus,
						)
					}

					http.Error(
						w,
						string(a.config.ServerErrorResponse.HTTPResponse),
						*a.config.ServerErrorResponse.HTTPStatus,
					)
				default:
					if err == sql.ErrNoRows {
						next.ServeHTTP(w, r)
					} else {
						if a.config.RecoverDB != nil {
							if db, err = a.config.RecoverDB(err); err == nil {
								a.db = db
								canRecover = true
								userBytes, err = a.queryForUser(r, a.db)

								if err == sql.ErrNoRows {
									next.ServeHTTP(w, r)
									return err
								}

								break
							}
						}

						http.Error(
							w,
							string(a.config.ServerErrorResponse.HTTPResponse),
							*a.config.ServerErrorResponse.HTTPStatus,
						)
					}
				}

				if !canRecover {
					return err
				}
			}

			err = json.Unmarshal(userBytes, &middlewareUser)

			if err != nil {
				http.Error(
					w,
					string(a.config.ServerErrorResponse.HTTPResponse),
					*a.config.ServerErrorResponse.HTTPStatus,
				)
				return err
			}

			return nil
		}

		serveWithEmail := func() {
			ctx := context.WithValue(r.Context(), UserCtxKey, userBytes)
			ctxWithEmail := context.WithValue(ctx, MiddlewareUserCtxKey, middlewareUser)
			next.ServeHTTP(w, r.WithContext(ctxWithEmail))
		}

		// If user sets SessionStore, then we try retrieving session from implemented
		// SessionStore which usually is a file system or in-memory database i.e. Redis
		if a.config.SessionStore != nil {
			session, err = a.config.SessionStore.Get(r, a.config.SessionConfig.SessionName)

			if err != nil {
				if err = setUser(); err != nil {
					return
				}

				serveWithEmail()
				return
			}

			// If session is considered new, that means
			// either current user is truly not logged in or cache was/is down
			if session.IsNew {
				// First we determine if user is sending a cookie with our user cookie key
				// If they are, try retrieving from db if AuthHandler#queryForUser is set
				// Else, continue to next handler
				if _, err = r.Cookie(a.config.SessionConfig.SessionName); err == nil {
					if err = setUser(); err != nil {
						return
					}

					// Here we test to see if our session backend is responsive
					// If it is, that means current user logged in while cache was down
					// and was using the database to grab their sessions but since session
					// backend is back up, we can grab current user's session from
					// database and set it to session backend and use that instead of database
					// for future requests
					if err = a.config.SessionStore.Ping(); err == nil && a.config.QueryForSession != nil {
						//canRecover := false
						sessionStr, err := a.config.QueryForSession(a.db, middlewareUser.ID)

						if err == nil {
							session, err = a.config.SessionStore.New(r, a.config.SessionConfig.SessionName)

							if err == nil {
								session.ID = sessionStr
								session.Values[a.config.SessionConfig.Keys.UserKey] = userBytes
								session.Save(r, w)
							}
						}
					}
				} else {
					next.ServeHTTP(w, r)
					return
				}
			} else {
				if val, ok := session.Values[a.config.SessionConfig.Keys.UserKey]; ok {
					userBytes = val.([]byte)
					err := json.Unmarshal(userBytes, &middlewareUser)

					if err != nil {
						//httputil.Logger.Errorf("invalid json from session: %s", err.Error())
						http.Error(
							w,
							string(a.config.ServerErrorResponse.HTTPResponse),
							*a.config.ServerErrorResponse.HTTPStatus,
						)
						return
					}
				} else {
					if err = setUser(); err != nil {
						return
					}
				}
			}
		} else {
			if err = setUser(); err != nil {
				return
			}
		}

		serveWithEmail()
	})
}

type GroupHandler struct {
	db             Querier
	queryForGroups QueryDB
	config         ServerErrorCacheConfig
}

func NewGroupHandler(
	db Querier,
	queryForGroups QueryDB,
	config ServerErrorCacheConfig,
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
			var db *sqlx.DB

			// Setting up default values from passed configs if none are set
			SetHTTPResponseDefaults(
				&g.config.ServerErrorResponse,
				http.StatusInternalServerError,
				[]byte(serverErrTxt),
			)
			user := user.(MiddlewareUser)
			groups := fmt.Sprintf(GroupKey, user.Email)

			setGroupFromDB := func() error {
				groupBytes, err = g.queryForGroups(r, g.db)

				if err != nil {
					isValid := false

					if err == sql.ErrNoRows {
						isValid = true
						next.ServeHTTP(w, r)
					} else {
						if g.config.RecoverDB != nil {
							if db, err = g.config.RecoverDB(err); err == nil {
								g.db = db
								isValid = true
								groupBytes, err = g.queryForGroups(r, g.db)

								if err == sql.ErrNoRows {
									next.ServeHTTP(w, r)
								}
							}
						}
					}

					if !isValid {
						http.Error(
							w,
							string(g.config.ServerErrorResponse.HTTPResponse),
							*g.config.ServerErrorResponse.HTTPStatus,
						)
					}

					return err
				}

				err = json.Unmarshal(groupBytes, &groupMap)

				if err != nil {
					http.Error(
						w,
						string(g.config.ServerErrorResponse.HTTPResponse),
						*g.config.ServerErrorResponse.HTTPStatus,
					)
					return err
				}

				return nil
			}

			// If cache is set, try to get group info from cache
			// Else query from db
			if g.config.Cache != nil {
				groupBytes, err = g.config.Cache.Get(groups)

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
						http.Error(
							w,
							string(g.config.ServerErrorResponse.HTTPResponse),
							*g.config.ServerErrorResponse.HTTPStatus,
						)
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
	db          Querier
	queryRoutes QueryDB
	pathRegex   PathRegex
	nonUserURLs map[string]bool
	config      ServerErrorCacheConfig
}

func NewRoutingHandler(
	db Querier,
	queryRoutes QueryDB,
	pathRegex PathRegex,
	nonUserURLs map[string]bool,
	config ServerErrorCacheConfig,
) *RoutingHandler {
	return &RoutingHandler{
		db:          db,
		queryRoutes: queryRoutes,
		pathRegex:   pathRegex,
		nonUserURLs: nonUserURLs,
		config:      config,
	}
}

func (routing *RoutingHandler) MiddlewareFunc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodOptions {
			var urlBytes []byte
			var urls map[string]bool
			var err error
			var db *sqlx.DB

			SetHTTPResponseDefaults(
				&routing.config.ClientErrorResponse,
				http.StatusForbidden,
				[]byte(forbiddenURLTxt),
			)
			SetHTTPResponseDefaults(
				&routing.config.ServerErrorResponse,
				http.StatusInternalServerError,
				[]byte(serverErrTxt),
			)

			// Queries from db and sets the bytes returned to url map
			setURLsFromDB := func() error {
				urlBytes, err = routing.queryRoutes(r, routing.db)

				if err != nil {
					isValid := false

					if err == sql.ErrNoRows {
						http.Error(
							w,
							string(routing.config.ClientErrorResponse.HTTPResponse),
							*routing.config.ClientErrorResponse.HTTPStatus,
						)
						return err
					}

					if routing.config.RecoverDB != nil {
						if db, err = routing.config.RecoverDB(err); err == nil {
							routing.db = db
							urlBytes, err = routing.queryRoutes(r, routing.db)

							if err == nil {
								isValid = true
							} else {
								if err == sql.ErrNoRows {
									http.Error(
										w,
										string(routing.config.ClientErrorResponse.HTTPResponse),
										*routing.config.ClientErrorResponse.HTTPStatus,
									)
									return err
								}
							}
						}
					}

					if !isValid {
						http.Error(
							w,
							string(routing.config.ServerErrorResponse.HTTPResponse),
							*routing.config.ServerErrorResponse.HTTPStatus,
						)
					}

					return err
				}

				err = json.Unmarshal(urlBytes, &urls)

				if err != nil {
					http.Error(
						w,
						string(routing.config.ServerErrorResponse.HTTPResponse),
						*routing.config.ServerErrorResponse.HTTPStatus,
					)
					return err
				}

				return nil
			}

			pathExp, err := routing.pathRegex(r)

			if err != nil {
				http.Error(
					w,
					string(routing.config.ServerErrorResponse.HTTPResponse),
					*routing.config.ServerErrorResponse.HTTPStatus,
				)
				return
			}

			allowedPath := false
			user := r.Context().Value(MiddlewareUserCtxKey)

			if user != nil {
				user := user.(MiddlewareUser)
				key := fmt.Sprintf(URLKey, user.Email)

				if routing.config.Cache != nil {
					urlBytes, err = routing.config.Cache.Get(key)

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
								http.Error(
									w,
									string(routing.config.ClientErrorResponse.HTTPResponse),
									*routing.config.ClientErrorResponse.HTTPStatus,
								)
								return
							}
						}
					} else {
						err = json.Unmarshal(urlBytes, &urls)

						if err != nil {
							http.Error(
								w,
								string(routing.config.ServerErrorResponse.HTTPResponse),
								*routing.config.ServerErrorResponse.HTTPStatus,
							)
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
				if _, ok := routing.nonUserURLs[pathExp]; ok {
					allowedPath = true
				}
			}

			// If returned urls do not match any urls user is allowed to
			// access, return with error response
			if !allowedPath {
				http.Error(
					w,
					string(routing.config.ClientErrorResponse.HTTPResponse),
					*routing.config.ClientErrorResponse.HTTPStatus,
				)
				return
			}

			var ctx context.Context

			if user != nil {
				ctx = context.WithValue(r.Context(), RoutingCtxKey, urls)
			} else {
				ctx = context.WithValue(r.Context(), RoutingCtxKey, routing.nonUserURLs)
			}

			next.ServeHTTP(w, r.WithContext(ctx))
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
