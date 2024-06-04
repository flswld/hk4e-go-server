package game

import (
	"math"
	"strings"

	"hk4e/common/constant"
	"hk4e/gdconf"
	"hk4e/gs/model"
	"hk4e/pkg/alg"
	"hk4e/pkg/logger"
	"hk4e/pkg/reflection"
	"hk4e/protocol/cmd"
	"hk4e/protocol/proto"

	pb "google.golang.org/protobuf/proto"
)

var cmdProtoMap *cmd.CmdProtoMap = nil
var attackResultTemplate *proto.AttackResult = nil

func DoForward[IET model.InvokeEntryType](player *model.Player, invokeHandler *model.InvokeHandler[IET],
	cmdId uint16, newNtf pb.Message, forwardField string,
	srcNtf pb.Message, copyFieldList []string) {
	if cmdProtoMap == nil {
		cmdProtoMap = cmd.NewCmdProtoMap()
	}
	world := WORLD_MANAGER.GetWorldById(player.WorldId)
	if world == nil {
		return
	}
	scene := world.GetSceneById(player.GetSceneId())
	if srcNtf != nil && copyFieldList != nil {
		for _, fieldName := range copyFieldList {
			reflection.CopyStructField(newNtf, srcNtf, fieldName)
		}
	}
	if invokeHandler.AllLen() == 0 && invokeHandler.AllExceptCurLen() == 0 && invokeHandler.HostLen() == 0 {
		return
	}
	if WORLD_MANAGER.IsAiWorld(world) {
		aiWorldAoi := world.GetAiWorldAoi()
		pos := GAME.GetPlayerPos(player)
		gid := aiWorldAoi.GetGidByPos(float32(pos.X), float32(pos.Y), float32(pos.Z))
		if gid == math.MaxUint32 {
			return
		}
		gridList := aiWorldAoi.GetSurrGridListByGid(gid, 1)
		for _, grid := range gridList {
			objectList := grid.GetObjectList()
			for uid, wa := range objectList {
				playerMap := world.GetAllPlayer()
				_, exist := playerMap[uint32(uid)]
				if !exist {
					logger.Error("remove not in world player cause by aoi bug, niw uid: %v, niw wa: %+v, uid: %v", uid, wa, player.PlayerId)
					delete(objectList, uid)
				}
			}
		}
	}
	if WORLD_MANAGER.IsAiWorld(world) && cmdId != cmd.CombatInvocationsNotify {
		if invokeHandler.AllLen() > 0 {
			reflection.SetStructFieldValue(newNtf, forwardField, invokeHandler.EntryListForwardAll)
			GAME.SendToSceneACV(scene, cmdId, player.ClientSeq, newNtf, 0, player.ClientVersion)
		}
		if invokeHandler.AllExceptCurLen() > 0 {
			reflection.SetStructFieldValue(newNtf, forwardField, invokeHandler.EntryListForwardAllExceptCur)
			GAME.SendToSceneACV(scene, cmdId, player.ClientSeq, newNtf, player.PlayerId, player.ClientVersion)
		}
		if invokeHandler.HostLen() > 0 {
			reflection.SetStructFieldValue(newNtf, forwardField, invokeHandler.EntryListForwardHost)
			GAME.SendToWorldH(world, cmdId, player.ClientSeq, newNtf)
		}
		return
	}
	if invokeHandler.AllLen() > 0 {
		reflection.SetStructFieldValue(newNtf, forwardField, invokeHandler.EntryListForwardAll)
		GAME.SendToSceneA(scene, cmdId, player.ClientSeq, newNtf, 0)
	}
	if invokeHandler.AllExceptCurLen() > 0 {
		reflection.SetStructFieldValue(newNtf, forwardField, invokeHandler.EntryListForwardAllExceptCur)
		GAME.SendToSceneA(scene, cmdId, player.ClientSeq, newNtf, player.PlayerId)
	}
	if invokeHandler.HostLen() > 0 {
		reflection.SetStructFieldValue(newNtf, forwardField, invokeHandler.EntryListForwardHost)
		GAME.SendToWorldH(world, cmdId, player.ClientSeq, newNtf)
	}
}

func (g *Game) UnionCmdNotify(player *model.Player, payloadMsg pb.Message) {
	ntf := payloadMsg.(*proto.UnionCmdNotify)
	_ = ntf
	if player.SceneLoadState != model.SceneEnterDone {
		return
	}
	DoForward[proto.CombatInvokeEntry](player, player.CombatInvokeHandler,
		cmd.CombatInvocationsNotify, new(proto.CombatInvocationsNotify), "InvokeList",
		nil, nil)
	DoForward[proto.AbilityInvokeEntry](player, player.AbilityInvokeHandler,
		cmd.AbilityInvocationsNotify, new(proto.AbilityInvocationsNotify), "Invokes",
		nil, nil)
	player.CombatInvokeHandler.Clear()
	player.AbilityInvokeHandler.Clear()
}

