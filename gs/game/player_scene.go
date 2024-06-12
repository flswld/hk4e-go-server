package game

import (
	"hk4e/pkg/alg"
	"math"
	"strconv"
	"time"

	"hk4e/common/constant"
	"hk4e/gdconf"
	"hk4e/gs/model"
	"hk4e/pkg/logger"
	"hk4e/pkg/object"
	"hk4e/pkg/random"
	"hk4e/protocol/cmd"
	"hk4e/protocol/proto"

	pb "google.golang.org/protobuf/proto"
)

// 场景模块 场景组 小组 实体 管理相关

const (
	ENTITY_MAX_BATCH_SEND_NUM = 1000 // 单次同步客户端的最大实体数量
)

/************************************************** 接口请求 **************************************************/

// EnterSceneReadyReq 准备进入场景
func (g *Game) EnterSceneReadyReq(player *model.Player, payloadMsg pb.Message) {
	req := payloadMsg.(*proto.EnterSceneReadyReq)
	logger.Debug("player enter scene ready, uid: %v", player.PlayerId)
	world := WORLD_MANAGER.GetWorldById(player.WorldId)
	if world == nil {
		logger.Error("get world is nil, worldId: %v, uid: %v", player.WorldId, player.PlayerId)
		return
	}

	if world.IsMultiplayerWorld() && world.IsPlayerFirstEnter(player) {
		playerPreEnterMpNotify := &proto.PlayerPreEnterMpNotify{
			State:    proto.PlayerPreEnterMpNotify_START,
			Uid:      player.PlayerId,
			Nickname: player.NickName,
		}
		g.SendToWorldH(world, cmd.PlayerPreEnterMpNotify, 0, playerPreEnterMpNotify)
	}

	ctx := world.GetEnterSceneContextByToken(req.EnterSceneToken)
	if ctx == nil {
		logger.Error("get enter scene context is nil, uid: %v", player.PlayerId)
		return
	}
	if ctx.OldSceneId != 0 {
		oldSceneId := ctx.OldSceneId
		oldPos := ctx.OldPos
		newSceneId := ctx.NewSceneId
		newPos := ctx.NewPos
		newRot := ctx.NewRot

		oldScene := world.GetSceneById(oldSceneId)
		newScene := world.GetSceneById(newSceneId)

		delEntityIdList := make([]uint32, 0)
		for entityId := range g.GetVisionEntity(oldScene, oldPos) {
			delEntityIdList = append(delEntityIdList, entityId)
		}
		g.RemoveSceneEntityNotifyToPlayer(player, proto.VisionType_VISION_MISS, delEntityIdList)

		activeAvatarEntity := world.GetPlayerActiveAvatarEntity(player)
		g.RemoveSceneEntityNotifyBroadcast(oldScene, proto.VisionType_VISION_REMOVE, []uint32{activeAvatarEntity.GetId()}, player.PlayerId)

		if !WORLD_MANAGER.IsAiWorld(world) {
			// 卸载旧位置附近的group
			otherPlayerNeighborGroupMap := make(map[uint32]*gdconf.Group)
			for _, otherPlayer := range oldScene.GetAllPlayer() {
				otherPlayerPos := g.GetPlayerPos(otherPlayer)
				for k, v := range g.GetNeighborGroup(oldSceneId, otherPlayerPos) {
					otherPlayerNeighborGroupMap[k] = v
				}
			}
			for _, groupConfig := range g.GetNeighborGroup(oldSceneId, oldPos) {
				if !world.IsMultiplayerWorld() {
					// 单人世界直接卸载group
					g.RemoveSceneGroup(player, oldScene, groupConfig)
				} else {
					// 多人世界group附近没有任何玩家则卸载
					_, exist := otherPlayerNeighborGroupMap[uint32(groupConfig.Id)]
					if !exist {
						g.RemoveSceneGroup(player, oldScene, groupConfig)
					}
				}
			}
		}

		player.SceneLoadState = model.SceneNone

		if player.SceneJump {
			oldScene.RemovePlayer(player)
			player.SetSceneId(newSceneId)
			player.SetPos(newPos)
			player.SetRot(newRot)
			newScene.AddPlayer(player)
		} else {
			player.SetPos(newPos)
			player.SetRot(newRot)
			for _, worldAvatar := range world.GetPlayerWorldAvatarList(player) {
				entityId := worldAvatar.GetAvatarEntityId()
				entity := oldScene.GetEntity(entityId)
				entity.SetPos(newPos)
				entity.SetRot(newRot)
			}
		}

		ntf := &proto.EnterScenePeerNotify{
			DestSceneId:     player.GetSceneId(),
			PeerId:          world.GetPlayerPeerId(player),
			HostPeerId:      world.GetPlayerPeerId(world.GetOwner()),
			EnterSceneToken: req.EnterSceneToken,
		}
		rsp := &proto.EnterSceneReadyRsp{
			EnterSceneToken: req.EnterSceneToken,
		}
		wait := g.LoadSceneBlockAsync(player, oldScene, newScene, oldPos, newPos, "EnterSceneReadyReq", &SceneBlockLoadInfoCtx{
			EnterScenePeerNotify: ntf,
			EnterSceneReadyRsp:   rsp,
		})
		if wait {
			return
		}
	}

	ntf := &proto.EnterScenePeerNotify{
		DestSceneId:     player.GetSceneId(),
		PeerId:          world.GetPlayerPeerId(player),
		HostPeerId:      world.GetPlayerPeerId(world.GetOwner()),
		EnterSceneToken: req.EnterSceneToken,
	}
	g.SendMsg(cmd.EnterScenePeerNotify, player.PlayerId, player.ClientSeq, ntf)

	rsp := &proto.EnterSceneReadyRsp{
		EnterSceneToken: req.EnterSceneToken,
	}
	g.SendMsg(cmd.EnterSceneReadyRsp, player.PlayerId, player.ClientSeq, rsp)
}

// SceneInitFinishReq 场景初始化完成
func (g *Game) SceneInitFinishReq(player *model.Player, payloadMsg pb.Message) {
	req := payloadMsg.(*proto.SceneInitFinishReq)
	logger.Debug("player scene init finish, uid: %v", player.PlayerId)
	world := WORLD_MANAGER.GetWorldById(player.WorldId)
	if world == nil {
		logger.Error("get world is nil, worldId: %v, uid: %v", player.WorldId, player.PlayerId)
		return
	}
	scene := world.GetSceneById(player.GetSceneId())

	if world.IsMultiplayerWorld() && world.IsPlayerFirstEnter(player) {
		guestBeginEnterSceneNotify := &proto.GuestBeginEnterSceneNotify{
			SceneId: player.GetSceneId(),
			Uid:     player.PlayerId,
		}
		g.SendToWorldA(world, cmd.GuestBeginEnterSceneNotify, 0, guestBeginEnterSceneNotify, player.PlayerId)
	}

	serverTimeNotify := &proto.ServerTimeNotify{
		ServerTime: uint64(time.Now().UnixMilli()),
	}
	g.SendMsg(cmd.ServerTimeNotify, player.PlayerId, player.ClientSeq, serverTimeNotify)

	if player.SceneJump {
		worldPlayerInfoNotify := &proto.WorldPlayerInfoNotify{
			PlayerInfoList: make([]*proto.OnlinePlayerInfo, 0),
			PlayerUidList:  make([]uint32, 0),
		}
		for _, worldPlayer := range world.GetAllPlayer() {
			onlinePlayerInfo := &proto.OnlinePlayerInfo{
				Uid:                 worldPlayer.PlayerId,
				Nickname:            worldPlayer.NickName,
				PlayerLevel:         worldPlayer.PropMap[constant.PLAYER_PROP_PLAYER_LEVEL],
				AvatarId:            worldPlayer.HeadImage,
				MpSettingType:       proto.MpSettingType(worldPlayer.PropMap[constant.PLAYER_PROP_PLAYER_MP_SETTING_TYPE]),
				NameCardId:          worldPlayer.GetDbSocial().NameCard,
				Signature:           worldPlayer.Signature,
				ProfilePicture:      &proto.ProfilePicture{AvatarId: worldPlayer.HeadImage},
				CurPlayerNumInWorld: uint32(world.GetWorldPlayerNum()),
			}
			worldPlayerInfoNotify.PlayerInfoList = append(worldPlayerInfoNotify.PlayerInfoList, onlinePlayerInfo)
			worldPlayerInfoNotify.PlayerUidList = append(worldPlayerInfoNotify.PlayerUidList, worldPlayer.PlayerId)
		}
		g.SendMsg(cmd.WorldPlayerInfoNotify, player.PlayerId, player.ClientSeq, worldPlayerInfoNotify)

		worldDataNotify := &proto.WorldDataNotify{
			WorldPropMap: make(map[uint32]*proto.PropValue),
		}
		// 世界等级
		worldDataNotify.WorldPropMap[1] = g.PacketPropValue(1, world.GetWorldLevel())
		// 是否多人游戏
		worldDataNotify.WorldPropMap[2] = g.PacketPropValue(2, object.ConvBoolToInt64(world.IsMultiplayerWorld()))
		g.SendMsg(cmd.WorldDataNotify, player.PlayerId, player.ClientSeq, worldDataNotify)

		// TODO 暂时先解锁全部场景和场景标签 看着喜庆
		playerWorldSceneInfoListNotify := &proto.PlayerWorldSceneInfoListNotify{
			InfoList: []*proto.PlayerWorldSceneInfo{
				{SceneId: 1, IsLocked: false, SceneTagIdList: []uint32{}},
				{SceneId: 3, IsLocked: false, SceneTagIdList: []uint32{}},
				{SceneId: 4, IsLocked: false, SceneTagIdList: []uint32{}},
				{SceneId: 5, IsLocked: false, SceneTagIdList: []uint32{}},
				{SceneId: 6, IsLocked: false, SceneTagIdList: []uint32{}},
				{SceneId: 7, IsLocked: false, SceneTagIdList: []uint32{}},
				{SceneId: 9, IsLocked: false, SceneTagIdList: []uint32{}},
			},
		}
		for _, info := range playerWorldSceneInfoListNotify.InfoList {
			for _, sceneTagDataConfig := range gdconf.GetSceneTagDataMap() {
				if uint32(sceneTagDataConfig.SceneId) == info.SceneId {
					info.SceneTagIdList = append(info.SceneTagIdList, uint32(sceneTagDataConfig.SceneTagId))
				}
			}
		}
		g.SendMsg(cmd.PlayerWorldSceneInfoListNotify, player.PlayerId, player.ClientSeq, playerWorldSceneInfoListNotify)

		g.SendMsg(cmd.SceneForceUnlockNotify, player.PlayerId, player.ClientSeq, new(proto.SceneForceUnlockNotify))

		hostPlayerNotify := &proto.HostPlayerNotify{
			HostUid:    world.GetOwner().PlayerId,
			HostPeerId: world.GetPlayerPeerId(world.GetOwner()),
		}
		g.SendMsg(cmd.HostPlayerNotify, player.PlayerId, player.ClientSeq, hostPlayerNotify)

		sceneTimeNotify := &proto.SceneTimeNotify{
			SceneId:   player.GetSceneId(),
			SceneTime: uint64(scene.GetSceneTime()),
		}
		g.SendMsg(cmd.SceneTimeNotify, player.PlayerId, player.ClientSeq, sceneTimeNotify)

		playerGameTimeNotify := &proto.PlayerGameTimeNotify{
			GameTime: world.GetGameTime(),
			Uid:      player.PlayerId,
		}
		g.SendMsg(cmd.PlayerGameTimeNotify, player.PlayerId, player.ClientSeq, playerGameTimeNotify)

		playerEnterSceneInfoNotify := &proto.PlayerEnterSceneInfoNotify{
			CurAvatarEntityId: world.GetPlayerActiveAvatarEntity(player).GetId(),
			EnterSceneToken:   req.EnterSceneToken,
			TeamEnterInfo: &proto.TeamEnterSceneInfo{
				TeamEntityId:        world.GetPlayerTeamEntityId(player),
				TeamAbilityInfo:     new(proto.AbilitySyncStateInfo),
				AbilityControlBlock: g.PacketTeamAbilityControlBlock(),
			},
			MpLevelEntityInfo: &proto.MPLevelEntityInfo{
				EntityId:        world.GetMpLevelEntityId(),
				AuthorityPeerId: world.GetPlayerPeerId(world.GetOwner()),
				AbilityInfo:     new(proto.AbilitySyncStateInfo),
			},
			AvatarEnterInfo: make([]*proto.AvatarEnterSceneInfo, 0),
		}
		dbAvatar := player.GetDbAvatar()
		for _, worldAvatar := range world.GetPlayerWorldAvatarList(player) {
			avatar := dbAvatar.GetAvatarById(worldAvatar.GetAvatarId())
			avatarEnterSceneInfo := &proto.AvatarEnterSceneInfo{
				AvatarGuid:     avatar.Guid,
				AvatarEntityId: world.GetPlayerWorldAvatarEntityId(player, worldAvatar.GetAvatarId()),
				WeaponGuid:     avatar.EquipWeapon.Guid,
				WeaponEntityId: world.GetPlayerWorldAvatarWeaponEntityId(player, worldAvatar.GetAvatarId()),
				AvatarAbilityInfo: &proto.AbilitySyncStateInfo{
					IsInited:           len(worldAvatar.PacketAbilityList()) != 0,
					DynamicValueMap:    nil,
					AppliedAbilities:   worldAvatar.PacketAbilityList(),
					AppliedModifiers:   worldAvatar.PacketModifierList(),
					MixinRecoverInfos:  nil,
					SgvDynamicValueMap: nil,
				},
				WeaponAbilityInfo: new(proto.AbilitySyncStateInfo),
			}
			playerEnterSceneInfoNotify.AvatarEnterInfo = append(playerEnterSceneInfoNotify.AvatarEnterInfo, avatarEnterSceneInfo)
		}
		g.SendMsg(cmd.PlayerEnterSceneInfoNotify, player.PlayerId, player.ClientSeq, playerEnterSceneInfoNotify)
	}

	// 天气未初始化 或 未锁定天气并且是场景跳跃 更新天气
	if player.WeatherInfo.WeatherAreaId == 0 || player.PropMap[constant.PLAYER_PROP_IS_WEATHER_LOCKED] == 0 {
		// 初始化天气区域id
		weatherAreaId := g.GetPlayerInWeatherAreaId(player, player.GetPos())
		if weatherAreaId == 0 {
			logger.Error("weather area id error, weatherAreaId: %v", weatherAreaId)
			return
		}
		// 获取天气气象
		climateType := g.GetWeatherAreaClimate(weatherAreaId)
		g.SetPlayerWeather(player, weatherAreaId, climateType, false)
	}

	g.UpdateWorldScenePlayerInfo(player, world)

	ctx := world.GetEnterSceneContextByToken(req.EnterSceneToken)
	if ctx == nil {
		logger.Error("get enter scene context is nil, uid: %v", player.PlayerId)
		return
	}
	// 进入的场景是地牢副本发送相关的包
	if ctx.OldDungeonPointId != 0 {
		g.GCGTavernInit(player) // GCG酒馆信息通知
		g.SendMsg(cmd.DungeonWayPointNotify, player.PlayerId, player.ClientSeq, &proto.DungeonWayPointNotify{})
		g.SendMsg(cmd.DungeonDataNotify, player.PlayerId, player.ClientSeq, &proto.DungeonDataNotify{})
	}

	if player.SceneEnterReason == uint32(proto.EnterReason_ENTER_REASON_REVIVAL) {
		for _, worldAvatar := range world.GetPlayerWorldAvatarList(player) {
			dbAvatar := player.GetDbAvatar()
			avatar := dbAvatar.GetAvatarById(worldAvatar.GetAvatarId())
			if avatar == nil {
				logger.Error("get avatar is nil, avatarId: %v", worldAvatar.GetAvatarId())
				continue
			}
			if avatar.LifeState != constant.LIFE_STATE_DEAD {
				continue
			}
			g.RevivePlayerAvatar(player, worldAvatar.GetAvatarId())
		}
	}

	rsp := &proto.SceneInitFinishRsp{
		EnterSceneToken: req.EnterSceneToken,
	}
	g.SendMsg(cmd.SceneInitFinishRsp, player.PlayerId, player.ClientSeq, rsp)

	player.SceneLoadState = model.SceneInitFinish
}

