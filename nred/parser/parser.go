
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


package parser

import (
	"strconv"
	"io"
	"strings"
	"unicode"
	"fmt"
	"bufio"
	"errors"
)

type charParser struct {
	name	string
	id	int
	not	bool
	chars	[]rune
	ranges	[][]rune
}
type charBuilder struct {
	name	string
	id	int
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
	name		string
	id		int
	commit		CommitType
	items		[]parser
	ranges		[][]int
	generalizations	[]int
	allChars	bool
}
type sequenceBuilder struct {
	name		string
	id		int
	commit		CommitType
	items		[]builder
	ranges		[][]int
	allChars	bool
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
		currentCount	int
		parsed		bool
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
		itemIndex	int
		currentCount	int
		nodes		[]*Node
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
	name		string
	id		int
	commit		CommitType
	options		[]parser
	generalizations	[]int
}
type choiceBuilder struct {
	name	string
	id	int
	commit	CommitType
	options	[]builder
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
		match		bool
		optionIndex	int
		foundMatch	bool
		failingParser	parser
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
	noMatch		[]*idSet
	match		[][]int
	isPending	[][]int
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
	reader		io.RuneReader
	offset		int
	readOffset	int
	consumed	int
	failOffset	int
	failingParser	parser
	readErr		error
	eof		bool
	results		*results
	tokens		[]rune
	matchLast	bool
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
	Name		string
	Nodes		[]*Node
	From, To	int
	tokens		[]rune
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
	None	CommitType	= 0
	Alias	CommitType	= 1 << iota
	Whitespace
	NoWhitespace
	FailPass
	Root
	userDefined
)

type formatFlags int

const (
	formatNone	formatFlags	= 0
	formatPretty	formatFlags	= 1 << iota
	formatIncludeComments
)

