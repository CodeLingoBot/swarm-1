package devtools

import (
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPlayMappings(t *testing.T) {
	cases := map[string]struct {
		mappings string
		expected Segment
	}{
		// "zero":   {mappings: "AAAA", expected: segment{0, 0, 0, 0}},
		// "1234":   {mappings: "ACEG", expected: segment{0, 1, 2, 3}},
		// "1234;;": {mappings: "ACEG", expected: segment{0, 1, 2, 3}},
		// ";":      {mappings: "AAAA", expected: segment{0, 0, 0, 0}},
		// "MED":    {mappings: "AAAA;BBBB;CCCC,ACCC,ABBB,XYZA;ADDD", expected: segment{0, 0, 0, 0}},
		"LONG": {mappings: ";;;;;;;;;YAGA;gBAAA;gBAIA,CAAC;gBAHiB,QAAE,GAAhB;oBACI,OAAO,qCAAqC,CAAA;gBAChD,CAAC;gBACL,YAAC;YAAD,CAAC,AAJD,IAIC;;QAAC,CAAC", expected: Segment{9, 0, 7, 3}},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			smap := &SourceMap{Mappings: tc.mappings}
			actual := smap.PlayMappings()
			assert.Equal(t, tc.expected, actual)
		})
	}
}

const firstJSON = `{
    "version": 3,
    "file": "First.js",
    "sourceRoot": "",
    "sources": [
        "First.ts"
    ],
    "names": [],
    "mappings": ";;;;;;;;;YAGA;gBAAA;gBAIA,CAAC;gBAHiB,QAAE,GAAhB;oBACI,OAAO,qCAAqC,CAAA;gBAChD,CAAC;gBACL,YAAC;YAAD,CAAC,AAJD,IAIC;;QAAC,CAAC"
}`

const firstMappingsNoChange = ";;;;;;;;;YAGA;gBAAA;gBAIA,CAAC;gBAHiB,QAAE,GAAhB;oBACI,OAAO,qCAAqC,CAAA;gBAChD,CAAC;gBACL,YAAC;YAAD,CAAC,AAJD,IAIC;;QAAC,CAAC"
const firstMappingsPlusOne1 = ";;;;;;;;;YCGA;gBAAA;gBAIA,CAAC;gBAHiB,QAAE,GAAhB;oBACI,OAAO,qCAAqC,CAAA;gBAChD,CAAC;gBACL,YAAC;YAAD,CAAC,AAJD,IAIC;;QAAC,CAAC"
const firstMappingsPlus1506 = ";;;;;;;;;Yk+CGA;gBAAA;gBAIA,CAAC;gBAHiB,QAAE,GAAhB;oBACI,OAAO,qCAAqC,CAAA;gBAChD,CAAC;gBACL,YAAC;YAAD,CAAC,AAJD,IAIC;;QAAC,CAAC"
const firstMappingsMinus369 = ";;;;;;;;;YjXGA;gBAAA;gBAIA,CAAC;gBAHiB,QAAE,GAAhB;oBACI,OAAO,qCAAqC,CAAA;gBAChD,CAAC;gBACL,YAAC;YAAD,CAAC,AAJD,IAIC;;QAAC,CAAC"

const thirdJSON = `{
    "version": 3,
    "file": "Third.js",
    "sourceRoot": "",
    "sources": [
        "Third.ts"
    ],
    "names": [],
    "mappings": ";;;;;;;;;YAGA;gBAAA;gBAIA,CAAC;gBAHiB,QAAE,GAAhB;oBACI,OAAO,qCAAqC,CAAA;gBAChD,CAAC;gBACL,YAAC;YAAD,CAAC,AAJD,IAIC;;QAAC,CAAC"
}`

const mapping1 = `;;;;YAGA;gBAAA;gBAIA,CAAC`
const mapping2 = `;;;;;;YACA;gBAAA;gBAKA,CAAC;`
const combinedMappings = `;;;;YAGA;gBAAA;gBAIA,CAAC;;;;;;YCCA;gBAAA;gBAKA,CAAC;`

