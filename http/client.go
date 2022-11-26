/**
 * @Time    :2022/10/19 14:26
 * @Author  :Xiaoyu.Zhang
 */

package httpclient

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"github.com/json-iterator/go"
	"github.com/melf-xyzh/go-client/commons"
	"github.com/melf-xyzh/go-client/model"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"
	"unsafe"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type HttpClient struct {
	RequestConfig
	rcd           *clientmod.HttpRequestRcd
	req           *http.Request
	res           *http.Response
	tableName     string
	once          sync.Once
	Error         error
	retryCount    int   // 重试次数统计
	retry         int   // 重试次数
	retryInterval []int // 重试间隔
}

// NewHttpClient
/**
 *  @Description: 创建Http连接池
 *  @receiver rc
 *  @return *http.Client
 */
func (rc *RequestConfig) NewHttpClient() *http.Client {
	rc.httpConnPoll = &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        100, // 最大空闲连接数
			MaxIdleConnsPerHost: 0,   // 单独的Host（ip+port）最大的空闲连接
			MaxConnsPerHost:     0,   // 单独的host最大链接限制（默认没有限制）
			// 解决 x509: certificate signed by unknown authority
			// 通过设置tls.Config的InsecureSkipVerify为true，client将不再对服务端的证书进行校验。
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			// 下面与源码一致
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			ForceAttemptHTTP2:     true,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,

			//DialTLSContext:  nil,
			//DisableKeepAlives:  false,
			//DisableCompression: false,
			//ResponseHeaderTimeout:  0,
			//TLSNextProto:           nil,
			//ProxyConnectHeader:     nil,
			//GetProxyConnectHeader:  nil,
			//MaxResponseHeaderBytes: 0,
			//WriteBufferSize:        0,
			//ReadBufferSize:         0,
		},
		Timeout: time.Duration(rc.TimeOut) * time.Second, // 请求超时时间
	}
	return rc.httpConnPoll
}

// getUrlPathWithParams
/**
 *  @Description: 获取完整的URL
 *  @receiver rc
 *  @param urlPath 基础的Url地址
 *  @param paramMap url参数列表
 *  @return urlPathNew 拼接后的url
 *  @return err 错误
 */
func (rc *RequestConfig) getUrlPathWithParams(urlPath string, paramMap map[string]string) (urlPathNew string, err error) {
	params := url.Values{}
	var parseURL *url.URL
	parseURL, err = url.Parse(urlPath)
	if err != nil {
		return "", err
	}
	if paramMap != nil {
		for k, v := range paramMap {
			params.Set(k, v)
		}
	}
	// 如果参数中有中文参数,这个方法会进行URLEncode
	parseURL.RawQuery = params.Encode()
	urlPathNew = parseURL.String()
	return urlPathNew, nil
}

// Request
/**
 *  @Description: 创建请求
 *  @receiver rc
 *  @param url
 *  @param method
 *  @param data
 */
func (rc *RequestConfig) Request(urlPath, method string, data interface{}, rcd *clientmod.HttpRequestRcd, opts ...RequestClientOptions) (hc *HttpClient) {
	// 接收可选参数
	pr := new(RequestClientPr)
	for _, opt := range opts {
		opt.setRequestClientOption(pr)
	}
	// 请求记录
	if rcd == nil {
		rcd = &clientmod.HttpRequestRcd{}
	}
	var tableName string
	if pr.TableName != "" {
		tableName = pr.TableName
	} else {
		tableName = rcd.TableName()
	}
	defer func() {
		hc.saveRcd()
	}()
	// 初始化封装之后的httpClient
	hc = &HttpClient{
		RequestConfig: *rc,
		rcd:           rcd,
		tableName:     tableName,
	}

	var jsonByte []byte
	rcd.BaseUrl = urlPath
	// 拼接url参数
	if pr.QueryMap != nil {
		urlPath, hc.Error = rc.getUrlPathWithParams(urlPath, pr.QueryMap)
		if hc.Error != nil {
			return
		}
		jsonByte, hc.Error = json.Marshal(pr.QueryMap)
		if hc.Error != nil {
			return hc
		}
		rcd.Query = string(jsonByte)
	}
	rcd.Path = urlPath
	log.Println("请求：" + rcd.Path)
	var body io.Reader
	switch data.(type) {
	case string:
		body = strings.NewReader(data.(string))
		rcd.Body = data.(string)
	default:
		jsonByte, hc.Error = json.Marshal(&data)
		if hc.Error != nil {
			return
		}
		body = bytes.NewBuffer(jsonByte)
		rcd.Body = string(jsonByte)
	}

	switch method {
	case "GET":
		// 发送Get请求
		hc.req, hc.Error = http.NewRequest("GET", urlPath, nil)
	case "POST":
		// 发送Get请求
		hc.req, hc.Error = http.NewRequest("POST", urlPath, body)
	case "PUT":
		// 发送Get请求
		hc.req, hc.Error = http.NewRequest("PUT", urlPath, body)
	case "DELETE":
		// 发送Get请求
		hc.req, hc.Error = http.NewRequest("DELETE", urlPath, body)
	case "PATCH":
		// 发送Get请求
		hc.req, hc.Error = http.NewRequest("PATCH", urlPath, body)
	default:
		hc.Error = errors.New("暂不支持的请求")
		return
	}
	if hc.Error != nil {
		return
	}
	rcd.Method = method

	// 添加请求头
	if pr.HeaderMap != nil {
		for k, v := range pr.HeaderMap {
			hc.req.Header.Add(k, v)
		}
	}
	jsonByte, hc.Error = json.Marshal(pr.HeaderMap)
	if hc.Error != nil {
		return hc
	}

	// 设置Content-Type
	if pr.ContentType != "" {
		hc.req.Header.Add("Content-Type", pr.ContentType)
		rcd.ContentType = pr.ContentType
	}
	rcd.RequestHeader = string(jsonByte)

	// 添加cookie
	if pr.Cookies != nil {
		for _, cookie := range pr.Cookies {
			hc.req.AddCookie(cookie)
		}
	}
	return hc
}

