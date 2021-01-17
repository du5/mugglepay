# mugglepay

#### 安装

```bash
go get github.com/du5/mugglepay@v1.0.6
```

#### 引用


```go
import "github.com/du5/mugglepay"
```

#### 创建订单
```go
func CreateOrder(c *gin.Context) {
    mgp := mugglepay.NewMugglepay("BitpayxApplicationKey")
    // host := "https://www.example.com"
    // 如需法币支付则必须设置正确的回调地址
    // mgp.CallBackUrl = host + "/payment/notify"
    // mgp.CancelUrl = host + "/user/code/return?merchantTradeNo="
    // mgp.SuccessUrl = host + "/user/code/return?merchantTradeNo="
    serverOrder, _ := mgp.CreateOrder(&mugglepay.Order{
		MerchantOrderId: orderId,
		PriceAmount:     money,
		// PriceCurrency:   "USD",
		// PayCurrency:     "ALIPAY",
		// PayCurrency:     "WECHAT",
		PayCurrency:     "",
		PriceCurrency:   "CNY",
		Title:           "订单标题",
		Description:     "订单描述",
    })
    // 支付宝/微信扫码链接，该函数仅 PayCurrency 为 ALIPAY/WECHAT 时可返回地址
    // 其他情况下均返回加密货币地址
    // aliqr := sorder.Invoice.GetAlipayUrl()
    c.Redirect(http.StatusFound, serverOrder.PaymentUrl)
}
```

#### 支付回调校验

```go
func Notify(c *gin.Context) {
	body, _ := c.GetRawData()
	var callback mugglepay.Callback
	if err := json.Unmarshal(body, &callback); err == nil {
        mgp := mugglepay.NewMugglepay("BitpayxApplicationKey")
        if mgp.VerifyOrder(&callback) {
            // code ... 
            c.JSON(200, gin.H{"status": 200})
            return
        }
    }
    c.JSON(200, gin.H{"status": 400})
}
```


#### 修改支付方式

```go
mgp := mugglepay.NewMugglepay("BitpayxApplicationKey")
sorder, _ := mgp.CheckOut(ServerOrderId, "P2P_BTC")
// 应付金额
money := sorder.Invoice.PayAmount
// 法币支付链接
// aliqr := sorder.Invoice.GetAlipayUrl()
// 虚拟货币交易地址
address := sorder.Invoice.Address
// 虚拟货币交易备注
memo := sorder.Invoice.Memo

```