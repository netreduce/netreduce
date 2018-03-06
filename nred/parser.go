/*
This file was generated with treerack (https://github.com/aryszka/treerack).

The contents of this file fall under different licenses.

The code between the "// head" and "// eo head" lines falls under the same
license as the source code of treerack (https://github.com/aryszka/treerack),
unless explicitly stated otherwise, if treerack's license allows changing the
license of this source code.

Treerack's license: MIT https://opensource.org/licenses/MIT
where YEAR=2017, COPYRIGHT HOLDER=Arpad Ryszka (arpad.ryszka@gmail.com)

The rest of the content of this file falls under the same license as the one
that the user of treerack generating this file declares for it, or it is
unlicensed.
*/

package nred

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"
)

type charParser struct {
	name   string
	id     int
	not    bool
	chars  []rune
	ranges [][]rune
}
type charBuilder struct {
	name string
	id   int
}

func (p *charParser) nodeName() string {
	return p.name
}
func (p *charParser) nodeID() int {
	return p.id
}
func (p *charParser) commitType() CommitType {
	return Alias
}
func matchChar(chars []rune, ranges [][]rune, not bool, char rune) bool {
	for _, ci := range chars {
		if ci == char {
			return !not
		}
	}
	for _, ri := range ranges {
		if char >= ri[0] && char <= ri[1] {
			return !not
		}
	}
	return not
}
func (p *charParser) match(t rune) bool {
	return matchChar(p.chars, p.ranges, p.not, t)
}
func (p *charParser) parse(c *context) {
	if tok, ok := c.token(); !ok || !p.match(tok) {
		if c.offset > c.failOffset {
			c.failOffset = c.offset
			c.failingParser = nil
		}
		c.fail(c.offset)
		return
	}
	c.success(c.offset + 1)
}
func (b *charBuilder) nodeName() string {
	return b.name
}
func (b *charBuilder) nodeID() int {
	return b.id
}
func (b *charBuilder) build(c *context) ([]*Node, bool) {
	return nil, false
}

type sequenceParser struct {
	name            string
	id              int
	commit          CommitType
	items           []parser
	ranges          [][]int
	generalizations []int
	allChars        bool
}
type sequenceBuilder struct {
	name     string
	id       int
	commit   CommitType
	items    []builder
	ranges   [][]int
	allChars bool
}

func (p *sequenceParser) nodeName() string {
	return p.name
}
func (p *sequenceParser) nodeID() int {
	return p.id
}
func (p *sequenceParser) commitType() CommitType {
	return p.commit
}
func (p *sequenceParser) parse(c *context) {
	if !p.allChars {
		if c.results.pending(c.offset, p.id) {
			c.fail(c.offset)
			return
		}
		c.results.markPending(c.offset, p.id)
	}
	var (
		currentCount int
		parsed       bool
	)
	itemIndex := 0
	from := c.offset
	to := c.offset
	for itemIndex < len(p.items) {
		p.items[itemIndex].parse(c)
		if !c.matchLast {
			if currentCount >= p.ranges[itemIndex][0] {
				itemIndex++
				currentCount = 0
				continue
			}
			if c.failingParser == nil && p.commit&userDefined != 0 && p.commit&Whitespace == 0 && p.commit&FailPass == 0 {
				c.failingParser = p
			}
			c.fail(from)
			if !p.allChars {
				c.results.unmarkPending(from, p.id)
			}
			return
		}
		parsed = c.offset > to
		if parsed {
			currentCount++
		}
		to = c.offset
		if !parsed || p.ranges[itemIndex][1] > 0 && currentCount == p.ranges[itemIndex][1] {
			itemIndex++
			currentCount = 0
		}
	}
	for _, g := range p.generalizations {
		if c.results.pending(from, g) {
			c.results.setMatch(from, g, to)
		}
	}
	if to > c.failOffset {
		c.failOffset = -1
		c.failingParser = nil
	}
	c.results.setMatch(from, p.id, to)
	c.success(to)
	if !p.allChars {
		c.results.unmarkPending(from, p.id)
	}
}
func (b *sequenceBuilder) nodeName() string {
	return b.name
}
func (b *sequenceBuilder) nodeID() int {
	return b.id
}
func (b *sequenceBuilder) build(c *context) ([]*Node, bool) {
	to, ok := c.results.longestMatch(c.offset, b.id)
	if !ok {
		return nil, false
	}
	from := c.offset
	parsed := to > from
	if b.allChars {
		c.offset = to
		if b.commit&Alias != 0 {
			return nil, true
		}
		return []*Node{{Name: b.name, From: from, To: to, tokens: c.tokens}}, true
	} else if parsed {
		c.results.dropMatchTo(c.offset, b.id, to)
	} else {
		if c.results.pending(c.offset, b.id) {
			return nil, false
		}
		c.results.markPending(c.offset, b.id)
	}
	var (
		itemIndex    int
		currentCount int
		nodes        []*Node
	)
	for itemIndex < len(b.items) {
		itemFrom := c.offset
		n, ok := b.items[itemIndex].build(c)
		if !ok {
			itemIndex++
			currentCount = 0
			continue
		}
		if c.offset > itemFrom {
			nodes = append(nodes, n...)
			currentCount++
			if b.ranges[itemIndex][1] > 0 && currentCount == b.ranges[itemIndex][1] {
				itemIndex++
				currentCount = 0
			}
			continue
		}
		if currentCount < b.ranges[itemIndex][0] {
			for i := 0; i < b.ranges[itemIndex][0]-currentCount; i++ {
				nodes = append(nodes, n...)
			}
		}
		itemIndex++
		currentCount = 0
	}
	if !parsed {
		c.results.unmarkPending(from, b.id)
	}
	if b.commit&Alias != 0 {
		return nodes, true
	}
	return []*Node{{Name: b.name, From: from, To: to, Nodes: nodes, tokens: c.tokens}}, true
}

