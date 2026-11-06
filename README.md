# wopan-sdk-go
 Wopan SDK for the Go programming language

## Installation

```bash
go get github.com/xhofe/wopan-sdk-go
```

## Usage

```go
package main

import (
	"fmt"
	"github.com/xhofe/wopan-sdk-go"
)

func main() {
	w := wopan.DefaultWithRefreshToken("91d4b946-xxxx-4909-bac1-d9914e45f2de")
	res, err := w.AppQueryUser()
	if err != nil {
		fmt.Printf("AppQueryUser() error = %v", err)
	} else {
		fmt.Printf("AppQueryUser() = %+v", res)
	}
}

```

## Login by SMS code（纯短信验证码登录，免密码）

```go
phone := "138xxxx5678"

// 1. 发送短信验证码（如返回 NeedCaptcha=true，按提示处理图形验证码后重发）
res, err := w.SendMessageCode(phone)
if err != nil {
	panic(err)
}
if res.NeedCaptcha {
	uuid := "随机串"
	fmt.Println("请查看图片并输入图形验证码:", wopan.GetVerifyCodeURL(uuid))
	var captcha string
	fmt.Scanln(&captcha)
	res, err = w.SendMessageCodeWithCaptcha(phone, uuid, captcha)
	if err != nil {
		panic(err)
	}
}

// 2. 输入手机收到的短信码，换取 token（自动装入客户端）
var smsCode string
fmt.Scanln(&smsCode)
login, err := w.LoginBySmsCode(phone, smsCode)
if err != nil {
	panic(err)
}
fmt.Printf("access_token=%s refresh_token=%s expires_in=%d\n",
	login.AccessToken, login.RefreshToken, login.ExpiresIn)

// 3. 正常使用
user, err := w.AppQueryUser()
```