// do
/**
 *  @Description: 带重试的请求
 *  @receiver hc
 *  @param req
 *  @return response
 *  @return err
 */
func (hc *HttpClient) do(req *http.Request) (response *http.Response, err error) {
	if hc.retry <= 1 {
		hc.retry = 1
	}
	for hc.retryCount = 1; hc.retryCount <= hc.retry; hc.retryCount++ {
		if hc.retryCount != 1 {
			hc.rcd.Remark = "第" + strconv.Itoa(hc.retryCount-1) + "请求重试"
			log.Println(hc.rcd.Remark)
		}
		response, err = hc.request(req)
		if err != nil {
			if hc.retryCount != hc.retry {
				hc.rcd.ID = ""
				hc.rcd.CreateTime = time.Now().Format("2006-01-02 15:04:05")
				time.Sleep(time.Duration(hc.retryInterval[hc.retryCount-1]) * time.Second)
			}
			continue
		} else {
			break
		}
	}
	return
}

// request
/**
 *  @Description: 请求
 *  @receiver hc
 *  @param req
 *  @return response
 *  @return err
 */
func (hc *HttpClient) request(req *http.Request) (response *http.Response, err error) {
	defer func() {
		hc.saveRcd()
	}()
	// 这里创建了一个设置了超时时间的context
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(hc.TimeOut)*time.Second)
	defer cancel()

	now := time.Now()
	hc.rcd.Timestamp = now.Format("2006-01-02 15:04:05")
	// 将有超时的context传递给client.do
	hc.res, hc.Error = hc.RequestConfig.httpConnPoll.Do(req.WithContext(ctx))
	response, err = hc.res, hc.Error
	if hc.Error != nil {
		select {
		// select 等待超时返回结果，当http请求时间超出我们设定的时间时，context就会中断请求
		case <-ctx.Done():
			hc.rcd.Latency = time.Now().Sub(now).Milliseconds()
			hc.Error = ctx.Err()
		}
	}
	return
}

// Retry
/**
 *  @Description: 重试
 *  @receiver hc
 *  @param retry 重试次数
 *  @param retryInterval 每次重试之间的间隔（如：[2,2,2,2,2]、[1,2,3,4,5]）
 *  @return *HttpClient
 */
func (hc *HttpClient) Retry(retry int, retryInterval []int) *HttpClient {
	if retry <= 0 {
		hc.retry = 1
	} else {
		hc.retry = retry
	}
	if len(hc.retryInterval) < retry-1 {
		hc.Error = errors.New("重试间隔时间不完整")
	}
	hc.retryInterval = retryInterval
	return hc
}

// GetBytes
/**
 *  @Description: 获取Bytes
 *  @receiver hc
 *  @param body
 *  @return *HttpClient
 */
func (hc *HttpClient) GetBytes(body *[]byte) *HttpClient {
	defer func() {
		hc.saveRcd()
	}()
	now := time.Now()
	hc.rcd.Timestamp = now.Format("2006-01-02 15:04:05")
	// 发起Http请求
	hc.res, hc.Error = hc.do(hc.req)
	if hc.Error != nil {
		return hc
	}
	hc.rcd.HttpVersion = hc.res.Proto
	hc.rcd.Latency = time.Now().Sub(now).Milliseconds()
	hc.rcd.HttpStatus = strconv.Itoa(hc.res.StatusCode)
	var jsonByte []byte
	jsonByte, hc.Error = json.Marshal(hc.res.Header)
	if hc.Error != nil {
		return hc
	}
	hc.rcd.ResponseHeader = string(jsonByte)
	// 这步是必要的，防止以后的内存泄漏，切记
	defer hc.res.Body.Close()
	*body, hc.Error = ioutil.ReadAll(hc.res.Body)
	if hc.Error != nil {
		return hc
	}
	hc.rcd.ResponseData = *(*string)(unsafe.Pointer(body))
	return hc
}

