package devtools

import (
	"log"
	"strings"
)

// SourceMapBuilder is used for compiling source maps from existing source map files
type SourceMapBuilder struct {
	filename string
	sources  []*sourceMap
}

// NewSourceMapBuilder creates a new sourceMapBuilder
func NewSourceMapBuilder(filename string, capacity int) *SourceMapBuilder {
	return &SourceMapBuilder{
		filename: filename,
		sources:  make([]*sourceMap, 0, capacity),
	}
}

// AddSourceMap adds a source map to be included in the build
func (smb *SourceMapBuilder) AddSourceMap(spacerLines int, fileLineCount int, path string, sourceMapContents string) {
	source := &sourceMap{
		spacerLines,
		fileLineCount,
		path,
		sourceMapContents,
	}
	smb.sources = append(smb.sources, source)
}

func (smb *SourceMapBuilder) String() string {
	var sb strings.Builder
	sb.WriteString(`{"version":3,"file":"`)
	sb.WriteString(smb.filename)
	sb.WriteString(`.js","sources":[`)
	first := true
	for _, source := range smb.sources {
		if first {
			first = false
		} else {
			sb.WriteByte(',')
		}
		sb.WriteString("\"" + source.path + "\"")
	}
	sb.WriteString(`],"mappings":"`)
	sb.WriteString(smb.GenerateMappings())
	sb.WriteString(`"}`)
	return sb.String()
}

// GenerateMappings outputs a string of the compiled sourcemap
func (smb *SourceMapBuilder) GenerateMappings() string {
	var sb strings.Builder
	var lastMappingsDelta = Segment{0, 0, 0, 0}
	var sourceMapLineCount int
	for _, source := range smb.sources {
		sb.WriteString(strings.Repeat(";", source.spacerLines))

		smap, err := ParseSourceMapJSON(source.contents)
		if err != nil {
			log.Printf("ERROR parsing source map: " + source.path)
			continue
		}

		mappings := smap.OffsetMappings(lastMappingsDelta)
		sourceMapLineCount, lastMappingsDelta = smap.PlayMappings()
		lastMappingsDelta.sourceFile = 1

		sb.WriteString(mappings)
		additionalSeparators := 1 + (source.fileLineCount - sourceMapLineCount)
		sb.WriteString(strings.Repeat(";", additionalSeparators))
	}
	return sb.String()
}
