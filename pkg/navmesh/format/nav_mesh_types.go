package format

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

type NavMeshBuildSettings struct {
	AgentTypeID           int32
	AgentRadius           float32
	AgentHeight           float32
	AgentSlope            float32
	AgentClimb            float32
	LedgeDropHeight       float32
	MaxJumpAcrossDistance float32

	// Advanced
	MinRegionArea     float32
	ManualCellSize    int32
	CellSize          float32
	ManualTileSize    int32
	TileSize          int32
	AccuratePlacement int32
}

type AABB struct {
	M_Center Vector3f
	M_Extent Vector3f
}

type Vector3f struct {
	X, Y, Z float32
}

func (this *Vector3f) Set(i int, v float32) {
	switch i {
	case 0:
		this.X = v
	case 1:
		this.Y = v
	case 2:
		this.Z = v
	}
}

type Quaternionf struct {
	X, Y, Z, W float32
}

type NavMeshData struct {
	M_NavMeshDataID        uint32
	M_NavMeshBuildSettings NavMeshBuildSettings
	M_NavMeshTiles         []NavMeshTileData
	M_HeightMeshes         []HeightMeshData
	M_AdditionalData       AddtionalJosnData

	M_SourceBounds AABB
	M_Rotation     Quaternionf
	M_Position     Vector3f
	M_AgentTypeID  int32
}

func (this *NavMeshData) SetAdditionalData(data *AddtionalJosnData) {
	this.M_AdditionalData = *data
}

type NavMeshTileData struct {
	M_MeshData []byte
	M_Hash     [16]byte
}

type HeightMeshData struct {
	M_Vertices []Vector3f
	M_Indices  []int32
	M_Nodes    []HeightMeshBVNode
	M_Bounds   AABB
}

type AutoOffMeshLinkData struct {
	M_Start         Vector3f `json:"startPos"`
	M_End           Vector3f `json:"endPos"`
	M_Radius        float32  `json:"radius"`
	M_LinkType      uint16   `json:"linkType"`      // Off-mesh poly flags.
	M_Area          byte     `json:"area"`          // Off-mesh poly  area ids.
	M_LinkDirection bool     `json:"biDirectional"` // Off-mesh connection direction flags (NavMeshLinkDirectionFlags)
}

type HeightMeshBVNode struct {
	Min, Max Vector3f
	I, N     int32
}

type SceneObsData struct {
	Name     string      `json:"name"`
	Center   Vector3f    `json:"center"`
	Position Vector3f    `json:"position"`
	Scale    Vector3f    `json:"lossyScale"`
	Rotation Quaternionf `json:"rotation"`
	Size     Vector3f    `json:"size"`
	Radius   float32     `json:"radius"`
	Height   float32     `json:"height"`
	Shape    int32       `json:"shape"`
}

type AddtionalJosnData struct {
	OffMeshLinks []AutoOffMeshLinkData `json:"offMeshLinks"`
	AreaCosts    []float32             `json:"areaCosts"`
	ObsLists     []SceneObsData        `json:"obsList"`
}

func (this *AddtionalJosnData) GetObstacle(s string) SceneObsData {
	for _, data := range this.ObsLists {
		if data.Name == s {
			return data
		}
	}
	return SceneObsData{}
}

func LoadFromGobFile(file string) (*NavMeshData, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var data NavMeshData
	err = gob.NewDecoder(f).Decode(&data)
	if err != nil {
		return nil, err
	} else {
		return &data, nil
	}
}

