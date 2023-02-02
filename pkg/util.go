package pkg

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/tidwall/gjson"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

type TradeMethod struct {
	Identifier string
}

type Advertiser struct {
	UserNo   string
	NickName string
	Email    string
}

func parseAdvertiser(g gjson.Result) Advertiser {
	return Advertiser{
		UserNo:   g.Get("userNo").Str,
		NickName: g.Get("nickName").Str,
		Email:    g.Get("email").Str,
	}
}

type Adv struct {
	AdvNo                       string
	Price                       float64
	IsTradable                  bool
	Asset                       string
	FiatUnit                    string
	TradeType                   string
	TradeMethods                []TradeMethod
	DynamicMaxSingleTransAmount float64
	MaxSingleTransAmount        float64
	MinSingleTransAmount        float64
	AdvStatus                   string
}

func parseAdv(g gjson.Result) Adv {
	r := Adv{
		AdvNo:                       g.Get("advNo").Str,
		Asset:                       g.Get("asset").Str,
		FiatUnit:                    g.Get("fiatUnit").Str,
		Price:                       g.Get("price").Float(),
		IsTradable:                  g.Get("isTradable").Bool(),
		DynamicMaxSingleTransAmount: g.Get("dynamicMaxSingleTransAmount").Float(),
		MinSingleTransAmount:        g.Get("minSingleTransAmount").Float(),
		MaxSingleTransAmount:        g.Get("maxSingleTransAmount").Float(),
		TradeType:                   g.Get("tradeType").Str,
		TradeMethods:                []TradeMethod{},
	}

	for _, v := range g.Get("tradeMethods").Array() {
		r.TradeMethods = append(r.TradeMethods, TradeMethod{
			Identifier: v.Get("identifier").Str,
		})
		v.Float()
	}
	return r
}

type Adver struct {
	Adv        Adv
	Advertiser Advertiser
}

func parseAdvStr(g gjson.Result) Adver {
	return Adver{Adv: parseAdv(g.Get("adv")), Advertiser: parseAdvertiser(g.Get("advertiser"))}
}

func SearchAdv(tradeType, asset, fiat string, page, rows, transAmount int, payTypes []string) ([]Adver, error) {
	//出售 USDT {"proMerchantAds":false,"page":1,"rows":10,"payTypes":[],"countries":[],"publisherType":null,"tradeType":"SELL","asset":"USDT","fiat":"CNY"}
	//购买 USDT {"proMerchantAds":false,"page":1,"rows":10,"payTypes":[],"countries":[],"publisherType":null,"tradeType":"BUY","asset":"USDT","fiat":"CNY"}

	urlPath := "https://p2p.binance.com/bapi/c2c/v2/friendly/c2c/adv/search"
	data := make(map[string]interface{})
	data["proMerchantAds"] = false
	data["page"] = page
	data["rows"] = rows
	data["payTypes"] = payTypes
	data["countries"] = []int{}
	data["publisherType"] = nil
	data["asset"] = asset
	data["fiat"] = fiat
	data["tradeType"] = tradeType
	data["transAmount"] = transAmount

	dataStr, _ := json.Marshal(data)

	req := NewRequest("POST", urlPath, dataStr)
	resp, err := (&http.Client{Timeout: 60 * time.Second}).Do(req)
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	resp.Body.Close()
	if resp.StatusCode == 200 {
		result := gjson.Parse(string(body))
		switch result.Get("code").Str {
		case "000000":
			if result.Get("success").Bool() == false {
				return nil, errors.New(fmt.Sprintf("StatusCode: [%v], Code: %s, message: %s", resp.StatusCode, result.Get("code").Str, result.Get("message").Str))
			} else {
				var advSlice = make([]Adver, 0)
				for _, v := range result.Get("data").Array() {
					g := parseAdvStr(v)
					advSlice = append(advSlice, g)
				}
				return advSlice, nil
			}
		default:
			return nil, errors.New(fmt.Sprintf("StatusCode: [%v], Code: %s, message: %s", resp.StatusCode, result.Get("code").Str, result.Get("message").Str))
		}
	} else {
		return nil, errors.New(fmt.Sprintf("StatusCode: [%v], Body: %s", resp.StatusCode, resp.Body))
	}
}

func NewRequest(method, url string, dataStr []byte) *http.Request {
	var body io.Reader = nil
	if dataStr != nil {
		body = bytes.NewReader(dataStr)
	}
	req, _ := http.NewRequest(method, url, body)
	req.Header.Set("Host", "binance.com")
	req.Header.Add("Origin", "https://p2p.binance.com")
	req.Header.Set("Content-type", "application/json;charset=UTF-8")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36 x-trace-id: 5605aa33-eee1-4ce2-aa07-ca2f9d3d47c5")

	return req
}

func PushSuccess(msg string, barkId string) error {
	urlPath := fmt.Sprintf("https://api.day.app/%s/%s?sound=minuet", barkId, msg)
	resp, err := http.Get(urlPath)
	if err != nil {
		return err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		return nil
	} else {
		return errors.New(fmt.Sprintf("[%v] %s", resp.StatusCode, body))
	}
}
