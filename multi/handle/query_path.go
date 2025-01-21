package handle

import (
	"bytes"
	"encoding/gob"
	"os"
	"runtime"

	"hk4e/pkg/alg"
	"hk4e/pkg/logger"
	"hk4e/pkg/navmesh"
	"hk4e/pkg/navmesh/format"
	"hk4e/protocol/cmd"
	"hk4e/protocol/proto"

	pb "google.golang.org/protobuf/proto"
)

func (h *Handle) QueryPath(userId uint32, gateAppId string, payloadMsg pb.Message) {
	req := payloadMsg.(*proto.QueryPathReq)
	logger.Debug("query path req: %v, uid: %v, gateAppId: %v", req, userId, gateAppId)
	for _, destinationPos := range req.DestinationPos {
		corners, ok := h.worldStatic.NavMeshPathfinding(req.SceneId, req.SourcePos, destinationPos)
		if ok {
			rsp := &proto.QueryPathRsp{
				QueryId:     req.QueryId,
				QueryStatus: proto.QueryPathRsp_STATUS_SUCC,
				Corners:     corners,
			}
			h.SendMsg(cmd.QueryPathRsp, userId, gateAppId, rsp)
			return
		}
	}
	rsp := &proto.QueryPathRsp{
		QueryId:     req.QueryId,
		QueryStatus: proto.QueryPathRsp_STATUS_FAIL,
	}
	h.SendMsg(cmd.QueryPathRsp, userId, gateAppId, rsp)
}

func (h *Handle) ObstacleModifyNotify(userId uint32, gateAppId string, payloadMsg pb.Message) {
	ntf := payloadMsg.(*proto.ObstacleModifyNotify)
	logger.Debug("obstacle modify ntf: %v, uid: %v, gateAppId: %v", ntf, userId, gateAppId)
	return
	navMeshManager, exist := h.worldStatic.navMeshManagerMap[ntf.SceneId]
	if !exist {
		return
	}
	navMeshObstacle, exist := h.worldStatic.navMeshObstacleMap[ntf.SceneId]
	if !exist {
		navMeshObstacle = new(NavMeshObstacle)
		navMeshObstacle.obstacleHandleMap = make(map[int32]int32)
		h.worldStatic.navMeshObstacleMap[ntf.SceneId] = navMeshObstacle
	}
	for _, obstacleId := range ntf.RemoveObstacleIds {
		handle, exist := navMeshObstacle.obstacleHandleMap[obstacleId]
		if !exist {
			continue
		}
		navMeshManager.RemoveObstacle(handle)
	}
	for _, pbObstacle := range ntf.AddObstacles {
		obstacle := navmesh.NewNavMeshObstacle(
			navmesh.NavMeshObstacleShape(pbObstacle.Shape),
			ConvPbVecToNavMeshVec(pbObstacle.Center),
			navmesh.Vector3_One,
			navmesh.NewQuaternionf(pbObstacle.Rotation.X, pbObstacle.Rotation.Y, pbObstacle.Rotation.Z, pbObstacle.Rotation.W),
		)
		obstacle.SetExtents(navmesh.NewVector3f(float32(pbObstacle.Extents.X), float32(pbObstacle.Extents.Y), float32(pbObstacle.Extents.Z)).Mulf(0.01))
		handle := navMeshManager.AddObstacle(obstacle)
		navMeshObstacle.obstacleHandleMap[pbObstacle.ObstacleId] = handle
	}
	navMeshManager.UpdateCarvingImmediately()
}

type NavMeshObstacle struct {
	obstacleHandleMap map[int32]int32
}

type WorldStatic struct {
	navMeshManagerMap  map[uint32]*navmesh.NavMeshManager
	navMeshObstacleMap map[uint32]*NavMeshObstacle
	// x y z -> if terrain exist
	terrain map[alg.MeshVector]struct{}
}

func NewWorldStatic() (r *WorldStatic) {
	r = new(WorldStatic)
	r.navMeshManagerMap = make(map[uint32]*navmesh.NavMeshManager)
	r.navMeshObstacleMap = make(map[uint32]*NavMeshObstacle)
	r.terrain = make(map[alg.MeshVector]struct{})
	return r
}

func (w *WorldStatic) InitTerrain() bool {
	fileList, err := os.ReadDir("./NavMesh")
	if err != nil {
		logger.Error("open navmesh dir error: %v", err)
	} else {
		navMeshDataMap := make(map[uint32]*format.NavMeshData)
		for _, file := range fileList {
			if file.IsDir() {
				continue
			}
			fileName := file.Name()
			navMeshDataFormat, err := format.LoadFromMhyFile("./NavMesh/" + fileName)
			if err != nil {
				logger.Error("parse navmesh file error: %v, fileName: %v", err, fileName)
				continue
			}
			sceneId := navMeshDataFormat.M_NavMeshDataID
			navMeshManager, exist := w.navMeshManagerMap[sceneId]
			if !exist {
				navMeshManager = navmesh.NewNavMeshManager()
				w.navMeshManagerMap[sceneId] = navMeshManager
			}
			if navMeshDataMap[sceneId] == nil {
				navMeshDataMap[sceneId] = navMeshDataFormat
			} else {
				navMeshDataMap[sceneId].M_NavMeshTiles = append(navMeshDataMap[sceneId].M_NavMeshTiles, navMeshDataFormat.M_NavMeshTiles...)
			}
			logger.Info("parse navmesh file ok, fileName: %v", fileName)
		}
		for sceneId, navMeshManager := range w.navMeshManagerMap {
			err = navMeshManager.LoadData(navmesh.NewDataFromFormat(navMeshDataMap[sceneId]))
			if err != nil {
				logger.Error("load navmesh data error: %v, sceneId: %v", err, sceneId)
				continue
			}
			logger.Info("load navmesh data ok, sceneId: %v", sceneId)
		}
	}
	data, err := os.ReadFile("./world_terrain.bin")
	if err != nil {
		logger.Error("read world terrain file error: %v", err)
	} else {
		decoder := gob.NewDecoder(bytes.NewReader(data))
		err = decoder.Decode(&w.terrain)
		if err != nil {
			logger.Error("unmarshal world terrain data error: %v", err)
		}
	}
	runtime.GC()
	return true
}

