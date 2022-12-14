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
	TimeOut            int          // 超时时间
	DB                 *gorm.DB     // 数据库对象
	InsecureSkipVerify bool         // InsecureSkipVerify为true，client将不再对服务端的证书进行校验
	httpConnPoll       *http.Client // 连接池
}
