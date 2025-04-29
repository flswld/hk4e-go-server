package controller

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"math"
	"net/http"
	"os"
	"slices"
	"time"

	"hk4e/common/config"
	"hk4e/common/region"
	httpapi "hk4e/dispatch/api"
	"hk4e/node/api"
	"hk4e/pkg/endec"
	"hk4e/pkg/random"
	"hk4e/protocol/proto"

	"github.com/flswld/halo/logger"
	"github.com/gin-gonic/gin"
	pb "google.golang.org/protobuf/proto"
)

// RegionCustomConfig 区服相关的配置 避免在http中使用Json格式
type RegionCustomConfig struct {
	CloseAntiDebug   bool `json:"close_antidebug"`  // 默认打开反调开关 默认false
	ForceKill        bool `json:"force_kill"`       // 默认false
	AntiDebugPc      bool `json:"antidebug_pc"`     // pc默认不开启反调 默认false
	AntiDebugIos     bool `json:"antidubug_ios"`    // ios默认不开启反调 默认false
	AntiDebugAndroid bool `json:"antidubug_androd"` // android默认不开启反调 默认false
}

// ClientCustomConfig 客户端版本定义的配置 客户端版本号对应的配置 需要兼容老的json格式
type ClientCustomConfig struct {
	Visitor        bool              `json:"visitor"`        // 游客功能
	SdkEnv         string            `json:"sdkenv"`         // sdk环境类型
	DebugMenu      bool              `json:"debugmenu"`      // debug菜单
	DebugLogSwitch []int32           `json:"debuglogswitch"` // 打开的log类型
	DebugLog       bool              `json:"debuglog"`       // log总开关
	DeviceList     map[string]string `json:"devicelist"`
	LoadJsonData   bool              `json:"loadjsondata"`  // 用json读取InLevel数据
	ShowException  bool              `json:"showexception"` // 是否显示异常提示框 默认为true
	CheckDevice    bool              `json:"checkdevice"`
	LoadPatch      bool              `json:"loadPatch"`
	RegionConfig   string            `json:"regionConfig"`
	DownloadMode   int32             `json:"downloadMode"`
	CodeSwitch     []int32           `json:"codeSwitch"`
	CoverSwitch    []int32           `json:"coverSwitch"`
}

// GetRegionList 一级dispatch信息
func GetRegionList(ec2b *random.Ec2b) *proto.QueryRegionListHttpRsp {
	regionList := new(proto.QueryRegionListHttpRsp)
	regionList.Retcode = 0
	serverList := make([]*proto.RegionSimpleInfo, 0)
	server := &proto.RegionSimpleInfo{
		Name:        "os_usa",
		Title:       "America",
		Type:        "DEV_PUBLIC",
		DispatchUrl: config.GetConfig().Hk4e.DispatchUrl,
	}
	serverList = append(serverList, server)
	regionList.RegionList = serverList
	dispatchEc2bData := ec2b.Bytes()
	regionList.ClientSecretKey = dispatchEc2bData // 客户端使用密钥
	dispatchXorKey := ec2b.XorKey()
	clientCustomConfig, _ := json.Marshal(&ClientCustomConfig{
		SdkEnv:         "2",
		CheckDevice:    false,
		LoadPatch:      false,
		ShowException:  false,
		RegionConfig:   "pm|fk|add",
		DownloadMode:   0,
		DebugMenu:      true,
		DebugLogSwitch: []int32{0},
		DebugLog:       true,
		CodeSwitch:     []int32{3628},
		CoverSwitch:    []int32{40},
	})
	endec.Xor(clientCustomConfig, dispatchXorKey)
	regionList.ClientCustomConfigEncrypted = clientCustomConfig // 加密后的客户端版本定义的配置
	regionList.EnableLoginPc = true
	return regionList
}

