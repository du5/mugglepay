package mugglepay

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

func NewMugglepay(key string) *Mugglepay {
	mgp := &Mugglepay{
		ApplicationKey: key,
		ApiUrl:         "https://api.mugglepay.com/v1",
	}
	return mgp
}

type Mugglepay struct {
	ApplicationKey string
	ApiUrl         string
	CallBackUrl    string
	CancelUrl      string
	SuccessUrl     string
}

type Order struct {
	OrderId         string  `json:"order_id"`
	UserId          int64   `json:"user_id"`
	MerchantOrderId string  `json:"merchant_order_id"`
	Title           string  `json:"title"`
	Description     string  `json:"description"`
	CallBackUrl     string  `json:"callback_url"`
	CancelUrl       string  `json:"cancel_url"`
	SuccessUrl      string  `json:"success_url"`
	PriceAmount     float64 `json:"price_amount"`
	PriceCurrency   string  `json:"price_currency"`
	Status          string  `json:"status"`
	Notified        string  `json:"notified"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`
	PayAmount       float64 `json:"pay_amount"`
	PayCurrency     string  `json:"pay_currency"`
	IsSelf          bool    `json:"is_self"`
	Mobile          bool    `json:"mobile"`
	Fast            bool    `json:"fast"`
	Token           string  `json:"token"`
	PaidAt          string  `json:"paid_at"`
	ReceiveCurrency string  `json:"receive_currency"`
}

type Invoice struct {
	InvoiceId       string  `json:"invoice_id"`
	OrderId         string  `json:"order_id"`
	PayAmount       float64 `json:"pay_amount"`
	PayCurrency     string  `json:"pay_currency"`
	Status          string  `json:"status"`
	CreatedAt       string  `json:"created_at"`
	CreatedAtT      int64   `json:"created_at_t"`
	ExpiredAt       string  `json:"expired_at"`
	ExpiredAtT      int64   `json:"expired_at_t"`
	MerchantOrderId string  `json:"merchant_order_id"`
	ReceiveAmount   float64 `json:"receive_amount"`
	ReceiveCurrency string  `json:"receive_currency"`
	Qrcode          string  `json:"qrcode"`
	QrcodeLg        string  `json:"qrcodeLg"`
	Address         string  `json:"address"`
	Memo            string  `json:"memo"`
}

type ServerOrder struct {
	Status     int      `json:"status"`
	Order      Order    `json:"order"`
	Merchant   Merchant `json:"merchant"`
	PaymentUrl string   `json:"payment_url"`
	Invoice    Invoice  `json:"invoice"`
	Permission string   `json:"permission"`
}

type Merchant struct {
	AcceptBtc         bool                   `json:"accept_btc"`
	AcceptUsdt        bool                   `json:"accept_usdt"`
	AcceptBch         bool                   `json:"accept_bch"`
	AcceptEth         bool                   `json:"accept_eth"`
	AcceptEos         bool                   `json:"accept_eos"`
	AcceptLtc         bool                   `json:"accept_ltc"`
	AcceptBnb         bool                   `json:"accept_bnb"`
	AcceptBusd        bool                   `json:"accept_busd"`
	AcceptCusd        bool                   `json:"accept_cusd"`
	AcceptAlipay      bool                   `json:"accept_alipay"`
	AcceptWechat      bool                   `json:"accept_wechat"`
	WalletUserHash    string                 `json:"wallet_user_hash"`
	WalletUserEnabled bool                   `json:"wallet_user_enabled"`
	EmailVerified     bool                   `json:"email_verified"`
	Price             map[string]interface{} `json:"price"`
	Permission        string                 `json:"permission"`
}

type Callback struct {
	MerchantOrderId string  `json:"merchant_order_id"`
	OrderId         string  `json:"order_id"`
	Status          string  `json:"status"`
	PriceAmount     float64 `json:"price_amount"`
	PriceCurrency   string  `json:"price_currency"`
	PayAmount       float64 `json:"pay_amount"`
	PayCurrency     string  `json:"pay_currency"`
	CreatedAt       string  `json:"created_at"`
	CreatedAtT      int64   `json:"created_at_t"`
	Token           string  `json:"token"`
	Meta            Meta    `json:"meta"`
}

type Meta struct {
	Payment     string `json:"payment"`
	TotalMmount string `json:"total_amount"`
	TradeNo     string `json:"trade_no"`
	OutTradeNo  string `json:"out_trade_no"`
}

// 创建订单，返回 ServerOrder
func (mgp *Mugglepay) CreateOrder(order *Order) (ServerOrder, error) {
	var sorder ServerOrder
	if mgp.ApplicationKey == "" {
		return sorder, errors.New("application key cannot be null")
	}
	if order.MerchantOrderId == "" {
		return sorder, errors.New("merchant order id cannot be null")
	}

	if order.PriceCurrency == "" {
		order.PriceCurrency = "CNY"
	}
	if mgp.CallBackUrl == "" {
		// 如果没有回调地址将无法使用法币支付，默认仅可用虚拟币
		order.PayCurrency = ""
	}

	order.CallBackUrl = mgp.CallBackUrl
	if mgp.CancelUrl != "" {
		order.CancelUrl = mgp.CancelUrl + order.MerchantOrderId
	}
	if mgp.SuccessUrl != "" {
		order.SuccessUrl = mgp.SuccessUrl + order.MerchantOrderId
	}

	// 签名
	order.sign(mgp.ApplicationKey)

	jsonOrder, _ := json.Marshal(order)
	reqest, _ := http.NewRequest("POST", fmt.Sprintf("%s/orders", mgp.ApiUrl), bytes.NewBuffer(jsonOrder))
	reqest.Header.Add("content-type", "application/json")
	http_unmarshal(reqest, &sorder, mgp.ApplicationKey)
	http_unmarshal(reqest, &sorder, mgp.ApplicationKey)
	return sorder, nil
}

