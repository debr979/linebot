package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/tidwall/gjson"
)

const (
	channelSecret = ""
	token         = ""
	IMGUR_TOKEN   = ""
	BITLY_TOKEN   = ""
)

var ch = make(chan string)

func init() {
	log.Print("linebot starting")
}
func main() {

	bot, err := linebot.New(
		channelSecret,
		token,
	)
	if err != nil {
		log.Print(err)
	}
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	botGroup := r.Group("/linebot")
	botGroup.POST("/callback", func(c *gin.Context) {
		events, err := bot.ParseRequest(c.Request)
		if err != nil {
			if err == linebot.ErrInvalidSignature {
				log.Print(err)
			}
			return
		}

		for _, event := range events {
			//事件判斷
			switch event.Type {

			case linebot.EventTypeMessage:
				//文字判斷
				switch message := event.Message.(type) {
				case *linebot.TextMessage:
					if strings.Contains(message.Text, ":") {
						strProc := strings.Split(message.Text, ":")
						if strProc[1] != "" {
							switch strings.ToLower(strProc[0]) {
							case "wiki":
								replyMsg := linebot.NewTextMessage("https://zh.wikipedia.org/wiki/" + strings.TrimSpace(strProc[1]))
								if _, err = bot.ReplyMessage(event.ReplyToken, replyMsg).Do(); err != nil {
									log.Print(err)
								}
								break
							case "url":
								shortUrl := urlToShort(strings.TrimSpace(strProc[1]) + ":" + strings.TrimSpace(strProc[2]))
								replyMsg := linebot.NewTextMessage(shortUrl)
								if _, err = bot.ReplyMessage(event.ReplyToken, replyMsg).Do(); err != nil {
									log.Print(err)
								}
								break
							case "stock":
								selectType := strings.TrimSpace(strProc[1])
								now := time.Now().Format("20060102")
								if len(strProc) > 2 {
									if strProc[2] != "" {
										now = strProc[2]
									}
								}
								result := getStockFundDaily(now, selectType)
								replyMsg := linebot.NewTextMessage(result)
								if _, err := bot.ReplyMessage(event.ReplyToken, replyMsg).Do(); err != nil {
									log.Println(err)
								}
								break
							}
						}
					} else {
						//cmds
						cmd := strings.ToLower(message.Text)
						switch cmd {
						case "uid":
							replyMsg := linebot.NewTextMessage("UID:" + event.Source.UserID)
							if _, err = bot.ReplyMessage(event.ReplyToken, replyMsg).Do(); err != nil {
								log.Print(err)
							}
							break
						case "gid":
							replyMsg := linebot.NewTextMessage("GID:" + event.Source.GroupID)
							if _, err = bot.ReplyMessage(event.ReplyToken, replyMsg).Do(); err != nil {
								log.Print(err)
							}
							break
						case "stock":
							tip := "請輸入stock:[類股代號]:[日期(選填)]\nex. stock:01:20210526 日期係影響個股每日籌碼。\n\n類股名稱：\n[01]水泥、[02]食品、\n[03]塑膠、[04]紡織、\n[05]電機、[06]電器、\n[07]化學生技醫療、[08]玻璃、\n[09]造紙、[10]鋼鐵、\n[11]橡膠、[12]汽車、\n[13]電子工業、[14]營造、\n[15]運輸、[16]觀光\n[17]金融、[18]百貨、\n[19]綜合、[20]其他、\n[21]化學、[22]生技、\n[23]油電燃氣、[24]半導體、\n[25]電腦週邊、[26]光電、\n[27]通信網路、[28]電子零件、\n[29]電子通路、[30]資訊服務、\n[31]其他電子、[0099P]ETF"
							replyMsg := linebot.NewTextMessage(tip)
							if _, err := bot.ReplyMessage(event.ReplyToken, replyMsg).Do(); err != nil {
								log.Println(err)
							}
							break
						case "hi", "help", "Doug", "doug", "123":
							description := "指令：[stock, url, img]\n 1、詳細說明請直接輸入stock\n2、將網址縮短輸入url:[網址]，例如:url:www.yahoo.com.tw\n3、img：將圖片上傳，轉為網址，不需輸入任何指令"
							replyMsg := linebot.NewTextMessage(description)
							if _, err := bot.ReplyMessage(event.ReplyToken, replyMsg).Do(); err != nil {
								log.Println(err)
							}
							break
						}
					}
					//圖片判斷
				case *linebot.ImageMessage:
					if event.Source.GroupID == "" {
						image, err := bot.GetMessageContent(message.ID).Do()
						if err != nil {
							log.Print(err)
						}
						link := upload(image.Content, IMGUR_TOKEN)
						replyMsg := linebot.NewTextMessage(link)
						if _, err = bot.ReplyMessage(event.ReplyToken, replyMsg).Do(); err != nil {
							log.Print(err)
						}
					}
				}
				//事件:被加入
			case linebot.EventTypeJoin:
				leftBtn := linebot.NewMessageAction("取消", "cancel")
				rightBtn := linebot.NewMessageAction("確定", "help")
				template := linebot.NewConfirmTemplate("查詢指令", leftBtn, rightBtn)
				replyMsg := linebot.NewTemplateMessage("查詢指令", template)
				if _, err = bot.ReplyMessage(event.ReplyToken, replyMsg).Do(); err != nil {
					log.Print(err)
				}
			}
			
		}
	})
	if err := r.Run(":8081"); err != nil {
		log.Print(err)
	}
}