func LoadFromJsonFile(file string) (*AddtionalJosnData, error) {
	var data AddtionalJosnData
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	err = json.NewDecoder(f).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func LoadFromTxtFile(file string) (*NavMeshData, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	reader := bufio.NewReader(f)
	data := new(NavMeshData)

	match1, err := regexp.Compile(`\s*([^\s]+) (.*) \((.+)\)`)
	if err != nil {
		return nil, err
	}

	match2, err := regexp.Compile(`\s*([^\s]+) \(([^\s]+)\) #[0-9]+: (.+)`)

	for {
		line, prefix, err := reader.ReadLine()
		if err == io.EOF {
			return data, nil
		} else if err != nil {
			return nil, err
		} else if prefix {
			return nil, fmt.Errorf("buffer is too small")
		}
		fields := strings.Fields(string(line))
		nfields := len(fields)
		if nfields > 0 && fields[nfields-1] == "NavMeshData" {
			err = readType(reader, match1, match2, reflect.ValueOf(data).Elem(), 1)
			return data, err
		}
	}
}

func readLine(reader *bufio.Reader, match *regexp.Regexp, tab int) (fieldName, fieldValue, fieldType string, err error) {
	var head []byte
	head, err = reader.Peek(tab + 1)
	if err != nil {
		return
	}

	if head[0] == '\r' || head[0] == '\n' {
		// 空行直接跳過
		reader.ReadLine()
		return readLine(reader, match, tab)
	}

	// 检查tab的数目，如果小于指定的则表示当前的type已经读完了
	for i := 0; i < tab; i++ {
		if head[i] != '\t' {
			err = io.EOF
			return
		}
	}

	// 如果当前的tab大于指定的，则表示这是个子类型
	if head[tab] == '\t' {
		err = io.EOF
		return
	}

	line, prefix, err := reader.ReadLine()
	if err != nil {
		return
	} else if prefix {
		err = fmt.Errorf("buffer is too small")
		return
	}

	fields := match.FindStringSubmatch(string(line))
	if len(fields) < 4 {
		err = fmt.Errorf("invalid line:%s", string(line))
		return
	} else {
		v := []rune(fields[1])
		v[0] = unicode.ToUpper(v[0])
		fieldName = string(v)
		fieldValue = fields[2]
		fieldType = fields[3]
	}
	return
}

func readType(reader *bufio.Reader, match1, match2 *regexp.Regexp, v reflect.Value, tab int) error {
	for {
		fieldName, fieldValue, fieldType, err := readLine(reader, match1, tab)
		if err == io.EOF {
			return nil
		} else if err != nil {
			return err
		}
		if fieldValue == "" {
			if !v.IsValid() || !v.FieldByName(fieldName).IsValid() {
				err := readType(reader, match1, match2, reflect.Value{}, tab+1)
				if err != nil {
					return err
				}
				continue
			}
			field := v.FieldByName(fieldName)
			switch fieldType {
			case "vector":
				_, value, _, err := readLine(reader, match1, tab+1)
				if err != nil {
					return err
				}
				length, _ := strconv.Atoi(value)
				d := reflect.MakeSlice(field.Type(), length, length)
				typ := field.Type().Elem().Kind()
				signed := true
				switch typ {
				case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
					signed = false
					fallthrough
				case reflect.Int8, reflect.Int16, reflect.Int, reflect.Int32, reflect.Int64:
					offset := 0
					for {
						_, _, buffer, err := readLine(reader, match2, tab+1)
						if err != nil {
							return err
						}
						data := strings.Fields(buffer)
						for i := 0; i < len(data); i++ {
							if signed {
								x, err := strconv.ParseInt(data[i], 10, 64)
								if err != nil {
									return err
								}
								d.Index(offset).SetInt(x)
							} else {
								x, err := strconv.ParseUint(data[i], 10, 64)
								if err != nil {
									return err
								}
								d.Index(offset).SetUint(x)
							}
							offset++
						}
						if offset == length {
							break
						}
					}
				default:
					for i := 0; i < length; i++ {
						a, _, _, err := readLine(reader, match1, tab+1)
						if a != "Data" {
							return fmt.Errorf("invalid vector format:%s", fieldName)
						}
						if err != nil {
							return err
						}
						err = readType(reader, match1, match2, d.Index(i), tab+2)
						if err != nil {
							return err
						}
					}
				}
				field.Set(d)
			case "Hash128":
				var hash [16]byte
				for i := 0; i < 16; i++ {
					_, value, _, err := readLine(reader, match1, tab+1)
					if err != nil {
						return err
					}
					b, err := strconv.Atoi(value)
					if err != nil {
						return err
					}
					hash[i] = byte(b)
				}
				field.Set(reflect.ValueOf(hash))
			default:
				err = readType(reader, match1, match2, field, tab+1)
				if err != nil {
					return err
				}
			}
		} else {
			if !v.IsValid() || !v.FieldByName(fieldName).IsValid() {
				continue
			}
			field := v.FieldByName(fieldName)
			if fieldType == "Vector3f" {
				data := strings.Fields(strings.Trim(fieldValue, "()"))
				if len(data) != 3 {
					return fmt.Errorf("invalid Vector3f %s", fieldValue)
				}
				v := new(Vector3f)
				for i := 0; i < 3; i++ {
					num, err := strconv.ParseFloat(data[i], 64)
					if err != nil {
						return err
					}
					v.Set(i, float32(num))
				}
				field.Set(reflect.ValueOf(*v))
				continue
			}
			switch field.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				num, err := strconv.ParseInt(fieldValue, 0, 64)
				if err != nil {
					return err
				}
				field.SetInt(num)
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				num, err := strconv.ParseUint(fieldValue, 0, 64)
				if err != nil {
					return err
				}
				field.SetUint(num)
			case reflect.Float32, reflect.Float64:
				num, err := strconv.ParseFloat(fieldValue, 64)
				if err != nil {
					return err
				}
				field.SetFloat(num)
			case reflect.String:
				field.SetString(fieldValue)
			default:
				return fmt.Errorf("invalid line:%s %s = %s", fieldName, fieldType, fieldValue)
			}
		}
	}
}

