package httpmw

type Middleware struct{}

func New() *Middleware {
	return &Middleware{}
}

func RequireContentType(contentType string) string {
	return contentType
}

func NewDefault() *Middleware {
	return &Middleware{}
}

var contentTypeJSON = "application/json"