// EnterSceneDoneReq 进入场景完成
func (g *Game) EnterSceneDoneReq(player *model.Player, payloadMsg pb.Message) {
	req := payloadMsg.(*proto.EnterSceneDoneReq)
	logger.Debug("player enter scene done, uid: %v", player.PlayerId)
	world := WORLD_MANAGER.GetWorldById(player.WorldId)
	if world == nil {
		logger.Error("get world is nil, worldId: %v, uid: %v", player.WorldId, player.PlayerId)
		return
	}
	scene := world.GetSceneById(player.GetSceneId())

	var visionType = proto.VisionType_VISION_NONE
	if player.SceneJump {
		visionType = proto.VisionType_VISION_BORN
	} else {
		visionType = proto.VisionType_VISION_TRANSPORT
	}

	activeAvatarId := world.GetPlayerActiveAvatarId(player)
	activeWorldAvatar := world.GetPlayerWorldAvatar(player, activeAvatarId)

	pos := player.GetPos()

	if WORLD_MANAGER.IsAiWorld(world) {
		aiWorldAoi := world.GetAiWorldAoi()
		logger.Debug("ai world aoi add player, newPos: %+v, uid: %v", pos, player.PlayerId)
		ok := aiWorldAoi.AddObjectToGridByPos(int64(player.PlayerId), activeWorldAvatar, float32(pos.X), float32(pos.Y), float32(pos.Z))
		if !ok {
			logger.Error("ai world aoi add player fail, uid: %v, pos: %+v", player.PlayerId, pos)
		}
	}

	g.AddSceneEntityNotify(player, visionType, []uint32{activeWorldAvatar.GetAvatarEntityId()}, true, false)

	if !WORLD_MANAGER.IsAiWorld(world) {
		// 加载附近的group
		for _, groupConfig := range g.GetNeighborGroup(scene.GetId(), pos) {
			g.AddSceneGroup(player, scene, groupConfig)
		}
		for _, triggerDataConfig := range gdconf.GetTriggerDataMap() {
			groupConfig := gdconf.GetSceneGroup(triggerDataConfig.GroupId)
			if groupConfig != nil {
				g.AddSceneGroup(player, scene, groupConfig)
			}
		}
	}

	// 同步客户端视野内的场景实体
	visionEntityMap := g.GetVisionEntity(scene, pos)
	entityIdList := make([]uint32, 0)
	for entityId, entity := range visionEntityMap {
		if WORLD_MANAGER.IsAiWorld(world) {
			_, ok := entity.(*AvatarEntity)
			if ok {
				continue
			}
		}
		entityIdList = append(entityIdList, entityId)
	}
	g.AddSceneEntityNotify(player, visionType, entityIdList, false, false)

	if WORLD_MANAGER.IsAiWorld(world) {
		aiWorldAoi := world.GetAiWorldAoi()
		otherWorldAvatarMap := aiWorldAoi.GetObjectListByPos(float32(pos.X), float32(pos.Y), float32(pos.Z), 1)
		entityIdList := make([]uint32, 0)
		for _, otherWorldAvatarAny := range otherWorldAvatarMap {
			otherWorldAvatar := otherWorldAvatarAny.(*WorldAvatar)
			entityIdList = append(entityIdList, otherWorldAvatar.GetAvatarEntityId())
		}
		g.AddSceneEntityNotify(player, visionType, entityIdList, false, false)
	}

	// 设置玩家天气
	sceneAreaWeatherNotify := &proto.SceneAreaWeatherNotify{
		WeatherAreaId: player.WeatherInfo.WeatherAreaId,
		ClimateType:   player.WeatherInfo.ClimateType,
	}
	g.SendMsg(cmd.SceneAreaWeatherNotify, player.PlayerId, player.ClientSeq, sceneAreaWeatherNotify)

	rsp := &proto.EnterSceneDoneRsp{
		EnterSceneToken: req.EnterSceneToken,
	}
	g.SendMsg(cmd.EnterSceneDoneRsp, player.PlayerId, player.ClientSeq, rsp)

	player.SceneLoadState = model.SceneEnterDone

	for _, otherPlayerId := range world.GetAllWaitPlayer() {
		// 房主第一次进入多人世界场景完成 开始通知等待列表中的玩家进入场景
		world.RemoveWaitPlayer(otherPlayerId)
		otherPlayer := USER_MANAGER.GetOnlineUser(otherPlayerId)
		if otherPlayer == nil {
			logger.Error("player is nil, uid: %v", otherPlayerId)
			continue
		}
		g.JoinOtherWorld(otherPlayer, player)
	}
}

// PostEnterSceneReq 进入场景后
func (g *Game) PostEnterSceneReq(player *model.Player, payloadMsg pb.Message) {
	req := payloadMsg.(*proto.PostEnterSceneReq)
	logger.Debug("player post enter scene, uid: %v", player.PlayerId)
	world := WORLD_MANAGER.GetWorldById(player.WorldId)
	if world == nil {
		logger.Error("get world is nil, worldId: %v, uid: %v", player.WorldId, player.PlayerId)
		return
	}

	if world.IsMultiplayerWorld() && world.IsPlayerFirstEnter(player) {
		guestPostEnterSceneNotify := &proto.GuestPostEnterSceneNotify{
			SceneId: player.GetSceneId(),
			Uid:     player.PlayerId,
		}
		g.SendToWorldA(world, cmd.GuestPostEnterSceneNotify, 0, guestPostEnterSceneNotify, player.PlayerId)
	}

	world.PlayerEnter(player.PlayerId)

	rsp := &proto.PostEnterSceneRsp{
		EnterSceneToken: req.EnterSceneToken,
	}
	g.SendMsg(cmd.PostEnterSceneRsp, player.PlayerId, player.ClientSeq, rsp)

	// 触发事件
	if PLUGIN_MANAGER.TriggerEvent(PluginEventIdPostEnterScene, &PluginEventPostEnterScene{
		PluginEvent: NewPluginEvent(),
		Player:      player,
		Req:         req,
	}) {
		return
	}
}

func (g *Game) SceneEntityDrownReq(player *model.Player, payloadMsg pb.Message) {
	req := payloadMsg.(*proto.SceneEntityDrownReq)

	world := WORLD_MANAGER.GetWorldById(player.WorldId)
	if world == nil {
		return
	}
	scene := world.GetSceneById(player.GetSceneId())
	g.KillEntity(player, scene, req.EntityId, proto.PlayerDieType_PLAYER_DIE_DRAWN)

	rsp := &proto.SceneEntityDrownRsp{
		EntityId: req.EntityId,
	}
	g.SendMsg(cmd.SceneEntityDrownRsp, player.PlayerId, player.ClientSeq, rsp)
}

func (g *Game) EntityForceSyncReq(player *model.Player, payloadMsg pb.Message) {
	req := payloadMsg.(*proto.EntityForceSyncReq)

	world := WORLD_MANAGER.GetWorldById(player.WorldId)
	if world == nil {
		return
	}
	scene := world.GetSceneById(player.GetSceneId())

	motionInfo := req.MotionInfo
	if motionInfo == nil {
		return
	}
	g.handleEntityMove(player, world, scene, req.EntityId, &model.Vector{
		X: float64(motionInfo.Pos.X),
		Y: float64(motionInfo.Pos.Y),
		Z: float64(motionInfo.Pos.Z),
	}, &model.Vector{
		X: float64(motionInfo.Rot.X),
		Y: float64(motionInfo.Rot.Y),
		Z: float64(motionInfo.Rot.Z),
	}, true, nil)

	rsp := &proto.EntityForceSyncRsp{
		SceneTime:  req.SceneTime,
		EntityId:   req.EntityId,
		FailMotion: motionInfo,
	}
	g.SendMsg(cmd.EntityForceSyncRsp, player.PlayerId, player.ClientSeq, rsp)
}

/************************************************** 游戏功能 **************************************************/

type SceneBlockLoadInfo struct {
	Uid            uint32
	SceneBlockList []*model.SceneBlock
	Origin         string
	Ctx            *SceneBlockLoadInfoCtx
}

type SceneBlockLoadInfoCtx struct {
	World                *World
	Scene                *Scene
	OldPos               *model.Vector
	NewPos               *model.Vector
	AvatarEntityId       uint32
	EnterScenePeerNotify *proto.EnterScenePeerNotify
	EnterSceneReadyRsp   *proto.EnterSceneReadyRsp
}

