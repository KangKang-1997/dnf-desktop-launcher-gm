package pvf

import (
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"hash"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/longbridgeapp/opencc"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
)

type cacheBuild struct {
	MD5            string
	Path           string
	FileSize       int64
	Encode         string
	StackableCount int
	EquipmentCount int
	ExpLevelCount  int
	Items          []cacheItem
	Data           map[string]any
	Logs           []string
}

type cacheItem struct {
	ID     int
	Type   string
	Name   string
	Detail map[string]any
}

type pvfHeader struct {
	path               string
	file               *os.File
	dirTree            []byte
	filePackIndexShift int64
	numFiles           uint32
}

type leaf struct {
	path           string
	fileLength     uint32
	fileCRC32      uint32
	relativeOffset uint32
}

type stringTable struct {
	raw    []byte
	length uint32
	encode string
	cache  map[int]string
	conv   func(string) string
}

type strTable struct {
	rows map[string]string
}

type lstFile struct {
	tableList [][2]any
	tableDict map[int]string
	strDict   map[int]strTable
	pvf       *tinyPVF
	baseDir   string
	encode    string
}

type tinyPVF struct {
	header          *pvfHeader
	fileTree        map[string]leaf
	fileContent     map[string][]byte
	strings         *stringTable
	nStrings        *lstFile
	encode          string
	traditionalConv func(string) string
}

func buildCache(pvfPath string, encodeName string) (cacheBuild, error) {
	encodeName, err := normalizePVFEncoding(encodeName)
	if err != nil {
		return cacheBuild{}, err
	}
	path := filepath.Clean(pvfPath)
	info, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cacheBuild{}, err
		}
		return cacheBuild{}, err
	}
	if info.IsDir() {
		return cacheBuild{}, fmt.Errorf("PVF path must point to a .pvf file")
	}
	if strings.ToLower(filepath.Ext(path)) != ".pvf" {
		return cacheBuild{}, fmt.Errorf("PVF path must point to a .pvf file")
	}
	md5Value, err := fileMD5(path)
	if err != nil {
		return cacheBuild{}, err
	}
	header, err := openPVFHeader(path)
	if err != nil {
		return cacheBuild{}, err
	}
	defer header.close()
	converter, err := opencc.New("t2s")
	if err != nil {
		return cacheBuild{}, err
	}
	pvf := &tinyPVF{
		header:      header,
		fileTree:    map[string]leaf{},
		fileContent: map[string][]byte{},
		encode:      encodeName,
		traditionalConv: func(value string) string {
			converted, err := converter.Convert(value)
			if err != nil {
				return value
			}
			return converted
		},
	}
	logs := []string{}
	log := func(format string, args ...any) {
		logs = append(logs, fmt.Sprintf(format, args...))
	}
	if err := pvf.loadLeafs(); err != nil {
		return cacheBuild{}, err
	}
	job, err := readJobDict(pvf)
	if err != nil {
		return cacheBuild{}, fmt.Errorf("job failed: %w", err)
	}
	expTable, err := readExpTable(pvf)
	if err != nil {
		return cacheBuild{}, fmt.Errorf("exp table failed: %w", err)
	}
	equipment, equipmentDetail, err := readNamedItems(pvf, "equipment/equipment.lst", "equipment", log)
	if err != nil {
		return cacheBuild{}, fmt.Errorf("equipment failed: %w", err)
	}
	stackable, stackableDetail, err := readNamedItems(pvf, "stackable/stackable.lst", "stackable", log)
	if err != nil {
		return cacheBuild{}, fmt.Errorf("stackable failed: %w", err)
	}
	avatarHidden, err := readHiddenAvatar(pvf)
	if err != nil {
		log("avatar hidden skipped: %s", err.Error())
		avatarHidden = []any{[]string{}, []string{}}
	}
	items := make([]cacheItem, 0, len(equipment)+len(stackable))
	for id, name := range stackable {
		items = append(items, cacheItem{
			ID:     id,
			Type:   "stackable",
			Name:   name,
			Detail: stackableDetail[id],
		})
	}
	for id, name := range equipment {
		items = append(items, cacheItem{
			ID:     id,
			Type:   "equipment",
			Name:   name,
			Detail: equipmentDetail[id],
		})
	}
	return cacheBuild{
		MD5:            md5Value,
		Path:           path,
		FileSize:       info.Size(),
		Encode:         encodeName,
		StackableCount: len(stackable),
		EquipmentCount: len(equipment),
		ExpLevelCount:  len(expTable),
		Items:          items,
		Data: map[string]any{
			"job":           job,
			"avatar_hidden": avatarHidden,
			"exp_table":     expTable,
		},
		Logs: logs,
	}, nil
}

