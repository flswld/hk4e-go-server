package net

import (
	"sync"
	"sync/atomic"
	"time"

	"hk4e/common/config"

	"hk4e/gate/client_proto"
	"hk4e/gate/kcp"
	hk4egatenet "hk4e/gate/net"
	"hk4e/pkg/logger"
	"hk4e/protocol/cmd"
	"hk4e/protocol/proto"

	pb "google.golang.org/protobuf/proto"
)

type Session struct {
	Conn                   *kcp.UDPSession
	XorKey                 []byte
	SendChan               chan *hk4egatenet.ProtoMsg
	RecvChan               chan *hk4egatenet.ProtoMsg
	ServerCmdProtoMap      *cmd.CmdProtoMap
	ClientCmdProtoMap      *client_proto.ClientCmdProtoMap
	ClientSeq              uint32
	DeadEvent              chan struct{}
	ClientVersionRandomKey string
	SecurityCmdBuffer      []byte
	Uid                    uint32
	CloseOnce              sync.Once
}

func NewSession(gateAddr string, dispatchKey []byte) (*Session, error) {
	conn, err := kcp.DialKCP(gateAddr)
	if err != nil {
		logger.Error("kcp client conn to server error: %v", err)
		return nil, err
	}
	kcp.SetByteCheckMode(int(config.GetConfig().Hk4e.ByteCheckMode))
	conn.SetACKNoDelay(true)
	conn.SetWriteDelay(false)
	conn.SetWindowSize(256, 256)
	conn.SetMtu(1200)
	r := &Session{
		Conn:                   conn,
		XorKey:                 dispatchKey,
		SendChan:               make(chan *hk4egatenet.ProtoMsg, 1000),
		RecvChan:               make(chan *hk4egatenet.ProtoMsg, 1000),
		ServerCmdProtoMap:      cmd.NewCmdProtoMap(),
		ClientCmdProtoMap:      nil,
		ClientSeq:              0,
		DeadEvent:              make(chan struct{}),
		ClientVersionRandomKey: "",
		SecurityCmdBuffer:      nil,
		Uid:                    0,
	}
	if config.GetConfig().Hk4e.ClientProtoProxyEnable {
		r.ClientCmdProtoMap = client_proto.NewClientCmdProtoMap()
	}
	go r.recvHandle()
	go r.sendHandle()
	return r, nil
}

func (s *Session) SendMsg(cmdId uint16, msg pb.Message) {
	s.SendChan <- &hk4egatenet.ProtoMsg{
		SessionId: 0,
		CmdId:     cmdId,
		HeadMessage: &proto.PacketHead{
			ClientSequenceId: atomic.AddUint32(&s.ClientSeq, 1),
			SentMs:           uint64(time.Now().UnixMilli()),
			EnetIsReliable:   1,
		},
		PayloadMessage: msg,
	}
}

func (s *Session) Close() {
	s.CloseOnce.Do(func() {
		_ = s.Conn.Close(kcp.EnetClientClose)
		close(s.DeadEvent)
	})
}

func (s *Session) recvHandle() {
	logger.Info("recv handle start")
	conn := s.Conn
	convId := conn.GetConv()
	recvBuf := make([]byte, hk4egatenet.PacketMaxLen)
	for {
		_ = conn.SetReadDeadline(time.Now().Add(time.Second * hk4egatenet.ConnRecvTimeout))
		recvLen, err := conn.Read(recvBuf)
		if err != nil {
			logger.Error("exit recv loop, conn read err: %v, convId: %v", err, convId)
			s.Close()
			break
		}
		recvData := recvBuf[:recvLen]
		kcpMsgList := make([]*hk4egatenet.KcpMsg, 0)
		hk4egatenet.DecodeBinToPayload(recvData, convId, &kcpMsgList, s.XorKey)
		for _, v := range kcpMsgList {
			protoMsgList := hk4egatenet.ProtoDecode(v, s.ServerCmdProtoMap, s.ClientCmdProtoMap)
			for _, vv := range protoMsgList {
				s.RecvChan <- vv
			}
		}
	}
}

func (s *Session) sendHandle() {
	logger.Info("send handle start")
	conn := s.Conn
	convId := conn.GetConv()
	for {
		protoMsg, ok := <-s.SendChan
		if !ok {
			logger.Error("exit send loop, send chan close, convId: %v", convId)
			s.Close()
			break
		}
		kcpMsg := hk4egatenet.ProtoEncode(protoMsg, s.ServerCmdProtoMap, s.ClientCmdProtoMap)
		if kcpMsg == nil {
			logger.Error("decode kcp msg is nil, convId: %v", convId)
			continue
		}
		bin := hk4egatenet.EncodePayloadToBin(kcpMsg, s.XorKey)
		_ = conn.SetWriteDeadline(time.Now().Add(time.Second * hk4egatenet.ConnSendTimeout))
		_, err := conn.Write(bin)
		if err != nil {
			logger.Error("exit send loop, conn write err: %v, convId: %v", err, convId)
			s.Close()
			break
		}
	}
}