type ParseError struct {
	Input		string
	Offset		int
	Line		int
	Column		int
	Definition	string
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


func Parse(r io.Reader) (*Node, error) {
var p135 = sequenceParser{id: 135, commit: 32,ranges: [][]int{{0, -1},{1, 1},{0, -1},},};var p133 = choiceParser{id: 133, commit: 2,};var p131 = sequenceParser{id: 131, commit: 70,name: "space",allChars: true,ranges: [][]int{{1, 1},{1, 1},},generalizations: []int{133,},};var p1 = charParser{id: 1,chars: []rune{32,8,12,13,9,11,},};p131.items = []parser{&p1,};var p132 = choiceParser{id: 132, commit: 70,name: "comment",generalizations: []int{133,},};var p21 = sequenceParser{id: 21, commit: 74,name: "line-comment",ranges: [][]int{{1, 1},{0, -1},{1, 1},{0, -1},},generalizations: []int{132,133,},};var p18 = sequenceParser{id: 18, commit: 10,allChars: true,ranges: [][]int{{1, 1},{1, 1},{1, 1},{1, 1},},};var p16 = charParser{id: 16,chars: []rune{47,},};var p17 = charParser{id: 17,chars: []rune{47,},};p18.items = []parser{&p16,&p17,};var p20 = sequenceParser{id: 20, commit: 10,allChars: true,ranges: [][]int{{1, 1},{1, 1},},};var p19 = charParser{id: 19,not: true,chars: []rune{10,},};p20.items = []parser{&p19,};p21.items = []parser{&p18,&p20,};var p36 = sequenceParser{id: 36, commit: 74,name: "block-comment",ranges: [][]int{{1, 1},{0, -1},{1, 1},{1, 1},{0, -1},{1, 1},},generalizations: []int{132,133,},};var p24 = sequenceParser{id: 24, commit: 10,allChars: true,ranges: [][]int{{1, 1},{1, 1},{1, 1},{1, 1},},};var p22 = charParser{id: 22,chars: []rune{47,},};var p23 = charParser{id: 23,chars: []rune{42,},};p24.items = []parser{&p22,&p23,};var p32 = choiceParser{id: 32, commit: 10,};var p26 = sequenceParser{id: 26, commit: 10,allChars: true,ranges: [][]int{{1, 1},{1, 1},},generalizations: []int{32,},};var p25 = charParser{id: 25,not: true,chars: []rune{42,},};p26.items = []parser{&p25,};var p31 = sequenceParser{id: 31, commit: 10,ranges: [][]int{{1, 1},{1, 1},{1, 1},{1, 1},},generalizations: []int{32,},};var p28 = sequenceParser{id: 28, commit: 10,allChars: true,ranges: [][]int{{1, 1},{1, 1},},};var p27 = charParser{id: 27,chars: []rune{42,},};p28.items = []parser{&p27,};var p30 = sequenceParser{id: 30, commit: 10,allChars: true,ranges: [][]int{{1, 1},{1, 1},},};var p29 = charParser{id: 29,not: true,chars: []rune{47,},};p30.items = []parser{&p29,};p31.items = []parser{&p28,&p30,};p32.options = []parser{&p26,&p31,};var p35 = sequenceParser{id: 35, commit: 10,allChars: true,ranges: [][]int{{1, 1},{1, 1},{1, 1},{1, 1},},};var p33 = charParser{id: 33,chars: []rune{42,},};var p34 = charParser{id: 34,chars: []rune{47,},};p35.items = []parser{&p33,&p34,};p36.items = []parser{&p24,&p32,&p35,};p132.options = []parser{&p21,&p36,};p133.options = []parser{&p131,&p132,};var p134 = sequenceParser{id: 134, commit: 66,name: "nred:wsroot",ranges: [][]int{{1, 1},},};var p130 = sequenceParser{id: 130, commit: 66,name: "definitions",ranges: [][]int{{0, 1},{0, 1},},};var p127 = sequenceParser{id: 127, commit: 2,ranges: [][]int{{1, 1},{0, -1},},};var p120 = choiceParser{id: 120, commit: 2,};var p3 = sequenceParser{id: 3, commit: 74,name: "nl",allChars: true,ranges: [][]int{{1, 1},{1, 1},},generalizations: []int{120,91,92,95,122,},};var p2 = charParser{id: 2,chars: []rune{10,},};p3.items = []parser{&p2,};var p8 = sequenceParser{id: 8, commit: 74,name: "colon",allChars: true,ranges: [][]int{{1, 1},{1, 1},},generalizations: []int{120,122,},};var p7 = charParser{id: 7,chars: []rune{59,},};p8.items = []parser{&p7,};p120.options = []parser{&p3,&p8,};var p126 = sequenceParser{id: 126, commit: 2,ranges: [][]int{{0, -1},{1, 1},},};p126.items = []parser{&p133,&p120,};p127.items = []parser{&p120,&p126,};var p129 = sequenceParser{id: 129, commit: 2,ranges: [][]int{{0, -1},{1, 1},{0, -1},},};var p125 = sequenceParser{id: 125, commit: 2,ranges: [][]int{{1, 1},{0, 1},},};var p121 = choiceParser{id: 121, commit: 2,};var p111 = sequenceParser{id: 111, commit: 64,name: "local",ranges: [][]int{{1, 1},{0, -1},{1, 1},{0, -1},{1, 1},{0, -1},{1, 1},{0, -1},{1, 1},{0, -1},{1, 1},{0, -1},{1, 1},},generalizations: []int{121,},};var p110 = sequenceParser{id: 110, commit: 10,allChars: true,ranges: [][]int{{1, 1},{1, 1},{1, 1},{1, 1},{1, 1},{1, 1},},};var p107 = charParser{id: 107,chars: []rune{108,},};var p108 = charParser{id: 108,chars: []rune{101,},};var p109 = charParser{id: 109,chars: []rune{116,},};p110.items = []parser{&p107,&p108,&p109,};var p6 = sequenceParser{id: 6, commit: 66,name: "nls",ranges: [][]int{{0, 1},},};var p5 = sequenceParser{id: 5, commit: 2,ranges: [][]int{{1, 1},{0, -1},},};var p4 = sequenceParser{id: 4, commit: 2,ranges: [][]int{{0, -1},{1, 1},},};p4.items = []parser{&p133,&p3,};p5.items = []parser{&p3,&p4,};p6.items = []parser{&p5,};var p87 = sequenceParser{id: 87, commit: 72,name: "symbol",ranges: [][]int{{1, 1},{0, -1},{1, 1},{0, -1},},generalizations: []int{88,106,},};var p85 = sequenceParser{id: 85, commit: 74,name: "symbol-tag",ranges: [][]int{{1, 1},{0, -1},{1, 1},{0, -1},},};var p82 = sequenceParser{id: 82, commit: 10,allChars: true,ranges: [][]int{{1, 1},{1, 1},},};var p81 = charParser{id: 81,chars: []rune{95,},ranges: [][]rune{{97, 122},{65, 90},},};p82.items = []parser{&p81,};var p84 = sequenceParser{id: 84, commit: 10,allChars: true,ranges: [][]int{{1, 1},{1, 1},},};var p83 = charParser{id: 83,chars: []rune{95,},ranges: [][]rune{{97, 122},{65, 90},{48, 57},},};p84.items = []parser{&p83,};p85.items = []parser{&p82,&p84,};var p86 = sequenceParser{id: 86, commit: 10,ranges: [][]int{{1, 1},{1, 1},{1, 1},{1, 1},{1, 1},{1, 1},{1, 1},{1, 1},},};var p10 = sequenceParser{id: 10, commit: 74,name: "dot",allChars: true,ranges: [][]int{{1, 1},{1, 1},},};var p9 = charParser{id: 9,chars: []rune{46,},};p10.items = []parser{&p9,};p86.items = []parser{&p6,&p10,&p6,&p85,};p87.items = []parser{&p85,&p86,};var p15 = sequenceParser{id: 15, commit: 66,name: "opteq",ranges: [][]int{{0, 1},},};var p14 = sequenceParser{id: 14, commit: 10,allChars: true,ranges: [][]int{{1, 1},{1, 1},},};var p13 = charParser{id: 13,chars: []rune{61,},};p14.items = []parser{&p13,};p15.items = []parser{&p14,};var p106 = choiceParser{id: 106, commit: 66,name: "expression",};var p88 = choiceParser{id: 88, commit: 66,name: "primitive-expression",generalizations: []int{106,},};var p54 = choiceParser{id: 54, commit: 64,name: "int",generalizations: []int{88,106,},};var p45 = sequenceParser{id: 45, commit: 74,name: "decimal",ranges: [][]int{{1, 1},{0, -1},{1, 1},{0, -1},},generalizations: []int{54,88,106,},};var p44 = sequenceParser{id: 44, commit: 10,allChars: true,ranges: [][]int{{1, 1},{1, 1},},};var p43 = charParser{id: 43,ranges: [][]rune{{49, 57},},};p44.items = []parser{&p43,};var p38 = sequenceParser{id: 38, commit: 66,name: "decimal-digit",allChars: true,ranges: [][]int{{1, 1},},};var p37 = charParser{id: 37,ranges: [][]rune{{48, 57},},};p38.items = []parser{&p37,};p45.items = []parser{&p44,&p38,};var p48 = sequenceParser{id: 48, commit: 74,name: "octal",ranges: [][]int{{1, 1},{1, -1},{1, 1},{1, -1},},generalizations: []int{54,88,106,},};var p47 = sequenceParser{id: 47, commit: 10,allChars: true,ranges: [][]int{{1, 1},{1, 1},},};var p46 = charParser{id: 46,chars: []rune{48,},};p47.items = []parser{&p46,};var p40 = sequenceParser{id: 40, commit: 66,name: "octal-digit",allChars: true,ranges: [][]int{{1, 1},},};var p39 = charParser{id: 39,ranges: [][]rune{{48, 55},},};p40.items = []parser{&p39,};p48.items = []parser{&p47,&p40,};var p53 = sequenceParser{id: 53, commit: 74,name: "hexa",ranges: [][]int{{1, 1},{1, 1},{1, -1},{1, 1},{1, 1},{1, -1},},generalizations: []int{54,88,106,},};var p50 = sequenceParser{id: 50, commit: 10,allChars: true,ranges: [][]int{{1, 1},{1, 1},},};var p49 = charParser{id: 49,chars: []rune{48,},};p50.items = []parser{&p49,};var p52 = sequenceParser{id: 52, commit: 10,allChars: true,ranges: [][]int{{1, 1},{1, 1},},};var p51 = charParser{id: 51,chars: []rune{120,88,},};p52.items = []parser{&p51,};var p42 = sequenceParser{id: 42, commit: 66,name: "hexa-digit",allChars: true,ranges: [][]int{{1, 1},},};var p41 = charParser{id: 41,ranges: [][]rune{{48, 57},{97, 102},{65, 70},},};p42.items = []parser{&p41,};p53.items = []parser{&p50,&p52,&p42,};p54.options = []parser{&p45,&p48,&p53,};var p67 = choiceParser{id: 67, commit: 72,name: "float",generalizations: []int{88,106,},};var p62 = sequenceParser{id: 62, commit: 10,ranges: [][]int{{1, -1},{1, 1},{0, -1},{0, 1},{1, -1},{1, 1},{0, -1},{0, 1},},generalizations: []int{67,88,106,},};var p61 = sequenceParser{id: 61, commit: 10,allChars: true,ranges: [][]int{{1, 1},{1, 1},},};var p60 = charParser{id: 60,chars: []rune{46,},};p61.items = []parser{&p60,};var p59 = sequenceParser{id: 59, commit: 74,name: "exponent",ranges: [][]int{{1, 1},{0, 1},{1, -1},{1, 1},{0, 1},{1, -1},},};var p56 = sequenceParser{id: 56, commit: 10,allChars: true,ranges: [][]int{{1, 1},{1, 1},},};var p55 = charParser{id: 55,chars: []rune{101,69,},};p56.items = []parser{&p55,};var p58 = sequenceParser{id: 58, commit: 10,allChars: true,ranges: [][]int{{1, 1},{1, 1},},};var p57 = charParser{id: 57,chars: []rune{43,45,},};p58.items = []parser{&p57,};p59.items = []parser{&p56,&p58,&p38,};p62.items = []parser{&p38,&p61,&p38,&p59,};var p65 = sequenceParser{id: 65, commit: 10,ranges: [][]int{{1, 1},{1, -1},{0, 1},{1, 1},{1, -1},{0, 1},},generalizations: []int{67,88,106,},};var p64 = sequenceParser{id: 64, commit: 10,allChars: true,ranges: [][]int{{1, 1},{1, 1},},};var p63 = charParser{id: 63,chars: []rune{46,},};p64.items = []parser{&p63,};p65.items = []parser{&p64,&p38,&p59,};var p66 = sequenceParser{id: 66, commit: 10,ranges: [][]int{{1, -1},{1, 1},{1, -1},{1, 1},},generalizations: []int{67,88,106,},};p66.items = []parser{&p38,&p59,};p67.options = []parser{&p62,&p65,&p66,};var p80 = sequenceParser{id: 80, commit: 72,name: "string",ranges: [][]int{{1, 1},{0, -1},{1, 1},{1, 1},{0, -1},{1, 1},},generalizations: []int{88,106,},};var p69 = sequenceParser{id: 69, commit: 10,allChars: true,ranges: [][]int{{1, 1},{1, 1},},};var p68 = charParser{id: 68,chars: []rune{34,},};p69.items = []parser{&p68,};var p77 = choiceParser{id: 77, commit: 10,};var p71 = sequenceParser{id: 71, commit: 10,allChars: true,ranges: [][]int{{1, 1},{1, 1},},generalizations: []int{77,},};var p70 = charParser{id: 70,not: true,chars: []rune{92,34,},};p71.items = []parser{&p70,};var p76 = sequenceParser{id: 76, commit: 10,ranges: [][]int{{1, 1},{1, 1},{1, 1},{1, 1},},generalizations: []int{77,},};var p73 = sequenceParser{id: 73, commit: 10,allChars: true,ranges: [][]int{{1, 1},{1, 1},},};var p72 = charParser{id: 72,chars: []rune{92,},};p73.items = []parser{&p72,};var p75 = sequenceParser{id: 75, commit: 10,allChars: true,ranges: [][]int{{1, 1},{1, 1},},};var p74 = charParser{id: 74,not: true,};p75.items = []parser{&p74,};p76.items = []parser{&p73,&p75,};p77.options = []parser{&p71,&p76,};var p79 = sequenceParser{id: 79, commit: 10,allChars: true,ranges: [][]int{{1, 1},{1, 1},},};var p78 = charParser{id: 78,chars: []rune{34,},};p79.items = []parser{&p78,};p80.items = []parser{&p69,&p77,&p79,};p88.options = []parser{&p54,&p67,&p80,&p87,};var p105 = sequenceParser{id: 105, commit: 64,name: "composite-expression",ranges: [][]int{{1, 1},{0, -1},{1, 1},{0, -1},{1, 1},{0, 1},{0, -1},{0, 1},{0, -1},{1, 1},},generalizations: []int{106,},};var p90 = sequenceParser{id: 90, commit: 10,allChars: true,ranges: [][]int{{1, 1},{1, 1},},};var p89 = charParser{id: 89,chars: []rune{40,},};p90.items = []parser{&p89,};var p104 = sequenceParser{id: 104, commit: 2,ranges: [][]int{{0, -1},{1, 1},{0, -1},},};var p91 = choiceParser{id: 91, commit: 2,};var p12 = sequenceParser{id: 12, commit: 74,name: "comma",allChars: true,ranges: [][]int{{1, 1},{1, 1},},generalizations: []int{91,92,95,},};var p11 = charParser{id: 11,chars: []rune{44,},};p12.items = []parser{&p11,};p91.options = []parser{&p3,&p12,};var p103 = sequenceParser{id: 103, commit: 2,ranges: [][]int{{0, -1},{1, 1},},};p103.items = []parser{&p133,&p91,};p104.items = []parser{&p133,&p91,&p103,};var p100 = sequenceParser{id: 100, commit: 2,ranges: [][]int{{1, 1},{0, 1},{0, 1},},};var p97 = sequenceParser{id: 97, commit: 2,ranges: [][]int{{0, -1},{1, 1},{0, -1},},};var p94 = sequenceParser{id: 94, commit: 2,ranges: [][]int{{1, 1},{0, -1},{0, -1},{1, 1},},};var p92 = choiceParser{id: 92, commit: 2,};p92.options = []parser{&p3,&p12,};var p93 = sequenceParser{id: 93, commit: 2,ranges: [][]int{{0, -1},{1, 1},},};p93.items = []parser{&p133,&p92,};p94.items = []parser{&p92,&p93,&p133,&p106,};var p96 = sequenceParser{id: 96, commit: 2,ranges: [][]int{{0, -1},{1, 1},},};p96.items = []parser{&p133,&p94,};p97.items = []parser{&p133,&p94,&p96,};var p99 = sequenceParser{id: 99, commit: 2,ranges: [][]int{{0, -1},{1, 1},{0, -1},},};var p95 = choiceParser{id: 95, commit: 2,};p95.options = []parser{&p3,&p12,};var p98 = sequenceParser{id: 98, commit: 2,ranges: [][]int{{0, -1},{1, 1},},};p98.items = []parser{&p133,&p95,};p99.items = []parser{&p133,&p95,&p98,};p100.items = []parser{&p106,&p97,&p99,};var p102 = sequenceParser{id: 102, commit: 10,allChars: true,ranges: [][]int{{1, 1},{1, 1},},};var p101 = charParser{id: 101,chars: []rune{41,},};p102.items = []parser{&p101,};p105.items = []parser{&p87,&p133,&p6,&p133,&p90,&p104,&p133,&p100,&p133,&p102,};p106.options = []parser{&p88,&p105,};p111.items = []parser{&p110,&p133,&p6,&p133,&p87,&p133,&p6,&p133,&p15,&p133,&p6,&p133,&p106,};var p119 = sequenceParser{id: 119, commit: 64,name: "export",ranges: [][]int{{1, 1},{0, -1},{1, 1},{0, -1},{1, 1},{0, -1},{1, 1},{0, -1},{1, 1},{0, -1},{1, 1},{0, -1},{1, 1},},generalizations: []int{121,},};var p118 = sequenceParser{id: 118, commit: 10,allChars: true,ranges: [][]int{{1, 1},{1, 1},{1, 1},{1, 1},{1, 1},{1, 1},{1, 1},{1, 1},{1, 1},{1, 1},{1, 1},{1, 1},},};var p112 = charParser{id: 112,chars: []rune{101,},};var p113 = charParser{id: 113,chars: []rune{120,},};var p114 = charParser{id: 114,chars: []rune{112,},};var p115 = charParser{id: 115,chars: []rune{111,},};var p116 = charParser{id: 116,chars: []rune{114,},};var p117 = charParser{id: 117,chars: []rune{116,},};p118.items = []parser{&p112,&p113,&p114,&p115,&p116,&p117,};p119.items = []parser{&p118,&p133,&p6,&p133,&p80,&p133,&p6,&p133,&p15,&p133,&p6,&p133,&p106,};p121.options = []parser{&p111,&p119,};var p124 = sequenceParser{id: 124, commit: 2,ranges: [][]int{{0, -1},{1, 1},{0, -1},},};var p122 = choiceParser{id: 122, commit: 2,};p122.options = []parser{&p3,&p8,};var p123 = sequenceParser{id: 123, commit: 2,ranges: [][]int{{0, -1},{1, 1},},};p123.items = []parser{&p133,&p122,};p124.items = []parser{&p133,&p122,&p123,};p125.items = []parser{&p121,&p124,};var p128 = sequenceParser{id: 128, commit: 2,ranges: [][]int{{0, -1},{1, 1},},};p128.items = []parser{&p133,&p125,};p129.items = []parser{&p133,&p125,&p128,};p130.items = []parser{&p127,&p129,};p134.items = []parser{&p130,};p135.items = []parser{&p133,&p134,&p133,};var b135 = sequenceBuilder{id: 135, commit: 32,name: "nred",ranges: [][]int{{0, -1},{1, 1},{0, -1},},};var b133 = choiceBuilder{id: 133, commit: 2,};var b131 = sequenceBuilder{id: 131, commit: 70,allChars: true,ranges: [][]int{{1, 1},{1, 1},},};var b1 = charBuilder{};b131.items = []builder{&b1,};var b132 = choiceBuilder{id: 132, commit: 70,};var b21 = sequenceBuilder{id: 21, commit: 74,ranges: [][]int{{1, 1},{0, -1},{1, 1},{0, -1},},};var b18 = sequenceBuilder{id: 18, commit: 10,allChars: true,ranges: [][]int{{1, 1},{1, 1},{1, 1},{1, 1},},};var b16 = charBuilder{};var b17 = charBuilder{};b18.items = []builder{&b16,&b17,};var b20 = sequenceBuilder{id: 20, commit: 10,allChars: true,ranges: [][]int{{1, 1},{1, 1},},};var b19 = charBuilder{};b20.items = []builder{&b19,};b21.items = []builder{&b18,&b20,};var b36 = sequenceBuilder{id: 36, commit: 74,ranges: [][]int{{1, 1},{0, -1},{1, 1},{1, 1},{0, -1},{1, 1},},};var b24 = sequenceBuilder{id: 24, commit: 10,allChars: true,ranges: [][]int{{1, 1},{1, 1},{1, 1},{1, 1},},};var b22 = charBuilder{};var b23 = charBuilder{};b24.items = []builder{&b22,&b23,};var b32 = choiceBuilder{id: 32, commit: 10,};var b26 = sequenceBuilder{id: 26, commit: 10,allChars: true,ranges: [][]int{{1, 1},{1, 1},},};var b25 = charBuilder{};b26.items = []builder{&b25,};var b31 = sequenceBuilder{id: 31, commit: 10,ranges: [][]int{{1, 1},{1, 1},{1, 1},{1, 1},},};var b28 = sequenceBuilder{id: 28, commit: 10,allChars: true,ranges: [][]int{{1, 1},{1, 1},},};var b27 = charBuilder{};b28.items = []builder{&b27,};var b30 = sequenceBuilder{id: 30, commit: 10,allChars: true,ranges: [][]int{{1, 1},{1, 1},},};var b29 = charBuilder{};b30.items = []builder{&b29,};b31.items = []builder{&b28,&b30,};b32.options = []builder{&b26,&b31,};var b35 = sequenceBuilder{id: 35, commit: 10,allChars: true,ranges: [][]int{{1, 1},{1, 1},{1, 1},{1, 1},},};var b33 = charBuilder{};var b34 = charBuilder{};b35.items = []builder{&b33,&b34,};b36.items = []builder{&b24,&b32,&b35,};b132.options = []builder{&b21,&b36,};b133.options = []builder{&b131,&b132,};var b134 = sequenceBuilder{id: 134, commit: 66,ranges: [][]int{{1, 1},},};var b130 = sequenceBuilder{id: 130, commit: 66,ranges: [][]int{{0, 1},{0, 1},},};var b127 = sequenceBuilder{id: 127, commit: 2,ranges: [][]int{{1, 1},{0, -1},},};var b120 = choiceBuilder{id: 120, commit: 2,};var b3 = sequenceBuilder{id: 3, commit: 74,allChars: true,ranges: [][]int{{1, 1},{1, 1},},};var b2 = charBuilder{};b3.items = []builder{&b2,};var b8 = sequenceBuilder{id: 8, commit: 74,allChars: true,ranges: [][]int{{1, 1},{1, 1},},};var b7 = charBuilder{};b8.items = []builder{&b7,};b120.options = []builder{&b3,&b8,};var b126 = sequenceBuilder{id: 126, commit: 2,ranges: [][]int{{0, -1},{1, 1},},};b126.items = []builder{&b133,&b120,};b127.items = []builder{&b120,&b126,};var b129 = sequenceBuilder{id: 129, commit: 2,ranges: [][]int{{0, -1},{1, 1},{0, -1},},};var b125 = sequenceBuilder{id: 125, commit: 2,ranges: [][]int{{1, 1},{0, 1},},};var b121 = choiceBuilder{id: 121, commit: 2,};var b111 = sequenceBuilder{id: 111, commit: 64,name: "local",ranges: [][]int{{1, 1},{0, -1},{1, 1},{0, -1},{1, 1},{0, -1},{1, 1},{0, -1},{1, 1},{0, -1},{1, 1},{0, -1},{1, 1},},};var b110 = sequenceBuilder{id: 110, commit: 10,allChars: true,ranges: [][]int{{1, 1},{1, 1},{1, 1},{1, 1},{1, 1},{1, 1},},};var b107 = charBuilder{};var b108 = charBuilder{};var b109 = charBuilder{};b110.items = []builder{&b107,&b108,&b109,};var b6 = sequenceBuilder{id: 6, commit: 66,ranges: [][]int{{0, 1},},};var b5 = sequenceBuilder{id: 5, commit: 2,ranges: [][]int{{1, 1},{0, -1},},};var b4 = sequenceBuilder{id: 4, commit: 2,ranges: [][]int{{0, -1},{1, 1},},};b4.items = []builder{&b133,&b3,};b5.items = []builder{&b3,&b4,};b6.items = []builder{&b5,};var b87 = sequenceBuilder{id: 87, commit: 72,name: "symbol",ranges: [][]int{{1, 1},{0, -1},{1, 1},{0, -1},},};var b85 = sequenceBuilder{id: 85, commit: 74,ranges: [][]int{{1, 1},{0, -1},{1, 1},{0, -1},},};var b82 = sequenceBuilder{id: 82, commit: 10,allChars: true,ranges: [][]int{{1, 1},{1, 1},},};var b81 = charBuilder{};b82.items = []builder{&b81,};var b84 = sequenceBuilder{id: 84, commit: 10,allChars: true,ranges: [][]int{{1, 1},{1, 1},},};var b83 = charBuilder{};b84.items = []builder{&b83,};b85.items = []builder{&b82,&b84,};var b86 = sequenceBuilder{id: 86, commit: 10,ranges: [][]int{{1, 1},{1, 1},{1, 1},{1, 1},{1, 1},{1, 1},{1, 1},{1, 1},},};var b10 = sequenceBuilder{id: 10, commit: 74,allChars: true,ranges: [][]int{{1, 1},{1, 1},},};var b9 = charBuilder{};b10.items = []builder{&b9,};b86.items = []builder{&b6,&b10,&b6,&b85,};b87.items = []builder{&b85,&b86,};var b15 = sequenceBuilder{id: 15, commit: 66,ranges: [][]int{{0, 1},},};var b14 = sequenceBuilder{id: 14, commit: 10,allChars: true,ranges: [][]int{{1, 1},{1, 1},},};var b13 = charBuilder{};b14.items = []builder{&b13,};b15.items = []builder{&b14,};var b106 = choiceBuilder{id: 106, commit: 66,};var b88 = choiceBuilder{id: 88, commit: 66,};var b54 = choiceBuilder{id: 54, commit: 64,name: "int",};var b45 = sequenceBuilder{id: 45, commit: 74,ranges: [][]int{{1, 1},{0, -1},{1, 1},{0, -1},},};var b44 = sequenceBuilder{id: 44, commit: 10,allChars: true,ranges: [][]int{{1, 1},{1, 1},},};var b43 = charBuilder{};b44.items = []builder{&b43,};var b38 = sequenceBuilder{id: 38, commit: 66,allChars: true,ranges: [][]int{{1, 1},},};var b37 = charBuilder{};b38.items = []builder{&b37,};b45.items = []builder{&b44,&b38,};var b48 = sequenceBuilder{id: 48, commit: 74,ranges: [][]int{{1, 1},{1, -1},{1, 1},{1, -1},},};var b47 = sequenceBuilder{id: 47, commit: 10,allChars: true,ranges: [][]int{{1, 1},{1, 1},},};var b46 = charBuilder{};b47.items = []builder{&b46,};var b40 = sequenceBuilder{id: 40, commit: 66,allChars: true,ranges: [][]int{{1, 1},},};var b39 = charBuilder{};b40.items = []builder{&b39,};b48.items = []builder{&b47,&b40,};var b53 = sequenceBuilder{id: 53, commit: 74,ranges: [][]int{{1, 1},{1, 1},{1, -1},{1, 1},{1, 1},{1, -1},},};var b50 = sequenceBuilder{id: 50, commit: 10,allChars: true,ranges: [][]int{{1, 1},{1, 1},},};var b49 = charBuilder{};b50.items = []builder{&b49,};var b52 = sequenceBuilder{id: 52, commit: 10,allChars: true,ranges: [][]int{{1, 1},{1, 1},},};var b51 = charBuilder{};b52.items = []builder{&b51,};var b42 = sequenceBuilder{id: 42, commit: 66,allChars: true,ranges: [][]int{{1, 1},},};var b41 = charBuilder{};b42.items = []builder{&b41,};b53.items = []builder{&b50,&b52,&b42,};b54.options = []builder{&b45,&b48,&b53,};var b67 = choiceBuilder{id: 67, commit: 72,name: "float",};var b62 = sequenceBuilder{id: 62, commit: 10,ranges: [][]int{{1, -1},{1, 1},{0, -1},{0, 1},{1, -1},{1, 1},{0, -1},{0, 1},},};var b61 = sequenceBuilder{id: 61, commit: 10,allChars: true,ranges: [][]int{{1, 1},{1, 1},},};var b60 = charBuilder{};b61.items = []builder{&b60,};var b59 = sequenceBuilder{id: 59, commit: 74,ranges: [][]int{{1, 1},{0, 1},{1, -1},{1, 1},{0, 1},{1, -1},},};var b56 = sequenceBuilder{id: 56, commit: 10,allChars: true,ranges: [][]int{{1, 1},{1, 1},},};var b55 = charBuilder{};b56.items = []builder{&b55,};var b58 = sequenceBuilder{id: 58, commit: 10,allChars: true,ranges: [][]int{{1, 1},{1, 1},},};var b57 = charBuilder{};b58.items = []builder{&b57,};b59.items = []builder{&b56,&b58,&b38,};b62.items = []builder{&b38,&b61,&b38,&b59,};var b65 = sequenceBuilder{id: 65, commit: 10,ranges: [][]int{{1, 1},{1, -1},{0, 1},{1, 1},{1, -1},{0, 1},},};var b64 = sequenceBuilder{id: 64, commit: 10,allChars: true,ranges: [][]int{{1, 1},{1, 1},},};var b63 = charBuilder{};b64.items = []builder{&b63,};b65.items = []builder{&b64,&b38,&b59,};var b66 = sequenceBuilder{id: 66, commit: 10,ranges: [][]int{{1, -1},{1, 1},{1, -1},{1, 1},},};b66.items = []builder{&b38,&b59,};b67.options = []builder{&b62,&b65,&b66,};var b80 = sequenceBuilder{id: 80, commit: 72,name: "string",ranges: [][]int{{1, 1},{0, -1},{1, 1},{1, 1},{0, -1},{1, 1},},};var b69 = sequenceBuilder{id: 69, commit: 10,allChars: true,ranges: [][]int{{1, 1},{1, 1},},};var b68 = charBuilder{};b69.items = []builder{&b68,};var b77 = choiceBuilder{id: 77, commit: 10,};var b71 = sequenceBuilder{id: 71, commit: 10,allChars: true,ranges: [][]int{{1, 1},{1, 1},},};var b70 = charBuilder{};b71.items = []builder{&b70,};var b76 = sequenceBuilder{id: 76, commit: 10,ranges: [][]int{{1, 1},{1, 1},{1, 1},{1, 1},},};var b73 = sequenceBuilder{id: 73, commit: 10,allChars: true,ranges: [][]int{{1, 1},{1, 1},},};var b72 = charBuilder{};b73.items = []builder{&b72,};var b75 = sequenceBuilder{id: 75, commit: 10,allChars: true,ranges: [][]int{{1, 1},{1, 1},},};var b74 = charBuilder{};b75.items = []builder{&b74,};b76.items = []builder{&b73,&b75,};b77.options = []builder{&b71,&b76,};var b79 = sequenceBuilder{id: 79, commit: 10,allChars: true,ranges: [][]int{{1, 1},{1, 1},},};var b78 = charBuilder{};b79.items = []builder{&b78,};b80.items = []builder{&b69,&b77,&b79,};b88.options = []builder{&b54,&b67,&b80,&b87,};var b105 = sequenceBuilder{id: 105, commit: 64,name: "composite-expression",ranges: [][]int{{1, 1},{0, -1},{1, 1},{0, -1},{1, 1},{0, 1},{0, -1},{0, 1},{0, -1},{1, 1},},};var b90 = sequenceBuilder{id: 90, commit: 10,allChars: true,ranges: [][]int{{1, 1},{1, 1},},};var b89 = charBuilder{};b90.items = []builder{&b89,};var b104 = sequenceBuilder{id: 104, commit: 2,ranges: [][]int{{0, -1},{1, 1},{0, -1},},};var b91 = choiceBuilder{id: 91, commit: 2,};var b12 = sequenceBuilder{id: 12, commit: 74,allChars: true,ranges: [][]int{{1, 1},{1, 1},},};var b11 = charBuilder{};b12.items = []builder{&b11,};b91.options = []builder{&b3,&b12,};var b103 = sequenceBuilder{id: 103, commit: 2,ranges: [][]int{{0, -1},{1, 1},},};b103.items = []builder{&b133,&b91,};b104.items = []builder{&b133,&b91,&b103,};var b100 = sequenceBuilder{id: 100, commit: 2,ranges: [][]int{{1, 1},{0, 1},{0, 1},},};var b97 = sequenceBuilder{id: 97, commit: 2,ranges: [][]int{{0, -1},{1, 1},{0, -1},},};var b94 = sequenceBuilder{id: 94, commit: 2,ranges: [][]int{{1, 1},{0, -1},{0, -1},{1, 1},},};var b92 = choiceBuilder{id: 92, commit: 2,};b92.options = []builder{&b3,&b12,};var b93 = sequenceBuilder{id: 93, commit: 2,ranges: [][]int{{0, -1},{1, 1},},};b93.items = []builder{&b133,&b92,};b94.items = []builder{&b92,&b93,&b133,&b106,};var b96 = sequenceBuilder{id: 96, commit: 2,ranges: [][]int{{0, -1},{1, 1},},};b96.items = []builder{&b133,&b94,};b97.items = []builder{&b133,&b94,&b96,};var b99 = sequenceBuilder{id: 99, commit: 2,ranges: [][]int{{0, -1},{1, 1},{0, -1},},};var b95 = choiceBuilder{id: 95, commit: 2,};b95.options = []builder{&b3,&b12,};var b98 = sequenceBuilder{id: 98, commit: 2,ranges: [][]int{{0, -1},{1, 1},},};b98.items = []builder{&b133,&b95,};b99.items = []builder{&b133,&b95,&b98,};b100.items = []builder{&b106,&b97,&b99,};var b102 = sequenceBuilder{id: 102, commit: 10,allChars: true,ranges: [][]int{{1, 1},{1, 1},},};var b101 = charBuilder{};b102.items = []builder{&b101,};b105.items = []builder{&b87,&b133,&b6,&b133,&b90,&b104,&b133,&b100,&b133,&b102,};b106.options = []builder{&b88,&b105,};b111.items = []builder{&b110,&b133,&b6,&b133,&b87,&b133,&b6,&b133,&b15,&b133,&b6,&b133,&b106,};var b119 = sequenceBuilder{id: 119, commit: 64,name: "export",ranges: [][]int{{1, 1},{0, -1},{1, 1},{0, -1},{1, 1},{0, -1},{1, 1},{0, -1},{1, 1},{0, -1},{1, 1},{0, -1},{1, 1},},};var b118 = sequenceBuilder{id: 118, commit: 10,allChars: true,ranges: [][]int{{1, 1},{1, 1},{1, 1},{1, 1},{1, 1},{1, 1},{1, 1},{1, 1},{1, 1},{1, 1},{1, 1},{1, 1},},};var b112 = charBuilder{};var b113 = charBuilder{};var b114 = charBuilder{};var b115 = charBuilder{};var b116 = charBuilder{};var b117 = charBuilder{};b118.items = []builder{&b112,&b113,&b114,&b115,&b116,&b117,};b119.items = []builder{&b118,&b133,&b6,&b133,&b80,&b133,&b6,&b133,&b15,&b133,&b6,&b133,&b106,};b121.options = []builder{&b111,&b119,};var b124 = sequenceBuilder{id: 124, commit: 2,ranges: [][]int{{0, -1},{1, 1},{0, -1},},};var b122 = choiceBuilder{id: 122, commit: 2,};b122.options = []builder{&b3,&b8,};var b123 = sequenceBuilder{id: 123, commit: 2,ranges: [][]int{{0, -1},{1, 1},},};b123.items = []builder{&b133,&b122,};b124.items = []builder{&b133,&b122,&b123,};b125.items = []builder{&b121,&b124,};var b128 = sequenceBuilder{id: 128, commit: 2,ranges: [][]int{{0, -1},{1, 1},},};b128.items = []builder{&b133,&b125,};b129.items = []builder{&b133,&b125,&b128,};b130.items = []builder{&b127,&b129,};b134.items = []builder{&b130,};b135.items = []builder{&b133,&b134,&b133,};

return parseInput(r, &p135, &b135)
}