func LoadFromMhyFile(file string) (*NavMeshData, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	bin, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	data := NewUnpackerData(bin)
	stru, err := data.BeginUnpack()
	if err != nil {
		return nil, err
	}
	return &stru.Struct, nil
}

// Unpacker

// Read Int8 from BufferUnpacker.Buffer
func (b *BufferUnpacker) ReadInt8() (int8, error) {
	var data int8
	err := binary.Read(b.Buffer, binary.LittleEndian, &data)
	return data, err
}

// Read UInt8 from BufferUnpacker.Buffer
func (b *BufferUnpacker) ReadUInt8() (uint8, error) {
	var data uint8
	err := binary.Read(b.Buffer, binary.LittleEndian, &data)
	return data, err
}

// Read Int16 from BufferUnpacker.Buffer
func (b *BufferUnpacker) ReadInt16() (int16, error) {
	var data int16
	err := binary.Read(b.Buffer, binary.LittleEndian, &data)
	return data, err
}

// Read UInt16 from BufferUnpacker.Buffer
func (b *BufferUnpacker) ReadUInt16() (uint16, error) {
	var data uint16
	err := binary.Read(b.Buffer, binary.LittleEndian, &data)
	return data, err
}

// Read UInt32 from BufferUnpacker.Buffer
func (b *BufferUnpacker) ReadUInt32() (uint32, error) {
	var data uint32
	err := binary.Read(b.Buffer, binary.LittleEndian, &data)
	return data, err
}

// Read Int32 from BufferUnpacker.Buffer
func (b *BufferUnpacker) ReadInt32() (int32, error) {
	var data int32
	err := binary.Read(b.Buffer, binary.LittleEndian, &data)
	return data, err
}

// Read Float32 from BufferUnpacker.Buffer
func (b *BufferUnpacker) ReadFloat32() (float32, error) {
	var data float32
	err := binary.Read(b.Buffer, binary.LittleEndian, &data)

	// 如果数据是NaN，则转换为0
	if math.IsNaN(float64(data)) {
		nanBits := math.Float32bits(data)
		nanHex := fmt.Sprintf("%08x", nanBits)
		log.Printf("[TRACE] Unreadable Float32 found: %v (Hex: %s)\n", data, nanHex)
		data = 0
	}

	return data, err
}