type choiceParser struct {
	name            string
	id              int
	commit          CommitType
	options         []parser
	generalizations []int
}
type choiceBuilder struct {
	name    string
	id      int
	commit  CommitType
	options []builder
}

func (p *choiceParser) nodeName() string {
	return p.name
}
func (p *choiceParser) nodeID() int {
	return p.id
}
func (p *choiceParser) commitType() CommitType {
	return p.commit
}
func (p *choiceParser) parse(c *context) {
	if c.fromResults(p) {
		return
	}
	if c.results.pending(c.offset, p.id) {
		c.fail(c.offset)
		return
	}
	c.results.markPending(c.offset, p.id)
	var (
		match         bool
		optionIndex   int
		foundMatch    bool
		failingParser parser
	)
	from := c.offset
	to := c.offset
	initialFailOffset := c.failOffset
	initialFailingParser := c.failingParser
	failOffset := initialFailOffset
	for {
		foundMatch = false
		optionIndex = 0
		for optionIndex < len(p.options) {
			p.options[optionIndex].parse(c)
			optionIndex++
			if !c.matchLast {
				if c.failOffset > failOffset {
					failOffset = c.failOffset
					failingParser = c.failingParser
				}
			}
			if !c.matchLast || match && c.offset <= to {
				c.offset = from
				continue
			}
			match = true
			foundMatch = true
			to = c.offset
			c.offset = from
			c.results.setMatch(from, p.id, to)
		}
		if !foundMatch {
			break
		}
	}
	if match {
		if failOffset > to {
			c.failOffset = failOffset
			c.failingParser = failingParser
		} else if to > initialFailOffset {
			c.failOffset = -1
			c.failingParser = nil
		} else {
			c.failOffset = initialFailOffset
			c.failingParser = initialFailingParser
		}
		c.success(to)
		c.results.unmarkPending(from, p.id)
		return
	}
	if failOffset > initialFailOffset {
		c.failOffset = failOffset
		c.failingParser = failingParser
		if c.failingParser == nil && p.commitType()&userDefined != 0 && p.commitType()&Whitespace == 0 && p.commitType()&FailPass == 0 {
			c.failingParser = p
		}
	}
	c.results.setNoMatch(from, p.id)
	c.fail(from)
	c.results.unmarkPending(from, p.id)
}
func (b *choiceBuilder) nodeName() string {
	return b.name
}
func (b *choiceBuilder) nodeID() int {
	return b.id
}
func (b *choiceBuilder) build(c *context) ([]*Node, bool) {
	to, ok := c.results.longestMatch(c.offset, b.id)
	if !ok {
		return nil, false
	}
	from := c.offset
	parsed := to > from
	if parsed {
		c.results.dropMatchTo(c.offset, b.id, to)
	} else {
		if c.results.pending(c.offset, b.id) {
			return nil, false
		}
		c.results.markPending(c.offset, b.id)
	}
	var option builder
	for _, o := range b.options {
		if c.results.hasMatchTo(c.offset, o.nodeID(), to) {
			option = o
			break
		}
	}
	n, _ := option.build(c)
	if !parsed {
		c.results.unmarkPending(from, b.id)
	}
	if b.commit&Alias != 0 {
		return n, true
	}
	return []*Node{{Name: b.name, From: from, To: to, Nodes: n, tokens: c.tokens}}, true
}

type idSet struct{ ids []uint }

func divModBits(id int) (int, int) {
	return id / strconv.IntSize, id % strconv.IntSize
}
func (s *idSet) set(id int) {
	d, m := divModBits(id)
	if d >= len(s.ids) {
		if d < cap(s.ids) {
			s.ids = s.ids[:d+1]
		} else {
			s.ids = s.ids[:cap(s.ids)]
			for i := cap(s.ids); i <= d; i++ {
				s.ids = append(s.ids, 0)
			}
		}
	}
	s.ids[d] |= 1 << uint(m)
}
func (s *idSet) unset(id int) {
	d, m := divModBits(id)
	if d >= len(s.ids) {
		return
	}
	s.ids[d] &^= 1 << uint(m)
}
func (s *idSet) has(id int) bool {
	d, m := divModBits(id)
	if d >= len(s.ids) {
		return false
	}
	return s.ids[d]&(1<<uint(m)) != 0
}

type results struct {
	noMatch   []*idSet
	match     [][]int
	isPending [][]int
}

