package routing

import "errors"

var ErrorMethodMismatch = errors.New("405 Method Not Allowed")
var ErrorNotFound = errors.New("404 Not Found")

type MatchInfo struct {
	MethodNotAllowed error

	Route  *Route
	Router *Router
}
