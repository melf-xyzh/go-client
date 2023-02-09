# go-client
Go-HttpClient的封装实现

### 功能

- [x] Get请求
- [x] Post请求
- [ ] 下载文件
  - [x] 普通下载
  - [ ] 断点续传
- [ ] 上传文件
  - [ ] 普通上传
  - [ ] 分片上传
- [x] 保存请求记录
  - [x] 分表

- [x] 超时设置
- [x] 请求重试

## 安装

```bash
go get -u github.com/melf-xyzh/go-client
```

### 快速入门

```go
type Res struct {
	Success bool `json:"success"`
	Data    struct {
		Id        int    `json:"id"`
		Name      string `json:"name"`
		Auther    string `json:"auther"`
		PicUrl    string `json:"picUrl"`
		Mp3Url    string `json:"mp3url"`
		AvatarUrl string `json:"avatarUrl"`
		Content   string `json:"content"`
	} `json:"data"`
}
```

#### 示例

```go
package main

import (
	"fmt"
	"github.com/melf-xyzh/go-client/http"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
    // 数据库链接信息
	dsn := "root:123456789@tcp(127.0.0.1:3306)/db1?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	//var rc httpclient.RequestConfig
	rc := httpclient.RequestConfig{
		TimeOut: 15,
		DB:      db,
	}
	rc.NewHttpClient()

	// 自定义表名
	tableName := httpclient.TableName("request_rcd")
	// 直接获取[]byte
	var data []byte
	rc.Request("https://api.vvhan.com/api/reping", "GET", nil, nil, tableName).GetBytes(&data)
	fmt.Println(string(data))

	// 将结果存入Map
	m := make(map[string]interface{})
	rc.Request("https://api.vvhan.com/api/reping", "GET", nil, nil, tableName).GetMap(m)
	for k, v := range m {
		fmt.Println(k, v)
	}

	// 将结果保存为字符串
	var res string
	rc.Request("https://api.vvhan.com/api/reping", "GET", nil, nil, tableName).GetString(&res)
	fmt.Println(res)

	// 将结果保存为结构体
	var resStruct Res
	rc.Request("https://api.vvhan.com/api/reping", "GET", nil, nil, tableName).GetStruct(&resStruct)
	fmt.Println(resStruct)
}
```

#### 带参数的Get请求

```go
// 将Query参数存入Map
query := httpclient.QueryMap{
    "format":"text",
}
// 自定义表名
tableName := httpclient.TableName("music")
// 按月分表
splitTable := httpclient.SplitTable(true)

// 定于用于接收返回结果的变量
var res string
// 请求
err = rc.Request("https://api.uomg.com/api/comments.163", "GET", nil,nil,query,tableName).GetString(&res).Error
if err != nil {
    panic(err)
}
fmt.Println(res)
```

#### 下载文件

```go
// 自定义表名
tableName := httpclient.TableName("download")
err = rc.Request("http://127.0.0.1:9001/downloadTemplate", "GET", nil,nil,tableName).DownloadFile("./","1.xlsx").Error
if err != nil {
    panic(err)
}
```

#### 重试

```go
// 直接获取[]byte
var data []byte
rc.Request("https://api.vvhan.com/api/reping", "GET", nil, nil, tableName).Retry(10, []int{1, 2, 3, 4, 5, 6, 7, 8, 9}).GetBytes(&data)
```