// LoadSceneBlockAsync 异步加载场景区块存档
func (g *Game) LoadSceneBlockAsync(player *model.Player, oldScene *Scene, newScene *Scene, oldPos *model.Vector, newPos *model.Vector, origin string, ctx *SceneBlockLoadInfoCtx) bool {
	oldSceneBlockAoi := WORLD_MANAGER.GetSceneBlockAoiMap()[oldScene.GetId()]
	if oldSceneBlockAoi == nil {
		logger.Error("scene not exist in aoi, sceneId: %v", oldScene.GetId())
		return false
	}
	newSceneBlockAoi := WORLD_MANAGER.GetSceneBlockAoiMap()[newScene.GetId()]
	if newSceneBlockAoi == nil {
		logger.Error("scene not exist in aoi, sceneId: %v", newScene.GetId())
		return false
	}
	oldGid := oldSceneBlockAoi.GetGidByPos(float32(oldPos.X), 0.0, float32(oldPos.Z))
	newGid := newSceneBlockAoi.GetGidByPos(float32(newPos.X), 0.0, float32(newPos.Z))
	if oldGid == newGid {
		return false
	}
	// 跨越了block格子
	logger.Debug("player cross scene block grid, oldGid: %v, newGid: %v, uid: %v", oldGid, newGid, player.PlayerId)
	if player.SceneBlockAsyncLoad {
		return false
	}
	logger.Info("async load player scene block from db, uid: %v", player.PlayerId)
	oldGridList := oldSceneBlockAoi.GetSurrGridListByGid(oldGid, 1)
	newGridList := newSceneBlockAoi.GetSurrGridListByGid(newGid, 1)
	delGridIdList := make([]uint32, 0)
	for _, oldGrid := range oldGridList {
		exist := false
		for _, newGrid := range newGridList {
			if oldGrid.GetGid() == newGrid.GetGid() {
				exist = true
				break
			}
		}
		if exist {
			continue
		}
		delGridIdList = append(delGridIdList, oldGrid.GetGid())
	}
	addGridIdList := make([]uint32, 0)
	for _, newGrid := range newGridList {
		exist := false
		for _, oldGrid := range oldGridList {
			if newGrid.GetGid() == oldGrid.GetGid() {
				exist = true
				break
			}
		}
		if exist {
			continue
		}
		addGridIdList = append(addGridIdList, newGrid.GetGid())
	}
	loadSceneBlockMap := make(map[uint32]*gdconf.Block)
	for _, addGridId := range addGridIdList {
		for _, blockAny := range newSceneBlockAoi.GetObjectListByGid(addGridId) {
			block := blockAny.(*gdconf.Block)
			_, exist := player.SceneBlockMap[uint32(block.Id)]
			if exist {
				continue
			}
			loadSceneBlockMap[uint32(block.Id)] = block
		}
	}
	if len(loadSceneBlockMap) == 0 {
		return false
	}
	player.SceneBlockAsyncLoad = true
	go func() {
		loadSceneBlockList := make([]*model.SceneBlock, 0)
		for _, block := range loadSceneBlockMap {
			sceneBlock, err := g.db.LoadSceneBlockByUidAndBlockId(player.PlayerId, uint32(block.Id))
			if err != nil {
				logger.Error("load scene block from db error: %v, uid: %v", err, player.PlayerId)
				continue
			}
			if sceneBlock == nil {
				sceneBlock = &model.SceneBlock{
					Uid:           player.PlayerId,
					BlockId:       uint32(block.Id),
					SceneGroupMap: make(map[uint32]*model.SceneGroup),
					IsNew:         true,
				}
			}
			loadSceneBlockList = append(loadSceneBlockList, sceneBlock)
		}
		LOCAL_EVENT_MANAGER.GetLocalEventChan() <- &LocalEvent{
			EventId: AsyncLoadSceneBlockFinish,
			Msg: &SceneBlockLoadInfo{
				Uid:            player.PlayerId,
				SceneBlockList: loadSceneBlockList,
				Origin:         origin,
				Ctx:            ctx,
			},
		}
	}()
	return true
}

func (g *Game) OnSceneBlockLoad(sceneBlockLoadInfo *SceneBlockLoadInfo) {
	player := USER_MANAGER.GetOnlineUser(sceneBlockLoadInfo.Uid)
	if player == nil {
		logger.Error("player is nil, uid: %v", sceneBlockLoadInfo.Uid)
		return
	}
	logger.Info("async load player scene block ok, uid: %v", player.PlayerId)
	for _, sceneBlock := range sceneBlockLoadInfo.SceneBlockList {
		player.SceneBlockMap[sceneBlock.BlockId] = sceneBlock
	}
	ctx := sceneBlockLoadInfo.Ctx
	switch sceneBlockLoadInfo.Origin {
	case "SceneBlockAoiPlayerMove":
		g.SceneBlockAoiPlayerMove(player, ctx.World, ctx.Scene, ctx.OldPos, ctx.NewPos, ctx.AvatarEntityId)
		entity := ctx.Scene.GetEntity(ctx.AvatarEntityId)
		entity.SetPos(ctx.NewPos)
	case "EnterSceneReadyReq":
		g.SendMsg(cmd.EnterScenePeerNotify, player.PlayerId, player.ClientSeq, ctx.EnterScenePeerNotify)
		g.SendMsg(cmd.EnterSceneReadyRsp, player.PlayerId, player.ClientSeq, ctx.EnterSceneReadyRsp)
	}
	player.SceneBlockAsyncLoad = false
}

func (g *Game) LoadSceneBlockSync(uid uint32, sceneId uint32, pos *model.Vector) map[uint32]*model.SceneBlock {
	sceneBlockAoi, exist := WORLD_MANAGER.GetSceneBlockAoiMap()[sceneId]
	if !exist {
		logger.Error("scene not exist in aoi, sceneId: %v", sceneId)
		return nil
	}
	loadSceneBlockMap := sceneBlockAoi.GetObjectListByPos(float32(pos.X), 0.0, float32(pos.Z), 1)
	sceneBlockMap := make(map[uint32]*model.SceneBlock)
	for _, blockAny := range loadSceneBlockMap {
		block := blockAny.(*gdconf.Block)
		sceneBlock, err := g.db.LoadSceneBlockByUidAndBlockId(uid, uint32(block.Id))
		if err != nil {
			logger.Error("load scene block from db error: %v, uid: %v", err, uid)
			return nil
		}
		if sceneBlock == nil {
			sceneBlock = &model.SceneBlock{
				Uid:           uid,
				BlockId:       uint32(block.Id),
				SceneGroupMap: make(map[uint32]*model.SceneGroup),
				IsNew:         true,
			}
		}
		sceneBlockMap[sceneBlock.BlockId] = sceneBlock
	}
	logger.Info("sync load player scene block ok, uid: %v", uid)
	return sceneBlockMap
}

func (g *Game) SaveSceneBlockSync(uid uint32, sceneBlockMap map[uint32]*model.SceneBlock) {
	for _, sceneBlock := range sceneBlockMap {
		err := g.db.SaveSceneBlock(sceneBlock)
		if err != nil {
			logger.Error("save scene block to db error: %v, uid: %v", err, uid)
			continue
		}
	}
	logger.Info("sync save player scene block ok, uid: %v", uid)
}

// AddSceneEntityNotifyToPlayer 添加的场景实体同步给玩家
func (g *Game) AddSceneEntityNotifyToPlayer(player *model.Player, visionType proto.VisionType, entityList []*proto.SceneEntityInfo) {
	ntf := &proto.SceneEntityAppearNotify{
		AppearType: visionType,
		EntityList: entityList,
	}
	// logger.Debug("[SceneEntityAppearNotify UC], type: %v, len: %v, uid: %v", ntf.AppearType, len(ntf.EntityList), player.PlayerId)
	g.SendMsg(cmd.SceneEntityAppearNotify, player.PlayerId, player.ClientSeq, ntf)
}

// AddSceneEntityNotifyBroadcast 添加的场景实体广播
func (g *Game) AddSceneEntityNotifyBroadcast(scene *Scene, visionType proto.VisionType, entityList []*proto.SceneEntityInfo, aecUid uint32) {
	ntf := &proto.SceneEntityAppearNotify{
		AppearType: visionType,
		EntityList: entityList,
	}
	world := scene.GetWorld()
	owner := world.GetOwner()
	// logger.Debug("[SceneEntityAppearNotify BC], type: %v, len: %v, uid: %v", ntf.AppearType, len(ntf.EntityList), owner.PlayerId)
	g.SendToSceneA(scene, cmd.SceneEntityAppearNotify, owner.ClientSeq, ntf, aecUid)
}

// RemoveSceneEntityNotifyToPlayer 移除的场景实体同步给玩家
func (g *Game) RemoveSceneEntityNotifyToPlayer(player *model.Player, visionType proto.VisionType, entityIdList []uint32) {
	ntf := &proto.SceneEntityDisappearNotify{
		EntityList:    entityIdList,
		DisappearType: visionType,
	}
	// logger.Debug("[SceneEntityDisappearNotify UC], type: %v, len: %v, uid: %v", ntf.DisappearType, len(ntf.EntityList), player.PlayerId)
	g.SendMsg(cmd.SceneEntityDisappearNotify, player.PlayerId, player.ClientSeq, ntf)
}

// RemoveSceneEntityNotifyBroadcast 移除的场景实体广播
func (g *Game) RemoveSceneEntityNotifyBroadcast(scene *Scene, visionType proto.VisionType, entityIdList []uint32, aecUid uint32) {
	ntf := &proto.SceneEntityDisappearNotify{
		EntityList:    entityIdList,
		DisappearType: visionType,
	}
	world := scene.GetWorld()
	owner := world.GetOwner()
	// logger.Debug("[SceneEntityDisappearNotify BC], type: %v, len: %v, uid: %v", ntf.DisappearType, len(ntf.EntityList), owner.PlayerId)
	g.SendToSceneA(scene, cmd.SceneEntityDisappearNotify, owner.ClientSeq, ntf, aecUid)
}

// AddSceneEntityNotify 添加的场景实体同步 封装接口
func (g *Game) AddSceneEntityNotify(player *model.Player, visionType proto.VisionType, entityIdList []uint32, broadcast bool, aec bool) {
	world := WORLD_MANAGER.GetWorldById(player.WorldId)
	if world == nil {
		return
	}
	scene := world.GetSceneById(player.GetSceneId())
	// 如果总数量太多则分包发送
	times := int(math.Ceil(float64(len(entityIdList)) / float64(ENTITY_MAX_BATCH_SEND_NUM)))
	for i := 0; i < times; i++ {
		begin := ENTITY_MAX_BATCH_SEND_NUM * i
		end := ENTITY_MAX_BATCH_SEND_NUM * (i + 1)
		if i == times-1 {
			end = len(entityIdList)
		}
		entityList := make([]*proto.SceneEntityInfo, 0)
		for _, entityId := range entityIdList[begin:end] {
			entity := scene.GetEntity(entityId)
			if entity == nil {
				logger.Error("get entity is nil, entityId: %v", entityId)
				continue
			}
			switch entity.(type) {
			case *AvatarEntity:
				avatarEntity := entity.(*AvatarEntity)
				scenePlayer := USER_MANAGER.GetOnlineUser(avatarEntity.GetUid())
				if scenePlayer == nil {
					logger.Error("get scene player is nil, world id: %v, scene id: %v", world.GetId(), scene.GetId())
					continue
				}
				sceneEntityInfoAvatar := g.PacketSceneEntityInfoAvatar(scene, scenePlayer, world.GetPlayerActiveAvatarId(scenePlayer))
				entityList = append(entityList, sceneEntityInfoAvatar)
			case *MonsterEntity:
				sceneEntityInfoMonster := g.PacketSceneEntityInfoMonster(scene, entity.GetId())
				entityList = append(entityList, sceneEntityInfoMonster)
			case *NpcEntity:
				sceneEntityInfoNpc := g.PacketSceneEntityInfoNpc(scene, entity.GetId())
				entityList = append(entityList, sceneEntityInfoNpc)
			case IGadgetEntity:
				sceneEntityInfoGadget := g.PacketSceneEntityInfoGadget(player, scene, entity.GetId())
				entityList = append(entityList, sceneEntityInfoGadget)
			default:
				logger.Error("not support entity: %v, stack: %v", entity, logger.Stack())
				continue
			}
		}
		if broadcast {
			if aec {
				g.AddSceneEntityNotifyBroadcast(scene, visionType, entityList, player.PlayerId)
			} else {
				g.AddSceneEntityNotifyBroadcast(scene, visionType, entityList, 0)
			}
		} else {
			g.AddSceneEntityNotifyToPlayer(player, visionType, entityList)
		}
	}
}

// EntityFightPropUpdateNotifyBroadcast 场景实体战斗属性变更通知广播
func (g *Game) EntityFightPropUpdateNotifyBroadcast(scene *Scene, entity IEntity) {
	ntf := &proto.EntityFightPropUpdateNotify{
		FightPropMap: entity.GetFightProp(),
		EntityId:     entity.GetId(),
	}
	g.SendToSceneA(scene, cmd.EntityFightPropUpdateNotify, 0, ntf, 0)
}

// KillPlayerAvatar 杀死玩家角色实体
func (g *Game) KillPlayerAvatar(player *model.Player, avatarId uint32, dieType proto.PlayerDieType) {
	dbAvatar := player.GetDbAvatar()
	avatar := dbAvatar.GetAvatarById(avatarId)
	if avatar == nil {
		logger.Error("get avatar is nil, avatarId: %v", avatarId)
		return
	}
	world := WORLD_MANAGER.GetWorldById(player.WorldId)
	if world == nil {
		return
	}
	worldAvatar := world.GetPlayerWorldAvatar(player, avatarId)
	if worldAvatar == nil {
		return
	}
	scene := world.GetSceneById(player.GetSceneId())
	entity := scene.GetEntity(worldAvatar.GetAvatarEntityId())

	avatar.LifeState = constant.LIFE_STATE_DEAD
	avatar.FightPropMap[constant.FIGHT_PROP_CUR_HP] = 0.0

	g.EntityFightPropUpdateNotifyBroadcast(scene, entity)

	activeAvatarId := world.GetPlayerActiveAvatarId(player)
	if avatarId == activeAvatarId {
		g.KillEntity(player, scene, entity.GetId(), dieType)
	} else {
		ntf := &proto.AvatarLifeStateChangeNotify{
			AvatarGuid:      avatar.Guid,
			LifeState:       uint32(avatar.LifeState),
			DieType:         dieType,
			MoveReliableSeq: entity.GetLastMoveReliableSeq(),
		}
		g.SendToWorldA(world, cmd.AvatarLifeStateChangeNotify, 0, ntf, 0)
	}
}