func upload(image io.Reader, token string) (link string) {
	APIURL := "https://api.imgur.com/3/image"
	var buf = new(bytes.Buffer)
	writer := multipart.NewWriter(buf)
	part, _ := writer.CreateFormFile("image", "dont care about name")
	io.Copy(part, image)
	writer.Close()
	req, _ := http.NewRequest("POST", APIURL, buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+token)
	client := &http.Client{}
	res, _ := client.Do(req)
	defer res.Body.Close()
	b, _ := ioutil.ReadAll(res.Body)
	success := gjson.Get(string(b), "success")
	if strings.Contains(success.Raw, "true") {
		link = gjson.Get(string(b), "data.link").Str
	} else {
		link = ""
	}
	return link
}

func urlToShort(longURL string) string {
	var response struct {
		Link string `json:"link"`
	}
	APIURL := "https://api-ssl.bitly.com/v4/bitlinks"
	postData := strings.NewReader(`{"long_url":"` + longURL + `"}`)

	log.Print(postData)
	req, err := http.NewRequest("POST", APIURL, postData)
	if err != nil {
		log.Print(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+BITLY_TOKEN)
	res, _ := http.DefaultClient.Do(req)
	defer func() {
		if err := res.Body.Close(); err != nil {
			log.Println(err)
		}
	}()

	b, _ := ioutil.ReadAll(res.Body)
	if err := json.Unmarshal(b, &response); err != nil {
		log.Println(err)
	}
	return response.Link
}

func getStockFundDaily(date string, selectType string) string {
	var result struct {
		Data struct {
			StockFundDailyList []struct {
				StockNum        string `json:"stock_num"`
				StockName       string `json:"stock_name"`
				Foreign         int    `json:"foreign"`
				InvestmentTrust int    `json:"investment_trust"`
				Dealer          int    `json:"dealer"`
				Total           int    `json:"total"`
			} `json:"stock_fund_daily_list"`
		} `json:"data"`
	}
	url := "https://sg.manager-shop.xyz/v1/stock/getStockFundDaily"
	data := fmt.Sprintf(`{"date":"%s","select_type":"%s"}`, date, selectType)
	payload := strings.NewReader(data)
	req, _ := http.NewRequest("POST", url, payload)
	req.Header.Add("Content-Type", "application/json")
	defer func() {
		_ = req.Body.Close()
	}()
	res, _ := http.DefaultClient.Do(req)
	body, _ := ioutil.ReadAll(res.Body)

	err := json.Unmarshal(body, &result)
	if err != nil {
		log.Println(err)
	}

	response := "[股號]名稱\n外資｜投信｜自營｜總和\n"

	for _, item := range result.Data.StockFundDailyList {
		data := fmt.Sprintf("[%s]%s\n外資：%d,投信：%d\n自營：%d,總和：%d\n", item.StockNum, item.StockName, item.Foreign, item.InvestmentTrust, item.Dealer, item.Total)

		response += data
	}

	return response
}
