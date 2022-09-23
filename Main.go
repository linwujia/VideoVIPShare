package main

import (
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"strings"
)

func main() {
	http.HandleFunc("/", handleRequest)
	//handle定义请求访问该服务器里的/healthz路径，就有下面healthz去处理，healthz一般为健康检查
	err := http.ListenAndServe("0.0.0.0:80", nil)
	if err != nil {
		log.Fatal(err)
	}
}

//定义handle处理函数，只要该healthz被调用，就会写入ok
func handleRequest(w http.ResponseWriter, request *http.Request) {

	/*_, urlValue := getSourceUrl(request.URL)
	url, _ := url.Parse(urlValue)*/

	request.URL.Scheme = "https"
	request.URL.Host = "www.baidu.com"

	newRequest, _ := NewRequest(request, request.URL.String())
	newRequest.Header.Set("Host", request.URL.Host)

	response, err := http.DefaultClient.Do(newRequest)
	if err != nil {
		w.WriteHeader(502)
		return
	}

	for _, value := range response.Cookies() {
		w.Header().Add(value.Name, value.Value)
	}

	for k, v := range response.Header {
		for _, vv := range v {
			w.Header().Add(k, vv)
		}
	}
	w.WriteHeader(response.StatusCode)

	all, err := io.ReadAll(response.Body)
	if err != nil {
		w.WriteHeader(502)
		return
	}

	w.Write(all)
}

func NewRequest(r *http.Request, url string) (request *http.Request, err error) {
	request, err = http.NewRequest(r.Method, url, r.Body)

	if r.Header != nil {
		request.Header = r.Header.Clone()
	}

	if r.Trailer != nil {
		request.Trailer = r.Trailer.Clone()
	}

	if s := r.TransferEncoding; s != nil {
		s2 := make([]string, len(s))
		copy(s2, s)
		request.TransferEncoding = s2
	}
	request.Form = cloneURLValues(r.Form)
	request.PostForm = cloneURLValues(r.PostForm)
	request.MultipartForm = cloneMultipartForm(r.MultipartForm)

	return
}

func cloneURLValues(v url.Values) url.Values {
	if v == nil {
		return nil
	}
	// http.Header and url.Values have the same representation, so temporarily
	// treat it like http.Header, which does have a clone:
	return url.Values(http.Header(v).Clone())
}

func cloneMultipartForm(f *multipart.Form) *multipart.Form {
	if f == nil {
		return nil
	}
	f2 := &multipart.Form{
		Value: (map[string][]string)(http.Header(f.Value).Clone()),
	}
	if f.File != nil {
		m := make(map[string][]*multipart.FileHeader)
		for k, vv := range f.File {
			vv2 := make([]*multipart.FileHeader, len(vv))
			for i, v := range vv {
				vv2[i] = cloneMultipartFileHeader(v)
			}
			m[k] = vv2
		}
		f2.File = m
	}
	return f2
}

func cloneMultipartFileHeader(fh *multipart.FileHeader) *multipart.FileHeader {
	if fh == nil {
		return nil
	}
	fh2 := new(multipart.FileHeader)
	*fh2 = *fh
	fh2.Header = textproto.MIMEHeader(http.Header(fh.Header).Clone())
	return fh2
}

func getSourceUrl(url *url.URL) (string, string) {
	splits := strings.Split(url.RawQuery, "&")
	for _, split := range splits {
		paramValue := strings.Split(split, "=")
		if paramValue[0] == "url" {
			return paramValue[0], paramValue[1]
		}
	}

	return "", ""
}
