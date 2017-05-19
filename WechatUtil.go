package main

import (
	"net/http"
	"io"
	"io/ioutil"
	"encoding/json"
	"time"
	"fmt"
	"sync"
	"errors"
)
//设置微信的AppId
const appId  = ""
//设置微信的APPSecret
const appSecret  = ""
//拼接字符串
const url_wechat  = "https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid="+appId+"&secret="+appSecret
//服务器监听端口
const listenPort  = 8088
//超时时间(微信默认7200秒过期 预留出100秒处理网络等)
const timeout  = 7100
const cachekey  = "access_token"

//微信信息结构
type WechatInfo struct {
	timestamp int64
	token string
}
//缓存 key值为accessToken
var wechatCache = make(map[string]WechatInfo,1)

var lock = sync.RWMutex{}

func getAccessTokenFromCacheNoLock(cacheKey string) (token string, err error) {
	data,ok := wechatCache[cachekey]
	if ok && data.timestamp > time.Now().Unix() - timeout{
		return data.token, nil
	}else{
		return "", errors.New("过期或不存在")
	}

}
func getAccessTokenFromCache(cacheKey string) (token string, err error) {
	lock.RLock()
	defer lock.RUnlock()

	data,ok := wechatCache[cachekey]

	if ok && data.timestamp > time.Now().Unix() - timeout{
		return data.token, nil
	}else{
		return "", errors.New("过期或不存在")
	}

}
func putAccessToken(info WechatInfo)  {
	wechatCache[cachekey] = info
}
//从网络中获取Token
func getAccessTokenFromWechat(token chan string)  {
	resp, err := http.Get(url_wechat)
	if err == nil{
		data, err := ioutil.ReadAll(resp.Body)
		defer resp.Body.Close()
		if err == nil {
			var jsonBean = make(map[string]interface{})
			json.Unmarshal(data, &jsonBean)
			data, ok := jsonBean[cachekey]
			tokenValue := data.(string)
			if(ok){
				token <- tokenValue
				putAccessToken(WechatInfo{time.Now().Unix(), tokenValue})
			}else {
				token <- ""
			}
		}else {
			fmt.Println(err)
		}
	}else {
		fmt.Println(err)
	}
}
func main()  {
	http.HandleFunc("/", func(writer http.ResponseWriter,request *http.Request) {
		token,err := getAccessTokenFromCache(cachekey)
		if(err != nil) {
			tokenChan := make(chan string)
			defer close(tokenChan)
			lock.Lock()
			defer lock.Unlock()
			token,err := getAccessTokenFromCacheNoLock(cachekey)
			if(err != nil){
				go getAccessTokenFromWechat(tokenChan)
				tokenNew := <-tokenChan
				io.WriteString(writer, tokenNew)
				fmt.Println("")
			}else {
				io.WriteString(writer, token)
			}
		}else {
			io.WriteString(writer, token)
		}
	})
	http.ListenAndServe(":8088", nil)
}