func ensureOffsetInts(ints [][]int, offset int) [][]int {
	if len(ints) > offset {
		return ints
	}
	if cap(ints) > offset {
		ints = ints[:offset+1]
		return ints
	}
	ints = ints[:cap(ints)]
	for i := len(ints); i <= offset; i++ {
		ints = append(ints, nil)
	}
	return ints
}
func ensureOffsetIDs(ids []*idSet, offset int) []*idSet {
	if len(ids) > offset {
		return ids
	}
	if cap(ids) > offset {
		ids = ids[:offset+1]
		return ids
	}
	ids = ids[:cap(ids)]
	for i := len(ids); i <= offset; i++ {
		ids = append(ids, nil)
	}
	return ids
}
func (r *results) setMatch(offset, id, to int) {
	r.match = ensureOffsetInts(r.match, offset)
	for i := 0; i < len(r.match[offset]); i += 2 {
		if r.match[offset][i] != id || r.match[offset][i+1] != to {
			continue
		}
		return
	}
	r.match[offset] = append(r.match[offset], id, to)
}
func (r *results) setNoMatch(offset, id int) {
	if len(r.match) > offset {
		for i := 0; i < len(r.match[offset]); i += 2 {
			if r.match[offset][i] != id {
				continue
			}
			return
		}
	}
	r.noMatch = ensureOffsetIDs(r.noMatch, offset)
	if r.noMatch[offset] == nil {
		r.noMatch[offset] = &idSet{}
	}
	r.noMatch[offset].set(id)
}
func (r *results) hasMatchTo(offset, id, to int) bool {
	if len(r.match) <= offset {
		return false
	}
	for i := 0; i < len(r.match[offset]); i += 2 {
		if r.match[offset][i] != id {
			continue
		}
		if r.match[offset][i+1] == to {
			return true
		}
	}
	return false
}
func (r *results) longestMatch(offset, id int) (int, bool) {
	if len(r.match) <= offset {
		return 0, false
	}
	var found bool
	to := -1
	for i := 0; i < len(r.match[offset]); i += 2 {
		if r.match[offset][i] != id {
			continue
		}
		if r.match[offset][i+1] > to {
			to = r.match[offset][i+1]
		}
		found = true
	}
	return to, found
}
func (r *results) longestResult(offset, id int) (int, bool, bool) {
	if len(r.noMatch) > offset && r.noMatch[offset] != nil && r.noMatch[offset].has(id) {
		return 0, false, true
	}
	to, ok := r.longestMatch(offset, id)
	return to, ok, ok
}
func (r *results) dropMatchTo(offset, id, to int) {
	for i := 0; i < len(r.match[offset]); i += 2 {
		if r.match[offset][i] != id {
			continue
		}
		if r.match[offset][i+1] == to {
			r.match[offset][i] = -1
			return
		}
	}
}
func (r *results) resetPending() {
	r.isPending = nil
}
func (r *results) pending(offset, id int) bool {
	if len(r.isPending) <= id {
		return false
	}
	for i := range r.isPending[id] {
		if r.isPending[id][i] == offset {
			return true
		}
	}
	return false
}
func (r *results) markPending(offset, id int) {
	r.isPending = ensureOffsetInts(r.isPending, id)
	for i := range r.isPending[id] {
		if r.isPending[id][i] == -1 {
			r.isPending[id][i] = offset
			return
		}
	}
	r.isPending[id] = append(r.isPending[id], offset)
}
func (r *results) unmarkPending(offset, id int) {
	for i := range r.isPending[id] {
		if r.isPending[id][i] == offset {
			r.isPending[id][i] = -1
			break
		}
	}
}

type context struct {
	reader        io.RuneReader
	offset        int
	readOffset    int
	consumed      int
	failOffset    int
	failingParser parser
	readErr       error
	eof           bool
	results       *results
	tokens        []rune
	matchLast     bool
}

func newContext(r io.RuneReader) *context {
	return &context{reader: r, results: &results{}, failOffset: -1}
}
func (c *context) read() bool {
	if c.eof || c.readErr != nil {
		return false
	}
	token, n, err := c.reader.ReadRune()
	if err != nil {
		if err == io.EOF {
			if n == 0 {
				c.eof = true
				return false
			}
		} else {
			c.readErr = err
			return false
		}
	}
	c.readOffset++
	if token == unicode.ReplacementChar {
		c.readErr = ErrInvalidUnicodeCharacter
		return false
	}
	c.tokens = append(c.tokens, token)
	return true
}
func (c *context) token() (rune, bool) {
	if c.offset == c.readOffset {
		if !c.read() {
			return 0, false
		}
	}
	return c.tokens[c.offset], true
}
func (c *context) fromResults(p parser) bool {
	to, m, ok := c.results.longestResult(c.offset, p.nodeID())
	if !ok {
		return false
	}
	if m {
		c.success(to)
	} else {
		c.fail(c.offset)
	}
	return true
}
func (c *context) success(to int) {
	c.offset = to
	c.matchLast = true
	if to > c.consumed {
		c.consumed = to
	}
}
func (c *context) fail(offset int) {
	c.offset = offset
	c.matchLast = false
}
func findLine(tokens []rune, offset int) (line, column int) {
	tokens = tokens[:offset]
	for i := range tokens {
		column++
		if tokens[i] == '\n' {
			column = 0
			line++
		}
	}
	return
}
func (c *context) parseError(p parser) error {
	definition := p.nodeName()
	flagIndex := strings.Index(definition, ":")
	if flagIndex > 0 {
		definition = definition[:flagIndex]
	}
	if c.failingParser == nil {
		c.failOffset = c.consumed
	}
	line, col := findLine(c.tokens, c.failOffset)
	return &ParseError{Offset: c.failOffset, Line: line, Column: col, Definition: definition}
}
func (c *context) finalizeParse(root parser) error {
	fp := c.failingParser
	if fp == nil {
		fp = root
	}
	to, match, found := c.results.longestResult(0, root.nodeID())
	if !found || !match || found && match && to < c.readOffset {
		return c.parseError(fp)
	}
	c.read()
	if c.eof {
		return nil
	}
	if c.readErr != nil {
		return c.readErr
	}
	return c.parseError(root)
}

type Node struct {
	Name     string
	Nodes    []*Node
	From, To int
	tokens   []rune
}

func (n *Node) Tokens() []rune {
	return n.tokens
}
func (n *Node) String() string {
	return fmt.Sprintf("%s:%d:%d:%s", n.Name, n.From, n.To, n.Text())
}
func (n *Node) Text() string {
	return string(n.Tokens()[n.From:n.To])
}

type CommitType int

const (
	None  CommitType = 0
	Alias CommitType = 1 << iota
	Whitespace
	NoWhitespace
	FailPass
	Root
	userDefined
)

type formatFlags int

const (
	formatNone   formatFlags = 0
	formatPretty formatFlags = 1 << iota
	formatIncludeComments
)

type ParseError struct {
	Input      string
	Offset     int
	Line       int
	Column     int
	Definition string
}
type parser interface {
	nodeName() string
	nodeID() int
	commitType() CommitType
	parse(*context)
}
type builder interface {
	nodeName() string
	nodeID() int
	build(*context) ([]*Node, bool)
}

var ErrInvalidUnicodeCharacter = errors.New("invalid unicode character")

