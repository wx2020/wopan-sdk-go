package wopan

import (
	"fmt"
)

// PcWebLoginData no encrypt
type PcWebLoginData struct {
	NeedSmsCode string `json:"needSmsCode"`
}

func (w *WoClient) PcWebLogin(phone, password string, opts ...RestyOption) (*PcWebLoginData, error) {
	var resp PcWebLoginData
	_, err := w.RequestApiUser(KeyPcWebLogin, Json{
		"phone":        phone,
		"password":     password,
		"uuid":         "",
		"verifyCode":   "",
		"clientSecret": DefaultClientSecret,
	}, JsonClientIDSecret, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// PcLoginVerifyCodeData no encrypt
type PcLoginVerifyCodeData struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

func (w *WoClient) PcLoginVerifyCode(phone, password, messageCode string, opts ...RestyOption) (*PcLoginVerifyCodeData, error) {
	var resp PcLoginVerifyCodeData
	_, err := w.RequestApiUser(KeyPcLoginVerifyCode, Json{
		"phone":        phone,
		"messageCode":  messageCode,
		"verifyCode":   nil,
		"uuid":         nil,
		"clientSecret": DefaultClientSecret,
		"password":     password,
	}, JsonClientIDSecret, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// AppLoginByMobileData no encrypt
type AppLoginByMobileData struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

// AppLoginByMobile 纯短信验证码登录（免密码）。
// 先调用 SendMessageCode 触发短信，再凭手机收到的验证码换取 token。
// 网页端「短信登录」页签使用的即是该接口（逆向自 chunk-2795042e）。
func (w *WoClient) AppLoginByMobile(phone, smsCode string, opts ...RestyOption) (*AppLoginByMobileData, error) {
	var resp AppLoginByMobileData
	_, err := w.RequestApiUser(KeyAppLoginByMobile, Json{
		"phone":        phone,
		"smsCode":      smsCode,
		"clientSecret": DefaultClientSecret,
	}, JsonClientIDSecret, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// LoginBySmsCode 纯短信验证码登录的便捷方法：登录成功后自动把 token 装入客户端。
func (w *WoClient) LoginBySmsCode(phone, smsCode string, opts ...RestyOption) (*AppLoginByMobileData, error) {
	resp, err := w.AppLoginByMobile(phone, smsCode, opts...)
	if err != nil {
		return nil, err
	}
	if resp.AccessToken != "" {
		w.SetAccessToken(resp.AccessToken)
	}
	if resp.RefreshToken != "" {
		w.SetRefreshToken(resp.RefreshToken)
	}
	return resp, nil
}

// SendMessageCodeResult no encrypt
type SendMessageCodeResult struct {
	// NeedCaptcha 为 true 表示服务端要求图形验证码（rsp_code 6006/6007）：
	// 用 GetVerifyCodeURL(uuid) 取图片，人工识别后调 SendMessageCodeWithCaptcha 重发
	NeedCaptcha bool
	// RspCode 原始 rsp_code，便于调用方细分处理
	RspCode string
}

// SendMessageCode 发送短信登录验证码（operateType=1）。
// 走独立端点 sendMessageCodeBase，非 dispatcher 协议。
func (w *WoClient) SendMessageCode(phone string, opts ...RestyOption) (*SendMessageCodeResult, error) {
	return w.sendMessageCode(phone, "", "", opts...)
}

// SendMessageCodeWithCaptcha 风控触发后，带图形验证码重发短信验证码。
// uuid 与 verifyCode 必须配对（uuid 用于服务端校验图形码，图片地址见 GetVerifyCodeURL）。
func (w *WoClient) SendMessageCodeWithCaptcha(phone, uuid, verifyCode string, opts ...RestyOption) (*SendMessageCodeResult, error) {
	return w.sendMessageCode(phone, uuid, verifyCode, opts...)
}

func (w *WoClient) sendMessageCode(phone, uuid, verifyCode string, opts ...RestyOption) (*SendMessageCodeResult, error) {
	param, err := w.EncryptParam(ChannelAPIUser, Json{
		"operateType": "1",
		"phone":       phone,
		"uuid":        uuid,
		"verifyCode":  verifyCode,
	})
	if err != nil {
		return nil, err
	}
	req := w.NewRequest()
	req.SetHeader("User-Agent", w.ua)
	req.SetHeaders(map[string]string{
		"Origin":       "https://pan.wo.cn",
		"Referer":      "https://pan.wo.cn/",
		"Content-Type": "application/json",
		// 浏览器指纹头（同 request.go，WAF 必须）
		"Accept":             "application/json, text/plain, */*",
		"sec-ch-ua":          `"Not.A/Brand";v="8", "Chromium";v="114", "Microsoft Edge";v="114"`,
		"sec-ch-ua-mobile":   "?0",
		"sec-ch-ua-platform": `"Windows"`,
		"Sec-Fetch-Dest":     "empty",
		"Sec-Fetch-Mode":     "cors",
		"Sec-Fetch-Site":     "same-site",
	})
	req.SetBody(Json{
		"func":     "pc_send",
		"clientId": DefaultClientID,
		"param":    param,
	}).SetResult(&Resp{})
	for _, opt := range opts {
		opt(req)
	}
	res, err := req.Post(fmt.Sprintf("%s/api-user/sendMessageCodeBase", DefaultBaseURL))
	if err != nil {
		return nil, err
	}
	if res.IsError() {
		return nil, fmt.Errorf("request failed with status: %s", res.Status())
	}
	var resp Resp
	if err := w.jsonUnmarshalFunc(res.Body(), &resp); err != nil {
		return nil, err
	}
	switch resp.Rsp.RspCode {
	case "0000":
		return &SendMessageCodeResult{RspCode: resp.Rsp.RspCode}, nil
	case "6006", "6007":
		return &SendMessageCodeResult{NeedCaptcha: true, RspCode: resp.Rsp.RspCode}, nil
	default:
		return nil, fmt.Errorf("sendMessageCode failed with rsp_code: %s, rsp_desc: %s", resp.Rsp.RspCode, resp.Rsp.RspDesc)
	}
}

// GetVerifyCodeURL 图形验证码图片地址（uuid 自行生成随机串，并配对传给 SendMessageCodeWithCaptcha）
func GetVerifyCodeURL(uuid string) string {
	return fmt.Sprintf("%s/api-user/getverifycode?uuid=%s", DefaultBaseURL, uuid)
}

type AppQueryUserData struct {
	UserId        string `json:"userId"`
	HeadUrl       string `json:"headUrl"`
	UserName      string `json:"userName"`
	Sex           string `json:"sex"`
	Birthday      string `json:"birthday"`
	IsModify      string `json:"isModify"`
	IsHeadModify  string `json:"isHeadModify"`
	IsSetPassword string `json:"isSetPassword"`
	RegisterTime  string `json:"registerTime"`
}

func (w *WoClient) AppQueryUser(opts ...RestyOption) (*AppQueryUserData, error) {
	var resp AppQueryUserData
	_, err := w.RequestApiUser(KeyAppQueryUser, Json{
		"accessToken": w.accessToken,
	}, JsonClientIDSecret, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// AppRefreshTokenData no encrypt
type AppRefreshTokenData struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
}

func (w *WoClient) AppRefreshToken(opts ...RestyOption) (*AppRefreshTokenData, error) {
	var resp AppRefreshTokenData
	_, err := w.RequestApiUser(KeyAppRefreshToken, Json{
		"refreshToken": w.refreshToken,
		"clientSecret": DefaultClientSecret,
	}, JsonClientIDSecret, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, err
}

func (w *WoClient) AppLogout(opts ...RestyOption) error {
	_, err := w.RequestApiUser(KeyAppLogout, Json{
		"accessToken": w.accessToken,
	}, JsonClientIDSecret, nil, opts...)
	return err
}
