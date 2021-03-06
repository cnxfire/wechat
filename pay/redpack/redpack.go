package redpack

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"github.com/cnxfire/wechat/pay/config"
	"github.com/cnxfire/wechat/util"
	"sort"
	"strconv"
)

var redpackGateway = "https://api.mch.weixin.qq.com/mmpaymkttransfers/sendredpack"
var groupredpackGateway = "https://api.mch.weixin.qq.com/mmpaymkttransfers/sendgroupredpack"

type RedPack struct {
	*config.Config
}

// NewRefund return an instance of refund package
func NewRedPack(cfg *config.Config) *RedPack {
	refund := RedPack{cfg}
	return &refund
}

//RedPackParams 调用参数
type RedPackParams struct {
	ActName     string
	MchBillno   string
	ReOpenid    string
	Remark      string
	SendName    string
	TotalAmount int
	TotalNum    int
	Wishing     string
	RootCa      string //ca证书
	SceneId     string //场景
}

//RedPackRequest 请求参数
type RedPackRequest struct {
	ActName     string `xml:"act_name"`     //必填，活动名称
	ClientIP    string `xml:"client_ip"`    //必填，调用接口的机器ip地址
	MchBillno   string `xml:"mch_billno"`   //必填，商户订单号
	MchID       string `xml:"mch_id"`       //必填，微信支付分配的商户号
	NonceStr    string `xml:"nonce_str"`    //必填,随机字符串，不超过32位
	ReOpenid    string `xml:"re_openid"`    //必填，接收红包者用户，用户在wxappid下的openid
	Remark      string `xml:"remark"`       //必填，备注信息
	SendName    string `xml:"send_name"`    //必填，红包发送者名称
	TotalAmount int    `xml:"total_amount"` //必填，付款金额，单位为分
	TotalNum    int    `xml:"total_num"`    //必填，红包发放人数
	Wishing     string `xml:"wishing"`      //必填，红包祝福语
	Wxappid     string `xml:"wxappid"`      //必填，微信公众号id
	Sign        string `xml:"sign"`         //必填，签名
	SceneId     string `xml:"scene_id"`     //非必填，红包使用场景
	//RiskInfo 	string		`xml:"risk_info"`    //非必填，用户操作的时间戳
	//ConsumeMchId string		`xml:"consume_mch_id"` //非必填，资金授权商户号
}

// RedPackResp 发送红包返回值
type RedPackResp struct {
	ReturnMsg   string `xml:"return_msg"`
	MchID       string `xml:"mch_id"`
	WxAppID     string `xml:"wxappid"`
	ReOpenid    string `xml:"re_openid"`
	TotalAmount int    `xml:"total_amount"`
	ReturnCode  string `xml:"return_code"`
	ResultCode  string `xml:"result_code"`
	ErrCode     string `xml:"err_code"`
	ErrCodeDes  string `xml:"err_code_des"`
	MchBillNo   string `xml:"mch_billno"`

	// 以下发送裂变红包的时候才会用到
	SendTime   string `xml:"send_time"`
	SendListID string `xml:"send_listid"`
}

//SendRedPack 发送普通红包
func (pcf *RedPack) SendRedPack(p *RedPackParams) (rsp RedPackResp, err error) {
	nonceStr := util.RandomStr(32)
	param := make(map[string]interface{})
	param["act_name"] = p.ActName
	param["client_ip"] = pcf.ClientIP
	param["mch_billno"] = p.MchBillno
	param["mch_id"] = pcf.PayMchID
	param["nonce_str"] = nonceStr
	param["re_openid"] = p.ReOpenid
	param["remark"] = p.Remark
	param["send_name"] = p.SendName
	param["total_amount"] = p.TotalAmount
	param["total_num"] = p.TotalNum
	param["wishing"] = p.Wishing
	param["wxappid"] = pcf.AppID
	param["scene_id"] = p.SceneId

	bizKey := "&key=" + pcf.PayKey
	str := orderParam(param, bizKey)
	sign := util.MD5Sum(str)
	request := RedPackRequest{
		ActName:     p.ActName,
		ClientIP:    pcf.ClientIP,
		MchBillno:   p.MchBillno,
		MchID:       pcf.PayMchID,
		NonceStr:    nonceStr,
		ReOpenid:    p.ReOpenid,
		Remark:      p.Remark,
		SendName:    p.SendName,
		TotalAmount: p.TotalAmount,
		TotalNum:    p.TotalNum,
		Wishing:     p.Wishing,
		Wxappid:     pcf.AppID,
		SceneId:     p.SceneId,
		Sign:        sign,
	}
	rawRet, err := util.PostXMLWithTLS(redpackGateway, request, p.RootCa, pcf.PayMchID)
	if err != nil {
		return
	}
	err = xml.Unmarshal(rawRet, &rsp)
	if err != nil {
		return
	}
	if rsp.ReturnCode == "SUCCESS" {
		if rsp.ResultCode == "SUCCESS" {
			err = nil
			return
		}
		err = fmt.Errorf("refund error, errcode=%s,errmsg=%s", rsp.ErrCode, rsp.ErrCodeDes)
		return
	}
	err = fmt.Errorf("[msg : xmlUnmarshalError] [rawReturn : %s] [params : %s] [sign : %s]",
		string(rawRet), str, sign)
	return
}

// order params
func orderParam(source interface{}, bizKey string) (returnStr string) {
	switch v := source.(type) {
	case map[string]string:
		keys := make([]string, 0, len(v))
		for k := range v {
			if k == "sign" {
				continue
			}
			keys = append(keys, k)
		}
		sort.Strings(keys)
		var buf bytes.Buffer
		for _, k := range keys {
			if v[k] == "" {
				continue
			}
			if buf.Len() > 0 {
				buf.WriteByte('&')
			}
			buf.WriteString(k)
			buf.WriteByte('=')
			buf.WriteString(v[k])
		}
		buf.WriteString(bizKey)
		returnStr = buf.String()
	case map[string]interface{}:
		keys := make([]string, 0, len(v))
		for k := range v {
			if k == "sign" {
				continue
			}
			keys = append(keys, k)
		}
		sort.Strings(keys)
		var buf bytes.Buffer
		for _, k := range keys {
			if v[k] == "" {
				continue
			}
			if buf.Len() > 0 {
				buf.WriteByte('&')
			}
			buf.WriteString(k)
			buf.WriteByte('=')
			switch vv := v[k].(type) {
			case string:
				buf.WriteString(vv)
			case int:
				buf.WriteString(strconv.FormatInt(int64(vv), 10))
			default:
				panic("params type not supported")
			}
		}
		buf.WriteString(bizKey)
		returnStr = buf.String()
	}
	return
}
