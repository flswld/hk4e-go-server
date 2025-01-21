package net

import (
	"context"
	"encoding/binary"
	"net"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"hk4e/common/config"
	"hk4e/common/mq"
	"hk4e/common/region"
	"hk4e/common/rpc"
	"hk4e/gate/client_proto"
	"hk4e/gate/dao"
	"hk4e/gate/kcp"
	"hk4e/node/api"
	"hk4e/pkg/logger"
	"hk4e/pkg/random"
	"hk4e/protocol/cmd"
)

// 网络连接管理

const (
	ConnEstFreqLimit      = 100        // 每秒连接建立频率限制
	RecvPacketFreqLimit   = 1000       // 客户端上行每秒发包频率限制
	SendPacketFreqLimit   = 1000       // 服务器下行每秒发包频率限制
	PacketMaxLen          = 343 * 1024 // 最大应用层包长度
	ConnRecvTimeout       = 30         // 收包超时时间 秒
	ConnSendTimeout       = 10         // 发包超时时间 秒
	MaxClientConnNumLimit = 1000       // 最大客户端连接数限制
	TcpNoDelay            = true       // 是否禁用tcp的nagle
)

var CLIENT_CONN_NUM int32 = 0 // 当前客户端连接数

const (
	ConnEventEst   = "ConnEventEst"
	ConnEventClose = "ConnEventClose"
)

type ConnEvent struct {
	SessionId    uint32
	EventId      string
	EventMessage any
}

type ConnManager struct {
	db                      *dao.Dao
	discoveryClient         *rpc.DiscoveryClient // 节点服务器rpc客户端
	messageQueue            *mq.MessageQueue     // 消息队列
	globalGsOnlineMap       map[uint32]string    // 全服玩家在线表
	globalGsOnlineMapLock   sync.RWMutex
	minLoadGsServerAppId    string
	minLoadMultiServerAppId string
	stopServerInfo          *api.StopServerInfo
	whiteList               *api.GetWhiteListRsp
	// 会话
	sessionIdCounter uint32
	sessionMap       map[uint32]*Session
	sessionUserIdMap map[uint32]*Session
	sessionMapLock   sync.RWMutex
	// 事件
	createSessionChan        chan *Session
	destroySessionChan       chan *Session
	connEventChan            chan *ConnEvent
	reLoginRemoteKickRegChan chan *RemoteKick
	// 协议
	serverCmdProtoMap *cmd.CmdProtoMap
	clientCmdProtoMap *client_proto.ClientCmdProtoMap
	// 密钥
	signRsaKey   []byte
	encRsaKeyMap map[string][]byte
	dispatchKey  []byte
}