func (g *Game) CombatInvocationsNotify(player *model.Player, payloadMsg pb.Message) {
	ntf := payloadMsg.(*proto.CombatInvocationsNotify)
	if player.SceneLoadState != model.SceneEnterDone {
		return
	}
	world := WORLD_MANAGER.GetWorldById(player.WorldId)
	if world == nil {
		return
	}
	scene := world.GetSceneById(player.GetSceneId())
	for _, entry := range ntf.InvokeList {
		switch entry.ArgumentType {
		case proto.CombatTypeArgument_COMBAT_EVT_BEING_HIT:
			evtBeingHitInfo := new(proto.EvtBeingHitInfo)
			err := pb.Unmarshal(entry.CombatData, evtBeingHitInfo)
			if err != nil {
				logger.Error("parse EvtBeingHitInfo error: %v", err)
				break
			}
			g.handleEvtBeingHit(player, scene, evtBeingHitInfo)
		case proto.CombatTypeArgument_ENTITY_MOVE:
			entityMoveInfo := new(proto.EntityMoveInfo)
			err := pb.Unmarshal(entry.CombatData, entityMoveInfo)
			if err != nil {
				logger.Error("parse EntityMoveInfo error: %v", err)
				break
			}
			motionInfo := entityMoveInfo.MotionInfo
			if motionInfo == nil || motionInfo.Pos == nil || motionInfo.Rot == nil {
				break
			}
			g.handleEntityMove(player, world, scene, entityMoveInfo.EntityId, &model.Vector{
				X: float64(motionInfo.Pos.X),
				Y: float64(motionInfo.Pos.Y),
				Z: float64(motionInfo.Pos.Z),
			}, &model.Vector{
				X: float64(motionInfo.Rot.X),
				Y: float64(motionInfo.Rot.Y),
				Z: float64(motionInfo.Rot.Z),
			}, false, entityMoveInfo)
			// 众里寻他千百度 蓦然回首 那人却在灯火阑珊处
			if motionInfo.State == proto.MotionState_MOTION_NOTIFY || motionInfo.State == proto.MotionState_MOTION_FIGHT {
				// 只要转发了这两个包的其中之一 客户端的动画就会被打断
				continue
			}
		case proto.CombatTypeArgument_COMBAT_ANIMATOR_PARAMETER_CHANGED:
			evtAnimatorParameterInfo := new(proto.EvtAnimatorParameterInfo)
			err := pb.Unmarshal(entry.CombatData, evtAnimatorParameterInfo)
			if err != nil {
				logger.Error("parse EvtAnimatorParameterInfo error: %v", err)
				break
			}
		case proto.CombatTypeArgument_COMBAT_ANIMATOR_STATE_CHANGED:
			evtAnimatorStateChangedInfo := new(proto.EvtAnimatorStateChangedInfo)
			err := pb.Unmarshal(entry.CombatData, evtAnimatorStateChangedInfo)
			if err != nil {
				logger.Error("parse EvtAnimatorStateChangedInfo error: %v", err)
				break
			}
		}
		player.CombatInvokeHandler.AddEntry(entry.ForwardType, entry)
	}
}

func (g *Game) handleEvtBeingHit(player *model.Player, scene *Scene, hitInfo *proto.EvtBeingHitInfo) {
	// 触发事件
	if PLUGIN_MANAGER.TriggerEvent(PluginEventIdEvtBeingHit, &PluginEventEvtBeingHit{
		PluginEvent: NewPluginEvent(),
		Player:      player,
		HitInfo:     hitInfo,
	}) {
		return
	}

	attackResult := hitInfo.AttackResult
	if attackResult == nil {
		return
	}
	if attackResultTemplate == nil {
		attackResultTemplate = attackResult
	}
	defEntity := scene.GetEntity(attackResult.DefenseId)
	if defEntity == nil {
		return
	}
	switch defEntity.(type) {
	case *AvatarEntity:
		avatarEntity := defEntity.(*AvatarEntity)
		g.SubPlayerAvatarHp(player.PlayerId, avatarEntity.GetAvatarId(), attackResult.Damage, false, proto.ChangHpReason_CHANGE_HP_SUB_MONSTER)
	case *MonsterEntity:
		monsterEntity := defEntity.(*MonsterEntity)
		if scene.GetMonsterWudi() {
			return
		}
		fightProp := defEntity.GetFightProp()
		currHp := fightProp[constant.FIGHT_PROP_CUR_HP]
		maxHp := fightProp[constant.FIGHT_PROP_MAX_HP]
		lastHp := currHp
		currHp -= attackResult.Damage
		if currHp < 0.0 {
			currHp = 0.0
		}
		fightProp[constant.FIGHT_PROP_CUR_HP] = currHp
		g.EntityFightPropUpdateNotifyBroadcast(scene, defEntity)
		if currHp == 0.0 {
			g.KillEntity(player, scene, defEntity.GetId(), proto.PlayerDieType_PLAYER_DIE_GM)
		}
		if defEntity.GetGroupId() == 0 {
			return
		}
		monsterDataConfig := gdconf.GetMonsterDataById(int32(monsterEntity.GetMonsterId()))
		if monsterDataConfig == nil {
			logger.Error("get monster data config is nil, monsterId: %v", monsterEntity.GetMonsterId())
			return
		}
		lastHpPercent := lastHp / maxHp * 100.0
		currHpPercent := currHp / maxHp * 100.0
		for _, hpDrop := range monsterDataConfig.HpDropList {
			if hpDrop.HpPercent >= int32(currHpPercent) && hpDrop.HpPercent <= int32(lastHpPercent) {
				g.monsterDrop(player, MonsterDropTypeHp, hpDrop.Id, defEntity)
			}
		}
	case IGadgetEntity:
		iGadgetEntity := defEntity.(IGadgetEntity)
		gadgetDataConfig := gdconf.GetGadgetDataById(int32(iGadgetEntity.GetGadgetId()))
		if gadgetDataConfig == nil {
			logger.Error("get gadget data config is nil, gadgetId: %v", iGadgetEntity.GetGadgetId())
			return
		}
		logger.Debug("[EvtBeingHit] GadgetData: %+v, EntityId: %v, uid: %v", gadgetDataConfig, defEntity.GetId(), player.PlayerId)
		if gadgetDataConfig.ServerLuaScript != "" {
			gadgetLuaConfig := gdconf.GetGadgetLuaConfigByName(gadgetDataConfig.ServerLuaScript)
			if gadgetLuaConfig == nil {
				logger.Error("get gadget lua config is nil, name: %v", gadgetDataConfig.ServerLuaScript)
				return
			}
			isHost := player.PlayerId == scene.GetWorld().GetOwner().PlayerId
			CallGadgetLuaFunc(gadgetLuaConfig.LuaState, "OnBeHurt", &LuaCtx{uid: player.PlayerId}, attackResult.ElementType, 0, isHost)
		}
	}
}

