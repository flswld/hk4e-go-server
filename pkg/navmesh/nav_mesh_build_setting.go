package navmesh

var (
	kMinAgentRadius float32 = 0.05
	kMinAgentHeight float32 = 0.001
	kMaxSlopeAngle  float32 = 60.0
	kMinCellSize    float32 = 0.01
)

const (
	kInvalidAgentTypeID int32 = -1
	kDefaultAgentTypeID int32 = 0
	kAreaCount          int32 = 32
)

type NavMeshBuildSettings struct {
	agentTypeID           int32
	agentRadius           float32
	agentHeight           float32
	agentSlope            float32
	agentClimb            float32
	ledgeDropHeight       float32
	maxJumpAcrossDistance float32

	// Advanced
	minRegionArea     float32
	manualCellSize    int32
	cellSize          float32
	manualTileSize    int32
	tileSize          int32
	accuratePlacement int32
}
