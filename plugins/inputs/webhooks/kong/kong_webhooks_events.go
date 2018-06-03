package kong

type Event interface {
	Tags() map[string]string
	Fields() map[string]interface{}
}

type RequestEvent struct {
	Upstream	string `json:"upstream_uri"`
	Request     string `json:"request"`
	Response    string `json:"response"`
}

func (re *RequestEvent) Tags() map[string]string {
	return map[string]string{
		"upstream_uri": re.Upstream,
		"request_uri": re.Request.uri,
		"request_method": re.Request.method,
		"client": re.Request.headers.Authorization,
		"response_status": re.Response.status,
		"response_size": re.Response.size
	}
}

func (re *RequestEvent) Fields() map[string]interface{} {
	return map[string]interface{}{
		"value": 1,
	}
}