func (g *Game) handleEntityMove(player *model.Player, world *World, scene *Scene, entityId uint32, pos, rot *model.Vector, force bool, moveInfo *proto.EntityMoveInfo) {
	entity := scene.GetEntity(entityId)
	if entity == nil {
		return
	}
	avatarEntity, ok := entity.(*AvatarEntity)
	if ok {
		// 玩家实体在移动
		if avatarEntity.GetUid() != player.PlayerId {
			return
		}
		oldPos := g.GetPlayerPos(player)
		if !WORLD_MANAGER.IsAiWorld(world) {
			if !world.IsValidScenePos(scene.GetId(), float32(pos.X), 0.0, float32(pos.Z)) {
				return
			}
			wait := g.LoadSceneBlockAsync(player, scene, scene, oldPos, pos, "SceneBlockAoiPlayerMove", &SceneBlockLoadInfoCtx{
				World:          world,
				Scene:          scene,
				OldPos:         oldPos,
				NewPos:         pos,
				AvatarEntityId: entity.GetId(),
			})
			if wait {
				return
			}
			g.SceneBlockAoiPlayerMove(player, world, scene, oldPos, pos, entity.GetId())
		} else {
			if !world.IsValidAiWorldPos(scene.GetId(), float32(pos.X), float32(pos.Y), float32(pos.Z)) {
				return
			}
			g.AiWorldAoiPlayerMove(player, world, scene, oldPos, pos)
		}
		// 场景天气区域变更检测
		g.SceneWeatherAreaCheck(player, oldPos, pos)
		// 更新玩家角色实体的位置信息
		for _, worldAvatar := range world.GetPlayerWorldAvatarList(player) {
			worldAvatarEntityId := worldAvatar.GetAvatarEntityId()
			worldAvatarEntity := scene.GetEntity(worldAvatarEntityId)
			if worldAvatarEntity == nil {
				logger.Error("world avatar entity is nil, worldAvatar: %+v, uid: %v", worldAvatar, player.PlayerId)
				continue
			}
			worldAvatarEntity.SetPos(pos)
			worldAvatarEntity.SetRot(rot)
			worldWeaponEntityId := worldAvatar.GetWeaponEntityId()
			worldWeaponEntity := scene.GetEntity(worldWeaponEntityId)
			if worldWeaponEntity == nil {
				logger.Error("world weapon entity is nil, worldAvatar: %+v, uid: %v", worldAvatar, player.PlayerId)
				continue
			}
			worldWeaponEntity.SetPos(pos)
			worldWeaponEntity.SetRot(rot)
		}
	}
	// 更新场景实体的位置信息
	entity.SetPos(pos)
	entity.SetRot(rot)
	if !force {
		motionInfo := moveInfo.MotionInfo
		switch entity.(type) {
		case *AvatarEntity:
			switch motionInfo.State {
			case proto.MotionState_MOTION_STANDBY, proto.MotionState_MOTION_WALK, proto.MotionState_MOTION_RUN, proto.MotionState_MOTION_DASH,
				58, 59, 60, 64, 65, 66:
				// 更新玩家安全位置
				player.SetPos(pos)
				player.SetRot(rot)
			}
			// 处理耐力消耗
			g.ImmediateStamina(player, motionInfo.State)
			// 坠落撞击扣血
			if player.Speed == nil {
				player.Speed = &model.Vector{X: float64(motionInfo.Speed.X), Y: float64(motionInfo.Speed.Y), Z: float64(motionInfo.Speed.Z)}
			}
			if motionInfo.State == proto.MotionState_MOTION_FALL_ON_GROUND || motionInfo.State == proto.MotionState_MOTION_FIGHT {
				oldSpeed := &alg.Vector3{X: float32(player.Speed.X), Y: float32(player.Speed.Y), Z: float32(player.Speed.Z)}
				newSpeed := &alg.Vector3{X: motionInfo.Speed.X, Y: motionInfo.Speed.Y, Z: motionInfo.Speed.Z}
				deltaSpeed := alg.Vector3Sub(oldSpeed, newSpeed)
				deltaSpeedMag := alg.Vector3Magnitude(deltaSpeed)
				if deltaSpeedMag > 20.0 {
					logger.Debug("player fall on ground, deltaSpeed: %+v, deltaSpeedMag: %v, uid: %v", deltaSpeed, deltaSpeedMag, player.PlayerId)
					// 速度矢量20-30部分线性映射到最大生命值百分比扣血
					rate := deltaSpeedMag - 20.0
					if rate > 10.0 {
						rate = 10.0
					}
					rate /= 10.0
					// 下落攻击最大生命值百分比扣血上限为40
					if motionInfo.State == proto.MotionState_MOTION_FIGHT {
						if rate > 0.4 {
							rate = 0.4
						}
					}
					fightProp := entity.GetFightProp()
					maxHp := fightProp[constant.FIGHT_PROP_MAX_HP]
					g.SubPlayerAvatarHp(player.PlayerId, avatarEntity.GetAvatarId(), maxHp*rate, false, proto.ChangHpReason_CHANGE_HP_SUB_FALL)
				}
			}
			player.Speed = &model.Vector{X: float64(motionInfo.Speed.X), Y: float64(motionInfo.Speed.Y), Z: float64(motionInfo.Speed.Z)}
		case IGadgetEntity:
			_, ok := entity.(*GadgetVehicleEntity)
			if ok {
				// 处理耐力消耗
				g.ImmediateStamina(player, motionInfo.State)
				// 处理载具销毁请求
				g.VehicleDestroyMotion(player, entity, motionInfo.State)
			}
		}
		entity.SetMoveState(uint16(motionInfo.State))
		entity.SetLastMoveSceneTimeMs(moveInfo.SceneTime)
		entity.SetLastMoveReliableSeq(moveInfo.ReliableSeq)
	}
}