// RevivePlayerAvatar 复活玩家活跃角色实体
func (g *Game) RevivePlayerAvatar(player *model.Player, avatarId uint32) {
	world := WORLD_MANAGER.GetWorldById(player.WorldId)
	if world == nil {
		return
	}
	scene := world.GetSceneById(player.GetSceneId())

	dbAvatar := player.GetDbAvatar()
	avatar := dbAvatar.GetAvatarById(avatarId)
	if avatar == nil {
		logger.Error("get avatar is nil, avatarId: %v", avatarId)
		return
	}

	avatar.LifeState = constant.LIFE_STATE_ALIVE
	avatar.FightPropMap[constant.FIGHT_PROP_CUR_HP] = 100.0

	g.UpdatePlayerAvatarFightProp(player.PlayerId, avatarId)

	ntf := &proto.AvatarLifeStateChangeNotify{
		AvatarGuid:      avatar.Guid,
		LifeState:       uint32(avatar.LifeState),
		DieType:         proto.PlayerDieType_PLAYER_DIE_NONE,
		MoveReliableSeq: 0,
	}
	g.SendToWorldA(world, cmd.AvatarLifeStateChangeNotify, 0, ntf, 0)

	worldAvatar := world.GetPlayerWorldAvatar(player, avatarId)
	if worldAvatar == nil {
		return
	}
	entity := scene.GetEntity(worldAvatar.GetAvatarEntityId())
	entity.SetLifeState(constant.LIFE_STATE_ALIVE)
	entity.GetFightProp()[constant.FIGHT_PROP_CUR_HP] = 100.0
}

// KillEntity 杀死实体
func (g *Game) KillEntity(player *model.Player, scene *Scene, entityId uint32, dieType proto.PlayerDieType) {
	entity := scene.GetEntity(entityId)
	if entity == nil {
		return
	}
	// 设置血量
	entity.SetLastDieType(int32(dieType))
	entity.SetLifeState(constant.LIFE_STATE_DEAD)
	entity.GetFightProp()[constant.FIGHT_PROP_CUR_HP] = 0.0
	ntf := &proto.LifeStateChangeNotify{
		EntityId:        entity.GetId(),
		LifeState:       uint32(entity.GetLifeState()),
		DieType:         dieType,
		MoveReliableSeq: entity.GetLastMoveReliableSeq(),
	}
	g.SendToSceneA(scene, cmd.LifeStateChangeNotify, 0, ntf, 0)

	_, ok := entity.(*AvatarEntity)
	if ok {
		return
	}

	// 删除实体
	g.EntityFightPropUpdateNotifyBroadcast(scene, entity)
	g.RemoveSceneEntityNotifyBroadcast(scene, proto.VisionType_VISION_DIE, []uint32{entity.GetId()}, 0)
	scene.DestroyEntity(entity.GetId())
	group := scene.GetGroupById(entity.GetGroupId())
	if group == nil {
		return
	}

	world := scene.GetWorld()
	owner := world.GetOwner()
	sceneGroup := owner.GetSceneGroupById(entity.GetGroupId())
	if sceneGroup == nil {
		return
	}
	sceneGroup.AddKill(entity.GetConfigId())

	group.DestroyEntity(entity.GetId())

	switch entity.(type) {
	case *MonsterEntity:
		// 随机掉落
		g.monsterDrop(player, MonsterDropTypeKill, 0, entity)
		// 怪物死亡触发器检测
		g.MonsterDieTriggerCheck(player, group, entity)
	case IGadgetEntity:
		iGadgetEntity := entity.(IGadgetEntity)
		// 物件死亡触发器检测
		g.GadgetDieTriggerCheck(player, group, entity)
		gadgetDataConfig := gdconf.GetGadgetDataById(int32(iGadgetEntity.GetGadgetId()))
		if gadgetDataConfig == nil {
			logger.Error("get gadget data config is nil, gadgetId: %v", iGadgetEntity.GetGadgetId())
			return
		}
		if gadgetDataConfig.ServerLuaScript != "" {
			gadgetLuaConfig := gdconf.GetGadgetLuaConfigByName(gadgetDataConfig.ServerLuaScript)
			if gadgetLuaConfig == nil {
				logger.Error("get gadget lua config is nil, name: %v", gadgetDataConfig.ServerLuaScript)
				return
			}
			CallGadgetLuaFunc(gadgetLuaConfig.LuaState, "OnDie",
				&LuaCtx{uid: player.PlayerId, targetEntityId: entity.GetId(), groupId: entity.GetGroupId()},
				0, 0)
		}

	}
}

// ChangeGadgetState 改变物件状态
func (g *Game) ChangeGadgetState(player *model.Player, entityId uint32, state uint32) {
	world := WORLD_MANAGER.GetWorldById(player.WorldId)
	if world == nil {
		return
	}
	scene := world.GetSceneById(player.GetSceneId())
	entity := scene.GetEntity(entityId)
	if entity == nil {
		logger.Error("get entity is nil, entityId: %v", entityId)
		return
	}
	iGadgetEntity, ok := entity.(IGadgetEntity)
	if !ok {
		logger.Error("entity is not gadget, entityId: %v", entityId)
		return
	}
	iGadgetEntity.SetGadgetState(state)
	ntf := &proto.GadgetStateNotify{
		GadgetEntityId:   entity.GetId(),
		GadgetState:      iGadgetEntity.GetGadgetState(),
		IsEnableInteract: true,
	}
	g.SendMsg(cmd.GadgetStateNotify, player.PlayerId, player.ClientSeq, ntf)

	groupId := entity.GetGroupId()
	group := scene.GetGroupById(groupId)
	if group == nil {
		logger.Error("group not exist, groupId: %v, uid: %v", groupId, player.PlayerId)
		return
	}

	owner := world.GetOwner()
	sceneGroup := owner.GetSceneGroupById(groupId)
	if sceneGroup == nil {
		return
	}
	sceneGroup.ChangeGadgetState(entity.GetConfigId(), uint8(iGadgetEntity.GetGadgetState()))

	// 物件状态变更触发器检测
	g.GadgetStateChangeTriggerCheck(player, group, entity, iGadgetEntity.GetGadgetState())
}

// GetVisionEntity 获取某位置视野内的全部实体
func (g *Game) GetVisionEntity(scene *Scene, pos *model.Vector) map[uint32]IEntity {
	allEntityMap := scene.GetAllEntity()
	visionEntity := make(map[uint32]IEntity)
	for _, entity := range allEntityMap {
		if !g.IsInVision(pos, entity.GetPos(), entity.GetVisionLevel()) {
			continue
		}
		avatarEntity, ok := entity.(*AvatarEntity)
		if ok {
			scenePlayer := USER_MANAGER.GetOnlineUser(avatarEntity.GetUid())
			if scenePlayer == nil {
				logger.Error("get scene player is nil, target uid: %v", avatarEntity.GetUid())
				continue
			}
			if !scene.GetWorld().IsPlayerActiveAvatarEntity(scenePlayer, entity.GetId()) {
				continue
			}
		}
		_, ok = entity.(*WeaponEntity)
		if ok {
			continue
		}
		visionEntity[entity.GetId()] = entity
	}
	return visionEntity
}

func (g *Game) IsInVision(p1 *model.Vector, p2 *model.Vector, visionLevel int) bool {
	vision, exist := constant.VISION_LEVEL[visionLevel]
	if !exist {
		return false
	}
	dx := int32(p1.X) - int32(p2.X)
	if dx < 0 {
		dx *= -1
	}
	dy := int32(p1.Z) - int32(p2.Z)
	if dy < 0 {
		dy *= -1
	}
	if uint32(dx) > vision.VisionRange || uint32(dy) > vision.VisionRange {
		return false
	}
	return true
}

// GetNeighborGroup 获取某位置附近的场景组
func (g *Game) GetNeighborGroup(sceneId uint32, pos *model.Vector) map[uint32]*gdconf.Group {
	sceneEntityAoi, exist := WORLD_MANAGER.GetSceneEntityAoiMap()[sceneId]
	if !exist {
		logger.Error("scene not exist in aoi, sceneId: %v", sceneId)
		return nil
	}
	neighborGroup := make(map[uint32]*gdconf.Group)
	for visionLevel, aoiManager := range sceneEntityAoi {
		vision := constant.VISION_LEVEL[visionLevel]
		objectMap := aoiManager.GetObjectListByPos(float32(pos.X), 0.0, float32(pos.Z), vision.VisionRange/vision.GridWidth+1)
		for objectId, obj := range objectMap {
			var objPos *gdconf.Vector = nil
			switch obj.(type) {
			case *gdconf.Monster:
				monster := obj.(*gdconf.Monster)
				objPos = monster.Pos
			case *gdconf.Npc:
				npc := obj.(*gdconf.Npc)
				objPos = npc.Pos
			case *gdconf.Gadget:
				gadget := obj.(*gdconf.Gadget)
				objPos = gadget.Pos
			case *gdconf.Region:
				region := obj.(*gdconf.Region)
				objPos = region.Pos
			}
			dx := int32(pos.X) - int32(objPos.X)
			if dx < 0 {
				dx *= -1
			}
			dy := int32(pos.Z) - int32(objPos.Z)
			if dy < 0 {
				dy *= -1
			}
			if uint32(dx) > vision.VisionRange+vision.GridWidth || uint32(dy) > vision.VisionRange+vision.GridWidth {
				continue
			}
			groupId := int32(objectId >> 32)
			_, exist = neighborGroup[uint32(groupId)]
			if exist {
				continue
			}
			groupConfig := gdconf.GetSceneGroup(groupId)
			if groupConfig.DynamicLoad {
				continue
			}
			neighborGroup[uint32(groupConfig.Id)] = groupConfig
		}
	}
	return neighborGroup
}

// TODO Group和Suite的初始化和加载卸载逻辑还没完全理清 所以现在这里写得略答辩

// AddSceneGroup 加载场景组
func (g *Game) AddSceneGroup(player *model.Player, scene *Scene, groupConfig *gdconf.Group) {
	group := scene.GetGroupById(uint32(groupConfig.Id))
	if group != nil {
		return
	}
	initSuiteId := groupConfig.GroupInitConfig.Suite
	_, exist := groupConfig.SuiteMap[initSuiteId]
	if !exist {
		logger.Error("invalid suiteId: %v, uid: %v", initSuiteId, player.PlayerId)
		return
	}
	// logger.Debug("add scene group, groupId: %v, initSuiteId: %v, uid: %v", groupConfig.Id, initSuiteId, player.PlayerId)
	g.AddSceneGroupSuiteCore(player, scene, uint32(groupConfig.Id), uint8(initSuiteId))
	if len(groupConfig.NpcMap) > 0 {
		ntf := &proto.GroupSuiteNotify{
			GroupMap: make(map[uint32]uint32),
		}
		ntf.GroupMap[uint32(groupConfig.Id)] = uint32(initSuiteId)
		g.SendMsg(cmd.GroupSuiteNotify, player.PlayerId, player.ClientSeq, ntf)
	}

	world := scene.GetWorld()
	owner := world.GetOwner()
	sceneGroup := owner.GetSceneGroupById(uint32(groupConfig.Id))
	if sceneGroup == nil {
		return
	}
	for _, variable := range groupConfig.VariableMap {
		exist := sceneGroup.CheckVariableExist(variable.Name)
		if exist && variable.NoRefresh {
			continue
		}
		sceneGroup.SetVariable(variable.Name, variable.Value)
	}

	group = scene.GetGroupById(uint32(groupConfig.Id))
	if group == nil {
		logger.Error("group not exist, groupId: %v, uid: %v", groupConfig.Id, player.PlayerId)
		return
	}
	// 场景组加载触发器检测
	g.GroupLoadTriggerCheck(player, group)
}

// RemoveSceneGroup 卸载场景组
func (g *Game) RemoveSceneGroup(player *model.Player, scene *Scene, groupConfig *gdconf.Group) {
	// logger.Debug("remove scene group, groupId: %v, uid: %v", groupConfig.Id, player.PlayerId)
	for _, triggerData := range gdconf.GetTriggerDataMap() {
		if groupConfig.Id == triggerData.GroupId {
			return
		}
	}
	group := scene.GetGroupById(uint32(groupConfig.Id))
	if group == nil {
		// logger.Error("group not exist, groupId: %v, uid: %v", groupConfig.Id, player.PlayerId)
		return
	}
	for suiteId := range group.GetAllSuite() {
		scene.RemoveGroupSuite(uint32(groupConfig.Id), suiteId)
	}
	if len(groupConfig.NpcMap) > 0 {
		ntf := &proto.GroupUnloadNotify{
			GroupList: []uint32{uint32(groupConfig.Id)},
		}
		g.SendMsg(cmd.GroupUnloadNotify, player.PlayerId, player.ClientSeq, ntf)
	}
}

