package kong

type Event interface {
	Tags() map[string]string
	Fields() map[string]interface{}
}

type Headers struct {
	XConsumerUsername string `json:"x-consumer-username"`
}

type Request struct {
	Uri     string  `json:"uri"`
	Method  string  `json:"method"`
	Headers Headers `json:"headers"`
}

type Response struct {
	Status string `json:"status"`
	Size   string `json:"size"`
}

type RequestEvent struct {
	Upstream string   `json:"upstream_uri"`
	Request  Request  `json:"request"`
	Response Response `json:"response"`
}

func (re *RequestEvent) Tags() map[string]string {
	return map[string]string{
		"upstream_uri":    re.Upstream,
		"request_uri":     re.Request.Uri,
		"request_method":  re.Request.Method,
		"client":          re.Request.Headers.XConsumerUsername,
		"response_status": re.Response.Status,
		"response_size":   re.Response.Size,
	}
}

func (re *RequestEvent) Fields() map[string]interface{} {
	return map[string]interface{}{
		"value": 1,
	}
}
