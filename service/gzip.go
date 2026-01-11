package service

import (
	"compress/gzip"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// gzipHandler
func gzipHandler(h http.Handler, allows ...string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 跨域访问
		if len(allows) > 0 {
			allowsMap := map[string]bool{}
			for _, allow := range allows {
				allowsMap[allow] = true
			}
			origin := getOrigin(r)
			if origin != "" {
				if IsAllowed(r, allowsMap) {
					// url parse
					url, _ := url.Parse(origin)
					origin = fmt.Sprintf("%s://%s", url.Scheme, url.Host)
					w.Header().Set("Access-Control-Allow-Origin", origin)
				}
			}
		}
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			h.ServeHTTP(w, r)
			return
		}

		// w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		defer gz.Close()

		gzWriter := &gzipResponseWriter{ResponseWriter: w, Writer: gz}
		h.ServeHTTP(gzWriter, r)
	}
}

type gzipResponseWriter struct {
	http.ResponseWriter
	*gzip.Writer
	skip bool
}

func (w *gzipResponseWriter) WriteHeader(code int) {
	if code != http.StatusOK {
        w.skip = true
    } else {
        w.Header().Set("Content-Encoding", "gzip")
        w.Header().Del("Content-Length")
    }
    w.ResponseWriter.WriteHeader(code)
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	if w.skip {
        return w.ResponseWriter.Write(b)
    }
	return w.Writer.Write(b)
}

func (w *gzipResponseWriter) Flush() {
	w.Writer.Flush()
}

func (w *gzipResponseWriter) Header() http.Header {
	return w.ResponseWriter.Header()
}