func (g *Game) SceneBlockAoiPlayerMove(player *model.Player, world *World, scene *Scene, oldPos *model.Vector, newPos *model.Vector, avatarEntityId uint32) {
	// 加载和卸载的group
	oldNeighborGroupMap := g.GetNeighborGroup(player.GetSceneId(), oldPos)
	newNeighborGroupMap := g.GetNeighborGroup(player.GetSceneId(), newPos)
	otherPlayerNeighborGroupMap := make(map[uint32]*gdconf.Group)
	for _, otherPlayer := range scene.GetAllPlayer() {
		otherPlayerPos := g.GetPlayerPos(otherPlayer)
		for k, v := range g.GetNeighborGroup(player.GetSceneId(), otherPlayerPos) {
			otherPlayerNeighborGroupMap[k] = v
		}
	}
	for groupId, groupConfig := range oldNeighborGroupMap {
		_, exist := newNeighborGroupMap[groupId]
		if exist {
			continue
		}
		// 旧有新没有的group即为卸载的
		if !world.IsMultiplayerWorld() {
			// 单人世界直接卸载group
			g.RemoveSceneGroup(player, scene, groupConfig)
		} else {
			// 多人世界group附近没有任何玩家则卸载
			_, exist = otherPlayerNeighborGroupMap[uint32(groupConfig.Id)]
			if !exist {
				g.RemoveSceneGroup(player, scene, groupConfig)
			}
		}
	}
	for groupId, groupConfig := range newNeighborGroupMap {
		_, exist := oldNeighborGroupMap[groupId]
		if exist {
			continue
		}
		// 新有旧没有的group即为加载的
		g.AddSceneGroup(player, scene, groupConfig)
	}
	// 消失和出现的场景实体
	oldVisionEntityMap := g.GetVisionEntity(scene, oldPos)
	newVisionEntityMap := g.GetVisionEntity(scene, newPos)
	delEntityIdList := make([]uint32, 0)
	for entityId := range oldVisionEntityMap {
		_, exist := newVisionEntityMap[entityId]
		if exist {
			continue
		}
		// 旧有新没有的实体即为消失的
		delEntityIdList = append(delEntityIdList, entityId)
	}
	addEntityIdList := make([]uint32, 0)
	for entityId := range newVisionEntityMap {
		_, exist := oldVisionEntityMap[entityId]
		if exist {
			continue
		}
		// 新有旧没有的实体即为出现的
		addEntityIdList = append(addEntityIdList, entityId)
	}
	// 同步客户端消失和出现的场景实体
	if len(delEntityIdList) > 0 {
		g.RemoveSceneEntityNotifyToPlayer(player, proto.VisionType_VISION_MISS, delEntityIdList)
		for _, delEntityId := range delEntityIdList {
			entity := scene.GetEntity(delEntityId)
			avatarEntity, ok := entity.(*AvatarEntity)
			if !ok {
				continue
			}
			otherPlayer := USER_MANAGER.GetOnlineUser(avatarEntity.GetUid())
			if otherPlayer == nil {
				logger.Error("get player is nil, target uid: %v, uid: %v", avatarEntity.GetUid(), player.PlayerId)
				continue
			}
			g.RemoveSceneEntityNotifyToPlayer(otherPlayer, proto.VisionType_VISION_MISS, []uint32{avatarEntityId})
		}
	}
	if len(addEntityIdList) > 0 {
		g.AddSceneEntityNotify(player, proto.VisionType_VISION_MEET, addEntityIdList, false, false)
		for _, addEntityId := range addEntityIdList {
			entity := scene.GetEntity(addEntityId)
			avatarEntity, ok := entity.(*AvatarEntity)
			if !ok {
				continue
			}
			otherPlayer := USER_MANAGER.GetOnlineUser(avatarEntity.GetUid())
			if otherPlayer == nil {
				logger.Error("get player is nil, target uid: %v, uid: %v", avatarEntity.GetUid(), player.PlayerId)
				continue
			}
			sceneEntityInfoAvatar := g.PacketSceneEntityInfoAvatar(scene, player, scene.GetEntity(avatarEntityId).(*AvatarEntity).GetAvatarId())
			g.AddSceneEntityNotifyToPlayer(otherPlayer, proto.VisionType_VISION_MEET, []*proto.SceneEntityInfo{sceneEntityInfoAvatar})
		}
	}
	// 场景区域触发器检测
	g.SceneRegionTriggerCheck(player, oldPos, newPos, avatarEntityId)
}

func (g *Game) AiWorldAoiPlayerMove(player *model.Player, world *World, scene *Scene, oldPos *model.Vector, newPos *model.Vector) {
	aiWorldAoi := world.GetAiWorldAoi()
	oldGid := aiWorldAoi.GetGidByPos(float32(oldPos.X), float32(oldPos.Y), float32(oldPos.Z))
	newGid := aiWorldAoi.GetGidByPos(float32(newPos.X), float32(newPos.Y), float32(newPos.Z))
	if oldGid != newGid {
		// 玩家跨越了格子
		logger.Debug("player cross ai world aoi grid, oldGid: %v, oldPos: %+v, newGid: %v, newPos: %+v, uid: %v",
			oldGid, oldPos, newGid, newPos, player.PlayerId)
		// 找出本次移动所带来的消失和出现的格子
		oldGridList := aiWorldAoi.GetSurrGridListByGid(oldGid, 1)
		newGridList := aiWorldAoi.GetSurrGridListByGid(newGid, 1)
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
		activeAvatarId := world.GetPlayerActiveAvatarId(player)
		activeWorldAvatar := world.GetPlayerWorldAvatar(player, activeAvatarId)
		// 老格子移除玩家
		logger.Debug("ai world aoi remove player, oldPos: %+v, uid: %v", oldPos, player.PlayerId)
		ok := aiWorldAoi.RemoveObjectFromGridByPos(int64(player.PlayerId), float32(oldPos.X), float32(oldPos.Y), float32(oldPos.Z))
		if !ok {
			logger.Error("ai world aoi remove player fail, uid: %v, pos: %+v", player.PlayerId, g.GetPlayerPos(player))
		}
		// 处理消失的格子
		for _, delGridId := range delGridIdList {
			// 通知自己 老格子里的其它玩家消失
			oldOtherWorldAvatarMap := aiWorldAoi.GetObjectListByGid(delGridId)
			delEntityIdList := make([]uint32, 0)
			for _, otherWorldAvatarAny := range oldOtherWorldAvatarMap {
				otherWorldAvatar := otherWorldAvatarAny.(*WorldAvatar)
				delEntityIdList = append(delEntityIdList, otherWorldAvatar.GetAvatarEntityId())
			}
			if len(delEntityIdList) > 0 {
				g.RemoveSceneEntityNotifyToPlayer(player, proto.VisionType_VISION_MISS, delEntityIdList)
			}
			// 通知老格子里的其它玩家 自己消失
			for otherPlayerId := range oldOtherWorldAvatarMap {
				otherPlayer := USER_MANAGER.GetOnlineUser(uint32(otherPlayerId))
				if otherPlayer == nil {
					logger.Error("get player is nil, target uid: %v, uid: %v", otherPlayerId, player.PlayerId)
					continue
				}
				g.RemoveSceneEntityNotifyToPlayer(otherPlayer, proto.VisionType_VISION_MISS, []uint32{activeWorldAvatar.GetAvatarEntityId()})
			}
		}
		// 处理出现的格子
		for _, addGridId := range addGridIdList {
			// 通知自己 新格子里的其他玩家出现
			newOtherWorldAvatarMap := aiWorldAoi.GetObjectListByGid(addGridId)
			addEntityIdList := make([]uint32, 0)
			for _, otherWorldAvatarAny := range newOtherWorldAvatarMap {
				otherWorldAvatar := otherWorldAvatarAny.(*WorldAvatar)
				addEntityIdList = append(addEntityIdList, otherWorldAvatar.GetAvatarEntityId())
			}
			if len(addEntityIdList) > 0 {
				g.AddSceneEntityNotify(player, proto.VisionType_VISION_MEET, addEntityIdList, false, false)
			}
			// 通知新格子里的其他玩家 自己出现
			for otherPlayerId := range newOtherWorldAvatarMap {
				otherPlayer := USER_MANAGER.GetOnlineUser(uint32(otherPlayerId))
				if otherPlayer == nil {
					logger.Error("get player is nil, target uid: %v, uid: %v", otherPlayerId, player.PlayerId)
					continue
				}
				sceneEntityInfoAvatar := g.PacketSceneEntityInfoAvatar(scene, player, activeAvatarId)
				g.AddSceneEntityNotifyToPlayer(otherPlayer, proto.VisionType_VISION_MEET, []*proto.SceneEntityInfo{sceneEntityInfoAvatar})
			}
		}
		// 新格子添加玩家
		logger.Debug("ai world aoi add player, newPos: %+v, uid: %v", newPos, player.PlayerId)
		ok = aiWorldAoi.AddObjectToGridByPos(int64(player.PlayerId), activeWorldAvatar, float32(newPos.X), float32(newPos.Y), float32(newPos.Z))
		if !ok {
			logger.Error("ai world aoi add player fail, uid: %v, pos: %+v", player.PlayerId, g.GetPlayerPos(player))
		}
	}
	// 消失和出现的场景实体
	oldVisionEntityMap := g.GetVisionEntity(scene, oldPos)
	newVisionEntityMap := g.GetVisionEntity(scene, newPos)
	delEntityIdList := make([]uint32, 0)
	for entityId, entity := range oldVisionEntityMap {
		_, ok := entity.(*AvatarEntity)
		if ok {
			continue
		}
		_, exist := newVisionEntityMap[entityId]
		if exist {
			continue
		}
		// 旧有新没有的实体即为消失的
		delEntityIdList = append(delEntityIdList, entityId)
	}
	addEntityIdList := make([]uint32, 0)
	for entityId, entity := range newVisionEntityMap {
		_, ok := entity.(*AvatarEntity)
		if ok {
			continue
		}
		_, exist := oldVisionEntityMap[entityId]
		if exist {
			continue
		}
		// 新有旧没有的实体即为出现的
		addEntityIdList = append(addEntityIdList, entityId)
	}
	// 同步客户端消失和出现的场景实体
	if len(delEntityIdList) > 0 {
		g.RemoveSceneEntityNotifyToPlayer(player, proto.VisionType_VISION_MISS, delEntityIdList)
	}
	if len(addEntityIdList) > 0 {
		g.AddSceneEntityNotify(player, proto.VisionType_VISION_MEET, addEntityIdList, false, false)
	}
}

