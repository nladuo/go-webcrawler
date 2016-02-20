package model

const (
	TYPE_HTTP            = "HTTP"
	TYPE_HTTPS           = "HTTPS"
	ErrProxyTypeNotExist = "The Proxy Type must be TYPE_HTTP or TYPE_HTTPS"
)

type Proxy struct {
	IP   string
	Port string
	Type string //http or https
}
