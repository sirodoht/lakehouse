package lakehousewiki

type ContextKey int

const (
	KeyUsername        ContextKey = iota
	KeyIsAuthenticated ContextKey = iota
)
