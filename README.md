# mugglepay

#### 安装

```bash
go get github.com/du5/mugglepay@v1.0.0
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
        PriceCurrency:   "CNY",
        Title:           "订单标题",
        Description:     "订单描述",
    })
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

