//line parser.y:6
package main

import __yyfmt__ "fmt"

//line parser.y:6
import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"unicode"
)

func init() {

	//  sanity test:
	//
	//  token __MIN_YYTOK is starting index into yacc generated token name
	//  table

	if yyToknames[3] != "__MIN_YYTOK" {
		panic("yyToknames[3] != __MIN_YYTOK: yacc may have changed")
	}
}

//line parser.y:32
type yySymType struct {
	yys int
	string
	uint8
	token int

	//  unix command execed by hoq
	command *command

	//  abstract syntax tree
	ast *ast
}

const __MIN_YYTOK = 57346
const COMMAND = 57347
const XCOMMAND = 57348
const EXIT_STATUS = 57349
const PATH = 57350
const CALL = 57351
const WHEN = 57352
const NAME = 57353
const STRING = 57354
const PARSE_ERROR = 57355
const EQ = 57356
const EQ_UINT8 = 57357
const EQ_STRING = 57358
const EQ_BOOL = 57359
const NEQ = 57360
const NEQ_UINT8 = 57361
const NEQ_STRING = 57362
const NEQ_BOOL = 57363
const DOLLAR = 57364
const UINT8 = 57365
const ARGV = 57366
const ARGV0 = 57367
const ARGV1 = 57368
const TO_STRING_UINT8 = 57369
const RE_MATCH = 57370
const RE_NMATCH = 57371
const NOT = 57372
const TRUE = 57373
const FALSE = 57374
const AND = 57375
const OR = 57376

var yyToknames = [...]string{
	"$end",
	"error",
	"$unk",
	"__MIN_YYTOK",
	"COMMAND",
	"XCOMMAND",
	"EXIT_STATUS",
	"PATH",
	"CALL",
	"WHEN",
	"NAME",
	"STRING",
	"PARSE_ERROR",
	"EQ",
	"EQ_UINT8",
	"EQ_STRING",
	"EQ_BOOL",
	"NEQ",
	"NEQ_UINT8",
	"NEQ_STRING",
	"NEQ_BOOL",
	"DOLLAR",
	"UINT8",
	"ARGV",
	"ARGV0",
	"ARGV1",
	"TO_STRING_UINT8",
	"RE_MATCH",
	"RE_NMATCH",
	"NOT",
	"TRUE",
	"FALSE",
	"AND",
	"OR",
	"'$'",
	"'.'",
	"'('",
	"')'",
	"','",
	"'{'",
	"'='",
	"';'",
	"'}'",
}
var yyStatenames = [...]string{}

const yyEofCode = 1
const yyErrCode = 2
const yyInitialStackSize = 16

//line parser.y:354
var keyword = map[string]int{
	"and":         AND,
	"call":        CALL,
	"command":     COMMAND,
	"exit_status": EXIT_STATUS,
	"false":       FALSE,
	"not":         NOT,
	"or":          OR,
	"path":        PATH,
	"true":        TRUE,
	"when":        WHEN,
}

type yyLexState struct {
	//  source code stream
	in io.RuneReader //  config source stream

	//  line number in source stream
	line_no uint64 //  lexical line number

	//  at end of stream
	eof bool //  seen eof in token stream

	//  lookahead on character
	peek rune //  lookahead in lexer

	//  error during parsing
	err error

	//  first statement in parse tree
	ast_head *ast

	//  track declared commands

	commands map[string]*command

	//  track called commands

	called map[string]bool

	//  track depends list used by tsort to build DAG of
	//  call relationships.

	depends []string

	call *command
}

func (l *yyLexState) pushback(c rune) {

	if l.peek != 0 {
		panic("pushback(): push before peek") /* impossible */
	}
	l.peek = c
	if c == '\n' {
		l.line_no--
	}
}

func (l *yyLexState) node(
	yy_tok int,
	go_type reflect.Kind,
	left, right, next *ast,
) *ast {

	return &ast{
		yy_tok:  yy_tok,
		go_type: go_type,
		left:    left,
		right:   right,
		next:    next,
		line_no: l.line_no,
	}
}

