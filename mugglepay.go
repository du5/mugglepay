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
}

type ServerOrder struct {
	Status     int     `json:"status"`
	Order      Order   `json:"order"`
	PaymentUrl string  `json:"payment_url"`
	Invoice    Invoice `json:"invoice"`
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

func (mgp *Mugglepay) CreateOrder(order *Order) (ServerOrder, error) {
	mgp.ApiUrl = mgp.ApiUrl + "/orders"

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
	order.Sign(mgp.ApplicationKey)

	jsonOrder, _ := json.Marshal(order)

	reqest, _ := http.NewRequest("POST", mgp.ApiUrl, bytes.NewBuffer(jsonOrder))
	reqest.Header.Add("token", mgp.ApplicationKey)
	reqest.Header.Add("content-type", "application/json")
	http_unmarshal(reqest, &sorder)

	return sorder, nil
}

func (order *Order) Sign(secret string) {
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
	order.Sign(mgp.ApplicationKey)
	// 校验签名
	if order.Token != callback.Token {
		return false
	}
	if callback.Status == "PAID" {
		return true
	}
	return false
}

func (mgp *Mugglepay) GetOrder(OrderId string) (ServerOrder, error) {
	var sorder ServerOrder
	if OrderId == "" {
		return sorder, errors.New("order id cannot be null")
	}
	mgp.ApiUrl = mgp.ApiUrl + "/orders/" + OrderId
	reqest, _ := http.NewRequest("GET", mgp.ApiUrl, nil)
	reqest.Header.Add("token", mgp.ApplicationKey)
	http_unmarshal(reqest, &sorder)
	return sorder, nil
}

func http_unmarshal(reqest *http.Request, sorder *ServerOrder) {
	client := &http.Client{}
	response, _ := client.Do(reqest)
	responseB := response.Body
	defer responseB.Close()
	body, _ := ioutil.ReadAll(responseB)
	bytes := []byte(body)
	_ = json.Unmarshal(bytes, &sorder)
}
