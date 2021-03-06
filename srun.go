package srun

import (
	"fmt"
	"github.com/astaxie/beego/logs"
	"log"
	"srun/form"
	"srun/hash"
	"srun/resp"
	"srun/utils"
	"strings"
)

const (
	challengeUrl       = "http://10.0.0.55/cgi-bin/get_challenge"
	portalUrl          = "http://10.0.0.55/cgi-bin/srun_portal"
	succeedUrl         = "http://10.0.0.55/srun_portal_pc_succeed.php"
	succeedUrlYidong   = "http://10.0.0.55/srun_portal_pc_succeed_yys.php"
	succeedUrlLiantong = "srun_portal_pc_succeed_yys_cucc.php"

	url = "http://10.0.0.55"
)

// api Login
// step 1: get acid
// step 2: get challenge
// step 3: do login
func Login(username, password string) (token, ip string) {
	// 先获取acid
	// 并检查是否已经联网
	acid, err := utils.GetAcid()
	if err != nil {
		log.Fatalln(err)
	}

	// 创建登录表单
	formLogin := form.Login(username, password, acid)

	//	get token
	qc := form.Challenge(username)

	rc := resp.Challenge{}
	if err := utils.GetJson(challengeUrl, qc, &rc); err != nil {
		logs.Error("请求错误")
		logs.Debug(err)
		return
	}

	token = rc.Challenge
	ip = rc.ClientIp

	formLogin.Set("ip", ip)
	info := hash.GenInfo(formLogin, token)
	formLogin.Set("info", info)
	formLogin.Set("password", hash.PwdHmd5("", token))
	formLogin.Set("chksum", hash.Checksum(formLogin, token))

	// response
	ra := resp.RAction{}
	err = utils.GetJson(portalUrl, formLogin, &ra)
	if err != nil {
		logs.Error("请求错误")
		logs.Debug(err)
		return
	}
	if ra.Res != "ok" {
		log.Println("登录失败:", ra.Res)
		log.Println("msg:", ra.ErrorMsg)
		logs.Debug(ra)
		return
	}

	log.Println("登录成功!")
	log.Println("ip:", ra.ClientIp)

	qs := form.Info(
		acid,
		formLogin.Get("username"),
		ra.ClientIp,
		token,
	)

	// 余量查询
	if strings.Contains(username, "@yidong") {
		fmt.Println("服务器:", "移动")
		utils.ParseHtml(succeedUrlYidong, qs)
	} else if strings.Contains(username, "@liantong") {
		fmt.Println("服务器:", "联通")
		utils.ParseHtml(succeedUrlLiantong, qs)
	} else {
		fmt.Println("服务器:", "校园网")
		utils.ParseHtml(succeedUrl, qs)
	}
	return
}

// api info
func Info(username, token, ip string) {
	qs := form.Info(
		1,
		username,
		ip,
		token,
	)
	utils.ParseHtml(succeedUrl, qs)
	return
}

// api logout
func Logout(username string) {
	q := form.Logout(username)

	ra := resp.RAction{}
	err := utils.GetJson(portalUrl, q, &ra)
	if err != nil {
		logs.Error("请求错误", err)
		logs.Debug(err)
		return
	}
	if ra.Error == "ok" {
		fmt.Println("下线成功！")
	} else {
		logs.Error("下线失败！")
		logs.Error(ra)
	}
}