func normalizePVFEncoding(value string) (string, error) {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" || value == "none" {
		value = "big5"
	}
	if value != "big5" && value != "gbk" {
		return "", fmt.Errorf("PVF encoding must be one of: big5, gbk")
	}
	return value, nil
}

func fileMD5(path string) (string, error) {
	fp, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer fp.Close()
	var h hash.Hash = md5.New()
	if _, err := io.Copy(h, fp); err != nil {
		return "", err
	}
	return strings.ToUpper(hex.EncodeToString(h.Sum(nil))), nil
}

func openPVFHeader(path string) (*pvfHeader, error) {
	fp, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	readU32 := func() (uint32, error) {
		var buf [4]byte
		if _, err := io.ReadFull(fp, buf[:]); err != nil {
			return 0, err
		}
		return binary.LittleEndian.Uint32(buf[:]), nil
	}
	uuidLen, err := readU32()
	if err != nil {
		fp.Close()
		return nil, err
	}
	if _, err := fp.Seek(int64(uuidLen), io.SeekCurrent); err != nil {
		fp.Close()
		return nil, err
	}
	if _, err := readU32(); err != nil {
		fp.Close()
		return nil, err
	}
	dirTreeLength, err := readU32()
	if err != nil {
		fp.Close()
		return nil, err
	}
	dirTreeCRC32, err := readU32()
	if err != nil {
		fp.Close()
		return nil, err
	}
	numFiles, err := readU32()
	if err != nil {
		fp.Close()
		return nil, err
	}
	headerEnd, err := fp.Seek(0, io.SeekCurrent)
	if err != nil {
		fp.Close()
		return nil, err
	}
	encryptedTree := make([]byte, dirTreeLength)
	if _, err := io.ReadFull(fp, encryptedTree); err != nil {
		fp.Close()
		return nil, err
	}
	return &pvfHeader{
		path:               path,
		file:               fp,
		dirTree:            decryptBytes(encryptedTree, dirTreeCRC32),
		filePackIndexShift: headerEnd + int64(dirTreeLength),
		numFiles:           numFiles,
	}, nil
}

func (h *pvfHeader) close() {
	if h.file != nil {
		_ = h.file.Close()
	}
}

func decryptBytes(input []byte, crc uint32) []byte {
	key := crc ^ 0x81A79011
	blocks := len(input) / 4
	out := make([]byte, blocks*4)
	for i := 0; i < blocks; i++ {
		offset := i * 4
		value := binary.LittleEndian.Uint32(input[offset:offset+4]) ^ key
		value = ((value & 0x0000003F) << 26) | ((value & 0xFFFFFFC0) >> 6)
		binary.LittleEndian.PutUint32(out[offset:offset+4], value)
	}
	return out
}

func (p *tinyPVF) loadLeafs() error {
	index := 0
	for i := uint32(0); i < p.header.numFiles; i++ {
		if index+24 > len(p.header.dirTree) {
			return fmt.Errorf("PVF directory tree is truncated")
		}
		index += 4
		pathLen := int(binary.LittleEndian.Uint32(p.header.dirTree[index : index+4]))
		index += 4
		if index+pathLen+12 > len(p.header.dirTree) {
			return fmt.Errorf("PVF directory entry is truncated")
		}
		path := strings.ToLower(strings.ReplaceAll(string(p.header.dirTree[index:index+pathLen]), "\\", "/"))
		index += pathLen
		if strings.HasPrefix(path, "/") {
			path = strings.TrimPrefix(path, "/")
		}
		fileLength := (binary.LittleEndian.Uint32(p.header.dirTree[index:index+4]) + 3) & 0xFFFFFFFC
		index += 4
		fileCRC32 := binary.LittleEndian.Uint32(p.header.dirTree[index : index+4])
		index += 4
		relativeOffset := binary.LittleEndian.Uint32(p.header.dirTree[index : index+4])
		index += 4
		p.fileTree[path] = leaf{
			path:           path,
			fileLength:     fileLength,
			fileCRC32:      fileCRC32,
			relativeOffset: relativeOffset,
		}
	}
	stringTableBytes, err := p.readDecrypted("stringtable.bin")
	if err != nil {
		return err
	}
	p.strings = newStringTable(stringTableBytes, p.encode, p.traditionalConv)
	nStringBytes, err := p.readDecrypted("n_string.lst")
	if err != nil {
		return err
	}
	p.nStrings, err = p.loadLstFromBytes(nStringBytes, "n_string.lst")
	return err
}

