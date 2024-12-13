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
	req := payloadMsg.(*proto.ObstacleModifyNotify)
	logger.Debug("obstacle modify req: %v, uid: %v, gateAppId: %v", req, userId, gateAppId)
}

type WorldStatic struct {
	navMeshManagerMap map[uint32]*navmesh.NavMeshManager
	// x y z -> if terrain exist
	terrain map[alg.MeshVector]struct{}
}

func NewWorldStatic() (r *WorldStatic) {
	r = new(WorldStatic)
	r.navMeshManagerMap = make(map[uint32]*navmesh.NavMeshManager)
	r.terrain = make(map[alg.MeshVector]struct{})
	return r
}

func (w *WorldStatic) InitTerrain() bool {
	fileList, err := os.ReadDir("./NavMesh")
	if err != nil {
		logger.Error("open navmesh dir error: %v", err)
	} else {
		for _, file := range fileList {
			if file.IsDir() {
				continue
			}
			fileName := file.Name()
			navMeshDataFormat, err := format.LoadFromMhyFile("./NavMesh/" + fileName)
			if err != nil {
				logger.Error("parse navmesh file error: %v", err)
				continue
			}
			navMeshManager, exist := w.navMeshManagerMap[navMeshDataFormat.M_NavMeshDataID]
			if !exist {
				navMeshManager = navmesh.NewNavMeshManager()
				w.navMeshManagerMap[navMeshDataFormat.M_NavMeshDataID] = navMeshManager
			}
			navMeshData := navmesh.NewDataFromFormat(navMeshDataFormat)
			err = navMeshManager.LoadData(navMeshData)
			if err != nil {
				logger.Error("load navmesh file error: %v", err)
				continue
			}
			logger.Info("load navmesh file ok, fileName: %v", fileName)
			runtime.GC()
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
			return false
		}
	}
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
	path := navmesh.NewNavMeshPath()
	navMeshManager, exist := w.navMeshManagerMap[sceneId]
	if !exist {
		logger.Error("navmesh scene not exist, sceneId: %v", sceneId)
		return nil, false
	}
	count := navMeshManager.CalculatePolygonPath(path, ConvPbVecToNavMeshVec(startPos), ConvPbVecToNavMeshVec(endPos), 30)
	if count == 0 {
		logger.Error("navmesh could not find path, sceneId: %v, startPos: %v, endPos: %v", sceneId, startPos, endPos)
		return nil, false
	}
	corners := make([]navmesh.Vector3f, count)
	count = navMeshManager.CalculatePathCorners(corners, count, path)
	if count == 0 {
		logger.Error("navmesh could not find path, sceneId: %v, startPos: %v, endPos: %v", sceneId, startPos, endPos)
		return nil, false
	}
	corners = corners[:count]
	return ConvNavMeshVecListToPbVecList(corners), true
}