func (w *WorldStatic) SaveTerrain() bool {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(w.terrain)
	if err != nil {
		logger.Error("marshal world terrain data error: %v", err)
		return false
	}
	err = os.WriteFile("./world_terrain.bin", buffer.Bytes(), 0644)
	if err != nil {
		logger.Error("write world terrain file error: %v", err)
		return false
	}
	return true
}

func (w *WorldStatic) GetTerrain(x int16, y int16, z int16) (exist bool) {
	pos := alg.MeshVector{X: x, Y: y, Z: z}
	_, exist = w.terrain[pos]
	return exist
}

func (w *WorldStatic) SetTerrain(x int16, y int16, z int16) {
	pos := alg.MeshVector{X: x, Y: y, Z: z}
	w.terrain[pos] = struct{}{}
}

func ConvPbVecToSvoVec(pbVec *proto.Vector) alg.MeshVector {
	return alg.MeshVector{X: int16(pbVec.X), Y: int16(pbVec.Y), Z: int16(pbVec.Z)}
}

func ConvSvoVecToPbVec(svoVec alg.MeshVector) *proto.Vector {
	return &proto.Vector{X: float32(svoVec.X), Y: float32(svoVec.Y), Z: float32(svoVec.Z)}
}

func ConvPbVecListToSvoVecList(pbVecList []*proto.Vector) []alg.MeshVector {
	ret := make([]alg.MeshVector, 0, len(pbVecList))
	for _, pbVec := range pbVecList {
		ret = append(ret, ConvPbVecToSvoVec(pbVec))
	}
	return ret
}

func ConvSvoVecListToPbVecList(svoVecList []alg.MeshVector) []*proto.Vector {
	ret := make([]*proto.Vector, 0, len(svoVecList))
	for _, svoVec := range svoVecList {
		ret = append(ret, ConvSvoVecToPbVec(svoVec))
	}
	return ret
}

func (w *WorldStatic) SvoPathfinding(sceneId uint32, startPos *proto.Vector, endPos *proto.Vector) ([]*proto.Vector, bool) {
	defer func() {
		if err := recover(); err != nil {
			logger.Error("svo pathfinding error, panic, startPos: %v, endPos: %v", startPos, endPos)
		}
	}()
	bfs := alg.NewBFS()
	bfs.InitMap(
		w.terrain,
		ConvPbVecToSvoVec(startPos),
		ConvPbVecToSvoVec(endPos),
		0,
	)
	pathVectorList := bfs.Pathfinding()
	if pathVectorList == nil {
		logger.Error("svo could not find path, sceneId: %v, startPos: %v, endPos: %v", sceneId, startPos, endPos)
		return nil, false
	}
	return ConvSvoVecListToPbVecList(pathVectorList), true
}

func ConvPbVecToNavMeshVec(pbVec *proto.Vector) navmesh.Vector3f {
	var ret navmesh.Vector3f
	ret.Set(pbVec.X, pbVec.Y, pbVec.Z)
	return ret
}

func ConvNavMeshVecToPbVec(navMeshVec navmesh.Vector3f) *proto.Vector {
	return &proto.Vector{X: navMeshVec.GetData(0), Y: navMeshVec.GetData(1), Z: navMeshVec.GetData(2)}
}

func ConvPbVecListToNavMeshVecList(pbVecList []*proto.Vector) []navmesh.Vector3f {
	ret := make([]navmesh.Vector3f, 0, len(pbVecList))
	for _, pbVec := range pbVecList {
		ret = append(ret, ConvPbVecToNavMeshVec(pbVec))
	}
	return ret
}

func ConvNavMeshVecListToPbVecList(navMeshVecList []navmesh.Vector3f) []*proto.Vector {
	ret := make([]*proto.Vector, 0, len(navMeshVecList))
	for _, navMeshVec := range navMeshVecList {
		ret = append(ret, ConvNavMeshVecToPbVec(navMeshVec))
	}
	return ret
}

func (w *WorldStatic) NavMeshPathfinding(sceneId uint32, startPos *proto.Vector, endPos *proto.Vector) ([]*proto.Vector, bool) {
	navMeshManager, exist := w.navMeshManagerMap[sceneId]
	if !exist {
		logger.Error("navmesh scene not exist, sceneId: %v", sceneId)
		return nil, false
	}
	var hit navmesh.NavMeshHit
	ok := navMeshManager.SamplePosition(&hit, ConvPbVecToNavMeshVec(startPos), 1)
	if !ok {
		logger.Error("navmesh could not find path, sceneId: %v, startPos: %v, endPos: %v", sceneId, startPos, endPos)
		return nil, false
	}
	source := hit.GetPosition()
	ok = navMeshManager.SamplePosition(&hit, ConvPbVecToNavMeshVec(endPos), 1)
	if !ok {
		logger.Error("navmesh could not find path, sceneId: %v, startPos: %v, endPos: %v", sceneId, startPos, endPos)
		return nil, false
	}
	target := hit.GetPosition()
	corners, fail := navMeshManager.CalculatePath(source, target, 100)
	if fail {
		logger.Error("navmesh could not find path, sceneId: %v, startPos: %v, endPos: %v", sceneId, startPos, endPos)
		return nil, false
	}
	return ConvNavMeshVecListToPbVecList(corners), true
}