// Read Hash from BufferUnpacker.Buffer, this type is 16 bytes
func (b *BufferUnpacker) ReadHash() ([16]byte, error) {
	var hashBytes [16]byte
	_, err := b.Buffer.Read(hashBytes[:])
	if err != nil {
		return [16]byte{}, err
	}
	return hashBytes, nil
}

func (b *BufferUnpacker) LeftBytes() int {
	return b.Buffer.Len()
}

// Workflow:
// (SceneID)NavMeshDataID | int32  | 4    bytes
// NavMeshTile            | struct | [:4] bytes for len
// NavMeshBuildSettings   | struct | [:4] bytes for len
// OffMeshLinks           | struct | [:4] bytes for len
func (b *BufferUnpacker) BeginUnpack() (*UnpackerReturn, error) {
	if err := b.UnpackNavMeshDataID(); err != nil {
		return nil, fmt.Errorf("Error reading m_NavMeshDataID: %v", err)
	}

	if err := b.UnpackNavMeshTilesData(); err != nil {
		return nil, fmt.Errorf("Error reading m_NavMeshTiles: %v", err)
	}

	if err := b.UnpackNavMeshBuildSettings(); err != nil {
		return nil, fmt.Errorf("Error reading m_NavMeshBuildSettings: %v", err)
	}

	if err := b.UnpackOffMeshLinks(); err != nil {
		return nil, fmt.Errorf("Error reading m_OffMeshLinks: %v", err)
	}

	return &UnpackerReturn{Struct: *b.Struct}, nil
}

func (b *BufferUnpacker) UnpackNavMeshDataID() error {
	meshID, err := b.ReadUInt32()
	if err != nil {
		return err
	}

	b.Struct.M_NavMeshDataID = meshID
	return nil
}

func (b *BufferUnpacker) UnpackNavMeshTilesData() error {
	var tiles []NavMeshTileData

	counts, err := b.ReadInt32()
	if err != nil {
		return err
	}

	for i := int32(0); i < counts; i++ {
		var tile NavMeshTileData

		size, err := b.ReadInt32()
		if err != nil {
			return err
		}

		data := make([]byte, int(size))
		if err := binary.Read(b.Buffer, binary.LittleEndian, &data); err != nil {
			return err
		}
		tile.M_MeshData = data

		hash, err := b.ReadHash()
		if err != nil {
			return err
		}
		tile.M_Hash = hash

		tiles = append(tiles, tile)
	}

	b.Struct.M_NavMeshTiles = tiles
	return nil
}

func (b *BufferUnpacker) UnpackNavMeshBuildSettings() error {
	settings := NavMeshBuildSettings{}
	if err := readStructFromBuffer(b, &settings); err != nil {
		return err
	}

	b.Struct.M_NavMeshBuildSettings = settings
	return nil
}

func (b *BufferUnpacker) UnpackOffMeshLinks() error {
	counts, err := b.ReadInt32()
	if err != nil {
		return err
	}

	var autoOffMeshLinks []AutoOffMeshLinkData
	for i := int32(0); i < counts; i++ {
		autoOffMeshLink := AutoOffMeshLinkData{}
		if err := readStructFromBuffer(b, &autoOffMeshLink); err != nil {
			return err
		}
		autoOffMeshLinks = append(autoOffMeshLinks, autoOffMeshLink)
	}

	b.Struct.M_AdditionalData.OffMeshLinks = autoOffMeshLinks
	return nil
}