func TestOffsetMappingsSourceFileIndex(t *testing.T) {
	cases := map[string]struct {
		json      string
		fileIndex int
		expected  string
	}{
		"no-change": {
			json:      firstJSON,
			fileIndex: 0,
			expected:  firstMappingsNoChange,
		},
		"increase-by-1": {
			json:      firstJSON,
			fileIndex: 1,
			expected:  firstMappingsPlusOne1,
		},
		"increase-by-1506": {
			json:      firstJSON,
			fileIndex: 1506,
			expected:  firstMappingsPlus1506,
		},
		"decrease-by-369": { // <-- this one is stupid, but meh \_/
			json:      firstJSON,
			fileIndex: -369,
			expected:  firstMappingsMinus369,
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			smap, err := ParseSourceMapJSON(tc.json)
			assert.Nil(t, err)
			actual := smap.OffsetMappingsSourceFileIndex(tc.fileIndex)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestFindFirstLVQ(t *testing.T) {
	cases := map[string]struct {
		mappings      string
		startPos      int
		expectedStart int
		expectedEnd   int
	}{
		"one":   {mappings: ";;;;;AAAA", startPos: 0, expectedStart: 5, expectedEnd: 9},
		"two":   {mappings: ";;AAAA;;;AZQA;bGAFA;", startPos: 2, expectedStart: 2, expectedEnd: 6},
		"none":  {mappings: ";;;;;;;;;;;;", startPos: 0, expectedStart: -1, expectedEnd: -1},
		"start": {mappings: "AAAA;;;;;", startPos: 9, expectedStart: 0, expectedEnd: 4},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			start, end := findFirstVLQ(tc.mappings)
			assert.Equal(t, tc.expectedStart, start)
			assert.Equal(t, tc.expectedEnd, end)
		})
	}
}

func TestNextNonSeparator(t *testing.T) {
	cases := map[string]struct {
		mappings string
		startPos int
		expected int
	}{
		"one":   {mappings: ";;;;;AAAA", startPos: 0, expected: 5},
		"two":   {mappings: ";;;;;AAAA", startPos: 2, expected: 5},
		"start": {mappings: "AAAC;AAAD;ZZZA", startPos: 0, expected: 0},
		"eof-1": {mappings: "AAAC;AAAD;;;;;", startPos: 9, expected: -1},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			actual := nextNonSeparator(tc.mappings, tc.startPos)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestNextSepartorOrEOF(t *testing.T) {
	cases := map[string]struct {
		mappings string
		startPos int
		expected int
	}{
		"start":  {mappings: ";;;;;AAAA", startPos: 0, expected: 0},
		"eof":    {mappings: ";;;;;AAAA", startPos: 5, expected: 9},
		"second": {mappings: "A;B", startPos: 0, expected: 1},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			actual := nextSeparatorOrEOF(tc.mappings, tc.startPos)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestParseSourceMap(t *testing.T) {
	value, err := ParseSourceMapJSON(firstJSON)
	assert.Nil(t, err)
	assert.Equal(t, 3, value.Version, "Version")
	assert.NotEmpty(t, value.File, "File")
	assert.Empty(t, value.SourceRoot, "SourceRoot")
	assert.Len(t, value.Sources, 1, "Sources")
	assert.Len(t, value.Names, 0, "Name")
	assert.NotEmpty(t, value.Mappings, "Mappings")

	parsed := parseMappings(value.Mappings)
	assert.Len(t, parsed, 19)
}

func TestReplaceFirstVLQ(t *testing.T) {
	cases := map[string]struct {
		mappings      string
		replacementFn vlqReplaceFn
		expected      string
	}{
		"one": {
			mappings: "YCCA",
			replacementFn: func(seg Segment) Segment {
				seg.generatedColumn++
				return seg
			},
			expected: "aCCA",
		},
		"upndown": {
			mappings: "AAAA",
			replacementFn: func(seg Segment) Segment {
				seg.generatedColumn++
				seg.sourceFile--
				seg.sourceLine++
				seg.sourceColumn--
				return seg
			},
			expected: "CDCD",
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			actual := replaceFirstVLQ(tc.mappings, tc.replacementFn)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

/*
YAGC = [12,0,3,1]
[
	12, // generated COLUMN (reset with each line, relative within same line)
	0,  // source FILE index (relative to last, except for first) <-- ONLY THING THAT NEEDS TO CHANGE
	4,  // source LINE index (relative to last, except for first)
	1,  // source COLUMN index (relative to last, except for first)
]
*/

func TestParseMapsString(t *testing.T) {
	mappings := ";;;;AAAA;KAAK;;;;"
	expected := []*line{
		nil,
		nil,
		nil,
		nil,
		&line{
			segments: []*Segment{
				&Segment{0, 0, 0, 0},
			},
		},
		&line{
			segments: []*Segment{
				&Segment{5, 0, 0, 5},
			},
		},
		nil,
		nil,
		nil,
		nil,
	}

	actual := parseMappings(mappings)
	assert.Equal(t, len(expected), len(actual), "The # of lines return from parseMaps(...) did not match")
	equal := assert.ObjectsAreEqual(expected, actual)
	if !equal {
		for _, line := range actual {
			log.Printf("%#v\n", line)
		}
		assert.True(t, equal, "The expected result of parseMaps(...) did not match the actual result.")
	}
}

func TestDecode(t *testing.T) {
	cases := map[string]struct {
		vlq      string
		expected []int
	}{
		"AAAC": {
			vlq:      "AAAC",
			expected: []int{0, 0, 0, 1},
		},
		"ADAA": {
			vlq:      "ADAA",
			expected: []int{0, -1, 0, 0},
		},
		"AAgBC": {
			vlq:      "AAgBC",
			expected: []int{0, 0, 16, 1},
		},
		"KAAK": {
			vlq:      "KAAK",
			expected: []int{5, 0, 0, 5},
		},
		"G9s6a8zns//+": {
			vlq:      "G9s6aAs8BzC",
			expected: []int{3, -439502, 0, 966, -41},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			actual := decode(tc.vlq)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestEncode(t *testing.T) {
	cases := map[string]struct {
		expected string
		nums     []int
	}{
		"AAAC": {
			nums:     []int{0, 0, 0, 1},
			expected: "AAAC",
		},
		"ADAA": {
			nums:     []int{0, -1, 0, 0},
			expected: "ADAA",
		},
		"AAgBC": {
			nums:     []int{0, 0, 16, 1},
			expected: "AAgBC",
		},
		"KAAK": {
			nums:     []int{5, 0, 0, 5},
			expected: "KAAK",
		},
		"G9s6a8zns//+": {
			nums:     []int{3, -439502, 0, 966, -41},
			expected: "G9s6aAs8BzC",
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			actual := encode(tc.nums)
			assert.Equal(t, tc.expected, actual)
		})
	}
}