// AddSceneGroupSuite 场景组中添加场景小组
func (g *Game) AddSceneGroupSuite(player *model.Player, groupId uint32, suiteId uint8) {
	// logger.Debug("add scene group suite, groupId: %v, suiteId: %v, uid: %v", groupId, suiteId, player.PlayerId)
	groupConfig := gdconf.GetSceneGroup(int32(groupId))
	if groupConfig == nil {
		logger.Error("get group config is nil, groupId: %v, uid: %v", groupId, player.PlayerId)
		return
	}
	_, exist := groupConfig.SuiteMap[int32(suiteId)]
	if !exist {
		logger.Error("invalid suite id: %v, uid: %v", suiteId, player.PlayerId)
		return
	}
	world := WORLD_MANAGER.GetWorldById(player.WorldId)
	if world == nil {
		return
	}
	scene := world.GetSceneById(player.GetSceneId())
	g.AddSceneGroupSuiteCore(player, scene, groupId, suiteId)
	group := scene.GetGroupById(groupId)
	suite := group.GetSuiteById(suiteId)
	entityIdList := make([]uint32, 0)
	for _, entity := range suite.GetAllEntity() {
		entityIdList = append(entityIdList, entity.GetId())
	}
	g.AddSceneEntityNotify(player, proto.VisionType_VISION_BORN, entityIdList, true, false)
}

// RemoveSceneGroupSuite 场景组中移除场景小组
func (g *Game) RemoveSceneGroupSuite(player *model.Player, groupId uint32, suiteId uint8) {
	world := WORLD_MANAGER.GetWorldById(player.WorldId)
	if world == nil {
		return
	}
	scene := world.GetSceneById(player.GetSceneId())
	group := scene.GetGroupById(groupId)
	if group != nil {
		suite := group.GetSuiteById(suiteId)
		if suite != nil {
			entityIdList := make([]uint32, 0)
			for _, entity := range suite.GetAllEntity() {
				entityIdList = append(entityIdList, entity.GetId())
			}
			g.RemoveSceneEntityNotifyBroadcast(scene, proto.VisionType_VISION_MISS, entityIdList, 0)
			scene.RemoveGroupSuite(groupId, suiteId)
		}
	}
}

// RefreshSceneGroupSuite 刷新场景小组
func (g *Game) RefreshSceneGroupSuite(player *model.Player, groupId uint32, suiteId uint8) {
	sceneGroup := player.GetSceneGroupById(groupId)
	if sceneGroup == nil {
		return
	}
	sceneGroup.RemoveAllKill()
	g.RemoveSceneGroupSuite(player, groupId, suiteId)
	g.AddSceneGroupSuite(player, groupId, suiteId)
}

func (g *Game) AddSceneGroupSuiteCore(player *model.Player, scene *Scene, groupId uint32, suiteId uint8) {
	groupConfig := gdconf.GetSceneGroup(int32(groupId))
	if groupConfig == nil {
		logger.Error("get scene group config is nil, groupId: %v", groupId)
		return
	}
	suiteConfig, exist := groupConfig.SuiteMap[int32(suiteId)]
	if !exist {
		logger.Error("invalid suiteId: %v", suiteId)
		return
	}
	world := scene.GetWorld()
	owner := world.GetOwner()
	sceneGroup := owner.GetSceneGroupById(groupId)
	if sceneGroup == nil {
		return
	}
	entityMap := make(map[uint32]IEntity)
	for _, monsterConfigId := range suiteConfig.MonsterConfigIdList {
		monsterConfig, exist := groupConfig.MonsterMap[monsterConfigId]
		if !exist {
			logger.Error("monster config not exist, monsterConfigId: %v", monsterConfigId)
			continue
		}
		isKill := sceneGroup.CheckIsKill(uint32(monsterConfig.ConfigId))
		if isKill {
			continue
		}
		entityId := g.CreateConfigEntity(scene, uint32(groupConfig.Id), monsterConfig)
		if entityId == 0 {
			continue
		}
		entity := scene.GetEntity(entityId)
		entityMap[entityId] = entity
	}
	for _, gadgetConfigId := range suiteConfig.GadgetConfigIdList {
		gadgetConfig, exist := groupConfig.GadgetMap[gadgetConfigId]
		if !exist {
			logger.Error("gadget config not exist, gadgetConfigId: %v", gadgetConfigId)
			continue
		}
		isKill := sceneGroup.CheckIsKill(uint32(gadgetConfig.ConfigId))
		if isKill {
			continue
		}
		entityId := g.CreateConfigEntity(scene, uint32(groupConfig.Id), gadgetConfig)
		if entityId == 0 {
			continue
		}
		entity := scene.GetEntity(entityId)
		entityMap[entityId] = entity
	}
	for _, npcConfig := range groupConfig.NpcMap {
		entityId := g.CreateConfigEntity(scene, uint32(groupConfig.Id), npcConfig)
		if entityId == 0 {
			continue
		}
		entity := scene.GetEntity(entityId)
		entityMap[entityId] = entity
	}
	scene.AddGroupSuite(groupId, suiteId, entityMap)
}

// CreateConfigEntity 创建配置表里的实体
func (g *Game) CreateConfigEntity(scene *Scene, groupId uint32, entityConfig any) uint32 {
	world := scene.GetWorld()
	owner := world.GetOwner()
	sceneGroup := owner.GetSceneGroupById(groupId)
	if sceneGroup == nil {
		return 0
	}
	switch entityConfig.(type) {
	case *gdconf.Monster:
		monster := entityConfig.(*gdconf.Monster)
		// TODO 怪物等级与世界等级的关联
		worldLevel := owner.PropMap[constant.PLAYER_PROP_PLAYER_WORLD_LEVEL]
		monsterLevel := uint8(monster.Level) * uint8(worldLevel+1)
		if monsterLevel > 100 {
			monsterLevel = 100
		}
		monsterEntity := scene.CreateEntityMonster(
			&model.Vector{X: float64(monster.Pos.X), Y: float64(monster.Pos.Y), Z: float64(monster.Pos.Z)},
			&model.Vector{X: float64(monster.Rot.X), Y: float64(monster.Rot.Y), Z: float64(monster.Rot.Z)},
			monsterLevel, uint32(monster.ConfigId), groupId, int(monster.VisionLevel),
		)
		monsterEntity.CreateMonsterEntity(uint32(monster.MonsterId))
		scene.CreateEntity(monsterEntity)
		return monsterEntity.GetId()
	case *gdconf.Npc:
		npc := entityConfig.(*gdconf.Npc)
		npcEntity := scene.CreateEntityNpc(
			&model.Vector{X: float64(npc.Pos.X), Y: float64(npc.Pos.Y), Z: float64(npc.Pos.Z)},
			&model.Vector{X: float64(npc.Rot.X), Y: float64(npc.Rot.Y), Z: float64(npc.Rot.Z)},
			uint32(npc.ConfigId), groupId,
		)
		npcEntity.CreateNpcEntity(uint32(npc.NpcId), 0, 0, 0)
		scene.CreateEntity(npcEntity)
		return npcEntity.GetId()
	case *gdconf.Gadget:
		gadget := entityConfig.(*gdconf.Gadget)
		gadgetDataConfig := gdconf.GetGadgetDataById(gadget.GadgetId)
		if gadgetDataConfig == nil {
			logger.Error("get gadget data config is nil, gadgetId: %v", gadget.GadgetId)
			return 0
		}
		switch gadgetDataConfig.Type {
		case constant.GADGET_TYPE_GATHER_POINT:
			gatherDataConfig := gdconf.GetGatherDataByPointType(gadget.PointType)
			if gatherDataConfig == nil {
				return 0
			}
			gadgetGatherEntity := scene.CreateEntityGadgetGather(
				&model.Vector{X: float64(gadget.Pos.X), Y: float64(gadget.Pos.Y), Z: float64(gadget.Pos.Z)},
				&model.Vector{X: float64(gadget.Rot.X), Y: float64(gadget.Rot.Y), Z: float64(gadget.Rot.Z)},
				uint32(gadget.ConfigId), groupId, int(gadget.VisionLevel), uint32(gatherDataConfig.GadgetId), uint32(constant.GADGET_STATE_DEFAULT),
			)
			gadgetGatherEntity.CreateGadgetGatherEntity(uint32(gatherDataConfig.ItemId))
			scene.CreateEntity(gadgetGatherEntity)
			return gadgetGatherEntity.GetId()
		case constant.GADGET_TYPE_WORKTOP:
			state := uint8(gadget.State)
			exist := sceneGroup.CheckGadgetExist(uint32(gadget.ConfigId))
			if exist {
				state = sceneGroup.GetGadgetState(uint32(gadget.ConfigId))
			}
			gadgetWorktopEntity := scene.CreateEntityGadgetWorktop(
				&model.Vector{X: float64(gadget.Pos.X), Y: float64(gadget.Pos.Y), Z: float64(gadget.Pos.Z)},
				&model.Vector{X: float64(gadget.Rot.X), Y: float64(gadget.Rot.Y), Z: float64(gadget.Rot.Z)},
				uint32(gadget.ConfigId), groupId, int(gadget.VisionLevel), uint32(gadget.GadgetId), uint32(state),
			)
			scene.CreateEntity(gadgetWorktopEntity)
			return gadgetWorktopEntity.GetId()
		default:
			state := uint8(gadget.State)
			exist := sceneGroup.CheckGadgetExist(uint32(gadget.ConfigId))
			if exist {
				state = sceneGroup.GetGadgetState(uint32(gadget.ConfigId))
			}
			gadgetNormalEntity := scene.CreateEntityGadgetNormal(
				&model.Vector{X: float64(gadget.Pos.X), Y: float64(gadget.Pos.Y), Z: float64(gadget.Pos.Z)},
				&model.Vector{X: float64(gadget.Rot.X), Y: float64(gadget.Rot.Y), Z: float64(gadget.Rot.Z)},
				uint32(gadget.ConfigId), groupId, int(gadget.VisionLevel), uint32(gadget.GadgetId), uint32(state),
			)
			scene.CreateEntity(gadgetNormalEntity)
			return gadgetNormalEntity.GetId()
		}
	}
	return 0
}

// SceneGroupCreateEntity 创建场景组配置物件实体
func (g *Game) SceneGroupCreateEntity(player *model.Player, groupId uint32, configId uint32, entityType uint8) {
	// 添加到初始小组
	groupConfig := gdconf.GetSceneGroup(int32(groupId))
	if groupConfig == nil {
		logger.Error("get group config is nil, groupId: %v, uid: %v", groupId, player.PlayerId)
		return
	}
	initSuiteId := groupConfig.GroupInitConfig.Suite
	_, exist := groupConfig.SuiteMap[initSuiteId]
	if !exist {
		logger.Error("invalid init suite id: %v, uid: %v", initSuiteId, player.PlayerId)
		return
	}
	// 添加场景实体
	world := WORLD_MANAGER.GetWorldById(player.WorldId)
	if world == nil {
		return
	}
	scene := world.GetSceneById(player.GetSceneId())
	var entityConfig any = nil
	switch entityType {
	case constant.ENTITY_TYPE_MONSTER:
		monsterConfig, exist := groupConfig.MonsterMap[int32(configId)]
		if !exist {
			logger.Error("monster config not exist, configId: %v", configId)
			return
		}
		entityConfig = monsterConfig
	case constant.ENTITY_TYPE_GADGET:
		gadgetConfig, exist := groupConfig.GadgetMap[int32(configId)]
		if !exist {
			logger.Error("gadget config not exist, configId: %v", configId)
			return
		}
		entityConfig = gadgetConfig
	default:
		logger.Error("unknown entity type: %v", entityType)
		return
	}
	entityId := g.CreateConfigEntity(scene, uint32(groupConfig.Id), entityConfig)
	if entityId == 0 {
		return
	}
	entity := scene.GetEntity(entityId)
	// 实体添加到场景小组
	scene.AddGroupSuite(groupId, uint8(initSuiteId), map[uint32]IEntity{entity.GetId(): entity})
	// 通知客户端
	g.AddSceneEntityNotify(player, proto.VisionType_VISION_BORN, []uint32{entityId}, true, false)
	// 触发器检测
	group := scene.GetGroupById(groupId)
	if group == nil {
		logger.Error("group not exist, groupId: %v, uid: %v", groupId, player.PlayerId)
		return
	}
	switch entityType {
	case constant.ENTITY_TYPE_MONSTER:
		// 怪物创建触发器检测
		g.MonsterCreateTriggerCheck(player, group, entity)
	case constant.ENTITY_TYPE_GADGET:
		// 物件创建触发器检测
		g.GadgetCreateTriggerCheck(player, group, entity)
	}
}

// CreateMonster 创建怪物实体
func (g *Game) CreateMonster(player *model.Player, pos *model.Vector, monsterId uint32, level uint8) uint32 {
	world := WORLD_MANAGER.GetWorldById(player.WorldId)
	if world == nil {
		return 0
	}
	scene := world.GetSceneById(player.GetSceneId())
	if scene == nil {
		return 0
	}
	if pos == nil {
		pos = g.GetPlayerPos(player)
		pos.X += random.GetRandomFloat64(-5.0, 5.0)
		pos.Z += random.GetRandomFloat64(-5.0, 5.0)
	}
	rot := new(model.Vector)
	rot.Y = random.GetRandomFloat64(0.0, 360.0)
	monsterEntity := scene.CreateEntityMonster(pos, rot, level, 0, 0, constant.VISION_LEVEL_NORMAL)
	monsterEntity.CreateMonsterEntity(monsterId)
	scene.CreateEntity(monsterEntity)
	g.AddSceneEntityNotify(player, proto.VisionType_VISION_BORN, []uint32{monsterEntity.GetId()}, true, false)
	return monsterEntity.GetId()
}