func readStructFromBuffer(b *BufferUnpacker, s interface{}) error {
	structValue := reflect.ValueOf(s)
	if structValue.Kind() != reflect.Ptr || structValue.IsNil() {
		return errors.New("expecting a non-nil pointer to a struct")
	}

	structType := structValue.Elem().Type()

	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		value := structValue.Elem().Field(i)

		// Read corresponding type from BufferUnpacker
		switch field.Type.Kind() {
		case reflect.Bool:
			continue
		case reflect.Int8:
			val, err := b.ReadInt8()
			if err != nil {
				return err
			}
			value.Set(reflect.ValueOf(val))
		case reflect.Uint8:
			val, err := b.ReadUInt8()
			if err != nil {
				return err
			}
			value.Set(reflect.ValueOf(val))
		case reflect.Int16:
			val, err := b.ReadInt16()
			if err != nil {
				return err
			}
			value.Set(reflect.ValueOf(val))
		case reflect.Uint16:
			val, err := b.ReadUInt16()
			if err != nil {
				return err
			}
			value.Set(reflect.ValueOf(val))
		case reflect.Int32:
			val, err := b.ReadInt32()
			if err != nil {
				return err
			}
			value.Set(reflect.ValueOf(val))
		case reflect.Uint32:
			val, err := b.ReadUInt32()
			if err != nil {
				return err
			}
			value.Set(reflect.ValueOf(val))
		case reflect.Float32:
			val, err := b.ReadFloat32()
			if err != nil {
				return err
			}
			value.Set(reflect.ValueOf(val))
		case reflect.Struct:
			nestedStructPtr := reflect.New(field.Type).Interface()
			if err := readStructFromBuffer(b, nestedStructPtr); err != nil {
				return err
			}
			value.Set(reflect.ValueOf(nestedStructPtr).Elem())
		case reflect.Array:
			// Check if the field type is an array
			if field.Type.Kind() != reflect.Array {
				return fmt.Errorf("unsupported field type: %v", field.Type.Kind())
			}

			// Get the length of the array
			arrayLen := field.Type.Len()

			// Create a slice to store the elements of the array
			array := reflect.New(field.Type).Elem()

			// Iterate through each element of the array
			for j := 0; j < arrayLen; j++ {
				element := array.Index(j)
				// Read corresponding type from BufferUnpacker
				switch field.Type.Elem().Kind() {
				case reflect.Int8:
					val, err := b.ReadInt8()
					if err != nil {
						return err
					}
					element.Set(reflect.ValueOf(val))
				case reflect.Uint8:
					val, err := b.ReadUInt8()
					if err != nil {
						return err
					}
					element.Set(reflect.ValueOf(val))
				case reflect.Int16:
					val, err := b.ReadInt16()
					if err != nil {
						return err
					}
					element.Set(reflect.ValueOf(val))
				case reflect.Uint16:
					val, err := b.ReadUInt16()
					if err != nil {
						return err
					}
					element.Set(reflect.ValueOf(val))
				case reflect.Int32:
					val, err := b.ReadInt32()
					if err != nil {
						return err
					}
					element.Set(reflect.ValueOf(val))
				case reflect.Uint32:
					val, err := b.ReadUInt32()
					if err != nil {
						return err
					}
					element.Set(reflect.ValueOf(val))
				case reflect.Float32:
					val, err := b.ReadFloat32()
					if err != nil {
						return err
					}
					element.Set(reflect.ValueOf(val))
				case reflect.Struct:
					nestedStructPtr := reflect.New(field.Type).Interface()
					if err := readStructFromBuffer(b, nestedStructPtr); err != nil {
						return err
					}
					element.Set(reflect.ValueOf(nestedStructPtr).Elem())
				default:
					return fmt.Errorf("unsupported array element type: %v", field.Type.Elem().Kind())
				}
			}
			value.Set(array)
		default:
			return fmt.Errorf("unsupported field type: %v", field.Type.Kind())
		}
	}

	return nil
}

// Basic Unpacker Data Struct
type BufferUnpacker struct {
	Buffer *bytes.Buffer
	Struct *NavMeshData
}

func NewUnpackerData(buffer []byte) *BufferUnpacker {
	return &BufferUnpacker{
		Buffer: bytes.NewBuffer(buffer),
		Struct: &NavMeshData{},
	}
}

type UnpackerReturn struct {
	Struct NavMeshData
}