func (g *Game) AbilityInvocationsNotify(player *model.Player, payloadMsg pb.Message) {
	ntf := payloadMsg.(*proto.AbilityInvocationsNotify)
	if player.SceneLoadState != model.SceneEnterDone {
		return
	}
	for _, entry := range ntf.Invokes {
		player.AbilityInvokeHandler.AddEntry(entry.ForwardType, entry)
	}
	for _, entry := range ntf.Invokes {
		g.handleAbilityInvoke(player, entry)
	}
}

func (g *Game) ClientAbilityInitFinishNotify(player *model.Player, payloadMsg pb.Message) {
	ntf := payloadMsg.(*proto.ClientAbilityInitFinishNotify)
	if player.SceneLoadState != model.SceneEnterDone {
		return
	}
	invokeHandler := model.NewInvokeHandler[proto.AbilityInvokeEntry]()
	for _, entry := range ntf.Invokes {
		invokeHandler.AddEntry(entry.ForwardType, entry)
	}
	DoForward[proto.AbilityInvokeEntry](player, invokeHandler,
		cmd.ClientAbilityInitFinishNotify, new(proto.ClientAbilityInitFinishNotify), "Invokes",
		ntf, []string{"EntityId"})
	for _, entry := range ntf.Invokes {
		g.handleAbilityInvoke(player, entry)
	}
}

func (g *Game) ClientAbilityChangeNotify(player *model.Player, payloadMsg pb.Message) {
	ntf := payloadMsg.(*proto.ClientAbilityChangeNotify)
	if player.SceneLoadState != model.SceneEnterDone {
		return
	}
	invokeHandler := model.NewInvokeHandler[proto.AbilityInvokeEntry]()
	for _, entry := range ntf.Invokes {
		invokeHandler.AddEntry(entry.ForwardType, entry)
	}
	DoForward[proto.AbilityInvokeEntry](player, invokeHandler,
		cmd.ClientAbilityChangeNotify, new(proto.ClientAbilityChangeNotify), "Invokes",
		ntf, []string{"IsInitHash", "EntityId"})
	for _, entry := range ntf.Invokes {
		g.handleAbilityInvoke(player, entry)
		world := WORLD_MANAGER.GetWorldById(player.WorldId)
		if world == nil {
			continue
		}
		worldAvatar := world.GetWorldAvatarByEntityId(entry.EntityId)
		if worldAvatar == nil {
			continue
		}
		switch entry.ArgumentType {
		case proto.AbilityInvokeArgument_ABILITY_META_ADD_NEW_ABILITY:
			abilityMetaAddAbility := new(proto.AbilityMetaAddAbility)
			err := pb.Unmarshal(entry.AbilityData, abilityMetaAddAbility)
			if err != nil {
				logger.Error("parse AbilityMetaAddAbility error: %v", err)
				continue
			}
			if abilityMetaAddAbility.Ability == nil {
				continue
			}
			worldAvatar.AddAbility(abilityMetaAddAbility.Ability)
		case proto.AbilityInvokeArgument_ABILITY_META_MODIFIER_CHANGE:
			abilityMetaModifierChange := new(proto.AbilityMetaModifierChange)
			err := pb.Unmarshal(entry.AbilityData, abilityMetaModifierChange)
			if err != nil {
				logger.Error("parse AbilityMetaModifierChange error: %v", err)
				continue
			}
			abilityAppliedModifier := &proto.AbilityAppliedModifier{
				ModifierLocalId:           abilityMetaModifierChange.ModifierLocalId,
				ParentAbilityName:         abilityMetaModifierChange.ParentAbilityName,
				ParentAbilityOverride:     abilityMetaModifierChange.ParentAbilityOverride,
				InstancedAbilityId:        entry.Head.InstancedAbilityId,
				InstancedModifierId:       entry.Head.InstancedModifierId,
				AttachedInstancedModifier: abilityMetaModifierChange.AttachedInstancedModifier,
				ApplyEntityId:             abilityMetaModifierChange.ApplyEntityId,
				IsAttachedParentAbility:   abilityMetaModifierChange.IsAttachedParentAbility,
				IsServerbuffModifier:      entry.Head.IsServerbuffModifier,
			}
			worldAvatar.AddModifier(abilityAppliedModifier)
		}
	}
}

