/**
 * @Time    :2022/10/19 14:56
 * @Author  :Xiaoyu.Zhang
 */

package clientmod

type HttpRequestRcd struct {
	ID          string `json:"id"                         gorm:"column:id;primary_key;type:varchar(36)"`
	CreateTime  string `json:"createTime"                 gorm:"column:create_time;type:varchar(19);index;"`
	UpdateTime  string `json:"updateTime,omitempty"       gorm:"column:update_time;type:varchar(19);"`
	Timestamp   string `json:"timestamp"                  gorm:"column:timestamp;comment:请求时间;type:varchar(20);"`
	BaseUrl     string `json:"baseUrl"                    gorm:"column:base_url;comment:请求地址;type:varchar(255);"`
	Path        string `json:"path"                       gorm:"column:path;comment:请求路径;type:varchar(255);"`
	Method      string `json:"method"                     gorm:"column:method;comment:请求方法;type:varchar(10);"`
	HttpVersion string `json:"httpVersion"                gorm:"column:http_version;comment:Http版本;type:varchar(10);"`
	Format      string `json:"format"                     gorm:"column:format;comment:数据格式;type:varchar(100);"`
	ContentType string `json:"contentType"                gorm:"column:content_type;comment:contentType;type:varchar(100);"`
	HttpStatus  string `json:"httpStatus"                 gorm:"column:http_status;comment:http请求响应状态;type:varchar(4);"`
	Latency     int64  `json:"latency"                    gorm:"column:latency;comment:延迟"`
	Err         string `json:"err"                        gorm:"column:err;comment:请求报错信息;type:text;"`
	HttpParams
	HttpSign
}

type HttpParams struct {
	Query          string `json:"query"                    gorm:"column:query;comment:query参数;type:text;"`
	Body           string `json:"body"                     gorm:"column:body;comment:body参数;type:text;"`
	BizData        string `json:"bizData"                  gorm:"column:biz_data;comment:请求业务参数;type:text;"`
	RequestHeader  string `json:"requestHeader"            gorm:"column:request_header;comment:请求头;type:text;"`
	ResponseData   string `json:"responseData"             gorm:"column:response_data;comment:响应体;type:text;"`
	ResponseHeader string `json:"responseHeader"           gorm:"column:response_header;comment:响应头;type:text;"`
	ReturnBizData  string `json:"returnBizData"            gorm:"column:return_biz_data;comment:返回业务参数;type:text;"`
}

type HttpSign struct {
	Sign        string `json:"sign"                gorm:"column:sign;comment:数字签名;type:text;"`
	SignContent string `json:"signContent"         gorm:"column:sign_content;comment:被签名内容;type:text;"`
	Key         string `json:"key"                 gorm:"column:key;comment:数字签名关键字;type:varchar(255);"`
}

// TableName 自定义表名
func (HttpRequestRcd) TableName() string {
	return "http_request_rcd"
}