// 签名
func (order *Order) sign(secret string) {
	q := url.Values{}
	q.Set("merchant_order_id", order.MerchantOrderId)
	q.Set("secret", secret)
	q.Set("type", "FIAT")
	order.Token = strings.ToLower(fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%x", md5.Sum([]byte(q.Encode())))+secret))))
}

// 校验订单 true: 已支付; false: 未支付/取消/欺诈
func (mgp *Mugglepay) VerifyOrder(callback *Callback) bool {
	if mgp.ApplicationKey == "" {
		return false
	}
	order := &Order{MerchantOrderId: callback.MerchantOrderId}
	order.sign(mgp.ApplicationKey)
	// 校验签名
	if order.Token != callback.Token {
		return false
	}
	if callback.Status == "PAID" {
		return true
	}
	return false
}

// 根据网关订单编号获取 ServerOrder
func (mgp *Mugglepay) GetOrder(OrderId string) (ServerOrder, error) {
	var sorder ServerOrder
	if OrderId == "" {
		return sorder, errors.New("order id cannot be null")
	}
	reqest, _ := http.NewRequest("GET", fmt.Sprintf("%s/orders/%s", mgp.ApiUrl, OrderId), nil)
	http_unmarshal(reqest, &sorder, mgp.ApplicationKey)
	return sorder, nil
}

// 构建 CURL 请求
func http_unmarshal(reqest *http.Request, sorder *ServerOrder, key string) {
	reqest.Header.Add("token", key)
	client := &http.Client{}
	response, _ := client.Do(reqest)
	responseB := response.Body
	defer responseB.Close()
	body, _ := ioutil.ReadAll(responseB)
	bytes := []byte(body)
	_ = json.Unmarshal(bytes, &sorder)
}

// 获取支付地址
func (sorder *ServerOrder) GetUrl() {
	getUrl := func(longurl, key string) string {
		var res string
		if u, err := url.Parse(longurl); err == nil {
			if p, err := url.ParseQuery(u.RawQuery); err == nil {
				if val, ok := p[key]; ok {
					res = val[0]
				}
			}
		}
		return res
	}
	switch sorder.Invoice.PayCurrency {
	case "ALIPAY":
		if rurl := getUrl(sorder.Invoice.Qrcode, "url"); rurl != "" {
			sorder.Invoice.Address = rurl
		} else {
			sorder.Invoice.Address = getUrl(sorder.Invoice.QrcodeLg, "mpurl")
		}
	case "WECHAT":
		sorder.Invoice.Address = sorder.Invoice.Qrcode
	case "EOS":
		sorder.Invoice.Address = "mgtestflight"
		sorder.Invoice.Memo = fmt.Sprintf("MP:%s", sorder.Invoice.OrderId)
	}
}

// 切换网关支付方式
func (mgp *Mugglepay) CheckOut(OrderId, PayCurrency string) (ServerOrder, error) {
	var sorder ServerOrder
	if OrderId == "" {
		return sorder, errors.New("order id cannot be null")
	}
	me := make(map[string]string)
	me["pay_currency"] = PayCurrency
	newpatC, _ := json.Marshal(me)
	reqest, _ := http.NewRequest("POST", fmt.Sprintf("%s/orders/%s/checkout", mgp.ApiUrl, OrderId), bytes.NewBuffer(newpatC))
	reqest.Header.Add("content-type", "application/json")
	http_unmarshal(reqest, &sorder, mgp.ApplicationKey)
	return sorder, nil
}

// 订单查询
func (mgp *Mugglepay) GetStatus(OrderId string) (ServerOrder, error) {
	var sorder ServerOrder
	if OrderId == "" {
		return sorder, errors.New("order id cannot be null")
	}
	reqest, _ := http.NewRequest("GET", fmt.Sprintf("%s/orders/%s/status", mgp.ApiUrl, OrderId), nil)
	http_unmarshal(reqest, &sorder, mgp.ApplicationKey)
	return sorder, nil
}

// 虚拟币: 我已支付
func (mgp *Mugglepay) Sent(OrderId string) (ServerOrder, error) {
	var sorder ServerOrder
	if OrderId == "" {
		return sorder, errors.New("order id cannot be null")
	}
	sorder, _ = mgp.GetOrder(OrderId)
	if sorder.Invoice.PayCurrency == "ALIPAY" || sorder.Invoice.PayCurrency == "WECHAT" {
		// 法币不可调用此 API
		return sorder, errors.New("tan 90°")
	}
	nilmap, _ := json.Marshal(make(map[string]interface{}))
	reqest, _ := http.NewRequest("POST", fmt.Sprintf("%s/orders/%s/sent", mgp.ApiUrl, OrderId), bytes.NewBuffer(nilmap))
	http_unmarshal(reqest, &sorder, mgp.ApplicationKey)
	return sorder, nil
}