func NewConnManager(db *dao.Dao, messageQueue *mq.MessageQueue, discovery *rpc.DiscoveryClient) (*ConnManager, error) {
	r := new(ConnManager)
	r.db = db
	r.discoveryClient = discovery
	r.messageQueue = messageQueue
	r.globalGsOnlineMap = make(map[uint32]string)
	r.minLoadGsServerAppId = ""
	r.minLoadMultiServerAppId = ""
	r.stopServerInfo = nil
	r.whiteList = nil
	r.sessionIdCounter = 0
	r.sessionMap = make(map[uint32]*Session)
	r.sessionUserIdMap = make(map[uint32]*Session)
	r.createSessionChan = make(chan *Session, 1000)
	r.destroySessionChan = make(chan *Session, 1000)
	r.connEventChan = make(chan *ConnEvent, 1000)
	r.reLoginRemoteKickRegChan = make(chan *RemoteKick, 1000)
	r.serverCmdProtoMap = cmd.NewCmdProtoMap()
	if config.GetConfig().Hk4e.ClientProtoProxyEnable {
		r.clientCmdProtoMap = client_proto.NewClientCmdProtoMap()
	}
	err := r.run()
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (c *ConnManager) run() error {
	// 读取密钥相关文件
	c.signRsaKey, c.encRsaKeyMap, _ = region.LoadRegionRsaKey()
	// key
	rsp, err := c.discoveryClient.GetRegionEc2B(context.TODO(), &api.NullMsg{})
	if err != nil {
		logger.Error("get region ec2b error: %v", err)
		return err
	}
	ec2b, err := random.LoadEc2bKey(rsp.Data)
	if err != nil {
		logger.Error("parse region ec2b error: %v", err)
		return err
	}
	regionEc2b := random.NewEc2b()
	regionEc2b.SetSeed(ec2b.Seed())
	c.dispatchKey = regionEc2b.XorKey()
	// kcp
	addr := "0.0.0.0:" + strconv.Itoa(int(config.GetConfig().Hk4e.KcpPort))
	kcpListener, err := kcp.ListenKCP(addr)
	if err != nil {
		logger.Error("listen kcp err: %v", err)
		return err
	}
	kcp.SetByteCheckMode(int(config.GetConfig().Hk4e.ByteCheckMode))
	logger.Info("listen kcp at addr: %v", addr)
	go c.kcpNetInfo()
	go c.acceptHandle(false, kcpListener, nil)
	if config.GetConfig().Hk4e.TcpModeEnable {
		// tcp
		addr := "0.0.0.0:" + strconv.Itoa(int(config.GetConfig().Hk4e.KcpPort))
		tcpAddr, err := net.ResolveTCPAddr("tcp4", addr)
		if err != nil {
			logger.Error("parse tcp addr err: %v", err)
			return err
		}
		tcpListener, err := net.ListenTCP("tcp4", tcpAddr)
		if err != nil {
			logger.Error("listen tcp err: %v", err)
			return err
		}
		logger.Info("listen tcp at addr: %v", addr)
		go c.acceptHandle(true, nil, tcpListener)
	}
	go c.forwardServerMsgToClientHandle()
	c.syncGlobalGsOnlineMap()
	go c.autoSyncGlobalGsOnlineMap()
	c.syncMinLoadServerAppid()
	go c.autoSyncMinLoadServerAppid()
	c.syncWhiteList()
	go c.autoSyncWhiteList()
	c.syncStopServerInfo()
	go c.autoSyncStopServerInfo()
	go func() {
		for {
			connEvent := <-c.connEventChan
			logger.Info("[Conn Event] connEvent: %+v", *connEvent)
		}
	}()
	return nil
}

func (c *ConnManager) Close() {
	c.closeAllConn()
}

func (c *ConnManager) kcpNetInfo() {
	ticker := time.NewTicker(time.Second * 60)
	kcpErrorCount := uint64(0)
	for {
		<-ticker.C
		snmp := kcp.DefaultSnmp.Copy()
		kcpErrorCount += snmp.KCPInErrors
		logger.Info("kcp send: %v B/s, kcp recv: %v B/s", snmp.BytesSent/60, snmp.BytesReceived/60)
		logger.Info("udp send: %v B/s, udp recv: %v B/s", snmp.OutBytes/60, snmp.InBytes/60)
		logger.Info("udp send: %v pps, udp recv: %v pps", snmp.OutPkts/60, snmp.InPkts/60)
		clientConnNum := atomic.LoadInt32(&CLIENT_CONN_NUM)
		logger.Info("conn num: %v, new conn num: %v, kcp error num: %v", clientConnNum, snmp.CurrEstab, kcpErrorCount)
		kcp.DefaultSnmp.Reset()
	}
}

// 接收新连接协程
func (c *ConnManager) acceptHandle(tcpMode bool, kcpListener *kcp.Listener, tcpListener *net.TCPListener) {
	logger.Info("accept handle start, tcpMode: %v", tcpMode)
	connEstFreqLimitCounter := 0
	connEstFreqLimitTimer := time.Now().UnixNano()
	for {
		var conn Conn = nil
		if !tcpMode {
			kcpConn, err := kcpListener.AcceptKCP()
			if err != nil {
				logger.Error("accept kcp err: %v", err)
				continue
			}
			kcpConn.SetACKNoDelay(true)
			kcpConn.SetWriteDelay(false)
			kcpConn.SetWindowSize(256, 256)
			kcpConn.SetMtu(1200)
			conn = kcpConn
		} else {
			tcpConn, err := tcpListener.AcceptTCP()
			if err != nil {
				logger.Error("accept tcp err: %v", err)
				continue
			}
			if TcpNoDelay {
				_ = tcpConn.SetNoDelay(true)
			}
			conn = NewTcpConn(tcpConn)
		}
		// 连接建立频率限制
		connEstFreqLimitCounter++
		if connEstFreqLimitCounter > ConnEstFreqLimit {
			now := time.Now().UnixNano()
			if now-connEstFreqLimitTimer > int64(time.Second) {
				connEstFreqLimitCounter = 0
				connEstFreqLimitTimer = now
			} else {
				logger.Error("conn est freq limit, now: %v conn/s", connEstFreqLimitCounter)
				_ = conn.Close(kcp.EnetServerKick)
				continue
			}
		}
		sessionId := uint32(0)
		if !tcpMode {
			sessionId = conn.GetSessionId()
		} else {
			sessionId = atomic.AddUint32(&c.sessionIdCounter, 1)
		}
		logger.Info("[ACCEPT] client connect, tcpMode: %v, sessionId: %v, conv: %v, addr: %v",
			tcpMode, sessionId, conn.GetConv(), conn.RemoteAddr())
		session := &Session{
			sessionId:        sessionId,
			conn:             conn,
			connState:        ConnEst,
			userId:           0,
			sendChan:         make(chan *ProtoMsg, 1000),
			seed:             0,
			xorKey:           c.dispatchKey,
			changeXorKeyFin:  false,
			gsServerAppId:    "",
			multiServerAppId: "",
			robotServerAppId: "",
			useMagicSeed:     false,
		}
		go c.recvHandle(session)
		go c.sendHandle(session)
		// 连接建立成功通知
		c.connEventChan <- &ConnEvent{
			SessionId:    session.sessionId,
			EventId:      ConnEventEst,
			EventMessage: session.conn.RemoteAddr(),
		}
		atomic.AddInt32(&CLIENT_CONN_NUM, 1)
	}
}

// Session 连接会话结构 只允许定义并发安全或者简单的基础数据结构
type Session struct {
	sessionId        uint32
	conn             Conn
	connState        uint32
	userId           uint32
	sendChan         chan *ProtoMsg
	seed             uint64
	xorKey           []byte
	changeXorKeyFin  bool
	gsServerAppId    string
	multiServerAppId string
	robotServerAppId string
	useMagicSeed     bool
}

// 接收协程
func (c *ConnManager) recvHandle(session *Session) {
	logger.Info("recv handle start, sessionId: %v", session.sessionId)
	conn := session.conn
	header := make([]byte, 4)
	payload := make([]byte, PacketMaxLen)
	pktFreqLimitCounter := 0
	pktFreqLimitTimer := time.Now().UnixNano()
	_, isKcpConn := conn.(*kcp.UDPSession)
	tcpConn, _ := conn.(*TCPConn)
	for {
		var bin []byte = nil
		if isKcpConn {
			_ = conn.SetReadDeadline(time.Now().Add(time.Second * ConnRecvTimeout))
			recvLen, err := conn.Read(payload)
			if err != nil {
				logger.Debug("exit recv loop, conn read err: %v, sessionId: %v", err, session.sessionId)
				c.closeConn(session, kcp.EnetServerKick)
				return
			}
			bin = payload[:recvLen]
		} else {
			// tcp流分割解析
			recvLen := 0
			for recvLen < 4 {
				_ = conn.SetReadDeadline(time.Now().Add(time.Second * ConnRecvTimeout))
				n, err := conn.Read(header[recvLen:])
				if err != nil {
					logger.Debug("exit recv loop, conn read err: %v, sessionId: %v", err, session.sessionId)
					c.closeConn(session, kcp.EnetServerKick)
					return
				}
				recvLen += n
			}
			msgLen := binary.BigEndian.Uint32(header)
			// tcp rtt探测
			if msgLen == 0 {
				_ = conn.SetWriteDeadline(time.Now().Add(time.Second * ConnSendTimeout))
				_, err := conn.Write([]byte{0x00, 0x00, 0x00, 0x00})
				if err != nil {
					logger.Debug("exit recv loop, conn write err: %v, sessionId: %v", err, session.sessionId)
					c.closeConn(session, kcp.EnetServerKick)
					return
				}
				continue
			}
			if msgLen == 0xffffffff {
				now := time.Now().UnixMilli()
				tcpConn.TCPRtt = uint32(now - tcpConn.TCPRttLastSendTime)
				logger.Debug("[TCP RTT] sessionId: %v, rtt: %v ms", session.sessionId, tcpConn.TCPRtt)
				continue
			}
			if msgLen > PacketMaxLen {
				logger.Error("exit recv loop, msg len too long, sessionId: %v", session.sessionId)
				c.closeConn(session, kcp.EnetServerKick)
				return
			}
			recvLen = 0
			for recvLen < int(msgLen) {
				_ = conn.SetReadDeadline(time.Now().Add(time.Second * ConnRecvTimeout))
				n, err := conn.Read(payload[recvLen:msgLen])
				if err != nil {
					logger.Debug("exit recv loop, conn read err: %v, sessionId: %v", err, session.sessionId)
					c.closeConn(session, kcp.EnetServerKick)
					return
				}
				recvLen += n
			}
			bin = payload[:msgLen]
		}
		// 收包频率限制
		pktFreqLimitCounter++
		if pktFreqLimitCounter > RecvPacketFreqLimit {
			now := time.Now().UnixNano()
			if now-pktFreqLimitTimer > int64(time.Second) {
				pktFreqLimitCounter = 0
				pktFreqLimitTimer = now
			} else {
				logger.Error("exit recv loop, client packet send freq too high, sessionId: %v, pps: %v",
					session.sessionId, pktFreqLimitCounter)
				c.closeConn(session, kcp.EnetPacketFreqTooHigh)
				return
			}
		}
		kcpMsgList := make([]*KcpMsg, 0)
		DecodeBinToPayload(bin, session.sessionId, &kcpMsgList, session.xorKey)
		for _, v := range kcpMsgList {
			protoMsgList := ProtoDecode(v, c.serverCmdProtoMap, c.clientCmdProtoMap)
			for _, vv := range protoMsgList {
				c.forwardClientMsgToServerHandle(vv, session)
			}
		}
	}
}

// 发送协程
func (c *ConnManager) sendHandle(session *Session) {
	logger.Info("send handle start, sessionId: %v", session.sessionId)
	conn := session.conn
	pktFreqLimitCounter := 0
	pktFreqLimitTimer := time.Now().UnixNano()
	tcpConn, isTcpConn := conn.(*TCPConn)
	for {
		protoMsg, ok := <-session.sendChan
		if !ok {
			logger.Debug("exit send loop, send chan close, sessionId: %v", session.sessionId)
			c.closeConn(session, kcp.EnetServerKick)
			return
		}
		kcpMsg := ProtoEncode(protoMsg, c.serverCmdProtoMap, c.clientCmdProtoMap)
		if kcpMsg == nil {
			logger.Error("encode kcp msg is nil, sessionId: %v", session.sessionId)
			continue
		}
		bin := EncodePayloadToBin(kcpMsg, session.xorKey)
		if isTcpConn {
			// tcp流分割的4个字节payload长度头部
			headLenData := make([]byte, 4)
			binary.BigEndian.PutUint32(headLenData, uint32(len(bin)))
			_ = conn.SetWriteDeadline(time.Now().Add(time.Second * ConnSendTimeout))
			_, err := conn.Write(headLenData)
			if err != nil {
				logger.Debug("exit send loop, conn write err: %v, sessionId: %v", err, session.sessionId)
				c.closeConn(session, kcp.EnetServerKick)
				return
			}
		}
		_ = conn.SetWriteDeadline(time.Now().Add(time.Second * ConnSendTimeout))
		_, err := conn.Write(bin)
		if err != nil {
			logger.Debug("exit send loop, conn write err: %v, sessionId: %v", err, session.sessionId)
			c.closeConn(session, kcp.EnetServerKick)
			return
		}
		// 发包频率限制
		pktFreqLimitCounter++
		if pktFreqLimitCounter > SendPacketFreqLimit {
			now := time.Now().UnixNano()
			if now-pktFreqLimitTimer > int64(time.Second) {
				pktFreqLimitCounter = 0
				pktFreqLimitTimer = now
			} else {
				logger.Error("exit send loop, server packet send freq too high, sessionId: %v, pps: %v",
					session.sessionId, pktFreqLimitCounter)
				c.closeConn(session, kcp.EnetPacketFreqTooHigh)
				return
			}
		}
		if session.changeXorKeyFin == false && protoMsg.CmdId == cmd.GetPlayerTokenRsp {
			// XOR密钥切换
			logger.Info("change session xor key, sessionId: %v", session.sessionId)
			session.changeXorKeyFin = true
			keyBlock := random.NewKeyBlock(session.seed, session.useMagicSeed)
			xorKey := keyBlock.XorKey()
			key := make([]byte, 4096)
			copy(key, xorKey[:])
			session.xorKey = key
		}
		if isTcpConn {
			// tcp rtt探测
			now := time.Now().UnixMilli()
			if now-tcpConn.TCPRttLastSendTime > 1000 {
				_ = conn.SetWriteDeadline(time.Now().Add(time.Second * ConnSendTimeout))
				_, err := conn.Write([]byte{0xff, 0xff, 0xff, 0xff})
				if err != nil {
					logger.Debug("exit send loop, conn write err: %v, sessionId: %v", err, session.sessionId)
					c.closeConn(session, kcp.EnetServerKick)
					return
				}
				tcpConn.TCPRttLastSendTime = now
			}
		}
	}
}

// 关闭所有连接
func (c *ConnManager) closeAllConn() {
	sessionList := make([]*Session, 0)
	c.sessionMapLock.RLock()
	for _, session := range c.sessionMap {
		sessionList = append(sessionList, session)
	}
	c.sessionMapLock.RUnlock()
	for _, session := range sessionList {
		if session == nil {
			continue
		}
		c.closeConn(session, kcp.EnetServerShutdown)
	}
	logger.Info("all conn has been force close")
}

// 关闭指定连接
func (c *ConnManager) closeConnBySessionId(sessionId uint32, reason uint32) {
	session := c.GetSession(sessionId)
	if session == nil {
		logger.Error("session not exist, sessionId: %v", sessionId)
		return
	}
	c.closeConn(session, reason)
	logger.Info("conn has been close, sessionId: %v", sessionId)
}

// 关闭连接
func (c *ConnManager) closeConn(session *Session, enetType uint32) {
	ok := atomic.CompareAndSwapUint32(&(session.connState), atomic.LoadUint32(&(session.connState)), ConnClose)
	if !ok {
		return
	}
	logger.Info("[CLOSE] client disconnect, sessionId: %v, conv: %v, addr: %v",
		session.sessionId, session.conn.GetConv(), session.conn.RemoteAddr())
	// 清理数据
	c.DeleteSession(session.sessionId, session.userId)
	// 关闭连接
	_ = session.conn.Close(enetType)
	// 连接关闭通知
	c.connEventChan <- &ConnEvent{
		SessionId:    session.sessionId,
		EventId:      ConnEventClose,
		EventMessage: session.conn.RemoteAddr(),
	}
	// 通知GS玩家下线
	connCtrlMsg := new(mq.ConnCtrlMsg)
	connCtrlMsg.UserId = session.userId
	c.messageQueue.SendToGs(session.gsServerAppId, &mq.NetMsg{
		MsgType:     mq.MsgTypeConnCtrl,
		EventId:     mq.UserOfflineNotify,
		ConnCtrlMsg: connCtrlMsg,
	})
	logger.Info("send to gs user offline, sessionId: %v, uid: %v", session.sessionId, connCtrlMsg.UserId)
	c.destroySessionChan <- session
	atomic.AddInt32(&CLIENT_CONN_NUM, -1)
}

func (c *ConnManager) GetSession(sessionId uint32) *Session {
	c.sessionMapLock.RLock()
	session, _ := c.sessionMap[sessionId]
	c.sessionMapLock.RUnlock()
	return session
}

func (c *ConnManager) GetSessionByUserId(userId uint32) *Session {
	c.sessionMapLock.RLock()
	session, _ := c.sessionUserIdMap[userId]
	c.sessionMapLock.RUnlock()
	return session
}

func (c *ConnManager) SetSession(session *Session, sessionId uint32, userId uint32) {
	c.sessionMapLock.Lock()
	c.sessionMap[sessionId] = session
	c.sessionUserIdMap[userId] = session
	c.sessionMapLock.Unlock()
}

func (c *ConnManager) DeleteSession(sessionId uint32, userId uint32) {
	c.sessionMapLock.Lock()
	delete(c.sessionMap, sessionId)
	delete(c.sessionUserIdMap, userId)
	c.sessionMapLock.Unlock()
}

func (c *ConnManager) autoSyncGlobalGsOnlineMap() {
	ticker := time.NewTicker(time.Second * 60)
	for {
		<-ticker.C
		c.syncGlobalGsOnlineMap()
	}
}

func (c *ConnManager) syncGlobalGsOnlineMap() {
	rsp, err := c.discoveryClient.GetGlobalGsOnlineMap(context.TODO(), nil)
	if err != nil {
		logger.Error("get global gs online map error: %v", err)
		return
	}
	copyMap := make(map[uint32]string)
	for k, v := range rsp.OnlineMap {
		copyMap[k] = v
	}
	copyMapLen := len(copyMap)
	c.globalGsOnlineMapLock.Lock()
	c.globalGsOnlineMap = copyMap
	c.globalGsOnlineMapLock.Unlock()
	logger.Info("sync global gs online map finish, len: %v", copyMapLen)
}

func (c *ConnManager) autoSyncMinLoadServerAppid() {
	ticker := time.NewTicker(time.Second * 15)
	for {
		<-ticker.C
		c.syncMinLoadServerAppid()
	}
}

func (c *ConnManager) syncMinLoadServerAppid() {
	gsServerAppId, err := c.discoveryClient.GetServerAppId(context.TODO(), &api.GetServerAppIdReq{
		ServerType: api.GS,
	})
	if err != nil {
		logger.Error("get gs server appid error: %v", err)
	} else {
		c.minLoadGsServerAppId = gsServerAppId.AppId
	}

	multiServerAppId, err := c.discoveryClient.GetServerAppId(context.TODO(), &api.GetServerAppIdReq{
		ServerType: api.MULTI,
	})
	if err != nil {
		c.minLoadMultiServerAppId = ""
	} else {
		c.minLoadMultiServerAppId = multiServerAppId.AppId
	}
}

func (c *ConnManager) autoSyncStopServerInfo() {
	ticker := time.NewTicker(time.Minute * 1)
	for {
		<-ticker.C
		c.syncStopServerInfo()
	}
}

func (c *ConnManager) syncStopServerInfo() {
	stopServerInfo, err := c.discoveryClient.GetStopServerInfo(context.TODO(), &api.NullMsg{})
	if err != nil {
		logger.Error("get stop server info error: %v", err)
		return
	}
	c.stopServerInfo = stopServerInfo
}

func (c *ConnManager) autoSyncWhiteList() {
	ticker := time.NewTicker(time.Minute * 1)
	for {
		<-ticker.C
		c.syncWhiteList()
	}
}

func (c *ConnManager) syncWhiteList() {
	whiteList, err := c.discoveryClient.GetWhiteList(context.TODO(), &api.NullMsg{})
	if err != nil {
		logger.Error("get white list error: %v", err)
		return
	}
	c.whiteList = whiteList
}

// tcp模式连接对象兼容层

type Conn interface {
	GetSessionId() uint32
	GetConv() uint32
	Close(e uint32) error
	RemoteAddr() net.Addr
	SetReadDeadline(t time.Time) error
	Read(b []byte) (int, error)
	SetWriteDeadline(t time.Time) error
	Write(b []byte) (int, error)
	GetRTO() uint32
	GetSRTT() int32
	GetSRTTVar() int32
}

type TCPConn struct {
	TCPConn            *net.TCPConn
	TCPRtt             uint32
	TCPRttLastSendTime int64
}

func NewTcpConn(tcpConn *net.TCPConn) Conn {
	return &TCPConn{
		TCPConn:            tcpConn,
		TCPRtt:             0,
		TCPRttLastSendTime: 0,
	}
}

func (c *TCPConn) GetSessionId() uint32 {
	return 0
}

func (c *TCPConn) GetConv() uint32 {
	return 0
}

func (c *TCPConn) Close(e uint32) error {
	return c.TCPConn.Close()
}

func (c *TCPConn) RemoteAddr() net.Addr {
	return c.TCPConn.RemoteAddr()
}

func (c *TCPConn) SetReadDeadline(t time.Time) error {
	return c.TCPConn.SetReadDeadline(t)
}

func (c *TCPConn) Read(b []byte) (int, error) {
	return c.TCPConn.Read(b)
}

func (c *TCPConn) SetWriteDeadline(t time.Time) error {
	return c.TCPConn.SetWriteDeadline(t)
}

func (c *TCPConn) Write(b []byte) (int, error) {
	return c.TCPConn.Write(b)
}

func (c *TCPConn) GetRTO() uint32 {
	return 0
}

func (c *TCPConn) GetSRTT() int32 {
	return int32(c.TCPRtt)
}

func (c *TCPConn) GetSRTTVar() int32 {
	return 0
}