func (g *Game) MassiveEntityElementOpBatchNotify(player *model.Player, payloadMsg pb.Message) {
	ntf := payloadMsg.(*proto.MassiveEntityElementOpBatchNotify)
	if player.SceneLoadState != model.SceneEnterDone {
		return
	}
	world := WORLD_MANAGER.GetWorldById(player.WorldId)
	if world == nil {
		return
	}
	scene := world.GetSceneById(player.GetSceneId())
	ntf.OpIdx = scene.GetMeeoIndex()
	scene.SetMeeoIndex(scene.GetMeeoIndex() + 1)
	g.SendToSceneA(scene, cmd.MassiveEntityElementOpBatchNotify, player.ClientSeq, ntf, 0)
}

func (g *Game) EvtDoSkillSuccNotify(player *model.Player, payloadMsg pb.Message) {
	ntf := payloadMsg.(*proto.EvtDoSkillSuccNotify)
	if player.SceneLoadState != model.SceneEnterDone {
		return
	}

	// 触发事件
	if PLUGIN_MANAGER.TriggerEvent(PluginEventIdEvtDoSkillSucc, &PluginEventEvtDoSkillSucc{
		PluginEvent: NewPluginEvent(),
		Player:      player,
		Ntf:         ntf,
	}) {
		return
	}

	// logger.Debug("EvtDoSkillSuccNotify: %+v", ntf)
	// 触发任务
	g.TriggerQuest(player, constant.QUEST_FINISH_COND_TYPE_SKILL, "", int32(ntf.SkillId))
	// 消耗元素能量
	world := WORLD_MANAGER.GetWorldById(player.WorldId)
	if world == nil {
		return
	}
	activeAvatarId := world.GetPlayerActiveAvatarId(player)
	dbAvatar := player.GetDbAvatar()
	avatar := dbAvatar.GetAvatarById(activeAvatarId)
	if avatar == nil {
		return
	}
	avatarSkillDataConfig := gdconf.GetAvatarEnergySkillConfig(avatar.SkillDepotId)
	if avatarSkillDataConfig == nil {
		return
	}
	if int32(ntf.SkillId) == avatarSkillDataConfig.AvatarSkillId {
		g.CostPlayerAvatarEnergy(player.PlayerId, activeAvatarId, 0, true)
	}
}

func (g *Game) EvtAvatarEnterFocusNotify(player *model.Player, payloadMsg pb.Message) {
	ntf := payloadMsg.(*proto.EvtAvatarEnterFocusNotify)
	if player.SceneLoadState != model.SceneEnterDone {
		return
	}
	// logger.Debug("EvtAvatarEnterFocusNotify: %+v", ntf)
	world := WORLD_MANAGER.GetWorldById(player.WorldId)
	if world == nil {
		return
	}
	scene := world.GetSceneById(player.GetSceneId())
	g.SendToSceneA(scene, cmd.EvtAvatarEnterFocusNotify, player.ClientSeq, ntf, 0)
}

func (g *Game) EvtAvatarUpdateFocusNotify(player *model.Player, payloadMsg pb.Message) {
	ntf := payloadMsg.(*proto.EvtAvatarUpdateFocusNotify)
	if player.SceneLoadState != model.SceneEnterDone {
		return
	}
	// logger.Debug("EvtAvatarUpdateFocusNotify: %+v", ntf)
	world := WORLD_MANAGER.GetWorldById(player.WorldId)
	if world == nil {
		return
	}
	scene := world.GetSceneById(player.GetSceneId())
	g.SendToSceneA(scene, cmd.EvtAvatarUpdateFocusNotify, player.ClientSeq, ntf, 0)
}

func (g *Game) EvtAvatarExitFocusNotify(player *model.Player, payloadMsg pb.Message) {
	ntf := payloadMsg.(*proto.EvtAvatarExitFocusNotify)
	if player.SceneLoadState != model.SceneEnterDone {
		return
	}
	// logger.Debug("EvtAvatarExitFocusNotify: %+v", ntf)
	world := WORLD_MANAGER.GetWorldById(player.WorldId)
	if world == nil {
		return
	}
	scene := world.GetSceneById(player.GetSceneId())
	g.SendToSceneA(scene, cmd.EvtAvatarExitFocusNotify, player.ClientSeq, ntf, 0)
}

func (g *Game) EvtEntityRenderersChangedNotify(player *model.Player, payloadMsg pb.Message) {
	ntf := payloadMsg.(*proto.EvtEntityRenderersChangedNotify)
	if player.SceneLoadState != model.SceneEnterDone {
		return
	}
	// logger.Debug("EvtEntityRenderersChangedNotify: %+v", ntf)
	world := WORLD_MANAGER.GetWorldById(player.WorldId)
	if world == nil {
		return
	}
	scene := world.GetSceneById(player.GetSceneId())
	g.SendToSceneA(scene, cmd.EvtEntityRenderersChangedNotify, player.ClientSeq, ntf, 0)
}

func (g *Game) EvtBulletDeactiveNotify(player *model.Player, payloadMsg pb.Message) {
	ntf := payloadMsg.(*proto.EvtBulletDeactiveNotify)
	if player.SceneLoadState != model.SceneEnterDone {
		return
	}
	// logger.Debug("EvtBulletDeactiveNotify: %+v", ntf)
	world := WORLD_MANAGER.GetWorldById(player.WorldId)
	if world == nil {
		return
	}
	scene := world.GetSceneById(player.GetSceneId())
	g.SendToSceneA(scene, cmd.EvtBulletDeactiveNotify, player.ClientSeq, ntf, 0)
}

