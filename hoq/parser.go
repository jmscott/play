//line parser.y:6
package main

import __yyfmt__ "fmt"

//line parser.y:6
import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"unicode"
)

//line parser.y:20
type yySymType struct {
	yys int
	string
	uint64

	//  unix command execed by hoq
	command *command

	//  abstract syntax tree
	ast *ast
}

const __MIN_YYTOK = 57346
const COMMAND = 57347
const COMMAND_REF = 57348
const PATH = 57349
const CALL = 57350
const WHEN = 57351
const NAME = 57352
const STRING = 57353
const PARSE_ERROR = 57354
const EQ = 57355
const NEQ = 57356
const RE_MATCH = 57357
const RE_NMATCH = 57358
const DOLLAR = 57359
const UINT64 = 57360
const AND = 57361
const OR = 57362
const NOT = 57363
const ARGV = 57364
const ARGV0 = 57365

var yyToknames = [...]string{
	"$end",
	"error",
	"$unk",
	"__MIN_YYTOK",
	"COMMAND",
	"COMMAND_REF",
	"PATH",
	"CALL",
	"WHEN",
	"NAME",
	"STRING",
	"PARSE_ERROR",
	"EQ",
	"NEQ",
	"RE_MATCH",
	"RE_NMATCH",
	"DOLLAR",
	"UINT64",
	"AND",
	"OR",
	"NOT",
	"ARGV",
	"ARGV0",
	"'$'",
	"','",
	"'('",
	"')'",
	"'{'",
	"'='",
	"';'",
	"'}'",
}
var yyStatenames = [...]string{}

const yyEofCode = 1
const yyErrCode = 2
const yyInitialStackSize = 16