func (p *tinyPVF) readDecrypted(path string) ([]byte, error) {
	path = normalizePath(path)
	if cached, ok := p.fileContent[path]; ok {
		return cached, nil
	}
	entry, ok := p.fileTree[path]
	if !ok {
		return nil, fmt.Errorf("PVF file not found: %s", path)
	}
	buf := make([]byte, entry.fileLength)
	if _, err := p.header.file.ReadAt(buf, p.header.filePackIndexShift+int64(entry.relativeOffset)); err != nil {
		return nil, err
	}
	out := decryptBytes(buf, entry.fileCRC32)
	p.fileContent[path] = out
	return out, nil
}

func (p *tinyPVF) loadLst(path string) (*lstFile, error) {
	content, err := p.readDecrypted(path)
	if err != nil {
		return nil, err
	}
	return p.loadLstFromBytes(content, path)
}

func (p *tinyPVF) loadLstFromBytes(content []byte, path string) (*lstFile, error) {
	baseDir := ""
	path = normalizePath(path)
	if slash := strings.LastIndex(path, "/"); slash >= 0 {
		baseDir = path[:slash]
	}
	lst := &lstFile{
		tableList: [][2]any{},
		tableDict: map[int]string{},
		strDict:   map[int]strTable{},
		pvf:       p,
		baseDir:   baseDir,
		encode:    p.encode,
	}
	for i := 2; i+10 <= len(content); i += 10 {
		firstType := content[i]
		firstValue := int(int32(binary.LittleEndian.Uint32(content[i+1 : i+5])))
		secondType := content[i+5]
		secondValue := int(int32(binary.LittleEndian.Uint32(content[i+6 : i+10])))
		index := 0
		stringIndex := 0
		if firstType == 2 {
			index = firstValue
		} else if firstType == 7 {
			stringIndex = firstValue
		}
		if secondType == 2 {
			index = secondValue
		} else if secondType == 7 {
			stringIndex = secondValue
		}
		text := p.strings.get(stringIndex)
		lst.tableList = append(lst.tableList, [2]any{index, text})
		lst.tableDict[index] = text
	}
	return lst, nil
}

func (l *lstFile) getStr(index int) (strTable, error) {
	if cached, ok := l.strDict[index]; ok {
		return cached, nil
	}
	path, ok := l.tableDict[index]
	if !ok {
		return strTable{rows: map[string]string{}}, nil
	}
	content, err := l.pvf.readDecrypted(path)
	if err != nil {
		return strTable{}, err
	}
	text := decodeText(content, l.encode)
	text = l.pvf.traditionalConv(text)
	rows := map[string]string{}
	for _, line := range strings.Split(text, "\n") {
		if !strings.Contains(line, ">") {
			continue
		}
		parts := strings.SplitN(line, ">", 2)
		rows[parts[0]] = strings.ReplaceAll(parts[1], "\r", "")
	}
	table := strTable{rows: rows}
	l.strDict[index] = table
	return table, nil
}

func newStringTable(content []byte, encode string, conv func(string) string) *stringTable {
	if len(content) < 4 {
		return &stringTable{raw: nil, encode: encode, cache: map[int]string{}, conv: conv}
	}
	return &stringTable{
		raw:    content[4:],
		length: binary.LittleEndian.Uint32(content[:4]),
		encode: encode,
		cache:  map[int]string{},
		conv:   conv,
	}
}

func (s *stringTable) get(index int) string {
	if value, ok := s.cache[index]; ok {
		return value
	}
	offset := index * 4
	if offset < 0 || offset+8 > len(s.raw) {
		return ""
	}
	start := int(binary.LittleEndian.Uint32(s.raw[offset : offset+4]))
	end := int(binary.LittleEndian.Uint32(s.raw[offset+4 : offset+8]))
	if start < 0 || end < start || end > len(s.raw) {
		return ""
	}
	value := decodeText(s.raw[start:end], s.encode)
	if s.conv != nil {
		value = s.conv(value)
	}
	s.cache[index] = value
	return value
}