func (g *Game) EvtBulletHitNotify(player *model.Player, payloadMsg pb.Message) {
	ntf := payloadMsg.(*proto.EvtBulletHitNotify)
	if player.SceneLoadState != model.SceneEnterDone {
		return
	}
	// logger.Debug("EvtBulletHitNotify: %+v", ntf)
	world := WORLD_MANAGER.GetWorldById(player.WorldId)
	if world == nil {
		return
	}
	scene := world.GetSceneById(player.GetSceneId())
	g.SendToSceneA(scene, cmd.EvtBulletHitNotify, player.ClientSeq, ntf, 0)

	// 触发事件
	if PLUGIN_MANAGER.TriggerEvent(PluginEventIdEvtBulletHit, &PluginEventEvtBulletHit{
		PluginEvent: NewPluginEvent(),
		Player:      player,
		Ntf:         ntf,
	}) {
		return
	}
}

func (g *Game) EvtBulletMoveNotify(player *model.Player, payloadMsg pb.Message) {
	ntf := payloadMsg.(*proto.EvtBulletMoveNotify)
	if player.SceneLoadState != model.SceneEnterDone {
		return
	}
	// logger.Debug("EvtBulletMoveNotify: %+v", ntf)
	world := WORLD_MANAGER.GetWorldById(player.WorldId)
	if world == nil {
		return
	}
	scene := world.GetSceneById(player.GetSceneId())
	g.SendToSceneA(scene, cmd.EvtBulletMoveNotify, player.ClientSeq, ntf, 0)
}

func (g *Game) EvtCreateGadgetNotify(player *model.Player, payloadMsg pb.Message) {
	ntf := payloadMsg.(*proto.EvtCreateGadgetNotify)
	if player.SceneLoadState != model.SceneEnterDone {
		return
	}
	// logger.Debug("EvtCreateGadgetNotify: %+v", ntf)
	world := WORLD_MANAGER.GetWorldById(player.WorldId)
	if world == nil {
		return
	}
	scene := world.GetSceneById(player.GetSceneId())
	if ntf.InitPos == nil {
		return
	}
	ok := scene.CreateEntityGadgetClient(&model.Vector{
		X: float64(ntf.InitPos.X),
		Y: float64(ntf.InitPos.Y),
		Z: float64(ntf.InitPos.Z),
	}, &model.Vector{
		X: float64(ntf.InitEulerAngles.X),
		Y: float64(ntf.InitEulerAngles.Y),
		Z: float64(ntf.InitEulerAngles.Z),
	}, ntf.EntityId, ntf.ConfigId, ntf.CampId, ntf.CampType, ntf.OwnerEntityId, ntf.TargetEntityId, ntf.PropOwnerEntityId)
	if !ok {
		return
	}
	g.AddSceneEntityNotify(player, proto.VisionType_VISION_BORN, []uint32{ntf.EntityId}, true, true)

	// 触发事件
	if PLUGIN_MANAGER.TriggerEvent(PluginEventIdEvtCreateGadget, &PluginEventEvtCreateGadget{
		PluginEvent: NewPluginEvent(),
		Player:      player,
		Ntf:         ntf,
	}) {
		return
	}
}

func (g *Game) EvtDestroyGadgetNotify(player *model.Player, payloadMsg pb.Message) {
	ntf := payloadMsg.(*proto.EvtDestroyGadgetNotify)
	if player.SceneLoadState != model.SceneEnterDone {
		return
	}
	// logger.Debug("EvtDestroyGadgetNotify: %+v", ntf)
	world := WORLD_MANAGER.GetWorldById(player.WorldId)
	if world == nil {
		return
	}
	scene := world.GetSceneById(player.GetSceneId())
	scene.DestroyEntity(ntf.EntityId)
	g.RemoveSceneEntityNotifyBroadcast(scene, proto.VisionType_VISION_MISS, []uint32{ntf.EntityId}, 0)
}

func (g *Game) EvtAiSyncSkillCdNotify(player *model.Player, payloadMsg pb.Message) {
	ntf := payloadMsg.(*proto.EvtAiSyncSkillCdNotify)
	if player.SceneLoadState != model.SceneEnterDone {
		return
	}
	// logger.Debug("EvtAiSyncSkillCdNotify: %+v", ntf)
	world := WORLD_MANAGER.GetWorldById(player.WorldId)
	if world == nil {
		return
	}
	scene := world.GetSceneById(player.GetSceneId())
	g.SendToSceneA(scene, cmd.EvtAiSyncSkillCdNotify, player.ClientSeq, ntf, 0)
}

func (g *Game) EvtAiSyncCombatThreatInfoNotify(player *model.Player, payloadMsg pb.Message) {
	ntf := payloadMsg.(*proto.EvtAiSyncCombatThreatInfoNotify)
	if player.SceneLoadState != model.SceneEnterDone {
		return
	}
	// logger.Debug("EvtAiSyncCombatThreatInfoNotify: %+v", ntf)
	world := WORLD_MANAGER.GetWorldById(player.WorldId)
	if world == nil {
		return
	}
	scene := world.GetSceneById(player.GetSceneId())
	g.SendToSceneA(scene, cmd.EvtAiSyncCombatThreatInfoNotify, player.ClientSeq, ntf, 0)
}

func (g *Game) EntityConfigHashNotify(player *model.Player, payloadMsg pb.Message) {
	ntf := payloadMsg.(*proto.EntityConfigHashNotify)
	_ = ntf
}

func (g *Game) MonsterAIConfigHashNotify(player *model.Player, payloadMsg pb.Message) {
	ntf := payloadMsg.(*proto.MonsterAIConfigHashNotify)
	_ = ntf
}

func (g *Game) SetEntityClientDataNotify(player *model.Player, payloadMsg pb.Message) {
	ntf := payloadMsg.(*proto.SetEntityClientDataNotify)
	g.SendMsg(cmd.SetEntityClientDataNotify, player.PlayerId, player.ClientSeq, ntf)
}

func (g *Game) EntityAiSyncNotify(player *model.Player, payloadMsg pb.Message) {
	ntf := payloadMsg.(*proto.EntityAiSyncNotify)
	entityAiSyncNotify := &proto.EntityAiSyncNotify{
		InfoList: make([]*proto.AiSyncInfo, 0),
	}
	for _, monsterId := range ntf.LocalAvatarAlertedMonsterList {
		entityAiSyncNotify.InfoList = append(entityAiSyncNotify.InfoList, &proto.AiSyncInfo{
			EntityId:        monsterId,
			HasPathToTarget: true,
			IsSelfKilling:   false,
		})
	}
	g.SendMsg(cmd.EntityAiSyncNotify, player.PlayerId, player.ClientSeq, entityAiSyncNotify)
}

