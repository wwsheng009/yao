package core

import (
	"fmt"
	"strings"

	jsoniter "github.com/json-iterator/go"
	"github.com/yaoapp/gou/process"
	"github.com/yaoapp/kun/any"
	"github.com/yaoapp/kun/log"
)

// Cache the cache
type Cache struct {
	Data          string
	Global        string
	Config        string
	Guard         string
	GuardRedirect string
	HTML          string
}

// Caches the caches
var Caches = map[string]*Cache{}

// NewRequestMock is the constructor for Request.
func NewRequestMock(mock *PageMock) *Request {
	if mock == nil {
		mock = &PageMock{Method: "GET"}
	}
	return &Request{
		Method:  mock.Method,
		Query:   mock.Query,
		Body:    mock.Body,
		Payload: mock.Payload,
		Referer: mock.Referer,
		Headers: mock.Headers,
		Params:  mock.Params,
		URL:     mock.URL,
	}
}

// ExecString get the data
func (r *Request) ExecString(data string) (Data, error) {
	var res Data
	err := jsoniter.UnmarshalFromString(data, &res)
	if err != nil {
		return nil, err
	}
	r.Exec(res)
	return res, nil
}

// Exec get the data
func (r *Request) Exec(m map[string]interface{}) error {
	ignores := map[string]bool{}
	for key, value := range m {
		if strings.HasPrefix(key, "$") && !ignores[key] {
			res, err := r.call(value)
			if err != nil {
				log.Error("[Request] Exec key:%s, value:%s, %s", key, value, err.Error())
				return err
			}
			newKey := key[1:]
			m[newKey] = res
			ignores[newKey] = true
			delete(m, key)
			continue
		}

		res, err := r.execValue(value)
		if err != nil {
			log.Error("[Request] Exec key:%s, value:%s, %s", key, value, err.Error())
			return err
		}
		m[key] = res

	}

	return nil
}

func (r *Request) execValue(value interface{}) (interface{}, error) {
	switch v := value.(type) {
	case string:

		if strings.HasPrefix(v, "$query.") {
			key := strings.TrimLeft(v, "$query.")
			if r.Query.Has(key) {
				return r.Query.Get(key), nil
			}
			return "", nil
		}

		if strings.HasPrefix(v, "$url.") {
			key := strings.TrimLeft(v, "$url.")
			switch key {
			case "path":
				return r.URL.Path, nil

			case "host":
				return r.URL.Host, nil

			case "domain":
				return r.URL.Domain, nil

			case "scheme":
				return r.URL.Scheme, nil
			}
			return "", nil
		}

		if strings.HasPrefix(v, "$header.") {
			key := strings.TrimLeft(v, "$header.")
			if r.Headers.Has(key) {
				return r.Headers.Get(key), nil
			}
			return "", nil
		}

		if strings.HasPrefix(v, "$param.") {
			key := strings.TrimLeft(v, "$param.")
			if value, has := r.Params[key]; has {
				return value, nil
			}
			return "", nil
		}

		if strings.HasPrefix(v, "$payload.") {
			key := strings.TrimLeft(v, "$payload.")
			if value, has := r.Payload[key]; has {
				return value, nil
			}
			return "", nil
		}

		if strings.HasPrefix(v, "$") {
			return r.call(strings.TrimLeft(v, "$"))
		}
		return v, nil

	case []interface{}:
		for i, item := range v {
			res, err := r.execValue(item)
			if err != nil {
				return nil, err
			}
			v[i] = res
		}
		return v, nil

	case []string:
		interfaceSlice := make([]interface{}, len(v))
		for i, item := range v {
			interfaceSlice[i] = item
		}
		return r.execValue(interfaceSlice)

	case map[string]interface{}:

		if _, ok := v["process"].(string); ok {
			if call, _ := v["__exec"].(bool); call {
				res, err := r.call(v)
				if err != nil {
					return nil, err
				}
				return res, nil
			}
		}

		err := r.Exec(v)
		if err != nil {
			return nil, err
		}
		return v, nil

	default:
		return v, nil
	}
}

func (r *Request) call(p interface{}) (interface{}, error) {

	processName := ""
	processArgs := []interface{}{r}
	switch v := p.(type) {
	case string:
		processName = v
		break

	case map[string]interface{}:
		if name, ok := v["process"].(string); ok {
			processName = name
		}

		if args, ok := v["args"].([]interface{}); ok {
			args, err := r.parseArgs(args)
			if err != nil {
				return nil, err
			}
			processArgs = append(args, processArgs...)
		}
	}

	if processName == "" {
		return nil, fmt.Errorf("process name is empty")
	}

	process, err := process.Of(processName, processArgs...)
	if err != nil {
		return nil, err
	}

	if r.Sid != "" {
		process.WithSID(r.Sid)
	}

	v, err := process.Exec()
	if err != nil {
		log.Error("[Request] process %s %s", processName, err.Error())
	}
	return v, err
}

func (r *Request) parseArgs(args []interface{}) ([]interface{}, error) {

	data := any.MapOf(map[string]interface{}{
		"param":   r.Params,
		"query":   r.Query,
		"payload": map[string]interface{}{},
		"header":  r.Headers,
		"theme":   r.Theme,
		"locale":  r.Locale,
		"url":     r.URL.Map(),
	}).Dot()

	for i, arg := range args {
		switch v := arg.(type) {

		case string:
			if !strings.HasPrefix(v, "$") {
				args[i] = v
				break
			}

			key := strings.TrimLeft(v, "$")
			args[i] = key
			if data.Has(key) {
				v := data.Get(key)
				args[i] = v
				if strings.HasPrefix(key, "query.") || strings.HasPrefix(key, "header.") {
					switch arg := v.(type) {
					case []interface{}:
						if len(arg) == 1 {
							args[i] = arg[0]
						}
					case []string:
						if len(arg) == 1 {
							args[i] = arg[0]
						}
					}
				}
			}
			break

		case int, int8, int16, int32, int64, float32, float64, bool, []string, []int, []int8, []int16, []int32, []int64, []float32, []float64, []bool:
			args[i] = v
			break

		case []interface{}:
			res, err := r.parseArgs(v)
			if err != nil {
				return nil, err
			}
			args[i] = res
			break

		case map[string]interface{}:
			res, err := r.parseArgs([]interface{}{v})
			if err != nil {
				return nil, err
			}
			args[i] = res[0]
			break
		}
	}

	return args, nil
}

// Map URL to map
func (url ReqeustURL) Map() Data {
	return map[string]interface{}{
		"url":    url.URL,
		"scheme": url.Scheme,
		"domain": url.Domain,
		"host":   url.Host,
		"path":   url.Path,
	}
}

// SetCache set the cache
func SetCache(file string, html string, data string, global string) *Cache {
	Caches[file] = &Cache{
		Data:   data,
		HTML:   html,
		Global: global,
	}
	return Caches[file]
}

// GetCache get the cache
func GetCache(file string) *Cache {
	if cache, has := Caches[file]; has {
		return cache
	}
	return nil
}

// RemoveCache remove the cache
func RemoveCache(file string) {
	delete(Caches, file)
}

// CleanCache clean the cache
func CleanCache() {
	Caches = map[string]*Cache{}
}