func (p *tinyPVF) readList(path string) ([]int, []any, error) {
	content, err := p.readDecrypted(path)
	if err != nil {
		return nil, nil, err
	}
	return p.contentToList(content)
}

func (p *tinyPVF) readDict(path string) (map[string]any, error) {
	types, values, err := p.readList(path)
	if err != nil {
		return nil, err
	}
	return listToDict(types, values), nil
}

func (p *tinyPVF) contentToList(content []byte) ([]int, []any, error) {
	if len(content) < 2 {
		return []int{}, []any{}, nil
	}
	unitNum := (len(content) - 2) / 5
	rawTypes := make([]int, unitNum)
	rawValues := make([]any, unitNum)
	for i := 0; i < unitNum; i++ {
		offset := 2 + i*5
		unitType := int(content[offset])
		rawTypes[i] = unitType
		if unitType == 4 {
			rawValues[i] = math.Float32frombits(binary.LittleEndian.Uint32(content[offset+1 : offset+5]))
		} else {
			rawValues[i] = int(int32(binary.LittleEndian.Uint32(content[offset+1 : offset+5])))
		}
	}
	types := []int{}
	values := []any{}
	for i, unitType := range rawTypes {
		switch unitType {
		case 2, 3:
			values = append(values, rawValues[i])
			types = append(types, unitType)
		case 4:
			values = append(values, rawValues[i])
			types = append(types, unitType)
		case 5, 6, 7, 8:
			index, _ := rawValues[i].(int)
			values = append(values, p.strings.get(index))
			types = append(types, unitType)
		case 9:
			index, _ := rawValues[i].(int)
			key := ""
			if i+1 < len(rawValues) {
				nextIndex, _ := rawValues[i+1].(int)
				key = p.strings.get(nextIndex)
			}
			table, err := p.nStrings.getStr(index)
			if err != nil {
				return nil, nil, err
			}
			values = append(values, table.rows[key])
			types = append(types, unitType)
		}
	}
	return types, values, nil
}

func listToDict(typeList []int, values []any) map[string]any {
	endMarks := map[string]bool{}
	for _, value := range values {
		text, ok := value.(string)
		if ok && strings.HasPrefix(text, "[/") && strings.HasSuffix(text, "]") {
			endMarks[strings.Replace(text, "/", "", 1)] = true
		}
	}
	result := map[string]any{}
	segment := []any{}
	segTypes := []int{}
	segmentKey := ""
	hasSegmentKey := false
	addSegment := func() {
		key := segmentKey
		if _, exists := result[key]; exists {
			suffix := 1
			for {
				candidate := fmt.Sprintf("%s-%d", key, suffix)
				if _, ok := result[candidate]; !ok {
					key = candidate
					break
				}
				suffix++
			}
		}
		if endMarks[segmentKey] && containsType(segTypes, 5) {
			result[key] = listToDict(segTypes, segment)
		} else {
			result[key] = segment
		}
	}
	for i, value := range values {
		if i >= len(typeList) {
			break
		}
		if typeList[i] == 5 {
			text, ok := value.(string)
			if ok {
				if !hasSegmentKey {
					if !strings.Contains(text, "/") {
						segmentKey = text
						hasSegmentKey = true
					}
					continue
				}
				if !endMarks[segmentKey] || strings.ReplaceAll(text, "/", "") == segmentKey {
					addSegment()
					hasSegmentKey = false
					segmentKey = ""
					if !strings.Contains(text, "/") {
						segmentKey = text
						hasSegmentKey = true
					}
					segment = []any{}
					segTypes = []int{}
					continue
				}
			}
		}
		segment = append(segment, value)
		segTypes = append(segTypes, typeList[i])
	}
	if hasSegmentKey {
		addSegment()
	}
	return result
}