func (l *yyLexState) op_name(yy_tok int) {
	switch yy_tok {

	}
}

func (l *yyLexState) bool_node(yy_tok int, left, right *ast) *ast {

	if left != nil && right != nil && left.go_type != right.go_type {
		l.error("operator %s: type mismatch: %s != %s",
			yyToknames[yy_tok-__MIN_YYTOK+3],
			left.go_type,
			right.go_type,
		)
		return nil
	}
	return l.node(yy_tok, reflect.Bool, left, right, nil)
}

func (l *yyLexState) scalar_node(yy_tok int, go_type reflect.Kind) *ast {

	return l.node(yy_tok, go_type, nil, nil, nil)
}

/*
 *  Read next UTF8 rune.
 */
func (l *yyLexState) get() (c rune, eof bool, err error) {

	if l.eof {
		return 0, true, nil
	}

	//  if we peeked ahead one char then return that char

	if l.peek != 0 {
		c = l.peek

		//  only push back one character.

		l.peek = 0
		if c == '\n' {
			l.line_no++
		}
		return c, false, nil
	}

	//  read the next character as a rune

	c, _, err = l.in.ReadRune()
	if err != nil {
		if err == io.EOF {
			l.eof = true
			return 0, true, nil
		}
		return 0, false, err
	}

	//  grumble about invalid code points

	if c == unicode.ReplacementChar {
		return 0, false, l.mkerror("get: invalid unicode sequence")
	}

	if c == '\n' {
		l.line_no++
	}

	return c, false, nil
}

func lookahead(l *yyLexState, peek rune, ifyes int, ifno int) (int, error) {

	next, eof, err := l.get()
	if err != nil {
		return 0, err
	}
	if next == peek {
		return ifyes, err
	}
	if !eof {
		l.pushback(next)
	}
	return ifno, nil
}

/*
 *  Skip '#' comment.
 *  The scan stops on the terminating newline or error
 */
func skip_comment(l *yyLexState) (err error) {
	var c rune
	var eof bool

	/*
	 *  Scan for newline, end of file, or error.
	 */
	for c, eof, err = l.get(); !eof && err == nil; c, eof, err = l.get() {
		if c == '\n' {
			return
		}
	}
	return err

}

func skip_space(l *yyLexState) (c rune, eof bool, err error) {

	for c, eof, err = l.get(); !eof && err == nil; c, eof, err = l.get() {
		if unicode.IsSpace(c) {
			continue
		}
		if c != '#' {
			return c, false, nil
		}

		/*
		 *  Skipping over # comment terminated by newline or EOF
		 */
		err = skip_comment(l)
		if err != nil {
			return 0, false, err
		}
	}
	return 0, eof, err
}

func (l *yyLexState) scan_uint8(yylval *yySymType, c rune) (err error) {
	var eof bool

	ui8 := string(c)
	count := 1

	/*
	 *  Scan a string of unicode numbers/digits and let Scanf parse the
	 *  actual digit string.
	 */
	for c, eof, err = l.get(); !eof && err == nil; c, eof, err = l.get() {
		count++
		if count > 20 {
			return l.mkerror("uint8 > 20 digits")
		}
		if c > 127 || !unicode.IsNumber(c) {
			break
		}
		ui8 += string(c)
	}
	if err != nil {
		return
	}
	if !eof {
		l.pushback(c) //  first character after ui8
	}

	var ui64 uint64
	ui64, err = strconv.ParseUint(ui8, 10, 8)

	if err == nil {
		if ui64 > 255 {
			err = errors.New(fmt.Sprintf("uint8 > 255: %d", ui64))
		} else {
			yylval.uint8 = uint8(ui64)
		}
	}
	return
}

/*
 *  scan a word from the input stream.
 *
 *  words have a leading ascii or '_' followed by 0 or more ascii letters,
 *  digits or '_' characters.  the word is mapped onto either a keyword, a
 *  command or the NAME token.  when the word is a NAME, then the 'string'
 *  field points the the actual name of the word.
 */
