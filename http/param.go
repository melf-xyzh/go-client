/**
 * @Time    :2022/11/9 10:50
 * @Author  :Xiaoyu.Zhang
 */

package httpclient

import "net/http"

type (
	// ContentType is an option used for RequestClientOptions
	ContentType string
	// QueryMap is an option used for RequestClientOptions
	QueryMap map[string]string
	// HeaderMap is an option used for RequestClientOptions
	HeaderMap map[string]string
	// Body is an option used for RequestClientOptions
	Body string
	// Cookies is an option used for RequestClientOptions
	Cookies []*http.Cookie
	// TableName is an option used for RequestClientOptions
	TableName string
	// SplitTable is an option used for RequestClientOptions
	SplitTable bool
)

type RequestClientOptions interface {
	setRequestClientOption(formatPr *RequestClientPr)
}

type RequestClientPr struct {
	ContentType string // ContentType
	Body        string // RequestBody
	TableName   string
	QueryMap    map[string]string
	HeaderMap   map[string]string
	Cookies     []*http.Cookie
	SplitTable  bool
}

func (o ContentType) setRequestClientOption(pr *RequestClientPr) {
	pr.ContentType = string(o)
}

func (o Body) setRequestClientOption(pr *RequestClientPr) {
	pr.Body = string(o)
}

func (o QueryMap) setRequestClientOption(pr *RequestClientPr) {
	pr.QueryMap = o
}

func (o HeaderMap) setRequestClientOption(pr *RequestClientPr) {
	pr.HeaderMap = o
}

func (o Cookies) setRequestClientOption(pr *RequestClientPr) {
	pr.Cookies = o
}

func (o TableName) setRequestClientOption(pr *RequestClientPr) {
	pr.TableName = string(o)
}

func (o SplitTable) setRequestClientOption(pr *RequestClientPr) {
	pr.SplitTable = bool(o)
}