// CreateGadget 创建物件实体
func (g *Game) CreateGadget(player *model.Player, pos *model.Vector, gadgetId uint32) uint32 {
	if gadgetId == 0 {
		logger.Error("create gadget id is zero, pos: %+v, uid: %v", pos, player.PlayerId)
		return 0
	}
	world := WORLD_MANAGER.GetWorldById(player.WorldId)
	if world == nil {
		return 0
	}
	scene := world.GetSceneById(player.GetSceneId())
	if pos == nil {
		pos = g.GetPlayerPos(player)
		pos.X += random.GetRandomFloat64(-5.0, 5.0)
		pos.Z += random.GetRandomFloat64(-5.0, 5.0)
	}
	rot := new(model.Vector)
	rot.Y = random.GetRandomFloat64(0.0, 360.0)
	gadgetNormalEntity := scene.CreateEntityGadgetNormal(pos, rot, 0, 0, constant.VISION_LEVEL_NORMAL, gadgetId, constant.GADGET_STATE_DEFAULT)
	scene.CreateEntity(gadgetNormalEntity)
	g.AddSceneEntityNotify(player, proto.VisionType_VISION_BORN, []uint32{gadgetNormalEntity.GetId()}, true, false)
	return gadgetNormalEntity.GetId()
}

// CreateDropGadget 创建掉落物的物件实体
func (g *Game) CreateDropGadget(player *model.Player, pos *model.Vector, gadgetId, itemId, count uint32) uint32 {
	if gadgetId == 0 {
		logger.Error("create gadget id is zero, pos: %+v, itemId: %v, count: %v, uid: %v", pos, itemId, count, player.PlayerId)
		return 0
	}
	world := WORLD_MANAGER.GetWorldById(player.WorldId)
	if world == nil {
		return 0
	}
	scene := world.GetSceneById(player.GetSceneId())
	if pos == nil {
		pos = g.GetPlayerPos(player)
		pos.X += random.GetRandomFloat64(-5.0, 5.0)
		pos.Z += random.GetRandomFloat64(-5.0, 5.0)
	}
	rot := new(model.Vector)
	rot.Y = random.GetRandomFloat64(0.0, 360.0)
	gadgetTrifleItemEntity := scene.CreateEntityGadgetTrifleItem(pos, rot, constant.VISION_LEVEL_NORMAL, gadgetId, constant.GADGET_STATE_DEFAULT)
	gadgetTrifleItemEntity.CreateGadgetTrifleItemEntity(itemId, count)
	scene.CreateEntity(gadgetTrifleItemEntity)
	g.AddSceneEntityNotify(player, proto.VisionType_VISION_BORN, []uint32{gadgetTrifleItemEntity.GetId()}, true, false)
	return gadgetTrifleItemEntity.GetId()
}

// GetPosIsInWeatherArea 获取坐标是否在指定的天气区域
func (g *Game) GetPosIsInWeatherArea(posX, posZ float64, sceneId, jsonWeatherAreaId uint32) bool {
	// 获取场景天气区域配置表
	weatherAreaData := gdconf.GetWeatherAreaMapBySceneIdAndWeatherAreaId(int32(sceneId), int32(jsonWeatherAreaId))
	if weatherAreaData == nil {
		logger.Error("weather area data config not exist, sceneId: %v, jsonWeatherAreaId: %v", sceneId, jsonWeatherAreaId)
		return false
	}
	// 判断坐标是否在指定的天气区域
	pos := &alg.Vector2{
		X: float32(posX),
		Z: float32(posZ),
	}
	return alg.Region2DPolygonContainPos(weatherAreaData.VectorPoints, pos)
}

// GetPlayerInWeatherAreaId 获取玩家所在的天气区域id
func (g *Game) GetPlayerInWeatherAreaId(player *model.Player, newPos *model.Vector) (weatherAreaId uint32) {
	// 获取场景天气区域配置表
	weatherAreaDataMap := gdconf.GetWeatherAreaMap()[int32(player.GetSceneId())]
	if weatherAreaDataMap == nil {
		logger.Error("weather area data config not exist, sceneId: %v", player.GetSceneId())
		return
	}
	// 寻找玩家所在范围内的天气区域
	var priority int32
	// 玩家所在的天气区域
	for _, area := range weatherAreaDataMap {
		// 获取天气数据配置表
		weatherDataMap := gdconf.GetWeatherDataMapByJsonWeatherAreaId(area.AreaId)
		if weatherDataMap == nil {
			// 有些天气不在配置表内
			logger.Error("weather data config not exist, weatherAreaId: %v", area.AreaId)
			continue
		}
		for _, weatherData := range weatherDataMap {
			// 确保高度不超过
			if weatherData.MaxHeight != 0 && player.GetPos().Y > float64(weatherData.MaxHeight) {
				continue
			}
			// 确保默认自动开启
			if weatherData.DefaultOpen != 1 {
				continue
			}
			// 确保处于天气区域内
			if !g.GetPosIsInWeatherArea(newPos.X, newPos.Z, player.GetSceneId(), uint32(weatherData.JsonWeatherAreaId)) {
				continue
			}
			// 优先级比较
			if weatherData.Priority > priority {
				weatherAreaId = uint32(weatherData.WeatherAreaId)
				priority = weatherData.Priority
			}
		}
	}
	return
}

// WeatherClimateRandom 随机天气气象
func (g *Game) WeatherClimateRandom(player *model.Player, weatherAreaId uint32) {
	// 天气气象锁定则跳过
	if player.PropMap[constant.PLAYER_PROP_IS_WEATHER_LOCKED] == 1 {
		return
	}
	// 获取天气气象
	climateType := g.GetWeatherAreaClimate(weatherAreaId)
	// 跳过相同的天气
	if climateType == player.WeatherInfo.ClimateType {
		return
	}
	g.SetPlayerWeather(player, player.WeatherInfo.WeatherAreaId, climateType, true)
}

// SceneWeatherAreaCheck 场景天气区域变更检测
func (g *Game) SceneWeatherAreaCheck(player *model.Player, oldPos *model.Vector, newPos *model.Vector) {
	// 如果玩家没移动就不检测变更
	if oldPos.X == newPos.X && oldPos.Z == newPos.Z {
		return
	}
	// 如果玩家还在历史区域内就不获取当前所在区域
	if g.GetPosIsInWeatherArea(newPos.X, newPos.Z, player.GetSceneId(), player.WeatherInfo.JsonWeatherAreaId) {
		return
	}
	// 获取当前所在的天气区域
	weatherAreaId := g.GetPlayerInWeatherAreaId(player, newPos)
	if weatherAreaId == 0 {
		logger.Error("weather area id error, weatherAreaId: %v", weatherAreaId)
		return
	}
	// 判断天气区域是否变更
	if player.WeatherInfo.WeatherAreaId == weatherAreaId {
		return
	}
	// 随机天气气象
	g.WeatherClimateRandom(player, weatherAreaId)
}

// GetWeatherAreaClimate 获取天气气象
func (g *Game) GetWeatherAreaClimate(weatherAreaId uint32) uint32 {
	// 获取天气数据配置表
	weatherData := gdconf.GetWeatherDataByWeatherAreaId(int32(weatherAreaId))
	if weatherData == nil {
		logger.Error("weather data config not exist, weatherAreaId: %v", weatherAreaId)
		return 0
	}
	// 如果指定了则使用指定的天气
	var weatherTemplateDataConfig *gdconf.WeatherTemplateData
	var weather int32
	if weatherData.UseDefaultWeather == 1 && weatherData.DefaultWeather != 0 {
		weather = weatherData.DefaultWeather
		weatherTemplateDataConfig = gdconf.GetWeatherTemplateDataByTemplateNameAndWeather(weatherData.TemplateName, weather)
	} else {
		// 随机取个天气类型
		weatherTemplateDataMap := gdconf.GetWeatherTemplateDataMap()[weatherData.TemplateName]
		if weatherTemplateDataMap == nil {
			logger.Error("weather template data map not exist, templateName: %v", weatherData.TemplateName)
			return 0
		}
		weatherTemplateList := make([]int32, 0, len(weatherTemplateDataMap))
		for key := range weatherTemplateDataMap {
			weatherTemplateList = append(weatherTemplateList, key)
		}
		weather = random.GetRandomInt32(1, int32(len(weatherTemplateList)))
		weatherTemplateDataConfig = weatherTemplateDataMap[weather]
	}
	// 确保指定的天气模版存在
	if weatherTemplateDataConfig == nil {
		logger.Error("weather template config not exist, templateName: %v, weather: %v", weatherData.TemplateName, weather)
		return 0
	}
	// 随机气象 轮盘赌选择法RWS
	climateWeightMap := map[uint32]int32{
		constant.CLIMATE_TYPE_SUNNY:        weatherTemplateDataConfig.Sunny,
		constant.CLIMATE_TYPE_CLOUDY:       weatherTemplateDataConfig.Cloudy,
		constant.CLIMATE_TYPE_RAIN:         weatherTemplateDataConfig.Rain,
		constant.CLIMATE_TYPE_THUNDERSTORM: weatherTemplateDataConfig.ThunderStorm,
		constant.CLIMATE_TYPE_SNOW:         weatherTemplateDataConfig.Snow,
		constant.CLIMATE_TYPE_MIST:         weatherTemplateDataConfig.Mist,
		constant.CLIMATE_TYPE_DESERT:       weatherTemplateDataConfig.Desert,
	}
	var weightAll int32
	for _, weight := range climateWeightMap {
		weightAll += weight
	}
	// logger.Debug("weather climate weightMap: %v, weightAll: %v", climateWeightMap, weightAll)
	randNum := random.GetRandomInt32(0, weightAll-1)
	sumWeight := int32(0)
	for climate, weight := range climateWeightMap {
		sumWeight += weight
		if sumWeight > randNum {
			return climate
		}
	}
	return 0
}

// SetPlayerWeather 设置玩家天气
func (g *Game) SetPlayerWeather(player *model.Player, weatherAreaId uint32, climateType uint32, sendNotify bool) {
	// 获取天气数据配置表
	weatherData := gdconf.GetWeatherDataByWeatherAreaId(int32(weatherAreaId))
	if weatherData == nil {
		logger.Error("weather data config not exist, weatherAreaId: %v", weatherAreaId)
		return
	}

	logger.Debug("weather climateType: %v, weatherAreaId: %v, jsonWeatherAreaId: %v, uid: %v", climateType, weatherAreaId, weatherData.JsonWeatherAreaId, player.PlayerId)

	// 记录数据
	player.WeatherInfo.WeatherAreaId = weatherAreaId
	player.WeatherInfo.JsonWeatherAreaId = uint32(weatherData.JsonWeatherAreaId)
	player.WeatherInfo.ClimateType = climateType

	if !sendNotify {
		return
	}
	sceneAreaWeatherNotify := &proto.SceneAreaWeatherNotify{
		WeatherAreaId: weatherAreaId,
		ClimateType:   climateType,
	}
	if player.SceneJump {
		g.SendMsg(cmd.SceneAreaWeatherNotify, player.PlayerId, player.ClientSeq, sceneAreaWeatherNotify)
	} else {
		g.SendMsg(cmd.SceneAreaWeatherNotify, player.PlayerId, 0, sceneAreaWeatherNotify)
	}
}

/************************************************** 打包封装 **************************************************/

func (g *Game) PacketPlayerEnterSceneNotifyLogin(player *model.Player) *proto.PlayerEnterSceneNotify {
	world := WORLD_MANAGER.GetWorldById(player.WorldId)
	if world == nil {
		logger.Error("get world is nil, worldId: %v, uid: %v", player.WorldId, player.PlayerId)
		return new(proto.PlayerEnterSceneNotify)
	}
	scene := world.GetSceneById(player.GetSceneId())
	enterSceneToken := world.AddEnterSceneContext(&EnterSceneContext{
		OldSceneId: 0,
		Uid:        player.PlayerId,
	})
	pos := player.GetPos()
	playerEnterSceneNotify := &proto.PlayerEnterSceneNotify{
		SceneId:                player.GetSceneId(),
		Pos:                    &proto.Vector{X: float32(pos.X), Y: float32(pos.Y), Z: float32(pos.Z)},
		SceneBeginTime:         uint64(scene.GetSceneCreateTime()),
		Type:                   proto.EnterType_ENTER_SELF,
		TargetUid:              player.PlayerId,
		EnterSceneToken:        enterSceneToken,
		WorldLevel:             player.PropMap[constant.PLAYER_PROP_PLAYER_WORLD_LEVEL],
		EnterReason:            player.SceneEnterReason,
		IsFirstLoginEnterScene: true,
		WorldType:              1,
		SceneTagIdList:         make([]uint32, 0),
	}
	playerEnterSceneNotify.SceneTransaction = strconv.Itoa(int(player.GetSceneId())) + "-" + g.NewTransaction(player.PlayerId)
	// TODO 暂时先解锁全部场景标签 看着喜庆
	for _, sceneTagDataConfig := range gdconf.GetSceneTagDataMap() {
		if uint32(sceneTagDataConfig.SceneId) == player.GetSceneId() {
			playerEnterSceneNotify.SceneTagIdList = append(playerEnterSceneNotify.SceneTagIdList, uint32(sceneTagDataConfig.SceneTagId))
		}
	}
	return playerEnterSceneNotify
}