func (l *yyLexState) scan_word(yylval *yySymType, c rune) (tok int, err error) {
	var eof bool

	w := string(c)
	count := 1

	//  Scan a string of ascii letters, numbers/digits and '_' character.

	for c, eof, err = l.get(); !eof && err == nil; c, eof, err = l.get() {
		if c > 127 ||
			(c != '_' &&
				!unicode.IsLetter(c) &&
				!unicode.IsNumber(c)) {
			break
		}
		count++
		if count > 128 {
			return 0, l.mkerror("name too many characters: max=128")
		}
		w += string(c)
	}
	if err != nil {
		return 0, err
	}

	//  pushback the first character after the end of the word

	if !eof {
		l.pushback(c)
	}

	//  language keyword?

	if keyword[w] > 0 {
		return keyword[w], nil
	}

	//  an executed command reference?

	if l.commands[w] != nil {
		yylval.command = l.commands[w]
		return XCOMMAND, nil
	}

	yylval.string = w
	return NAME, nil
}

//  simple utf8 string scanning with trivial character escaping.
//  this string scan is not compatible with the golang string.

func (l *yyLexState) scan_string(yylval *yySymType) (eof bool, err error) {
	var c rune
	s := ""

	for c, eof, err = l.get(); !eof && err == nil; c, eof, err = l.get() {

		//  double quotes always clsoe the string
		if c == '"' {
			yylval.string = s
			return false, nil
		}

		//  no new-line, carriage return, tab or slosh in string

		switch c {
		case '\n':
			return false, l.mkerror("new line in string")
		case '\r':
			return false, l.mkerror("carriage return in string")
		case '\t':
			return false, l.mkerror("tab in string")
		case '\\':
			return false, l.mkerror("backslash in string")
		}
		s += string(c)
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

//  Lex() called by automatically generated yacc go code

func (l *yyLexState) Lex(yylval *yySymType) (tok int) {

	if l.err != nil {
		return PARSE_ERROR
	}
	if l.eof {
		return 0
	}
	c, eof, err := skip_space(l)
	if err != nil {
		goto PARSE_ERROR
	}
	if eof {
		return 0
	}

	//  ascii outside of strings, for time being
	if c > 127 {
		goto PARSE_ERROR
	}

	//  scan a word

	if (unicode.IsLetter(c)) || c == '_' {
		tok, err = l.scan_word(yylval, c)
		if err != nil {
			goto PARSE_ERROR
		}
		return tok
	}

	//  scan an unsigned int 64

	if unicode.IsNumber(c) {
		err = l.scan_uint8(yylval, c)
		if err != nil {
			goto PARSE_ERROR
		}
		return UINT8
	}

	//  scan a string

	if c == '"' {
		lno := l.line_no // reset line number on error

		eof, err = l.scan_string(yylval)
		if err != nil {
			goto PARSE_ERROR
		}
		if eof {
			l.line_no = lno
			err = l.mkerror("unexpected end of file in string")
			goto PARSE_ERROR
		}
		return STRING
	}

	//  peek ahead for '==' or '-'

	if c == '=' {
		tok, err = lookahead(l, '=', EQ, int('='))
		if err != nil {
			goto PARSE_ERROR
		}
		return tok
	}

	//  peak ahead for not equals, '!=' or not matches regular expression,
	//  '!~'

	if c == '!' {
		tok, err = lookahead(l, '=', NEQ, int('!'))
		if err != nil {
			goto PARSE_ERROR
		}
		if tok == NEQ {
			return NEQ
		}
		tok, err = lookahead(l, '~', RE_NMATCH, int('!'))
		if err != nil {
			goto PARSE_ERROR
		}
		return tok
	}

	//  peak ahead for regular expression match '~~';  otherwise '~'

	if c == '~' {
		tok, err = lookahead(l, '~', RE_MATCH, int('~'))
		if err != nil {
			goto PARSE_ERROR
		}
		return tok
	}

	return int(c)

PARSE_ERROR:
	l.err = err
	return PARSE_ERROR
}

func (l *yyLexState) mkerror(format string, args ...interface{}) error {

	return errors.New(fmt.Sprintf("%s near line %d",
		fmt.Sprintf(format, args...),
		l.line_no,
	))
}

func (l *yyLexState) error(format string, args ...interface{}) {

	l.Error(fmt.Sprintf(format, args...))
}

func (l *yyLexState) Error(msg string) {

	if l.err == nil { //  only report first error
		l.err = l.mkerror("%s", msg)
	}
}

//  enter the yacc dragon

func parse(in io.Reader) (_ *ast, depend_order []string, err error) {

	l := &yyLexState{
		line_no:  1,
		in:       bufio.NewReader(in),
		commands: make(map[string]*command),
		called:   make(map[string]bool),
	}

	yyParse(l)

	//  added unreferenced calls() to dependency list

	var find_unreferenced_CALL func(a *ast)

	find_unreferenced_CALL = func(a *ast) {

		if a == nil {
			return
		}
		if a.yy_tok == CALL && a.command.depend_ref_count == 0 {
			n := a.command.name
			l.depends = append(l.depends, fmt.Sprintf("%s %s", n, n))
		}
		find_unreferenced_CALL(a.left)
		find_unreferenced_CALL(a.right)
		find_unreferenced_CALL(a.next)
	}
	find_unreferenced_CALL(l.ast_head)

	return l.ast_head, tsort(l.depends), l.err
}

//line yacctab:1
var yyExca = [...]int{
	-1, 1,
	1, -1,
	-2, 0,
}

const yyNprod = 26
const yyPrivate = 57344

var yyTokenNames []string
var yyStates []string

const yyLast = 65

var yyAct = [...]int{

	15, 49, 31, 50, 37, 12, 32, 8, 26, 25,
	11, 34, 33, 24, 6, 39, 27, 28, 10, 21,
	7, 29, 30, 35, 36, 18, 48, 40, 41, 42,
	43, 44, 45, 46, 3, 47, 19, 31, 4, 9,
	51, 32, 13, 22, 16, 17, 14, 38, 20, 31,
	23, 27, 28, 32, 1, 0, 29, 30, 2, 0,
	5, 0, 0, 27, 28,
}
var yyPact = [...]int{

	29, 29, -1000, 3, 14, -1000, -33, -1000, 10, -27,
	-36, 13, 1, -29, -31, 23, -1000, -1000, -1000, -1000,
	-11, -25, 13, 13, -38, 5, 13, 13, 13, 13,
	13, 13, 13, -1000, 28, -1000, -12, -42, -39, 13,
	23, -1000, -1000, 35, 35, -1000, -1000, -1000, -1000, -1000,
	-1000, 23,
}
var yyPgo = [...]int{

	0, 58, 54, 47, 0, 46, 42, 39,
}
var yyR1 = [...]int{

	0, 2, 2, 4, 4, 4, 4, 4, 4, 4,
	4, 4, 4, 4, 4, 4, 4, 5, 5, 6,
	6, 3, 3, 1, 7, 1,
}
var yyR2 = [...]int{

	0, 1, 2, 1, 1, 1, 1, 2, 3, 3,
	3, 3, 3, 3, 3, 2, 3, 1, 3, 0,
	1, 0, 2, 8, 0, 8,
}
var yyChk = [...]int{

	-1000, -2, -1, 5, 9, -1, 11, 6, 40, -7,
	8, 37, 41, -6, -5, -4, 31, 32, 12, 23,
	35, 6, 30, 37, 12, 38, 39, 28, 29, 33,
	34, 14, 18, 23, 36, -4, -4, 42, -3, 10,
	-4, -4, -4, -4, -4, -4, -4, 7, 38, 43,
	42, -4,
}
var yyDef = [...]int{

	0, -2, 1, 0, 0, 2, 0, 24, 0, 0,
	0, 19, 0, 0, 20, 17, 3, 4, 5, 6,
	0, 0, 0, 0, 0, 21, 0, 0, 0, 0,
	0, 0, 0, 7, 0, 15, 0, 0, 0, 0,
	18, 9, 10, 11, 12, 13, 14, 8, 16, 23,
	25, 22,
}
var yyTok1 = [...]int{

	1, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 35, 3, 3, 3,
	37, 38, 3, 3, 39, 3, 36, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 42,
	3, 41, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 40, 3, 43,
}
var yyTok2 = [...]int{

	2, 3, 4, 5, 6, 7, 8, 9, 10, 11,
	12, 13, 14, 15, 16, 17, 18, 19, 20, 21,
	22, 23, 24, 25, 26, 27, 28, 29, 30, 31,
	32, 33, 34,
}
var yyTok3 = [...]int{
	0,
}

var yyErrorMessages = [...]struct {
	state int
	token int
	msg   string
}{}

//line yaccpar:1

/*	parser for yacc output	*/

var (
	yyDebug        = 0
	yyErrorVerbose = false
)

type yyLexer interface {
	Lex(lval *yySymType) int
	Error(s string)
}

type yyParser interface {
	Parse(yyLexer) int
	Lookahead() int
}

type yyParserImpl struct {
	lval  yySymType
	stack [yyInitialStackSize]yySymType
	char  int
}

func (p *yyParserImpl) Lookahead() int {
	return p.char
}

func yyNewParser() yyParser {
	return &yyParserImpl{}
}

const yyFlag = -1000

func yyTokname(c int) string {
	if c >= 1 && c-1 < len(yyToknames) {
		if yyToknames[c-1] != "" {
			return yyToknames[c-1]
		}
	}
	return __yyfmt__.Sprintf("tok-%v", c)
}

func yyStatname(s int) string {
	if s >= 0 && s < len(yyStatenames) {
		if yyStatenames[s] != "" {
			return yyStatenames[s]
		}
	}
	return __yyfmt__.Sprintf("state-%v", s)
}

func yyErrorMessage(state, lookAhead int) string {
	const TOKSTART = 4

	if !yyErrorVerbose {
		return "syntax error"
	}

	for _, e := range yyErrorMessages {
		if e.state == state && e.token == lookAhead {
			return "syntax error: " + e.msg
		}
	}

	res := "syntax error: unexpected " + yyTokname(lookAhead)

	// To match Bison, suggest at most four expected tokens.
	expected := make([]int, 0, 4)

	// Look for shiftable tokens.
	base := yyPact[state]
	for tok := TOKSTART; tok-1 < len(yyToknames); tok++ {
		if n := base + tok; n >= 0 && n < yyLast && yyChk[yyAct[n]] == tok {
			if len(expected) == cap(expected) {
				return res
			}
			expected = append(expected, tok)
		}
	}

	if yyDef[state] == -2 {
		i := 0
		for yyExca[i] != -1 || yyExca[i+1] != state {
			i += 2
		}

		// Look for tokens that we accept or reduce.
		for i += 2; yyExca[i] >= 0; i += 2 {
			tok := yyExca[i]
			if tok < TOKSTART || yyExca[i+1] == 0 {
				continue
			}
			if len(expected) == cap(expected) {
				return res
			}
			expected = append(expected, tok)
		}

		// If the default action is to accept or reduce, give up.
		if yyExca[i+1] != 0 {
			return res
		}
	}

	for i, tok := range expected {
		if i == 0 {
			res += ", expecting "
		} else {
			res += " or "
		}
		res += yyTokname(tok)
	}
	return res
}

func yylex1(lex yyLexer, lval *yySymType) (char, token int) {
	token = 0
	char = lex.Lex(lval)
	if char <= 0 {
		token = yyTok1[0]
		goto out
	}
	if char < len(yyTok1) {
		token = yyTok1[char]
		goto out
	}
	if char >= yyPrivate {
		if char < yyPrivate+len(yyTok2) {
			token = yyTok2[char-yyPrivate]
			goto out
		}
	}
	for i := 0; i < len(yyTok3); i += 2 {
		token = yyTok3[i+0]
		if token == char {
			token = yyTok3[i+1]
			goto out
		}
	}

out:
	if token == 0 {
		token = yyTok2[1] /* unknown char */
	}
	if yyDebug >= 3 {
		__yyfmt__.Printf("lex %s(%d)\n", yyTokname(token), uint(char))
	}
	return char, token
}

func yyParse(yylex yyLexer) int {
	return yyNewParser().Parse(yylex)
}

func (yyrcvr *yyParserImpl) Parse(yylex yyLexer) int {
	var yyn int
	var yyVAL yySymType
	var yyDollar []yySymType
	_ = yyDollar // silence set and not used
	yyS := yyrcvr.stack[:]

	Nerrs := 0   /* number of errors */
	Errflag := 0 /* error recovery flag */
	yystate := 0
	yyrcvr.char = -1
	yytoken := -1 // yyrcvr.char translated into internal numbering
	defer func() {
		// Make sure we report no lookahead when not parsing.
		yystate = -1
		yyrcvr.char = -1
		yytoken = -1
	}()
	yyp := -1
	goto yystack

ret0:
	return 0

ret1:
	return 1

yystack:
	/* put a state and value onto the stack */
	if yyDebug >= 4 {
		__yyfmt__.Printf("char %v in %v\n", yyTokname(yytoken), yyStatname(yystate))
	}

	yyp++
	if yyp >= len(yyS) {
		nyys := make([]yySymType, len(yyS)*2)
		copy(nyys, yyS)
		yyS = nyys
	}
	yyS[yyp] = yyVAL
	yyS[yyp].yys = yystate

yynewstate:
	yyn = yyPact[yystate]
	if yyn <= yyFlag {
		goto yydefault /* simple state */
	}
	if yyrcvr.char < 0 {
		yyrcvr.char, yytoken = yylex1(yylex, &yyrcvr.lval)
	}
	yyn += yytoken
	if yyn < 0 || yyn >= yyLast {
		goto yydefault
	}
	yyn = yyAct[yyn]
	if yyChk[yyn] == yytoken { /* valid shift */
		yyrcvr.char = -1
		yytoken = -1
		yyVAL = yyrcvr.lval
		yystate = yyn
		if Errflag > 0 {
			Errflag--
		}
		goto yystack
	}

yydefault:
	/* default state action */
	yyn = yyDef[yystate]
	if yyn == -2 {
		if yyrcvr.char < 0 {
			yyrcvr.char, yytoken = yylex1(yylex, &yyrcvr.lval)
		}

		/* look through exception table */
		xi := 0
		for {
			if yyExca[xi+0] == -1 && yyExca[xi+1] == yystate {
				break
			}
			xi += 2
		}
		for xi += 2; ; xi += 2 {
			yyn = yyExca[xi+0]
			if yyn < 0 || yyn == yytoken {
				break
			}
		}
		yyn = yyExca[xi+1]
		if yyn < 0 {
			goto ret0
		}
	}
	if yyn == 0 {
		/* error ... attempt to resume parsing */
		switch Errflag {
		case 0: /* brand new error */
			yylex.Error(yyErrorMessage(yystate, yytoken))
			Nerrs++
			if yyDebug >= 1 {
				__yyfmt__.Printf("%s", yyStatname(yystate))
				__yyfmt__.Printf(" saw %s\n", yyTokname(yytoken))
			}
			fallthrough

		case 1, 2: /* incompletely recovered error ... try again */
			Errflag = 3

			/* find a state where "error" is a legal shift action */
			for yyp >= 0 {
				yyn = yyPact[yyS[yyp].yys] + yyErrCode
				if yyn >= 0 && yyn < yyLast {
					yystate = yyAct[yyn] /* simulate a shift of "error" */
					if yyChk[yystate] == yyErrCode {
						goto yystack
					}
				}

				/* the current p has no shift on "error", pop stack */
				if yyDebug >= 2 {
					__yyfmt__.Printf("error recovery pops state %d\n", yyS[yyp].yys)
				}
				yyp--
			}
			/* there is no state on the stack with an error shift ... abort */
			goto ret1

		case 3: /* no shift yet; clobber input char */
			if yyDebug >= 2 {
				__yyfmt__.Printf("error recovery discards %s\n", yyTokname(yytoken))
			}
			if yytoken == yyEofCode {
				goto ret1
			}
			yyrcvr.char = -1
			yytoken = -1
			goto yynewstate /* try again in the same state */
		}
	}

	/* reduction by production yyn */
	if yyDebug >= 2 {
		__yyfmt__.Printf("reduce %v in:\n\t%v\n", yyn, yyStatname(yystate))
	}

	yynt := yyn
	yypt := yyp
	_ = yypt // guard against "declared and not used"

	yyp -= yyR2[yyn]
	// yyp is now the index of $0. Perform the default action. Iff the
	// reduced production is ε, $1 is possibly out of range.
	if yyp+1 >= len(yyS) {
		nyys := make([]yySymType, len(yyS)*2)
		copy(nyys, yyS)
		yyS = nyys
	}
	yyVAL = yyS[yyp+1]

	/* consult goto table to find next state */
	yyn = yyR1[yyn]
	yyg := yyPgo[yyn]
	yyj := yyg + yyS[yyp].yys + 1

	if yyj >= yyLast {
		yystate = yyAct[yyg]
	} else {
		yystate = yyAct[yyj]
		if yyChk[yystate] != -yyn {
			yystate = yyAct[yyg]
		}
	}
	// dummy call; replaced with literal code
	switch yynt {

	case 1:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line parser.y:82
		{
			yylex.(*yyLexState).ast_head = yyDollar[1].ast
		}
	case 2:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line parser.y:87
		{
			s := yyDollar[1].ast

			//  linearly find the last statement

			for ; s.next != nil; s = s.next {
			}

			s.next = yyDollar[2].ast
		}
	case 3:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line parser.y:100
		{
			yyVAL.ast = yylex.(*yyLexState).scalar_node(TRUE, reflect.Bool)
			yyVAL.ast.bool = true
		}
	case 4:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line parser.y:106
		{
			yyVAL.ast = yylex.(*yyLexState).scalar_node(FALSE, reflect.Bool)
			yyVAL.ast.bool = true
		}
	case 5:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line parser.y:112
		{
			yyVAL.ast = yylex.(*yyLexState).scalar_node(STRING, reflect.String)
			yyVAL.ast.string = yyDollar[1].string
		}
	case 6:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line parser.y:118
		{
			yyVAL.ast = yylex.(*yyLexState).scalar_node(UINT8, reflect.Uint8)
			yyVAL.ast.uint8 = yyDollar[1].uint8
		}
	case 7:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line parser.y:124
		{
			yyVAL.ast = yylex.(*yyLexState).scalar_node(DOLLAR, reflect.String)
			yyVAL.ast.uint8 = yyDollar[2].uint8
		}
	case 8:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line parser.y:130
		{
			l := yylex.(*yyLexState)
			cmd := yyDollar[1].command

			if cmd == l.call {
				l.error("command cannot call itself: %s", cmd.name)
				return 0
			}

			if l.called[cmd.name] == false {
				l.error("command '%s' referenced before call", cmd.name)
				return 0
			}
			if cmd.depend_ref_count == 255 {
				l.error("%s: too many dependencies: > 255", cmd.name)
				return 0
			}
			cmd.depend_ref_count++

			//  record for detection of cycles in the invocation graph

			l.depends = append(
				l.depends,
				fmt.Sprintf("%s %s", l.call.name, yyDollar[1].command.name),
			)

			yyVAL.ast = yylex.(*yyLexState).scalar_node(EXIT_STATUS, reflect.Uint8)
			yyVAL.ast.command = cmd
		}
	case 9:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line parser.y:161
		{
			l := yylex.(*yyLexState)

			yyVAL.ast = l.bool_node(RE_MATCH, yyDollar[1].ast, yyDollar[3].ast)
			if yyVAL.ast == nil {
				return 0
			}

			if yyDollar[1].ast.go_type != reflect.String {
				l.error("~~ operator requires string operands")
				return 0
			}
		}
	case 10:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line parser.y:176
		{
			l := yylex.(*yyLexState)
			yyVAL.ast = l.bool_node(RE_MATCH, yyDollar[1].ast, yyDollar[3].ast)

			if yyVAL.ast == nil {
				return 0
			}

			if yyDollar[1].ast.go_type != reflect.String {
				l.error("!~ operator requires string operands")
				return 0
			}
		}
	case 11:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line parser.y:191
		{
			l := yylex.(*yyLexState)
			yyVAL.ast = yylex.(*yyLexState).bool_node(AND, yyDollar[1].ast, yyDollar[3].ast)
			if yyVAL.ast == nil {
				return 0
			}

			if yyDollar[1].ast.go_type != reflect.Bool {
				l.error("logical and requires boolean operands")
				return 0
			}
		}
	case 12:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line parser.y:205
		{
			l := yylex.(*yyLexState)
			yyVAL.ast = yylex.(*yyLexState).bool_node(OR, yyDollar[1].ast, yyDollar[3].ast)
			if yyVAL.ast == nil {
				return 0
			}

			if yyDollar[1].ast.go_type != reflect.Bool {
				l.error("logical or requires boolean operands")
				return 0
			}
		}
	case 13:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line parser.y:219
		{
			yyVAL.ast = yylex.(*yyLexState).bool_node(EQ, yyDollar[1].ast, yyDollar[3].ast)
			if yyVAL.ast == nil {
				return 0
			}
		}
	case 14:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line parser.y:227
		{
			yyVAL.ast = yylex.(*yyLexState).bool_node(NEQ, yyDollar[1].ast, yyDollar[3].ast)
			if yyVAL.ast == nil {
				return 0
			}
		}
	case 15:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line parser.y:235
		{
			l := yylex.(*yyLexState)
			if yyDollar[2].ast.go_type != reflect.Bool {
				l.error("logical not requires boolean operand")
				return 0
			}
			yyVAL.ast = yylex.(*yyLexState).bool_node(NOT, yyDollar[2].ast, nil)
		}
	case 16:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line parser.y:245
		{
			yyVAL.ast = yyDollar[2].ast
		}
	case 18:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line parser.y:255
		{
			argc := uint16(0)
			ae := yyDollar[1].ast

			//  find the last expression in the list

			for ; ae.next != nil; ae = ae.next {

				if argc >= 255 {
					yylex.(*yyLexState).error(
						"too many expressions in list: > 255",
					)
					return 0
				}
				argc++
			}
			ae.next = yyDollar[3].ast
		}
	case 19:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line parser.y:277
		{
			yyVAL.ast = nil
		}
	case 20:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line parser.y:282
		{
			yyVAL.ast = &ast{
				yy_tok: ARGV,
				left:   yyDollar[1].ast,
			}

			//  count the arguments

			for ae := yyDollar[1].ast; ae != nil; ae = ae.next {
				yyVAL.ast.uint8++
			}
		}
	case 21:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line parser.y:298
		{
			yyVAL.ast = nil
		}
	case 22:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line parser.y:303
		{
			yyVAL.ast = yylex.(*yyLexState).bool_node(WHEN, yyDollar[2].ast, nil)
		}
	case 23:
		yyDollar = yyS[yypt-8 : yypt+1]
		//line parser.y:312
		{
			l := yylex.(*yyLexState)

			if yyDollar[6].string == "" {
				l.error("command %s: path is zero length", yyDollar[2].string)
				return 0
			}

			l.commands[yyDollar[2].string] = &command{
				name: yyDollar[2].string,
				path: yyDollar[6].string,
			}
			yyVAL.ast = &ast{
				yy_tok:  COMMAND,
				command: l.commands[yyDollar[2].string],
			}
		}
	case 24:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line parser.y:331
		{
			//  dependency graph needs command being called

			yylex.(*yyLexState).call = yyDollar[2].command
		}
	case 25:
		yyDollar = yyS[yypt-8 : yypt+1]
		//line parser.y:337
		{
			l := yylex.(*yyLexState)
			n := yyDollar[2].command.name
			if l.called[n] {
				l.error("command '%s' called more than once", n)
				return 0
			}
			l.called[n] = true

			yyVAL.ast = &ast{
				yy_tok:  CALL,
				command: yyDollar[2].command,
				left:    yyDollar[5].ast,
				right:   yyDollar[7].ast,
			}
		}
	}
	goto yystack /* stack new state and value */
}