// GetRegionCurr 二级dispatch信息
func GetRegionCurr(ec2b *random.Ec2b, gateServerAddr *api.GateServerAddr, stopServerInfo *api.StopServerInfo) *proto.QueryCurrRegionHttpRsp {
	regionCurr := new(proto.QueryCurrRegionHttpRsp)
	// region_info retcode == 0 || RET_STOP_SERVER
	// force_udpate retcode == RET_CLIENT_FORCE_UPDATE
	// stop_server retcode == RET_STOP_SERVER
	if !stopServerInfo.StopServer {
		regionCurr.Retcode = 0 // 错误码
	} else {
		regionCurr.Retcode = int32(proto.Retcode_RET_STOP_SERVER)
		regionCurr.Detail = &proto.QueryCurrRegionHttpRsp_StopServer{
			StopServer: &proto.StopServerInfo{
				StopBeginTime: stopServerInfo.StartTime,
				StopEndTime:   stopServerInfo.EndTime,
				Url:           "https://hk4e.flswld.com",
				ContentMsg:    "服务器维护中",
			},
		}
	}
	regionCurr.Msg = "" // 错误信息
	dispatchEc2bData := ec2b.Bytes()
	regionCurr.RegionInfo = &proto.RegionInfo{
		GateserverIp:   gateServerAddr.KcpAddr,
		GateserverPort: gateServerAddr.KcpPort,
		SecretKey:      dispatchEc2bData, // 第一条协议加密密钥
	}
	regionCurr.ClientSecretKey = dispatchEc2bData // 客户端使用密钥
	dispatchXorKey := ec2b.XorKey()
	regionCustomConfig, _ := json.Marshal(&RegionCustomConfig{
		CloseAntiDebug:   true,
		ForceKill:        false,
		AntiDebugPc:      false,
		AntiDebugIos:     false,
		AntiDebugAndroid: false,
	})
	endec.Xor(regionCustomConfig, dispatchXorKey)
	regionCurr.RegionCustomConfigEncrypted = regionCustomConfig // 加密后的区服定义的配置
	clientCustomConfig, _ := json.Marshal(&ClientCustomConfig{
		SdkEnv:         "2",
		CheckDevice:    false,
		LoadPatch:      false,
		ShowException:  true,
		RegionConfig:   "pm|fk|add",
		DownloadMode:   0,
		DebugMenu:      true,
		DebugLogSwitch: []int32{0},
		DebugLog:       true,
		CodeSwitch:     []int32{3628},
		CoverSwitch:    []int32{40},
	})
	endec.Xor(clientCustomConfig, dispatchXorKey)
	regionCurr.ClientRegionCustomConfigEncrypted = clientCustomConfig // 加密后的客户端区服定义的配置
	return regionCurr
}

type GateServerInfo struct {
	addr      *api.GateServerAddr
	timestamp int64
}

// dispatch请求错误响应
func (c *Controller) dispatchReqErrorRsp(ctx *gin.Context, dispatchLevel uint8) {
	if dispatchLevel == 1 {
		// query_region_list
		_, _ = ctx.Writer.WriteString("CP///////////wE=")
	} else if dispatchLevel == 2 {
		// query_cur_region
		_, _ = ctx.Writer.WriteString("CAESGE5vdCBGb3VuZCB2ZXJzaW9uIGNvbmZpZw==")
	}
}

// 一级dispatch
func (c *Controller) queryRegionList(ctx *gin.Context) {
	ctx.Header("Content-type", "text/html; charset=UTF-8")
	regionList := GetRegionList(c.ec2b)
	regionListData, err := pb.Marshal(regionList)
	if err != nil {
		logger.Error("pb marshal QueryRegionListHttpRsp error: %v", err)
		c.dispatchReqErrorRsp(ctx, 1)
		return
	}
	regionListBase64 := base64.StdEncoding.EncodeToString(regionListData)
	_, _ = ctx.Writer.WriteString(regionListBase64)
}

