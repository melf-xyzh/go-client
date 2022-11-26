/**
 * @Time    :2022/10/31 9:57
 * @Author  :Xiaoyu.Zhang
 */

package httpclient

import (
	"gorm.io/gorm"
	"net/http"
)

type RequestConfig struct {
	TimeOut      int          // 超时时间
	DB           *gorm.DB     // 数据库对象
	httpConnPoll *http.Client // 连接池
}