func (pe *ParseError) Error() string {
	return fmt.Sprintf("%s:%d:%d:parse failed, parsing: %s", pe.Input, pe.Line+1, pe.Column+1, pe.Definition)
}
func parseInput(r io.Reader, p parser, b builder) (*Node, error) {
	c := newContext(bufio.NewReader(r))
	p.parse(c)
	if c.readErr != nil {
		return nil, c.readErr
	}
	if err := c.finalizeParse(p); err != nil {
		if perr, ok := err.(*ParseError); ok {
			perr.Input = "<input>"
		}
		return nil, err
	}
	c.offset = 0
	c.results.resetPending()
	n, _ := b.build(c)
	return n[0], nil
}

func parse(r io.Reader) (*Node, error) {
	var p131 = sequenceParser{id: 131, commit: 32, ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var p129 = choiceParser{id: 129, commit: 2}
	var p127 = sequenceParser{id: 127, commit: 70, name: "space", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{129}}
	var p1 = charParser{id: 1, chars: []rune{32, 8, 12, 13, 9, 11}}
	p127.items = []parser{&p1}
	var p128 = choiceParser{id: 128, commit: 70, name: "comment", generalizations: []int{129}}
	var p19 = sequenceParser{id: 19, commit: 74, name: "line-comment", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}, generalizations: []int{128, 129}}
	var p16 = sequenceParser{id: 16, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var p14 = charParser{id: 14, chars: []rune{47}}
	var p15 = charParser{id: 15, chars: []rune{47}}
	p16.items = []parser{&p14, &p15}
	var p18 = sequenceParser{id: 18, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p17 = charParser{id: 17, not: true, chars: []rune{10}}
	p18.items = []parser{&p17}
	p19.items = []parser{&p16, &p18}
	var p34 = sequenceParser{id: 34, commit: 74, name: "block-comment", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {1, 1}, {0, -1}, {1, 1}}, generalizations: []int{128, 129}}
	var p22 = sequenceParser{id: 22, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var p20 = charParser{id: 20, chars: []rune{47}}
	var p21 = charParser{id: 21, chars: []rune{42}}
	p22.items = []parser{&p20, &p21}
	var p30 = choiceParser{id: 30, commit: 10}
	var p24 = sequenceParser{id: 24, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{30}}
	var p23 = charParser{id: 23, not: true, chars: []rune{42}}
	p24.items = []parser{&p23}
	var p29 = sequenceParser{id: 29, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{30}}
	var p26 = sequenceParser{id: 26, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p25 = charParser{id: 25, chars: []rune{42}}
	p26.items = []parser{&p25}
	var p28 = sequenceParser{id: 28, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p27 = charParser{id: 27, not: true, chars: []rune{47}}
	p28.items = []parser{&p27}
	p29.items = []parser{&p26, &p28}
	p30.options = []parser{&p24, &p29}
	var p33 = sequenceParser{id: 33, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var p31 = charParser{id: 31, chars: []rune{42}}
	var p32 = charParser{id: 32, chars: []rune{47}}
	p33.items = []parser{&p31, &p32}
	p34.items = []parser{&p22, &p30, &p33}
	p128.options = []parser{&p19, &p34}
	p129.options = []parser{&p127, &p128}
	var p130 = sequenceParser{id: 130, commit: 66, name: "nred:wsroot", ranges: [][]int{{1, 1}}}
	var p126 = sequenceParser{id: 126, commit: 66, name: "definitions", ranges: [][]int{{0, 1}, {0, 1}}}
	var p123 = sequenceParser{id: 123, commit: 2, ranges: [][]int{{1, 1}, {0, -1}}}
	var p116 = choiceParser{id: 116, commit: 2}
	var p3 = sequenceParser{id: 3, commit: 74, name: "nl", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{116, 87, 88, 91, 118}}
	var p2 = charParser{id: 2, chars: []rune{10}}
	p3.items = []parser{&p2}
	var p8 = sequenceParser{id: 8, commit: 74, name: "colon", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{116, 118}}
	var p7 = charParser{id: 7, chars: []rune{59}}
	p8.items = []parser{&p7}
	p116.options = []parser{&p3, &p8}
	var p122 = sequenceParser{id: 122, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p122.items = []parser{&p129, &p116}
	p123.items = []parser{&p116, &p122}
	var p125 = sequenceParser{id: 125, commit: 2, ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var p121 = sequenceParser{id: 121, commit: 2, ranges: [][]int{{1, 1}, {0, 1}}}
	var p117 = choiceParser{id: 117, commit: 2}
	var p107 = sequenceParser{id: 107, commit: 64, name: "local", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}}, generalizations: []int{117}}
	var p106 = sequenceParser{id: 106, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var p103 = charParser{id: 103, chars: []rune{108}}
	var p104 = charParser{id: 104, chars: []rune{101}}
	var p105 = charParser{id: 105, chars: []rune{116}}
	p106.items = []parser{&p103, &p104, &p105}
	var p6 = sequenceParser{id: 6, commit: 66, name: "nls", ranges: [][]int{{0, 1}}}
	var p5 = sequenceParser{id: 5, commit: 2, ranges: [][]int{{1, 1}, {0, -1}}}
	var p4 = sequenceParser{id: 4, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p4.items = []parser{&p129, &p3}
	p5.items = []parser{&p3, &p4}
	p6.items = []parser{&p5}
	var p83 = sequenceParser{id: 83, commit: 72, name: "symbol", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}, generalizations: []int{84, 102}}
	var p80 = sequenceParser{id: 80, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p79 = charParser{id: 79, chars: []rune{95}, ranges: [][]rune{{97, 122}, {65, 90}}}
	p80.items = []parser{&p79}
	var p82 = sequenceParser{id: 82, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p81 = charParser{id: 81, chars: []rune{95}, ranges: [][]rune{{97, 122}, {65, 90}, {48, 57}}}
	p82.items = []parser{&p81}
	p83.items = []parser{&p80, &p82}
	var p13 = sequenceParser{id: 13, commit: 66, name: "opteq", ranges: [][]int{{0, 1}}}
	var p12 = sequenceParser{id: 12, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p11 = charParser{id: 11, chars: []rune{61}}
	p12.items = []parser{&p11}
	p13.items = []parser{&p12}
	var p102 = choiceParser{id: 102, commit: 64, name: "expression"}
	var p84 = choiceParser{id: 84, commit: 66, name: "primitive-expression", generalizations: []int{102}}
	var p52 = choiceParser{id: 52, commit: 64, name: "int", generalizations: []int{84, 102}}
	var p43 = sequenceParser{id: 43, commit: 74, name: "decimal", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}, generalizations: []int{52, 84, 102}}
	var p42 = sequenceParser{id: 42, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p41 = charParser{id: 41, ranges: [][]rune{{49, 57}}}
	p42.items = []parser{&p41}
	var p36 = sequenceParser{id: 36, commit: 66, name: "decimal-digit", allChars: true, ranges: [][]int{{1, 1}}}
	var p35 = charParser{id: 35, ranges: [][]rune{{48, 57}}}
	p36.items = []parser{&p35}
	p43.items = []parser{&p42, &p36}
	var p46 = sequenceParser{id: 46, commit: 74, name: "octal", ranges: [][]int{{1, 1}, {1, -1}, {1, 1}, {1, -1}}, generalizations: []int{52, 84, 102}}
	var p45 = sequenceParser{id: 45, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p44 = charParser{id: 44, chars: []rune{48}}
	p45.items = []parser{&p44}
	var p38 = sequenceParser{id: 38, commit: 66, name: "octal-digit", allChars: true, ranges: [][]int{{1, 1}}}
	var p37 = charParser{id: 37, ranges: [][]rune{{48, 55}}}
	p38.items = []parser{&p37}
	p46.items = []parser{&p45, &p38}
	var p51 = sequenceParser{id: 51, commit: 74, name: "hexa", ranges: [][]int{{1, 1}, {1, 1}, {1, -1}, {1, 1}, {1, 1}, {1, -1}}, generalizations: []int{52, 84, 102}}
	var p48 = sequenceParser{id: 48, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p47 = charParser{id: 47, chars: []rune{48}}
	p48.items = []parser{&p47}
	var p50 = sequenceParser{id: 50, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p49 = charParser{id: 49, chars: []rune{120, 88}}
	p50.items = []parser{&p49}
	var p40 = sequenceParser{id: 40, commit: 66, name: "hexa-digit", allChars: true, ranges: [][]int{{1, 1}}}
	var p39 = charParser{id: 39, ranges: [][]rune{{48, 57}, {97, 102}, {65, 70}}}
	p40.items = []parser{&p39}
	p51.items = []parser{&p48, &p50, &p40}
	p52.options = []parser{&p43, &p46, &p51}
	var p65 = choiceParser{id: 65, commit: 72, name: "float", generalizations: []int{84, 102}}
	var p60 = sequenceParser{id: 60, commit: 10, ranges: [][]int{{1, -1}, {1, 1}, {0, -1}, {0, 1}, {1, -1}, {1, 1}, {0, -1}, {0, 1}}, generalizations: []int{65, 84, 102}}
	var p59 = sequenceParser{id: 59, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p58 = charParser{id: 58, chars: []rune{46}}
	p59.items = []parser{&p58}
	var p57 = sequenceParser{id: 57, commit: 74, name: "exponent", ranges: [][]int{{1, 1}, {0, 1}, {1, -1}, {1, 1}, {0, 1}, {1, -1}}}
	var p54 = sequenceParser{id: 54, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p53 = charParser{id: 53, chars: []rune{101, 69}}
	p54.items = []parser{&p53}
	var p56 = sequenceParser{id: 56, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p55 = charParser{id: 55, chars: []rune{43, 45}}
	p56.items = []parser{&p55}
	p57.items = []parser{&p54, &p56, &p36}
	p60.items = []parser{&p36, &p59, &p36, &p57}
	var p63 = sequenceParser{id: 63, commit: 10, ranges: [][]int{{1, 1}, {1, -1}, {0, 1}, {1, 1}, {1, -1}, {0, 1}}, generalizations: []int{65, 84, 102}}
	var p62 = sequenceParser{id: 62, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p61 = charParser{id: 61, chars: []rune{46}}
	p62.items = []parser{&p61}
	p63.items = []parser{&p62, &p36, &p57}
	var p64 = sequenceParser{id: 64, commit: 10, ranges: [][]int{{1, -1}, {1, 1}, {1, -1}, {1, 1}}, generalizations: []int{65, 84, 102}}
	p64.items = []parser{&p36, &p57}
	p65.options = []parser{&p60, &p63, &p64}
	var p78 = sequenceParser{id: 78, commit: 72, name: "string", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {1, 1}, {0, -1}, {1, 1}}, generalizations: []int{84, 102}}
	var p67 = sequenceParser{id: 67, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p66 = charParser{id: 66, chars: []rune{34}}
	p67.items = []parser{&p66}
	var p75 = choiceParser{id: 75, commit: 10}
	var p69 = sequenceParser{id: 69, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{75}}
	var p68 = charParser{id: 68, not: true, chars: []rune{92, 34}}
	p69.items = []parser{&p68}
	var p74 = sequenceParser{id: 74, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}, generalizations: []int{75}}
	var p71 = sequenceParser{id: 71, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p70 = charParser{id: 70, chars: []rune{92}}
	p71.items = []parser{&p70}
	var p73 = sequenceParser{id: 73, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p72 = charParser{id: 72, not: true}
	p73.items = []parser{&p72}
	p74.items = []parser{&p71, &p73}
	p75.options = []parser{&p69, &p74}
	var p77 = sequenceParser{id: 77, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p76 = charParser{id: 76, chars: []rune{34}}
	p77.items = []parser{&p76}
	p78.items = []parser{&p67, &p75, &p77}
	p84.options = []parser{&p52, &p65, &p78, &p83}
	var p101 = sequenceParser{id: 101, commit: 66, name: "composite-expression", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}, {0, 1}, {0, -1}, {0, 1}, {0, -1}, {1, 1}}, generalizations: []int{102}}
	var p86 = sequenceParser{id: 86, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p85 = charParser{id: 85, chars: []rune{40}}
	p86.items = []parser{&p85}
	var p100 = sequenceParser{id: 100, commit: 2, ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var p87 = choiceParser{id: 87, commit: 2}
	var p10 = sequenceParser{id: 10, commit: 74, name: "comma", allChars: true, ranges: [][]int{{1, 1}, {1, 1}}, generalizations: []int{87, 88, 91}}
	var p9 = charParser{id: 9, chars: []rune{44}}
	p10.items = []parser{&p9}
	p87.options = []parser{&p3, &p10}
	var p99 = sequenceParser{id: 99, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p99.items = []parser{&p129, &p87}
	p100.items = []parser{&p129, &p87, &p99}
	var p96 = sequenceParser{id: 96, commit: 2, ranges: [][]int{{1, 1}, {0, 1}, {0, 1}}}
	var p93 = sequenceParser{id: 93, commit: 2, ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var p90 = sequenceParser{id: 90, commit: 2, ranges: [][]int{{1, 1}, {0, -1}, {0, -1}, {1, 1}}}
	var p88 = choiceParser{id: 88, commit: 2}
	p88.options = []parser{&p3, &p10}
	var p89 = sequenceParser{id: 89, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p89.items = []parser{&p129, &p88}
	p90.items = []parser{&p88, &p89, &p129, &p102}
	var p92 = sequenceParser{id: 92, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p92.items = []parser{&p129, &p90}
	p93.items = []parser{&p129, &p90, &p92}
	var p95 = sequenceParser{id: 95, commit: 2, ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var p91 = choiceParser{id: 91, commit: 2}
	p91.options = []parser{&p3, &p10}
	var p94 = sequenceParser{id: 94, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p94.items = []parser{&p129, &p91}
	p95.items = []parser{&p129, &p91, &p94}
	p96.items = []parser{&p102, &p93, &p95}
	var p98 = sequenceParser{id: 98, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var p97 = charParser{id: 97, chars: []rune{41}}
	p98.items = []parser{&p97}
	p101.items = []parser{&p83, &p129, &p6, &p129, &p86, &p100, &p129, &p96, &p129, &p98}
	p102.options = []parser{&p84, &p101}
	p107.items = []parser{&p106, &p129, &p6, &p129, &p83, &p129, &p6, &p129, &p13, &p129, &p6, &p129, &p102}
	var p115 = sequenceParser{id: 115, commit: 64, name: "export", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}}, generalizations: []int{117}}
	var p114 = sequenceParser{id: 114, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var p108 = charParser{id: 108, chars: []rune{101}}
	var p109 = charParser{id: 109, chars: []rune{120}}
	var p110 = charParser{id: 110, chars: []rune{112}}
	var p111 = charParser{id: 111, chars: []rune{111}}
	var p112 = charParser{id: 112, chars: []rune{114}}
	var p113 = charParser{id: 113, chars: []rune{116}}
	p114.items = []parser{&p108, &p109, &p110, &p111, &p112, &p113}
	p115.items = []parser{&p114, &p129, &p6, &p129, &p78, &p129, &p6, &p129, &p13, &p129, &p6, &p129, &p102}
	p117.options = []parser{&p107, &p115}
	var p120 = sequenceParser{id: 120, commit: 2, ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var p118 = choiceParser{id: 118, commit: 2}
	p118.options = []parser{&p3, &p8}
	var p119 = sequenceParser{id: 119, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p119.items = []parser{&p129, &p118}
	p120.items = []parser{&p129, &p118, &p119}
	p121.items = []parser{&p117, &p120}
	var p124 = sequenceParser{id: 124, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	p124.items = []parser{&p129, &p121}
	p125.items = []parser{&p129, &p121, &p124}
	p126.items = []parser{&p123, &p125}
	p130.items = []parser{&p126}
	p131.items = []parser{&p129, &p130, &p129}
	var b131 = sequenceBuilder{id: 131, commit: 32, name: "nred", ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var b129 = choiceBuilder{id: 129, commit: 2}
	var b127 = sequenceBuilder{id: 127, commit: 70, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b1 = charBuilder{}
	b127.items = []builder{&b1}
	var b128 = choiceBuilder{id: 128, commit: 70}
	var b19 = sequenceBuilder{id: 19, commit: 74, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}}
	var b16 = sequenceBuilder{id: 16, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b14 = charBuilder{}
	var b15 = charBuilder{}
	b16.items = []builder{&b14, &b15}
	var b18 = sequenceBuilder{id: 18, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b17 = charBuilder{}
	b18.items = []builder{&b17}
	b19.items = []builder{&b16, &b18}
	var b34 = sequenceBuilder{id: 34, commit: 74, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {1, 1}, {0, -1}, {1, 1}}}
	var b22 = sequenceBuilder{id: 22, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b20 = charBuilder{}
	var b21 = charBuilder{}
	b22.items = []builder{&b20, &b21}
	var b30 = choiceBuilder{id: 30, commit: 10}
	var b24 = sequenceBuilder{id: 24, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b23 = charBuilder{}
	b24.items = []builder{&b23}
	var b29 = sequenceBuilder{id: 29, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b26 = sequenceBuilder{id: 26, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b25 = charBuilder{}
	b26.items = []builder{&b25}
	var b28 = sequenceBuilder{id: 28, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b27 = charBuilder{}
	b28.items = []builder{&b27}
	b29.items = []builder{&b26, &b28}
	b30.options = []builder{&b24, &b29}
	var b33 = sequenceBuilder{id: 33, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b31 = charBuilder{}
	var b32 = charBuilder{}
	b33.items = []builder{&b31, &b32}
	b34.items = []builder{&b22, &b30, &b33}
	b128.options = []builder{&b19, &b34}
	b129.options = []builder{&b127, &b128}
	var b130 = sequenceBuilder{id: 130, commit: 66, ranges: [][]int{{1, 1}}}
	var b126 = sequenceBuilder{id: 126, commit: 66, ranges: [][]int{{0, 1}, {0, 1}}}
	var b123 = sequenceBuilder{id: 123, commit: 2, ranges: [][]int{{1, 1}, {0, -1}}}
	var b116 = choiceBuilder{id: 116, commit: 2}
	var b3 = sequenceBuilder{id: 3, commit: 74, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b2 = charBuilder{}
	b3.items = []builder{&b2}
	var b8 = sequenceBuilder{id: 8, commit: 74, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b7 = charBuilder{}
	b8.items = []builder{&b7}
	b116.options = []builder{&b3, &b8}
	var b122 = sequenceBuilder{id: 122, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b122.items = []builder{&b129, &b116}
	b123.items = []builder{&b116, &b122}
	var b125 = sequenceBuilder{id: 125, commit: 2, ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var b121 = sequenceBuilder{id: 121, commit: 2, ranges: [][]int{{1, 1}, {0, 1}}}
	var b117 = choiceBuilder{id: 117, commit: 2}
	var b107 = sequenceBuilder{id: 107, commit: 64, name: "local", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}}}
	var b106 = sequenceBuilder{id: 106, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b103 = charBuilder{}
	var b104 = charBuilder{}
	var b105 = charBuilder{}
	b106.items = []builder{&b103, &b104, &b105}
	var b6 = sequenceBuilder{id: 6, commit: 66, ranges: [][]int{{0, 1}}}
	var b5 = sequenceBuilder{id: 5, commit: 2, ranges: [][]int{{1, 1}, {0, -1}}}
	var b4 = sequenceBuilder{id: 4, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b4.items = []builder{&b129, &b3}
	b5.items = []builder{&b3, &b4}
	b6.items = []builder{&b5}
	var b83 = sequenceBuilder{id: 83, commit: 72, name: "symbol", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}}
	var b80 = sequenceBuilder{id: 80, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b79 = charBuilder{}
	b80.items = []builder{&b79}
	var b82 = sequenceBuilder{id: 82, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b81 = charBuilder{}
	b82.items = []builder{&b81}
	b83.items = []builder{&b80, &b82}
	var b13 = sequenceBuilder{id: 13, commit: 66, ranges: [][]int{{0, 1}}}
	var b12 = sequenceBuilder{id: 12, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b11 = charBuilder{}
	b12.items = []builder{&b11}
	b13.items = []builder{&b12}
	var b102 = choiceBuilder{id: 102, commit: 64, name: "expression"}
	var b84 = choiceBuilder{id: 84, commit: 66}
	var b52 = choiceBuilder{id: 52, commit: 64, name: "int"}
	var b43 = sequenceBuilder{id: 43, commit: 74, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}}}
	var b42 = sequenceBuilder{id: 42, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b41 = charBuilder{}
	b42.items = []builder{&b41}
	var b36 = sequenceBuilder{id: 36, commit: 66, allChars: true, ranges: [][]int{{1, 1}}}
	var b35 = charBuilder{}
	b36.items = []builder{&b35}
	b43.items = []builder{&b42, &b36}
	var b46 = sequenceBuilder{id: 46, commit: 74, ranges: [][]int{{1, 1}, {1, -1}, {1, 1}, {1, -1}}}
	var b45 = sequenceBuilder{id: 45, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b44 = charBuilder{}
	b45.items = []builder{&b44}
	var b38 = sequenceBuilder{id: 38, commit: 66, allChars: true, ranges: [][]int{{1, 1}}}
	var b37 = charBuilder{}
	b38.items = []builder{&b37}
	b46.items = []builder{&b45, &b38}
	var b51 = sequenceBuilder{id: 51, commit: 74, ranges: [][]int{{1, 1}, {1, 1}, {1, -1}, {1, 1}, {1, 1}, {1, -1}}}
	var b48 = sequenceBuilder{id: 48, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b47 = charBuilder{}
	b48.items = []builder{&b47}
	var b50 = sequenceBuilder{id: 50, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b49 = charBuilder{}
	b50.items = []builder{&b49}
	var b40 = sequenceBuilder{id: 40, commit: 66, allChars: true, ranges: [][]int{{1, 1}}}
	var b39 = charBuilder{}
	b40.items = []builder{&b39}
	b51.items = []builder{&b48, &b50, &b40}
	b52.options = []builder{&b43, &b46, &b51}
	var b65 = choiceBuilder{id: 65, commit: 72, name: "float"}
	var b60 = sequenceBuilder{id: 60, commit: 10, ranges: [][]int{{1, -1}, {1, 1}, {0, -1}, {0, 1}, {1, -1}, {1, 1}, {0, -1}, {0, 1}}}
	var b59 = sequenceBuilder{id: 59, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b58 = charBuilder{}
	b59.items = []builder{&b58}
	var b57 = sequenceBuilder{id: 57, commit: 74, ranges: [][]int{{1, 1}, {0, 1}, {1, -1}, {1, 1}, {0, 1}, {1, -1}}}
	var b54 = sequenceBuilder{id: 54, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b53 = charBuilder{}
	b54.items = []builder{&b53}
	var b56 = sequenceBuilder{id: 56, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b55 = charBuilder{}
	b56.items = []builder{&b55}
	b57.items = []builder{&b54, &b56, &b36}
	b60.items = []builder{&b36, &b59, &b36, &b57}
	var b63 = sequenceBuilder{id: 63, commit: 10, ranges: [][]int{{1, 1}, {1, -1}, {0, 1}, {1, 1}, {1, -1}, {0, 1}}}
	var b62 = sequenceBuilder{id: 62, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b61 = charBuilder{}
	b62.items = []builder{&b61}
	b63.items = []builder{&b62, &b36, &b57}
	var b64 = sequenceBuilder{id: 64, commit: 10, ranges: [][]int{{1, -1}, {1, 1}, {1, -1}, {1, 1}}}
	b64.items = []builder{&b36, &b57}
	b65.options = []builder{&b60, &b63, &b64}
	var b78 = sequenceBuilder{id: 78, commit: 72, name: "string", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {1, 1}, {0, -1}, {1, 1}}}
	var b67 = sequenceBuilder{id: 67, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b66 = charBuilder{}
	b67.items = []builder{&b66}
	var b75 = choiceBuilder{id: 75, commit: 10}
	var b69 = sequenceBuilder{id: 69, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b68 = charBuilder{}
	b69.items = []builder{&b68}
	var b74 = sequenceBuilder{id: 74, commit: 10, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b71 = sequenceBuilder{id: 71, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b70 = charBuilder{}
	b71.items = []builder{&b70}
	var b73 = sequenceBuilder{id: 73, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b72 = charBuilder{}
	b73.items = []builder{&b72}
	b74.items = []builder{&b71, &b73}
	b75.options = []builder{&b69, &b74}
	var b77 = sequenceBuilder{id: 77, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b76 = charBuilder{}
	b77.items = []builder{&b76}
	b78.items = []builder{&b67, &b75, &b77}
	b84.options = []builder{&b52, &b65, &b78, &b83}
	var b101 = sequenceBuilder{id: 101, commit: 66, ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}, {0, 1}, {0, -1}, {0, 1}, {0, -1}, {1, 1}}}
	var b86 = sequenceBuilder{id: 86, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b85 = charBuilder{}
	b86.items = []builder{&b85}
	var b100 = sequenceBuilder{id: 100, commit: 2, ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var b87 = choiceBuilder{id: 87, commit: 2}
	var b10 = sequenceBuilder{id: 10, commit: 74, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b9 = charBuilder{}
	b10.items = []builder{&b9}
	b87.options = []builder{&b3, &b10}
	var b99 = sequenceBuilder{id: 99, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b99.items = []builder{&b129, &b87}
	b100.items = []builder{&b129, &b87, &b99}
	var b96 = sequenceBuilder{id: 96, commit: 2, ranges: [][]int{{1, 1}, {0, 1}, {0, 1}}}
	var b93 = sequenceBuilder{id: 93, commit: 2, ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var b90 = sequenceBuilder{id: 90, commit: 2, ranges: [][]int{{1, 1}, {0, -1}, {0, -1}, {1, 1}}}
	var b88 = choiceBuilder{id: 88, commit: 2}
	b88.options = []builder{&b3, &b10}
	var b89 = sequenceBuilder{id: 89, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b89.items = []builder{&b129, &b88}
	b90.items = []builder{&b88, &b89, &b129, &b102}
	var b92 = sequenceBuilder{id: 92, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b92.items = []builder{&b129, &b90}
	b93.items = []builder{&b129, &b90, &b92}
	var b95 = sequenceBuilder{id: 95, commit: 2, ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var b91 = choiceBuilder{id: 91, commit: 2}
	b91.options = []builder{&b3, &b10}
	var b94 = sequenceBuilder{id: 94, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b94.items = []builder{&b129, &b91}
	b95.items = []builder{&b129, &b91, &b94}
	b96.items = []builder{&b102, &b93, &b95}
	var b98 = sequenceBuilder{id: 98, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}}}
	var b97 = charBuilder{}
	b98.items = []builder{&b97}
	b101.items = []builder{&b83, &b129, &b6, &b129, &b86, &b100, &b129, &b96, &b129, &b98}
	b102.options = []builder{&b84, &b101}
	b107.items = []builder{&b106, &b129, &b6, &b129, &b83, &b129, &b6, &b129, &b13, &b129, &b6, &b129, &b102}
	var b115 = sequenceBuilder{id: 115, commit: 64, name: "export", ranges: [][]int{{1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}, {0, -1}, {1, 1}}}
	var b114 = sequenceBuilder{id: 114, commit: 10, allChars: true, ranges: [][]int{{1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}}}
	var b108 = charBuilder{}
	var b109 = charBuilder{}
	var b110 = charBuilder{}
	var b111 = charBuilder{}
	var b112 = charBuilder{}
	var b113 = charBuilder{}
	b114.items = []builder{&b108, &b109, &b110, &b111, &b112, &b113}
	b115.items = []builder{&b114, &b129, &b6, &b129, &b78, &b129, &b6, &b129, &b13, &b129, &b6, &b129, &b102}
	b117.options = []builder{&b107, &b115}
	var b120 = sequenceBuilder{id: 120, commit: 2, ranges: [][]int{{0, -1}, {1, 1}, {0, -1}}}
	var b118 = choiceBuilder{id: 118, commit: 2}
	b118.options = []builder{&b3, &b8}
	var b119 = sequenceBuilder{id: 119, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b119.items = []builder{&b129, &b118}
	b120.items = []builder{&b129, &b118, &b119}
	b121.items = []builder{&b117, &b120}
	var b124 = sequenceBuilder{id: 124, commit: 2, ranges: [][]int{{0, -1}, {1, 1}}}
	b124.items = []builder{&b129, &b121}
	b125.items = []builder{&b129, &b121, &b124}
	b126.items = []builder{&b123, &b125}
	b130.items = []builder{&b126}
	b131.items = []builder{&b129, &b130, &b129}

	return parseInput(r, &p131, &b131)
}