func readJobDict(p *tinyPVF) (map[int]map[int]string, error) {
	characters, err := p.loadLst("character/character.lst")
	if err != nil {
		return nil, err
	}
	result := map[int]map[int]string{}
	for _, pair := range characters.tableList {
		id, _ := pair[0].(int)
		path, _ := pair[1].(string)
		detail, err := p.readDict(joinPVFPath(characters.baseDir, path))
		if err != nil {
			continue
		}
		growNames := listFromDetail(detail["[growtype name]"])
		growTypes := map[int]string{}
		for i, value := range growNames {
			growTypes[i] = fmt.Sprint(value)
		}
		result[id] = growTypes
	}
	return result, nil
}

func readExpTable(p *tinyPVF) ([]int, error) {
	_, values, err := p.readList("character/exptable.tbl")
	if err != nil {
		return nil, err
	}
	result := []int{}
	for _, value := range values {
		if number, ok := value.(int); ok {
			result = append(result, number)
		}
	}
	return result, nil
}

func readNamedItems(p *tinyPVF, listPath string, itemType string, log func(string, ...any)) (map[int]string, map[int]map[string]any, error) {
	lst, err := p.loadLst(listPath)
	if err != nil {
		return nil, nil, err
	}
	names := map[int]string{}
	details := map[int]map[string]any{}
	failures := 0
	for _, pair := range lst.tableList {
		id, _ := pair[0].(int)
		path, _ := pair[1].(string)
		fpath := joinPVFPath(lst.baseDir, path)
		detail, err := p.readDict(fpath)
		if err != nil {
			failures++
			names[id] = fpath
			details[id] = map[string]any{}
			continue
		}
		details[id] = detail
		name := joinValues(listFromDetail(detail["[name]"]))
		if name == "" && itemType == "equipment" {
			name = "[无名称]"
		}
		if name == "" {
			name = fpath
		}
		if len([]rune(name)) > 255 {
			name = string([]rune(name)[:255])
		}
		names[id] = name
	}
	if failures > 0 {
		log("%s load failures: %d", itemType, failures)
	}
	return names, details, nil
}

func readHiddenAvatar(p *tinyPVF) ([]any, error) {
	_, values, err := p.readList("etc/avatar_roulette/avatarfixedhiddenoptionlist.etc")
	if err != nil {
		return nil, err
	}
	upper := false
	rare := false
	upperRows := []string{}
	rareRows := []string{}
	for _, value := range values {
		text := fmt.Sprint(value)
		switch text {
		case "[upper]":
			upper = true
			continue
		case "[/upper]":
			upper = false
			continue
		case "[rare]":
			rare = true
			continue
		case "[/rare]":
			rare = false
			continue
		}
		if strings.Contains(text, "[") && len(text) >= 2 {
			trimmed := strings.TrimSuffix(strings.TrimPrefix(text, "["), "]")
			if upper {
				upperRows = append(upperRows, trimmed)
			}
			if rare {
				rareRows = append(rareRows, trimmed)
			}
		}
	}
	return []any{upperRows, rareRows}, nil
}

func decodeText(content []byte, encodeName string) string {
	var (
		out []byte
		err error
	)
	switch strings.ToLower(encodeName) {
	case "gbk":
		out, err = simplifiedchinese.GBK.NewDecoder().Bytes(content)
	default:
		out, err = traditionalchinese.Big5.NewDecoder().Bytes(content)
	}
	if err != nil {
		return string(content)
	}
	return string(out)
}

func normalizePath(path string) string {
	path = strings.ToLower(strings.ReplaceAll(path, "\\", "/"))
	return strings.TrimPrefix(path, "/")
}

func joinPVFPath(baseDir string, path string) string {
	if baseDir == "" {
		return normalizePath(path)
	}
	return normalizePath(baseDir + "/" + path)
}

func listFromDetail(value any) []any {
	switch rows := value.(type) {
	case []any:
		return rows
	case []string:
		result := make([]any, 0, len(rows))
		for _, item := range rows {
			result = append(result, item)
		}
		return result
	default:
		return []any{}
	}
}

func joinValues(values []any) string {
	parts := make([]string, 0, len(values))
	for _, value := range values {
		parts = append(parts, fmt.Sprint(value))
	}
	return strings.Join(parts, "")
}

func detailJSON(detail map[string]any) (string, error) {
	if detail == nil {
		detail = map[string]any{}
	}
	payload, err := json.Marshal(detail)
	if err != nil {
		return "", err
	}
	return string(payload), nil
}

func containsType(values []int, target int) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