func (g *Game) PacketPlayerEnterSceneNotifyTp(
	player *model.Player,
	enterType proto.EnterType,
	sceneId uint32,
	pos *model.Vector,
	dungeonId uint32,
	enterSceneToken uint32,
) *proto.PlayerEnterSceneNotify {
	return g.PacketPlayerEnterSceneNotifyCore(player, player, enterType, sceneId, pos, dungeonId, enterSceneToken)
}

func (g *Game) PacketPlayerEnterSceneNotifyMp(
	player *model.Player,
	targetPlayer *model.Player,
	enterType proto.EnterType,
	sceneId uint32,
	pos *model.Vector,
	enterSceneToken uint32,
) *proto.PlayerEnterSceneNotify {
	return g.PacketPlayerEnterSceneNotifyCore(player, targetPlayer, enterType, sceneId, pos, 0, enterSceneToken)
}

func (g *Game) PacketPlayerEnterSceneNotifyCore(
	player *model.Player,
	targetPlayer *model.Player,
	enterType proto.EnterType,
	sceneId uint32,
	pos *model.Vector,
	dungeonId uint32,
	enterSceneToken uint32,
) *proto.PlayerEnterSceneNotify {
	world := WORLD_MANAGER.GetWorldById(targetPlayer.WorldId)
	if world == nil {
		logger.Error("get world is nil, worldId: %v, uid: %v", player.WorldId, player.PlayerId)
		return new(proto.PlayerEnterSceneNotify)
	}
	scene := world.GetSceneById(targetPlayer.GetSceneId())
	playerEnterSceneNotify := &proto.PlayerEnterSceneNotify{
		PrevSceneId:     player.GetSceneId(),
		PrevPos:         &proto.Vector{X: float32(player.GetPos().X), Y: float32(player.GetPos().Y), Z: float32(player.GetPos().Z)},
		SceneId:         sceneId,
		Pos:             &proto.Vector{X: float32(pos.X), Y: float32(pos.Y), Z: float32(pos.Z)},
		SceneBeginTime:  uint64(scene.GetSceneCreateTime()),
		Type:            enterType,
		TargetUid:       targetPlayer.PlayerId,
		EnterSceneToken: enterSceneToken,
		WorldLevel:      targetPlayer.PropMap[constant.PLAYER_PROP_PLAYER_WORLD_LEVEL],
		EnterReason:     player.SceneEnterReason,
		WorldType:       1,
		DungeonId:       dungeonId,
		SceneTagIdList:  make([]uint32, 0),
	}
	playerEnterSceneNotify.SceneTransaction = strconv.Itoa(int(sceneId)) + "-" + g.NewTransaction(player.PlayerId)
	// TODO 暂时先解锁全部场景标签 看着喜庆
	for _, sceneTagDataConfig := range gdconf.GetSceneTagDataMap() {
		if uint32(sceneTagDataConfig.SceneId) == sceneId {
			playerEnterSceneNotify.SceneTagIdList = append(playerEnterSceneNotify.SceneTagIdList, uint32(sceneTagDataConfig.SceneTagId))
		}
	}
	return playerEnterSceneNotify
}

func (g *Game) PacketFightPropMapToPbFightPropList(fightPropMap map[uint32]float32) []*proto.FightPropPair {
	fightPropList := []*proto.FightPropPair{
		{PropType: constant.FIGHT_PROP_BASE_HP, PropValue: fightPropMap[constant.FIGHT_PROP_BASE_HP]},
		{PropType: constant.FIGHT_PROP_BASE_ATTACK, PropValue: fightPropMap[constant.FIGHT_PROP_BASE_ATTACK]},
		{PropType: constant.FIGHT_PROP_BASE_DEFENSE, PropValue: fightPropMap[constant.FIGHT_PROP_BASE_DEFENSE]},
		{PropType: constant.FIGHT_PROP_CRITICAL, PropValue: fightPropMap[constant.FIGHT_PROP_CRITICAL]},
		{PropType: constant.FIGHT_PROP_CRITICAL_HURT, PropValue: fightPropMap[constant.FIGHT_PROP_CRITICAL_HURT]},
		{PropType: constant.FIGHT_PROP_CHARGE_EFFICIENCY, PropValue: fightPropMap[constant.FIGHT_PROP_CHARGE_EFFICIENCY]},
		{PropType: constant.FIGHT_PROP_CUR_HP, PropValue: fightPropMap[constant.FIGHT_PROP_CUR_HP]},
		{PropType: constant.FIGHT_PROP_MAX_HP, PropValue: fightPropMap[constant.FIGHT_PROP_MAX_HP]},
		{PropType: constant.FIGHT_PROP_CUR_ATTACK, PropValue: fightPropMap[constant.FIGHT_PROP_CUR_ATTACK]},
		{PropType: constant.FIGHT_PROP_CUR_DEFENSE, PropValue: fightPropMap[constant.FIGHT_PROP_CUR_DEFENSE]},
	}
	return fightPropList
}

func (g *Game) PacketSceneEntityInfoAvatar(scene *Scene, player *model.Player, avatarId uint32) *proto.SceneEntityInfo {
	entity := scene.GetEntity(scene.GetWorld().GetPlayerWorldAvatarEntityId(player, avatarId))
	if entity == nil {
		return new(proto.SceneEntityInfo)
	}
	pos := &proto.Vector{
		X: float32(entity.GetPos().X),
		Y: float32(entity.GetPos().Y),
		Z: float32(entity.GetPos().Z),
	}
	worldAvatar := scene.GetWorld().GetWorldAvatarByEntityId(entity.GetId())
	dbAvatar := player.GetDbAvatar()
	avatar := dbAvatar.GetAvatarById(worldAvatar.GetAvatarId())
	if avatar == nil {
		logger.Error("get avatar is nil, avatarId: %v", worldAvatar.GetAvatarId())
		return new(proto.SceneEntityInfo)
	}
	sceneEntityInfo := &proto.SceneEntityInfo{
		EntityType: proto.ProtEntityType_PROT_ENTITY_AVATAR,
		EntityId:   entity.GetId(),
		MotionInfo: &proto.MotionInfo{
			Pos: pos,
			Rot: &proto.Vector{
				X: float32(entity.GetRot().X),
				Y: float32(entity.GetRot().Y),
				Z: float32(entity.GetRot().Z),
			},
			Speed: &proto.Vector{},
			State: proto.MotionState(entity.GetMoveState()),
		},
		PropList: []*proto.PropPair{
			{Type: uint32(constant.PLAYER_PROP_LEVEL), PropValue: g.PacketPropValue(constant.PLAYER_PROP_LEVEL, avatar.Level)},
			{Type: uint32(constant.PLAYER_PROP_EXP), PropValue: g.PacketPropValue(constant.PLAYER_PROP_EXP, avatar.Exp)},
			{Type: uint32(constant.PLAYER_PROP_BREAK_LEVEL), PropValue: g.PacketPropValue(constant.PLAYER_PROP_BREAK_LEVEL, avatar.Promote)},
			{Type: uint32(constant.PLAYER_PROP_SATIATION_VAL), PropValue: g.PacketPropValue(constant.PLAYER_PROP_SATIATION_VAL, avatar.Satiation)},
			{Type: uint32(constant.PLAYER_PROP_SATIATION_PENALTY_TIME), PropValue: g.PacketPropValue(constant.PLAYER_PROP_SATIATION_PENALTY_TIME, avatar.SatiationPenalty)},
		},
		FightPropList:    g.PacketFightPropMapToPbFightPropList(avatar.FightPropMap),
		LifeState:        uint32(avatar.LifeState),
		AnimatorParaList: make([]*proto.AnimatorParameterValueInfoPair, 0),
		Entity: &proto.SceneEntityInfo_Avatar{
			Avatar: g.PacketSceneAvatarInfo(scene, player, avatarId),
		},
		EntityClientData: new(proto.EntityClientData),
		EntityAuthorityInfo: &proto.EntityAuthorityInfo{
			AbilityInfo: &proto.AbilitySyncStateInfo{
				IsInited:           len(worldAvatar.PacketAbilityList()) != 0,
				DynamicValueMap:    nil,
				AppliedAbilities:   worldAvatar.PacketAbilityList(),
				AppliedModifiers:   worldAvatar.PacketModifierList(),
				MixinRecoverInfos:  nil,
				SgvDynamicValueMap: nil,
			},
			RendererChangedInfo: new(proto.EntityRendererChangedInfo),
			AiInfo: &proto.SceneEntityAiInfo{
				IsAiOpen: true,
				BornPos:  pos,
			},
			BornPos: pos,
		},
		LastMoveSceneTimeMs: entity.GetLastMoveSceneTimeMs(),
		LastMoveReliableSeq: entity.GetLastMoveReliableSeq(),
	}
	return sceneEntityInfo
}

func (g *Game) PacketSceneEntityInfoMonster(scene *Scene, entityId uint32) *proto.SceneEntityInfo {
	entity := scene.GetEntity(entityId)
	if entity == nil {
		return new(proto.SceneEntityInfo)
	}
	pos := &proto.Vector{
		X: float32(entity.GetPos().X),
		Y: float32(entity.GetPos().Y),
		Z: float32(entity.GetPos().Z),
	}
	sceneEntityInfo := &proto.SceneEntityInfo{
		EntityType: proto.ProtEntityType_PROT_ENTITY_MONSTER,
		EntityId:   entity.GetId(),
		MotionInfo: &proto.MotionInfo{
			Pos: pos,
			Rot: &proto.Vector{
				X: float32(entity.GetRot().X),
				Y: float32(entity.GetRot().Y),
				Z: float32(entity.GetRot().Z),
			},
			Speed: &proto.Vector{},
			State: proto.MotionState(entity.GetMoveState()),
		},
		PropList: []*proto.PropPair{
			{Type: uint32(constant.PLAYER_PROP_LEVEL), PropValue: g.PacketPropValue(constant.PLAYER_PROP_LEVEL, int64(entity.GetLevel()))},
		},
		FightPropList:    g.PacketFightPropMapToPbFightPropList(entity.GetFightProp()),
		LifeState:        uint32(entity.GetLifeState()),
		AnimatorParaList: make([]*proto.AnimatorParameterValueInfoPair, 0),
		Entity: &proto.SceneEntityInfo_Monster{
			Monster: g.PacketSceneMonsterInfo(entity),
		},
		EntityClientData: new(proto.EntityClientData),
		EntityAuthorityInfo: &proto.EntityAuthorityInfo{
			AbilityInfo:         new(proto.AbilitySyncStateInfo),
			RendererChangedInfo: new(proto.EntityRendererChangedInfo),
			AiInfo: &proto.SceneEntityAiInfo{
				IsAiOpen: true,
				BornPos:  pos,
			},
			BornPos: pos,
		},
	}
	return sceneEntityInfo
}

func (g *Game) PacketSceneEntityInfoNpc(scene *Scene, entityId uint32) *proto.SceneEntityInfo {
	entity := scene.GetEntity(entityId)
	if entity == nil {
		return new(proto.SceneEntityInfo)
	}
	pos := &proto.Vector{
		X: float32(entity.GetPos().X),
		Y: float32(entity.GetPos().Y),
		Z: float32(entity.GetPos().Z),
	}
	sceneEntityInfo := &proto.SceneEntityInfo{
		EntityType: proto.ProtEntityType_PROT_ENTITY_NPC,
		EntityId:   entity.GetId(),
		MotionInfo: &proto.MotionInfo{
			Pos: pos,
			Rot: &proto.Vector{
				X: float32(entity.GetRot().X),
				Y: float32(entity.GetRot().Y),
				Z: float32(entity.GetRot().Z),
			},
			Speed: &proto.Vector{},
			State: proto.MotionState(entity.GetMoveState()),
		},
		PropList: []*proto.PropPair{
			{Type: uint32(constant.PLAYER_PROP_LEVEL), PropValue: g.PacketPropValue(constant.PLAYER_PROP_LEVEL, int64(entity.GetLevel()))},
		},
		FightPropList:    g.PacketFightPropMapToPbFightPropList(entity.GetFightProp()),
		LifeState:        uint32(entity.GetLifeState()),
		AnimatorParaList: make([]*proto.AnimatorParameterValueInfoPair, 0),
		Entity: &proto.SceneEntityInfo_Npc{
			Npc: g.PacketSceneNpcInfo(entity.(*NpcEntity)),
		},
		EntityClientData: new(proto.EntityClientData),
		EntityAuthorityInfo: &proto.EntityAuthorityInfo{
			AbilityInfo:         new(proto.AbilitySyncStateInfo),
			RendererChangedInfo: new(proto.EntityRendererChangedInfo),
			AiInfo: &proto.SceneEntityAiInfo{
				IsAiOpen: true,
				BornPos:  pos,
			},
			BornPos: pos,
		},
	}
	return sceneEntityInfo
}