// saveRcd
/**
 *  @Description: 保存请求
 *  @receiver hc
 *  @return *HttpClient
 */
func (hc *HttpClient) saveRcd() *HttpClient {
	if hc.RequestConfig.DB == nil {
		hc.Error = errors.New("数据库链接不存在")
	}
	var dbErr error
	// 只对表创建一次
	hc.once.Do(func() {
		// 没有表,创建表
		dbErr = hc.DB.Table(hc.tableName).AutoMigrate(&clientmod.HttpRequestRcd{})
	})
	if dbErr != nil {
		hc.Error = errors.New("创建表失败:" + dbErr.Error())
		return hc
	}

	if hc.rcd.ID == "" {
		hc.rcd.ID = commons.UUID()
		hc.rcd.CreateTime = time.Now().Format("2006-01-02 15:04:05")
		hc.rcd.UpdateTime = time.Now().Format("2006-01-02 15:04:05")
		hc.Error = hc.DB.Table(hc.tableName).Create(&hc.rcd).Error
		if hc.Error != nil {
			log.Println("更新数据失败", hc.Error)
			return hc
		}
	} else {
		hc.rcd.UpdateTime = time.Now().Format("2006-01-02 15:04:05")
		hc.Error = hc.DB.Table(hc.tableName).Updates(&hc.rcd).Error
		if hc.Error != nil {
			log.Println("更新数据失败", hc.Error)
			return hc
		}
	}
	return hc
}

// SaveReturnBizData
/**
 *  @Description: 保存业务
 *  @receiver hc
 *  @param result
 */
func (hc *HttpClient) SaveReturnBizData(result interface{}) {
	var jsonByte []byte
	jsonByte, hc.Error = json.Marshal(&result)
	if hc.Error != nil {
		return
	}
	hc.rcd.ReturnBizData = string(jsonByte)
	hc.Error = hc.saveRcd().Error
	return
}

// GetMap
/**
 *  @Description: 获取Map
 *  @receiver hc
 *  @param m
 *  @return *HttpClient
 */
func (hc *HttpClient) GetMap(m map[string]interface{}) *HttpClient {
	var body []byte
	hc.Error = hc.GetBytes(&body).Error
	if hc.Error != nil {
		return hc
	}
	hc.Error = json.Unmarshal(body, &m)
	return hc
}

// GetStruct
/**
 *  @Description: 将结果保存到结构体
 *  @receiver hc
 *  @param data
 *  @return *HttpClient
 */
func (hc *HttpClient) GetStruct(data interface{}) *HttpClient {
	var body []byte
	hc.Error = hc.GetBytes(&body).Error
	if hc.Error != nil {
		return hc
	}
	hc.Error = json.Unmarshal(body, &data)
	return hc
}

// GetString
/**
 *  @Description: 获取字符串
 *  @receiver hc
 *  @param str
 *  @return *HttpClient
 */
func (hc *HttpClient) GetString(str *string) *HttpClient {
	var body []byte
	hc.Error = hc.GetBytes(&body).Error
	if hc.Error != nil {
		return hc
	}
	// byte数组直接转成string，优化内存
	*str = *(*string)(unsafe.Pointer(&body))
	return hc
}

// DownloadFile
/**
 *  @Description: 下载文件
 *  @receiver hc
 *  @param filePath 文件保存地址
 *  @param fileName 文件名
 *  @return *HttpClient
 */
func (hc *HttpClient) DownloadFile(filePath, fileName string) *HttpClient {
	defer func() {
		hc.saveRcd()
	}()
	now := time.Now()
	hc.rcd.Timestamp = now.Format("2006-01-02 15:04:05")
	// 发起Http请求
	hc.res, hc.Error = hc.do(hc.req)
	if hc.Error != nil {
		return hc
	}
	hc.rcd.HttpVersion = hc.res.Proto
	hc.rcd.Latency = time.Now().Sub(now).Milliseconds()
	hc.rcd.HttpStatus = strconv.Itoa(hc.res.StatusCode)
	var jsonByte []byte
	jsonByte, hc.Error = json.Marshal(hc.res.Header)
	if hc.Error != nil {
		return hc
	}
	hc.rcd.ResponseHeader = string(jsonByte)

	// 这步是必要的，防止以后的内存泄漏，切记
	defer hc.res.Body.Close()
	savePath := path.Join(filePath, fileName)
	// 创建一个文件用于保存
	var out *os.File
	out, hc.Error = os.Create(savePath)
	if hc.Error != nil {
		return hc
	}
	defer out.Close()
	// 然后将响应流和文件流对接起来
	_, hc.Error = io.Copy(out, hc.res.Body)
	return hc
}
