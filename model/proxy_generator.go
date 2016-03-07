package model

type ProxyGenerator interface {
	GetProxy() Proxy
	ChangeProxy(proxy Proxy) //to tell the generator the proxy is not available to download.
}