//line parser.y:249
var keyword = map[string]int{
	"command": COMMAND,
	"path":    PATH,
	"call":    CALL,
	"when":    WHEN,
	"or":      OR,
	"and":     AND,
	"not":     NOT,
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

func (l *yyLexState) scan_uint64(yylval *yySymType, c rune) (err error) {
	var eof bool

	ui64 := string(c)
	count := 1

	/*
	 *  Scan a string of unicode numbers/digits and let Scanf parse the
	 *  actual digit string.
	 */
	for c, eof, err = l.get(); !eof && err == nil; c, eof, err = l.get() {
		count++
		if count > 20 {
			return l.mkerror("uint64 > 20 digits")
		}
		if c > 127 || !unicode.IsNumber(c) {
			break
		}
		ui64 += string(c)
	}
	if err != nil {
		return
	}
	if !eof {
		l.pushback(c) //  first character after ui64
	}

	yylval.uint64, err = strconv.ParseUint(ui64, 10, 64)
	return
}

/*
 *  Words are a leading ascii or '_' followed by 0 or more ascii letters, digits
 *  and '_' characters.  The word is mapped onto either a keyword or
 *  the NAME token.  When the word is a NAME, then the 'string' field points
 *  the the actual name.
 */
func (l *yyLexState) scan_word(yylval *yySymType, c rune) (tok int, err error) {
	var eof bool

	w := string(c)
	count := 1

	/*
	 *  Scan a string of ascii letters, numbers/digits and '_' character.
	 */
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
	if !eof {
		l.pushback(c) /* first character after word */
	}

	//  keyword?
	if keyword[w] > 0 {
		return keyword[w], nil
	}

	//  command reference?
	if l.commands[w] != nil {
		yylval.command = l.commands[w]
		return COMMAND_REF, nil
	}

	yylval.string = w
	return NAME, nil
}

//  simple utf8 string scanning and with trivial character escaping.
//  this string scan is not compatible with the golang string

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
		err = l.scan_uint64(yylval, c)
		if err != nil {
			goto PARSE_ERROR
		}
		return UINT64
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

	//  peek ahead for ==

	if c == '=' {
		tok, err = lookahead(l, '=', EQ, int('='))
		if err != nil {
			goto PARSE_ERROR
		}
		return tok
	}

	//  peak ahead for not equals (!=) or not matches regular expression

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

	//  peak ahead for regular expression match
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

func parse() (ast *ast, err error) {

	l := &yyLexState{
		line_no:  1,
		in:       bufio.NewReader(os.Stdin),
		commands: make(map[string]*command),
	}

	yyParse(l)

	return l.ast_head, l.err
}

//line yacctab:1
var yyExca = [...]int{
	-1, 1,
	1, -1,
	-2, 0,
}

const yyNprod = 22
const yyPrivate = 57344

var yyTokenNames []string
var yyStates []string

const yyLast = 50

var yyAct = [...]int{

	30, 26, 31, 25, 24, 33, 32, 15, 16, 8,
	13, 17, 15, 42, 9, 18, 19, 28, 20, 23,
	14, 6, 29, 33, 32, 14, 36, 37, 38, 39,
	34, 35, 22, 3, 40, 41, 4, 43, 44, 45,
	46, 10, 7, 2, 11, 5, 12, 27, 21, 1,
}
var yyPact = [...]int{

	28, 28, -1000, 11, 36, -1000, -19, -12, 34, 1,
	-21, -16, -10, -1000, -2, -1000, 7, 23, 1, -1000,
	-26, -27, -4, -1000, -29, -1000, 4, -1000, -4, -4,
	13, -1000, -4, -4, 4, -14, 1, 1, 1, 1,
	4, 4, -1000, -1000, -1000, -1000, -1000,
}
var yyPgo = [...]int{

	0, 43, 49, 48, 47, 1, 0, 46, 44,
}
var yyR1 = [...]int{

	0, 2, 2, 6, 6, 7, 7, 8, 8, 4,
	4, 4, 4, 5, 5, 5, 5, 5, 3, 3,
	1, 1,
}
var yyR2 = [...]int{

	0, 1, 2, 2, 1, 1, 3, 0, 1, 3,
	3, 3, 3, 1, 3, 3, 2, 3, 0, 2,
	8, 7,
}
var yyChk = [...]int{

	-1000, -2, -1, 5, 8, -1, 10, 6, 28, 26,
	7, -8, -7, -6, 24, 11, 29, 27, 25, 18,
	11, -3, 9, -6, 30, 30, -5, -4, 21, 26,
	-6, 31, 20, 19, -5, -5, 13, 14, 15, 16,
	-5, -5, 27, -6, -6, -6, -6,
}
var yyDef = [...]int{

	0, -2, 1, 0, 0, 2, 0, 0, 0, 7,
	0, 0, 8, 5, 0, 4, 0, 18, 0, 3,
	0, 0, 0, 6, 0, 21, 19, 13, 0, 0,
	0, 20, 0, 0, 16, 0, 0, 0, 0, 0,
	14, 15, 17, 9, 10, 11, 12,
}
var yyTok1 = [...]int{

	1, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 24, 3, 3, 3,
	26, 27, 3, 3, 25, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 30,
	3, 29, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 28, 3, 31,
}
var yyTok2 = [...]int{

	2, 3, 4, 5, 6, 7, 8, 9, 10, 11,
	12, 13, 14, 15, 16, 17, 18, 19, 20, 21,
	22, 23,
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
		//line parser.y:63
		{
			yylex.(*yyLexState).ast_head = yyDollar[1].ast
		}
	case 2:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line parser.y:68
		{
			s := yyDollar[1].ast

			//  linearly find the last statement

			for ; s.next != nil; s = s.next {
			}

			s.next = yyDollar[2].ast
		}
	case 3:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line parser.y:81
		{
			yyVAL.ast = &ast{
				yy_tok: DOLLAR,
				uint64: yyDollar[2].uint64,
			}
		}
	case 4:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line parser.y:89
		{
			yyVAL.ast = &ast{
				yy_tok: STRING,
				string: yyDollar[1].string,
			}
		}
	case 6:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line parser.y:101
		{
			s := yyDollar[1].ast

			//  linearly find the last statement

			for ; s.next != nil; s = s.next {
			}

			s.next = yyDollar[3].ast
		}
	case 7:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line parser.y:114
		{
			yyVAL.ast = &ast{
				yy_tok: ARGV0,
			}
		}
	case 8:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line parser.y:121
		{
			yyVAL.ast = &ast{
				yy_tok: ARGV,
				left:   yyDollar[1].ast,
			}
		}
	case 9:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line parser.y:131
		{
			yyVAL.ast = &ast{
				yy_tok: EQ,
				left:   yyDollar[1].ast,
				right:  yyDollar[3].ast,
			}
		}
	case 10:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line parser.y:140
		{
			yyVAL.ast = &ast{
				yy_tok: NEQ,
				left:   yyDollar[1].ast,
				right:  yyDollar[3].ast,
			}
		}
	case 11:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line parser.y:149
		{
			yyVAL.ast = &ast{
				yy_tok: RE_MATCH,
				left:   yyDollar[1].ast,
				right:  yyDollar[3].ast,
			}
		}
	case 12:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line parser.y:158
		{
			yyVAL.ast = &ast{
				yy_tok: RE_NMATCH,
				left:   yyDollar[1].ast,
				right:  yyDollar[3].ast,
			}
		}
	case 14:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line parser.y:171
		{
			yyVAL.ast = &ast{
				yy_tok: OR,
				left:   yyDollar[1].ast,
				right:  yyDollar[3].ast,
			}
		}
	case 15:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line parser.y:180
		{
			yyVAL.ast = &ast{
				yy_tok: AND,
				left:   yyDollar[1].ast,
				right:  yyDollar[3].ast,
			}
		}
	case 16:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line parser.y:189
		{
			yyVAL.ast = &ast{
				yy_tok: NOT,
				left:   yyDollar[2].ast,
			}
		}
	case 17:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line parser.y:197
		{
			yyVAL.ast = yyDollar[2].ast
		}
	case 18:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line parser.y:204
		{
			yyVAL.ast = nil
		}
	case 19:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line parser.y:209
		{
			yyVAL.ast = &ast{
				yy_tok: WHEN,
				right:  yyDollar[2].ast,
			}
		}
	case 20:
		yyDollar = yyS[yypt-8 : yypt+1]
		//line parser.y:221
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
	case 21:
		yyDollar = yyS[yypt-7 : yypt+1]
		//line parser.y:240
		{
			yyVAL.ast = &ast{
				yy_tok:  CALL,
				command: yyDollar[2].command,
				left:    yyDollar[4].ast,
				right:   yyDollar[6].ast,
			}
		}
	}
	goto yystack /* stack new state and value */
}
