package main

import (
	"crypto/sha1"
	"encoding/xml"
	"fmt"
	"github.com/bitly/go-simplejson"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"
)

const (
	TOKEN  = "xiaofei"
	APIKEY = "9aa081ff30e207f96c46dfdd21a7b65f"
	//TUURL  = "http://127.0.0.1:8989/postpage"
	TUURL = "http://www.tuling123.com/openapi/api"
	//-------menukey自定义菜单按键值
	MENUKEY_1 = "123"
	MENUKEY_2 = ""
	MENUKEY_3 = ""
)

type Message struct {
	XMLName      xml.Name `xml:"xml"`
	ToUserName   string   `xml:",omitempty"`
	FromUserName string   `xml:",omitempty"`
	CreateTime   int64    `xml:",omitempty"`
	MsgType      string   `xml:",omitempty"`
	Content      string   `xml:",omitempty"`
	//-----------------accpet---------------
	Event        string  `xml:",omitempty"`
	EventKey     string  `xml:",omitempty"`
	PicUrl       string  `xml:",omitempty"`
	MediaId      string  `xml:",omitempty"`
	MsgId        string  `xml:",omitempty"`
	Format       string  `xml:",omitempty"` //语音格式
	Recognition  string  `xml:",omitempty"` //语音识别结果
	ThumbMediaId string  `xml:",omitempty"` //ThumbMediaId 视频媒体id
	Location_X   float64 `xml:",omitempty"` //Location_X地理位置维度
	Location_Y   float64 `xml:",omitempty"` //Location_Y
	Scale        string  `xml:",omitempty"` //地图缩放大小
	Label        string  `xml:",omitempty"` //地理位置信息
	Title        string  `xml:",omitempty"` //消息标题
	Description  string  `xml:",omitempty"` //消息Description
	Url          string  `xml:",omitempty"` //消息url
	//-----------------response--------------

}

/**
 * [validtion 验证函数]
 * @param  {[type]} w http.ResponseWriter [description]
 * @param  {[type]} r *http.Request)      (e            bool, s string [description]
 * @return {[type]}   [description]
 */
func validtion(w http.ResponseWriter, r *http.Request) (e bool, s string) {
	sign := r.FormValue("signature")
	timestamp := r.FormValue("timestamp")
	nonce := r.FormValue("nonce")
	echostr := r.FormValue("echostr")
	result := []string{TOKEN, timestamp, nonce}
	sort.Sort(sort.StringSlice(result))

	//拼接字符串
	endstr := ""
	for _, v := range result {
		endstr = endstr + v
	}
	//sha1 hash
	h := sha1.New()
	h.Write([]byte(endstr))
	verify := h.Sum(nil)
	ok := fmt.Sprintf("%x", verify)
	//以字符串形式保存verify，注意其中不掺杂其他换行符之类否则字符串判定时为false
	/*println(ok)
	println(sign)
	print(echostr)*/
	if strings.EqualFold(ok, sign) {
		return true, echostr
	} else {
		return false, ""
	}

}

func (message Message) reText() (response Message) {
	//图灵机器人介入
	info := `{"key":"` + APIKEY + `","info":"` + message.Content + `","userid":"` + message.FromUserName + `"}`
	reader := strings.NewReader(info)
	//字符串读取到内存中
	res, _ := http.Post(TUURL, "application/json", reader)
	defer res.Body.Close()
	result, _ := ioutil.ReadAll(res.Body)
	//log.Println(string(result))
	js, _ := simplejson.NewJson(result)

	switch code, _ := js.Get("code").Int(); code {
	//处理文本消息
	case 100000:
		text, _ := js.Get("text").String()
		response = Message{ToUserName: message.FromUserName,
			FromUserName: message.ToUserName,
			CreateTime:   time.Now().Unix(),
			MsgType:      "text",
			Content:      text}
		return response
	//处理链接消息
	case 200000:
		text, _ := js.Get("text").String()
		url, _ := js.Get("url").String()
		response = Message{ToUserName: message.FromUserName,
			FromUserName: message.ToUserName,
			CreateTime:   time.Now().Unix(),
			MsgType:      "text",
			Content:      text + url}
		return response
	//菜谱及新闻 含有列表
	case 302000, 308000:
		text, _ := js.Get("text").String()
		//list, _ := js.Get("list").Map()

		response = Message{ToUserName: message.FromUserName,
			FromUserName: message.ToUserName,
			CreateTime:   time.Now().Unix(),
			MsgType:      "text",
			Content:      text}
		return response
	default:
		println(code)
	}
	return
}

func (message Message) reEvent() (response Message) {
	switch message.Event {
	case "subscribe":
		response = Message{ToUserName: message.FromUserName,
			FromUserName: message.ToUserName,
			CreateTime:   time.Now().Unix(),
			MsgType:      "text",
			Content:      "欢迎订阅哦"}
	case "unsubscribe":
		response = Message{ToUserName: message.FromUserName,
			FromUserName: message.ToUserName,
			CreateTime:   time.Now().Unix(),
			MsgType:      "text",
			Content:      "亲爱的你回来吧，有什么建议可以告诉我哦，只是你不要走"}
	default:
		if message.EventKey == MENUKEY_1 {

		}

	}
	return response
}
func messageHandler(message Message) (response Message) {
	switch message.MsgType {
	case "text":
		response = message.reText()
		return response
	case "event":
		response = message.reEvent()
		return response
	default:
		response = Message{ToUserName: message.FromUserName,
			FromUserName: message.ToUserName,
			CreateTime:   time.Now().Unix(),
			MsgType:      "text",
			Content:      "not ok"}
		return response
	}

}
func mainHandler(w http.ResponseWriter, r *http.Request) {
	if len(r.FormValue("echostr")) != 0 {
		//验证函数
		ok, s := validtion(w, r)
		println(s)
		println(ok)
		if ok {
			io.WriteString(w, s)
		}
		return
	}
	var message Message
	//执行回复操作
	result, _ := ioutil.ReadAll(r.Body)
	log.Println(string(result))
	if err := xml.Unmarshal(result, &message); err != nil {
		panic(err)
	}
	re := messageHandler(message)
	out, err := xml.MarshalIndent(re, "", "")
	if err != nil {
		log.Printf("error: %v\n", err)
	}
	log.Println(string(out))
	io.WriteString(w, string(out))

}
func main() {
	http.HandleFunc("/", mainHandler)
	log.Fatal(http.ListenAndServe(":80", nil))
}
