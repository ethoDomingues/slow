package slow

func NewMiddleware(f ...Func) Middlewares {
	return Middlewares(f)
}

type Middlewares []Func
