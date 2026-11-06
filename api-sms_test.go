package wopan_test

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/xhofe/wopan-sdk-go"
)

// TestWoClient_SmsLogin 真机冒烟：手机号由环境变量 WOPAN_PHONE 提供。
// 流程：SendMessageCode → 等待人工输入短信码（stdin，测试超时内）→ LoginBySmsCode → AppQueryUser 验证。
//
//	go test -run TestWoClient_SmsLogin -v -timeout 10m
//	WOPAN_PHONE=138xxxx5678 go test -run TestWoClient_SmsLogin -v -timeout 10m
func TestWoClient_SmsLogin(t *testing.T) {
	phone := os.Getenv("WOPAN_PHONE")
	if phone == "" {
		t.Skip("WOPAN_PHONE 未设置，跳过真机测试")
	}
	w := wopan.Default()
	if proxy := os.Getenv("WOPAN_PROXY"); proxy != "" {
		w.SetProxy(proxy)
	}

	// 1. 发送短信验证码
	res, err := w.SendMessageCode(phone)
	if err != nil {
		t.Fatalf("SendMessageCode() error = %v", err)
	}
	if res.NeedCaptcha {
		uuid := fmt.Sprintf("%d", time.Now().UnixNano())
		t.Logf("服务端要求图形验证码（rsp_code=%s）", res.RspCode)
		t.Logf("请打开图片查看并输入图形验证码: %s", wopan.GetVerifyCodeURL(uuid))
		fmt.Print("输入图形验证码: ")
		captcha := readLine(t)
		res, err = w.SendMessageCodeWithCaptcha(phone, uuid, captcha)
		if err != nil {
			t.Fatalf("SendMessageCodeWithCaptcha() error = %v", err)
		}
		if res.NeedCaptcha {
			t.Fatalf("图形验证码校验仍未通过（rsp_code=%s）", res.RspCode)
		}
	}
	t.Log("短信验证码已发送，请查收")

	// 2. 等待人工输入短信码
	fmt.Print("输入收到的短信验证码: ")
	smsCode := readLine(t)

	// 3. 纯短信登录（免密码）
	login, err := w.LoginBySmsCode(phone, smsCode)
	if err != nil {
		t.Fatalf("LoginBySmsCode() error = %v", err)
	}
	t.Logf("LoginBySmsCode() = access_token=%s refresh_token=%s expires_in=%d",
		login.AccessToken, login.RefreshToken, login.ExpiresIn)
	if login.AccessToken == "" {
		t.Fatal("access_token 为空")
	}
	if login.RefreshToken != "" {
		t.Logf("✓ 响应含 refresh_token（可静默续期）")
	} else {
		t.Logf("✗ 响应不含 refresh_token（token 过期后需重新短信登录）")
	}

	// 4. 验证 token 有效性
	user, err := w.AppQueryUser()
	if err != nil {
		t.Fatalf("AppQueryUser() error = %v", err)
	}
	t.Logf("AppQueryUser() = userId=%s userName=%s", user.UserId, user.UserName)
	if !strings.HasPrefix(user.UserId, phone[:3]) || !strings.HasSuffix(user.UserId, phone[len(phone)-4:]) {
		t.Logf("⚠ userId=%s 与登录手机号 %s 掩码不一致，请人工核对", user.UserId, phone)
	}
}

func readLine(t *testing.T) string {
	sc := bufio.NewScanner(os.Stdin)
	if !sc.Scan() {
		t.Fatal("读取输入失败")
	}
	return strings.TrimSpace(sc.Text())
}