// 二级dispatch
func (c *Controller) queryCurRegion(ctx *gin.Context) {
	versionName := ctx.Query("version")
	if versionName == "" {
		c.dispatchReqErrorRsp(ctx, 2)
		return
	}
	version, versionStr := region.GetClientVersionByName(versionName)
	if version == 0 {
		c.dispatchReqErrorRsp(ctx, 2)
		return
	}
	now := time.Now().Unix()
	gateServerInfo, exist := c.gateServerMap.Load(versionStr)
	if !exist {
		addr, err := c.discoveryClient.GetGateServerAddr(ctx.Request.Context(), &api.GetGateServerAddrReq{
			GameVersion: versionStr,
		})
		if err != nil {
			logger.Error("get gate server addr error: %v", err)
			c.dispatchReqErrorRsp(ctx, 2)
			return
		}
		gateServerInfo = &GateServerInfo{
			addr:      addr,
			timestamp: now,
		}
		c.gateServerMap.Store(versionStr, gateServerInfo)
	}
	if now-gateServerInfo.(*GateServerInfo).timestamp > 60 {
		addr, err := c.discoveryClient.GetGateServerAddr(ctx.Request.Context(), &api.GetGateServerAddrReq{
			GameVersion: versionStr,
		})
		if err != nil {
			logger.Error("get gate server addr error: %v", err)
		} else {
			gateServerInfo = &GateServerInfo{
				addr:      addr,
				timestamp: now,
			}
			c.gateServerMap.Store(versionStr, gateServerInfo)
		}
	}
	gateServerAddr := gateServerInfo.(*GateServerInfo).addr
	regionCurr := GetRegionCurr(c.ec2b, gateServerAddr, &api.StopServerInfo{StopServer: false})
	if c.stopServerInfo.StopServer {
		clientIp := ctx.ClientIP()
		if !slices.Contains[[]string, string](c.whiteList.IpAddrList, clientIp) {
			regionCurr = GetRegionCurr(c.ec2b, gateServerAddr, c.stopServerInfo)
		}
	}
	regionCurrData, err := pb.Marshal(regionCurr)
	if err != nil {
		logger.Error("pb marshal QueryCurrRegionHttpRsp error: %v", err)
		c.dispatchReqErrorRsp(ctx, 2)
		return
	}
	if version < 275 {
		ctx.Header("Content-type", "text/html; charset=UTF-8")
		regionCurrBase64 := base64.StdEncoding.EncodeToString(regionCurrData)
		_, _ = ctx.Writer.WriteString(regionCurrBase64)
		return
	}
	// 2.8版本开始的加密和签名
	keyId := ctx.Query("key_id")
	encPubPrivKey, exist := c.encRsaKeyMap[keyId]
	if !exist {
		logger.Error("can not found key id: %v", keyId)
		c.dispatchReqErrorRsp(ctx, 2)
		return
	}
	chunkSize := 256 - 11
	regionInfoLength := len(regionCurrData)
	numChunks := int(math.Ceil(float64(regionInfoLength) / float64(chunkSize)))
	encryptedRegionInfo := make([]byte, 0)
	for i := 0; i < numChunks; i++ {
		from := i * chunkSize
		to := int(math.Min(float64((i+1)*chunkSize), float64(regionInfoLength)))
		chunk := regionCurrData[from:to]
		pubKey, err := endec.RsaParsePubKeyByPrivKey(encPubPrivKey)
		if err != nil {
			logger.Error("parse rsa pub key error: %v", err)
			c.dispatchReqErrorRsp(ctx, 2)
			return
		}
		privKey, err := endec.RsaParsePrivKey(encPubPrivKey)
		if err != nil {
			logger.Error("parse rsa priv key error: %v", err)
			c.dispatchReqErrorRsp(ctx, 2)
			return
		}
		encrypt, err := endec.RsaEncrypt(chunk, pubKey)
		if err != nil {
			logger.Error("rsa enc error: %v", err)
			c.dispatchReqErrorRsp(ctx, 2)
			return
		}
		decrypt, err := endec.RsaDecrypt(encrypt, privKey)
		if err != nil {
			logger.Error("rsa dec error: %v", err)
			c.dispatchReqErrorRsp(ctx, 2)
			return
		}
		if bytes.Compare(decrypt, chunk) != 0 {
			logger.Error("rsa dec test fail")
			c.dispatchReqErrorRsp(ctx, 2)
			return
		}
		encryptedRegionInfo = append(encryptedRegionInfo, encrypt...)
	}
	signPrivkey, err := endec.RsaParsePrivKey(c.signRsaKey)
	if err != nil {
		logger.Error("parse rsa priv key error: %v", err)
		c.dispatchReqErrorRsp(ctx, 2)
		return
	}
	signData, err := endec.RsaSign(regionCurrData, signPrivkey)
	if err != nil {
		logger.Error("rsa sign error: %v", err)
		c.dispatchReqErrorRsp(ctx, 2)
		return
	}
	ok, err := endec.RsaVerify(regionCurrData, signData, &signPrivkey.PublicKey)
	if err != nil {
		logger.Error("rsa verify error: %v", err)
		c.dispatchReqErrorRsp(ctx, 2)
		return
	}
	if !ok {
		logger.Error("rsa verify test fail")
		c.dispatchReqErrorRsp(ctx, 2)
		return
	}
	rsp := &httpapi.QueryCurRegionRspJson{
		Content: base64.StdEncoding.EncodeToString(encryptedRegionInfo),
		Sign:    base64.StdEncoding.EncodeToString(signData),
	}
	ctx.JSON(http.StatusOK, rsp)
}

func (c *Controller) querySecurityFile(ctx *gin.Context) {
	// 很早以前2.6.0版本的时候抓包为了完美还原写的 不清楚有没有副作用暂时不要了
	return
	file, err := os.ReadFile("static/security_file")
	if err != nil {
		logger.Error("open security_file error")
		return
	}
	ctx.Header("Content-type", "text/html; charset=UTF-8")
	_, _ = ctx.Writer.WriteString(string(file))
}