func (g *Game) handleAbilityInvoke(player *model.Player, entry *proto.AbilityInvokeEntry) {
	world := WORLD_MANAGER.GetWorldById(player.WorldId)
	if world == nil {
		return
	}
	scene := world.GetSceneById(player.GetSceneId())
	entity := scene.GetEntity(entry.EntityId)
	if entity == nil {
		return
	}
	if entry.Head.LocalId != 0 {
		logger.Debug("[LocalAbilityInvoke] type: %v, localId: %v, entityId: %v, uid: %v", entry.ArgumentType, entry.Head.LocalId, entity.GetId(), player.PlayerId)
		ability := entity.GetAbility(entry.Head.InstancedAbilityId)
		if ability == nil {
			logger.Error("get ability is nil, instancedAbilityId: %v, entityId: %v, uid: %v", entry.Head.InstancedAbilityId, entity.GetId(), player.PlayerId)
			return
		}
		abilityDataConfig := gdconf.GetAbilityDataByName(ability.abilityName)
		if abilityDataConfig == nil {
			logger.Error("get ability data config is nil, abilityName: %v", ability.abilityName)
			return
		}
		if entry.ArgumentType == proto.AbilityInvokeArgument_ABILITY_NONE || strings.Contains(entry.ArgumentType.String(), "ACTION") {
			actionDataConfig := abilityDataConfig.GetActionDataByLocalId(entry.Head.LocalId)
			if actionDataConfig == nil {
				logger.Error("get action data config is nil, localId: %v", entry.Head.LocalId)
				return
			}
			entity.AbilityAction(ability, actionDataConfig, entity)
		} else if strings.Contains(entry.ArgumentType.String(), "MIXIN") {
			mixinDataConfig := abilityDataConfig.GetMixinDataByLocalId(entry.Head.LocalId)
			if mixinDataConfig == nil {
				logger.Error("get mixin data config is nil, localId: %v", entry.Head.LocalId)
				return
			}
		} else {
			logger.Error("???")
		}
		return
	}
	switch entry.ArgumentType {
	case proto.AbilityInvokeArgument_ABILITY_META_OVERRIDE_PARAM:
		abilityScalarValueEntry := new(proto.AbilityScalarValueEntry)
		err := pb.Unmarshal(entry.AbilityData, abilityScalarValueEntry)
		if err != nil {
			logger.Error("parse AbilityScalarValueEntry error: %v", err)
			return
		}
		ability := entity.GetAbility(entry.Head.InstancedAbilityId)
		if ability == nil {
			logger.Error("get ability is nil, instancedAbilityId: %v, entityId: %v, uid: %v", entry.Head.InstancedAbilityId, entity.GetId(), player.PlayerId)
			return
		}
		if abilityScalarValueEntry.ValueType != proto.AbilityScalarType_ABILITY_SCALAR_TYPE_FLOAT {
			logger.Error("override param type not support, type: %v, uid: %v", abilityScalarValueEntry.ValueType, player.PlayerId)
			return
		}
		key := abilityScalarValueEntry.Key.GetHash()
		value := abilityScalarValueEntry.Value.(*proto.AbilityScalarValueEntry_FloatValue).FloatValue
		ability.abilitySpecialOverrideMap[key] = value
	case proto.AbilityInvokeArgument_ABILITY_META_REINIT_OVERRIDEMAP:
		reInitOverrideMap := new(proto.AbilityMetaReInitOverrideMap)
		err := pb.Unmarshal(entry.AbilityData, reInitOverrideMap)
		if err != nil {
			logger.Error("parse AbilityMetaReInitOverrideMap error: %v", err)
			return
		}
		ability := entity.GetAbility(entry.Head.InstancedAbilityId)
		if ability == nil {
			logger.Error("get ability is nil, instancedAbilityId: %v, entityId: %v, uid: %v", entry.Head.InstancedAbilityId, entity.GetId(), player.PlayerId)
			return
		}
		for _, abilityScalarValueEntry := range reInitOverrideMap.OverrideMap {
			if abilityScalarValueEntry.ValueType != proto.AbilityScalarType_ABILITY_SCALAR_TYPE_FLOAT {
				logger.Error("override param type not support, type: %v, uid: %v", abilityScalarValueEntry.ValueType, player.PlayerId)
				return
			}
			key := abilityScalarValueEntry.Key.GetHash()
			value := abilityScalarValueEntry.Value.(*proto.AbilityScalarValueEntry_FloatValue).FloatValue
			ability.abilitySpecialOverrideMap[key] = value
		}
	case proto.AbilityInvokeArgument_ABILITY_META_ADD_NEW_ABILITY:
		addAbility := new(proto.AbilityMetaAddAbility)
		err := pb.Unmarshal(entry.AbilityData, addAbility)
		if err != nil {
			logger.Error("parse AbilityMetaAddAbility error: %v", err)
			return
		}
		abilityNameHash := addAbility.Ability.AbilityName.GetHash()
		if abilityNameHash == 0 {
			logger.Error("ability name hash is 0, ability: %+v, entityId: %v, uid: %v", addAbility.Ability, entity.GetId(), player.PlayerId)
			return
		}
		abilityDataConfig := gdconf.GetAbilityDataByHash(abilityNameHash)
		if abilityDataConfig == nil {
			logger.Error("get abilityDataConfig is nil, abilityNameHash: %v", abilityNameHash)
			return
		}
		entity.AddAbility(abilityDataConfig.AbilityName, addAbility.Ability.InstancedAbilityId)
	case proto.AbilityInvokeArgument_ABILITY_META_MODIFIER_CHANGE:
		modifierChange := new(proto.AbilityMetaModifierChange)
		err := pb.Unmarshal(entry.AbilityData, modifierChange)
		if err != nil {
			logger.Error("parse AbilityMetaModifierChange error: %v", err)
			return
		}
		if modifierChange.Action == proto.ModifierAction_ADDED {
			ability := entity.GetAbility(entry.Head.InstancedAbilityId)
			if ability == nil {
				logger.Error("get ability is nil, instancedAbilityId: %v, entityId: %v, uid: %v", entry.Head.InstancedAbilityId, entity.GetId(), player.PlayerId)
				return
			}
			entity.AddModifier(ability, entry.Head.InstancedModifierId, uint32(entry.Head.ModifierConfigLocalId))
		} else if modifierChange.Action == proto.ModifierAction_REMOVED {
			entity.RemoveModifier(entry.Head.InstancedModifierId)
		}
	}
}