func (g *Game) PacketSceneEntityInfoGadget(player *model.Player, scene *Scene, entityId uint32) *proto.SceneEntityInfo {
	entity := scene.GetEntity(entityId)
	if entity == nil {
		return new(proto.SceneEntityInfo)
	}
	pos := &proto.Vector{
		X: float32(entity.GetPos().X),
		Y: float32(entity.GetPos().Y),
		Z: float32(entity.GetPos().Z),
	}
	sceneEntityInfo := &proto.SceneEntityInfo{
		EntityType: proto.ProtEntityType_PROT_ENTITY_GADGET,
		EntityId:   entity.GetId(),
		MotionInfo: &proto.MotionInfo{
			Pos: pos,
			Rot: &proto.Vector{
				X: float32(entity.GetRot().X),
				Y: float32(entity.GetRot().Y),
				Z: float32(entity.GetRot().Z),
			},
			Speed: &proto.Vector{},
			State: proto.MotionState(entity.GetMoveState()),
		},
		PropList: []*proto.PropPair{
			{Type: uint32(constant.PLAYER_PROP_LEVEL), PropValue: g.PacketPropValue(constant.PLAYER_PROP_LEVEL, 1)},
		},
		FightPropList:    g.PacketFightPropMapToPbFightPropList(entity.GetFightProp()),
		LifeState:        uint32(entity.GetLifeState()),
		AnimatorParaList: make([]*proto.AnimatorParameterValueInfoPair, 0),
		EntityClientData: new(proto.EntityClientData),
		EntityAuthorityInfo: &proto.EntityAuthorityInfo{
			AbilityInfo:         new(proto.AbilitySyncStateInfo),
			RendererChangedInfo: new(proto.EntityRendererChangedInfo),
			AiInfo: &proto.SceneEntityAiInfo{
				IsAiOpen: true,
				BornPos:  pos,
			},
			BornPos: pos,
		},
	}
	switch entity.(type) {
	case *GadgetNormalEntity:
		sceneEntityInfo.Entity = &proto.SceneEntityInfo_Gadget{
			Gadget: g.PacketSceneGadgetInfoNormal(entity.(*GadgetNormalEntity)),
		}
	case *GadgetTrifleItemEntity:
		sceneEntityInfo.Entity = &proto.SceneEntityInfo_Gadget{
			Gadget: g.PacketSceneGadgetInfoTrifleItem(player, entity.(*GadgetTrifleItemEntity)),
		}
	case *GadgetGatherEntity:
		sceneEntityInfo.Entity = &proto.SceneEntityInfo_Gadget{
			Gadget: g.PacketSceneGadgetInfoGather(entity.(*GadgetGatherEntity)),
		}
	case *GadgetWorktopEntity:
		sceneEntityInfo.Entity = &proto.SceneEntityInfo_Gadget{
			Gadget: g.PacketSceneGadgetInfoWorktop(entity.(*GadgetWorktopEntity)),
		}
	case *GadgetClientEntity:
		sceneEntityInfo.Entity = &proto.SceneEntityInfo_Gadget{
			Gadget: g.PacketSceneGadgetInfoClient(entity.(*GadgetClientEntity)),
		}
	case *GadgetVehicleEntity:
		sceneEntityInfo.Entity = &proto.SceneEntityInfo_Gadget{
			Gadget: g.PacketSceneGadgetInfoVehicle(entity.(*GadgetVehicleEntity)),
		}
	}
	return sceneEntityInfo
}

func (g *Game) PacketSceneAvatarInfo(scene *Scene, player *model.Player, avatarId uint32) *proto.SceneAvatarInfo {
	dbAvatar := player.GetDbAvatar()
	avatar := dbAvatar.GetAvatarById(avatarId)
	if avatar == nil {
		logger.Error("get avatar is nil, avatarId: %v", avatarId)
		return new(proto.SceneAvatarInfo)
	}
	equipIdList := make([]uint32, len(avatar.EquipReliquaryMap)+1)
	for _, reliquary := range avatar.EquipReliquaryMap {
		equipIdList = append(equipIdList, reliquary.ItemId)
	}
	equipIdList = append(equipIdList, avatar.EquipWeapon.ItemId)
	reliquaryList := make([]*proto.SceneReliquaryInfo, 0, len(avatar.EquipReliquaryMap))
	for _, reliquary := range avatar.EquipReliquaryMap {
		reliquaryList = append(reliquaryList, &proto.SceneReliquaryInfo{
			ItemId:       reliquary.ItemId,
			Guid:         reliquary.Guid,
			Level:        uint32(reliquary.Level),
			PromoteLevel: uint32(reliquary.Promote),
		})
	}
	world := WORLD_MANAGER.GetWorldById(player.WorldId)
	sceneAvatarInfo := &proto.SceneAvatarInfo{
		Uid:          player.PlayerId,
		AvatarId:     avatarId,
		Guid:         avatar.Guid,
		PeerId:       world.GetPlayerPeerId(player),
		EquipIdList:  equipIdList,
		SkillDepotId: avatar.SkillDepotId,
		Weapon: &proto.SceneWeaponInfo{
			EntityId:    scene.GetWorld().GetPlayerWorldAvatarWeaponEntityId(player, avatarId),
			GadgetId:    uint32(gdconf.GetItemDataById(int32(avatar.EquipWeapon.ItemId)).GadgetId),
			ItemId:      avatar.EquipWeapon.ItemId,
			Guid:        avatar.EquipWeapon.Guid,
			Level:       uint32(avatar.EquipWeapon.Level),
			AbilityInfo: new(proto.AbilitySyncStateInfo),
		},
		ReliquaryList:          reliquaryList,
		SkillLevelMap:          avatar.SkillLevelMap,
		TalentIdList:           avatar.TalentIdList,
		InherentProudSkillList: gdconf.GetAvatarInherentProudSkillList(avatar.SkillDepotId, avatar.Promote),
		WearingFlycloakId:      avatar.FlyCloak,
		CostumeId:              avatar.Costume,
		BornTime:               uint32(avatar.BornTime),
		TeamResonanceList:      make([]uint32, 0), // 队伍元素共鸣
	}
	return sceneAvatarInfo
}

func (g *Game) PacketSceneMonsterInfo(entity IEntity) *proto.SceneMonsterInfo {
	blockId := uint32(0)
	if entity.GetGroupId() != 0 {
		groupConfig := gdconf.GetSceneGroup(int32(entity.GetGroupId()))
		blockId = uint32(groupConfig.BlockId)
	}
	sceneMonsterInfo := &proto.SceneMonsterInfo{
		MonsterId:       entity.(*MonsterEntity).GetMonsterId(),
		AuthorityPeerId: 1,
		BornType:        proto.MonsterBornType_MONSTER_BORN_DEFAULT,
		BlockId:         blockId,
		TitleId:         blockId,
		SpecialNameId:   0,
	}
	return sceneMonsterInfo
}

func (g *Game) PacketSceneNpcInfo(entity *NpcEntity) *proto.SceneNpcInfo {
	sceneNpcInfo := &proto.SceneNpcInfo{
		NpcId:         entity.GetNpcId(),
		RoomId:        entity.GetRoomId(),
		ParentQuestId: entity.GetParentQuestId(),
		BlockId:       entity.GetBlockId(),
	}
	return sceneNpcInfo
}

func (g *Game) PacketSceneGadgetInfoNormal(gadgetNormalEntity *GadgetNormalEntity) *proto.SceneGadgetInfo {
	sceneGadgetInfo := &proto.SceneGadgetInfo{
		GadgetId:         gadgetNormalEntity.GetGadgetId(),
		GroupId:          gadgetNormalEntity.GetGroupId(),
		ConfigId:         gadgetNormalEntity.GetConfigId(),
		GadgetState:      gadgetNormalEntity.GetGadgetState(),
		IsEnableInteract: true,
		AuthorityPeerId:  1,
	}
	return sceneGadgetInfo
}

func (g *Game) PacketSceneGadgetInfoTrifleItem(player *model.Player, gadgetTrifleItemEntity *GadgetTrifleItemEntity) *proto.SceneGadgetInfo {
	sceneGadgetInfo := &proto.SceneGadgetInfo{
		GadgetId:         gadgetTrifleItemEntity.GetGadgetId(),
		GroupId:          gadgetTrifleItemEntity.GetGroupId(),
		ConfigId:         gadgetTrifleItemEntity.GetConfigId(),
		GadgetState:      gadgetTrifleItemEntity.GetGadgetState(),
		IsEnableInteract: true,
		AuthorityPeerId:  1,
	}
	dbItem := player.GetDbItem()
	sceneGadgetInfo.Content = &proto.SceneGadgetInfo_TrifleItem{
		TrifleItem: &proto.Item{
			ItemId: gadgetTrifleItemEntity.GetItemId(),
			Guid:   dbItem.GetItemGuid(gadgetTrifleItemEntity.GetItemId()),
			Detail: &proto.Item_Material{
				Material: &proto.Material{
					Count: gadgetTrifleItemEntity.GetCount(),
				},
			},
		},
	}
	return sceneGadgetInfo
}

func (g *Game) PacketSceneGadgetInfoGather(gadgetGatherEntity *GadgetGatherEntity) *proto.SceneGadgetInfo {
	sceneGadgetInfo := &proto.SceneGadgetInfo{
		GadgetId:         gadgetGatherEntity.GetGadgetId(),
		GroupId:          gadgetGatherEntity.GetGroupId(),
		ConfigId:         gadgetGatherEntity.GetConfigId(),
		GadgetState:      gadgetGatherEntity.GetGadgetState(),
		IsEnableInteract: true,
		AuthorityPeerId:  1,
	}
	sceneGadgetInfo.Content = &proto.SceneGadgetInfo_GatherGadget{
		GatherGadget: &proto.GatherGadgetInfo{
			ItemId:        gadgetGatherEntity.GetItemId(),
			IsForbidGuest: false,
		},
	}
	return sceneGadgetInfo
}

func (g *Game) PacketSceneGadgetInfoWorktop(gadgetWorktopEntity *GadgetWorktopEntity) *proto.SceneGadgetInfo {
	sceneGadgetInfo := &proto.SceneGadgetInfo{
		GadgetId:         gadgetWorktopEntity.GetGadgetId(),
		GroupId:          gadgetWorktopEntity.GetGroupId(),
		ConfigId:         gadgetWorktopEntity.GetConfigId(),
		GadgetState:      gadgetWorktopEntity.GetGadgetState(),
		IsEnableInteract: true,
		AuthorityPeerId:  1,
	}
	sceneGadgetInfo.Content = &proto.SceneGadgetInfo_Worktop{
		Worktop: &proto.WorktopInfo{
			OptionList:        object.ConvMapKeyToList[uint32, struct{}](gadgetWorktopEntity.GetOptionMap()),
			IsGuestCanOperate: false,
		},
	}
	return sceneGadgetInfo
}

func (g *Game) PacketSceneGadgetInfoClient(gadgetClientEntity *GadgetClientEntity) *proto.SceneGadgetInfo {
	sceneGadgetInfo := &proto.SceneGadgetInfo{
		GadgetId:         gadgetClientEntity.GetGadgetId(),
		OwnerEntityId:    gadgetClientEntity.GetOwnerEntityId(),
		AuthorityPeerId:  1,
		IsEnableInteract: true,
		Content: &proto.SceneGadgetInfo_ClientGadget{
			ClientGadget: &proto.ClientGadgetInfo{
				CampId:         gadgetClientEntity.GetCampId(),
				CampType:       gadgetClientEntity.GetCampType(),
				OwnerEntityId:  gadgetClientEntity.GetOwnerEntityId(),
				TargetEntityId: gadgetClientEntity.GetTargetEntityId(),
			},
		},
		PropOwnerEntityId: gadgetClientEntity.GetPropOwnerEntityId(),
	}
	return sceneGadgetInfo
}

func (g *Game) PacketSceneGadgetInfoVehicle(gadgetVehicleEntity *GadgetVehicleEntity) *proto.SceneGadgetInfo {
	player := USER_MANAGER.GetOnlineUser(gadgetVehicleEntity.GetOwnerUid())
	if player == nil {
		logger.Error("player is nil, userId: %v", gadgetVehicleEntity.GetOwnerUid())
		return new(proto.SceneGadgetInfo)
	}
	sceneGadgetInfo := &proto.SceneGadgetInfo{
		GadgetId:         gadgetVehicleEntity.GetGadgetId(),
		AuthorityPeerId:  gadgetVehicleEntity.GetScene().GetWorld().GetPlayerPeerId(player),
		IsEnableInteract: true,
		Content: &proto.SceneGadgetInfo_VehicleInfo{
			VehicleInfo: &proto.VehicleInfo{
				MemberList: make([]*proto.VehicleMember, 0, len(gadgetVehicleEntity.GetMemberMap())),
				OwnerUid:   gadgetVehicleEntity.GetOwnerUid(),
				CurStamina: gadgetVehicleEntity.GetCurStamina(),
			},
		},
	}
	return sceneGadgetInfo
}

func (g *Game) PacketDelTeamEntityNotify(world *World, player *model.Player) *proto.DelTeamEntityNotify {
	delTeamEntityNotify := &proto.DelTeamEntityNotify{
		SceneId:         player.GetSceneId(),
		DelEntityIdList: []uint32{world.GetPlayerTeamEntityId(player)},
	}
	return delTeamEntityNotify
}
