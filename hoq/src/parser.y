//  Yacc grammar for 'hoq' language.
//  Enter the dragon:
//
//	http://www.amazon.com/Compilers-Principles-Techniques-Tools-2nd/dp/0321486811

%{
package main

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

%}

//  go values associated with seen patterns during parsing

%union {
	string
	uint8
	token	int
	sarray	[]string

	//  unix command execed by hoq

	command		*command

	//  abstract syntax tree

	ast		*ast
}

//  lowest numbered yytoken.  must be first in list.
%token	__MIN_YYTOK

//  tokens are integers, returned by yyLex and stored in nodes of abstract
//  syntax tree.

%token	COMMAND  EXIT_STATUS
%token	PATH
%token	EXEC  WHEN
%token	PARSE_ERROR
%token	EQ  EQ_UINT8  EQ_STRING  EQ_BOOL
%token	NEQ  NEQ_UINT8  NEQ_STRING  NEQ_BOOL
%token	DOLLAR0
%token	ARGV  ARGV0  ARGV1
%token	TO_STRING_UINT8  TO_STRING_BOOL
%token	RE_MATCH  RE_NMATCH
%token	NOT
%token	TRUE  FALSE

%token	<string>	STRING  NAME
%token	<command>	XCOMMAND
%token	<uint8>		UINT8
%token	<ast>		DOLLAR

//  complex patterns seen in input stream produced by yyLex.

%type	<ast>		statement  statement_list
%type	<ast>		qualification 
%type	<ast>		exp  exp_list
%type	<ast>		argv
%type	<sarray>	string_list  command_argv

//  precedence rules for reducing patterns in input stream

%left	AND  OR
%left	EQ  NEQ  RE_MATCH  RE_NMATCH
%right	NOT  '$'

%%

statement_list:
	  statement
	  {
	  	yylex.(*yyLexState).ast_head = $1
	  }
	|
	  statement_list statement
	  {
	  	s := $1

		//  linearly find the last statement

		for ;  s.next != nil;  s = s.next {}

		s.next = $2
	  }
	;

exp:
	  TRUE
	  {
	  	$$ = yylex.(*yyLexState).scalar_node(TRUE, reflect.Bool)
		$$.bool = true
	  }
	|
	  FALSE
	  {
	  	$$ = yylex.(*yyLexState).scalar_node(FALSE, reflect.Bool)
		$$.bool = true
	  }
	|
	  STRING
	  {
	  	$$ = yylex.(*yyLexState).scalar_node(STRING, reflect.String)
		$$.string = $1
	  }
	|
	  UINT8
	  {
	  	$$ = yylex.(*yyLexState).scalar_node(UINT8, reflect.Uint8)
		$$.uint8 = $1
	  }
	|
	  '$'  UINT8
	  {
	  	$$ = yylex.(*yyLexState).scalar_node(DOLLAR, reflect.String)
		$$.uint8 = $2
	  }
	|
	  XCOMMAND  '.'  EXIT_STATUS
	  {
		l := yylex.(*yyLexState)
		cmd := $1

		if cmd == l.exec {
			l.error("command cannot exec itself: %s", cmd.name)
			return 0
		}

		if l.execed[cmd.name] == false {
			l.error("command '%s' referenced before exec", cmd.name)
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
				fmt.Sprintf("%s %s", l.exec.name, $1.name),
			)

	  	$$ = yylex.(*yyLexState).scalar_node(EXIT_STATUS, reflect.Uint8)
		$$.command = cmd
	  }
	|
	  exp  RE_MATCH  exp
	  {
		l := yylex.(*yyLexState)

		$$ = l.bool_node(RE_MATCH, $1, $3)
		if $$ == nil {
			return 0
		}

		if $1.go_type != reflect.String {
			l.error("~~ operator requires string operands")
			return 0
		}
	  }
	|
	  exp  RE_NMATCH  exp
	  {
		l := yylex.(*yyLexState)
		$$ = l.bool_node(RE_NMATCH, $1, $3)

		if $$ == nil {
			return 0
		}

		if $1.go_type != reflect.String {
			l.error("!~ operator requires string operands")
			return 0
		}
	  }
	|
	  exp  AND  exp
	  {
	  	l := yylex.(*yyLexState)
		$$ = yylex.(*yyLexState).bool_node(AND, $1, $3)
		if $$ == nil {
			return 0
		}

		if $1.go_type != reflect.Bool {
			l.error("logical and requires boolean operands")
			return 0
		}
	  }
	|
	  exp  OR  exp
	  {
	  	l := yylex.(*yyLexState)
		$$ = yylex.(*yyLexState).bool_node(OR, $1, $3)
		if $$ == nil {
			return 0
		}

		if $1.go_type != reflect.Bool {
			l.error("logical or requires boolean operands")
			return 0
		}
	  }
	|
	  exp  EQ  exp
	  {
		$$ = yylex.(*yyLexState).bool_node(EQ, $1, $3)
		if $$ == nil {
			return 0
		}
	  }
	|
	  exp  NEQ  exp
	  {
		$$ = yylex.(*yyLexState).bool_node(NEQ, $1, $3)
		if $$ == nil {
			return 0
		}
	  }
	|
	  NOT  exp
	  {
	  	l := yylex.(*yyLexState)
		if $2.go_type != reflect.Bool {
			l.error("logical not requires boolean operand")
			return 0
		}
	  	$$ = yylex.(*yyLexState).bool_node(NOT, $2, nil) 
	  }
	|
	  '('  exp  ')'
	  {
	  	$$ = $2
	  }
	;
	  

exp_list:
	  exp
	|
	  exp_list  ','  exp
	  {
		argc := uint16(0)
	  	ae := $1

		//  find the last expression in the list

		for ;  ae.next != nil;  ae = ae.next {

			if argc >= 255 {
				yylex.(*yyLexState).error(
					"too many expressions in list: > 255",
				)
				return 0
			}
			argc++
		}
		ae.next = $3
	  }
	;

string_list:
	  STRING
	  {
	  	$$ = make([]string, 1)
		($$)[0] = $1
	  }
	|
	  string_list  ','  STRING
	  {
	  	$$ = append($1, $3)
	  }
	;

argv:
	  /* empty */
	  {
	  	$$ = nil
	  }
	|
	  exp_list
	  {
	  	$$ = &ast{
			yy_tok:	ARGV,
			left:	$1,
		}

		//  count the arguments

		for ae := $1;  ae != nil;  ae = ae.next {
			$$.uint8++
		}
	  }
	;
	
qualification:
	  /*  empty  */
	  {
	  	$$ = nil
	  }
	|
	  WHEN   exp
	  {
	  	$$ = yylex.(*yyLexState).bool_node(WHEN, $2, nil)
	  }
	;

command_argv:
	  /* empty */
	  {
	  	$$ = nil
	  }
	|
	  '('  ')'
	  {
	  	$$ = nil
	  }
	|
	  '('  string_list ')'
	  {
	  	$$ = $2
	  }
	;
	
statement:
	  //  the {} block ought to be optional.

	  COMMAND  NAME  command_argv  '{'
	  	PATH  '='  STRING  ';'
	  '}'
	  {
		l := yylex.(*yyLexState)

		if $7 == "" {
			l.error("command %s: path is zero length", $2)
			return 0
		}

		l.commands[$2] = &command{
					name:		$2,
					path:		$7,
					init_argv:	$3,
				}
		$$ = &ast{
			yy_tok:		COMMAND,
			command:	l.commands[$2],
		}
	  }
	|
	  EXEC  XCOMMAND
	  {
		//  dependency graph needs command being executed

	  	yylex.(*yyLexState).exec = $2
	  }
	  '('  argv  ')'  qualification  ';'
	  {
	  	l := yylex.(*yyLexState)
		n := $2.name
		if l.execed[n] {
			l.error("command '%s' execed more than once", n)
			return 0
		}
		l.execed[n] = true

	  	$$ = &ast{
			yy_tok:		EXEC,
			command:	$2,
			left:		$5,
			right:		$7,
		}
	  }
	;
%%

var keyword = map[string]int{
	"and":			AND,
	"argv":			ARGV,
	"exec":			EXEC,
	"command":		COMMAND,
	"exit_status":		EXIT_STATUS,
	"false":		FALSE,
	"not":			NOT,
	"or":			OR,
	"path":			PATH,
	"true":			TRUE,
	"when":			WHEN,
}

type yyLexState struct {
	//  source code stream
	in				io.RuneReader	//  config source stream

	//  line number in source stream
	line_no				uint64	   //  lexical line number

	//  at end of stream
	eof				bool       //  seen eof in token stream

	//  lookahead on character
	peek				rune       //  lookahead in lexer

	//  error during parsing
	err				error

	//  first statement in parse tree
	ast_head			*ast
	
	//  track declared commands

	commands			map[string]*command

	//  track execed commands

	execed				map[string]bool

	//  track depends list used by tsort to build DAG of
	//  exec relationships.

	depends []string

	exec	*command
}

func (l *yyLexState) pushback(c rune) {

	if l.peek != 0 {
		panic("pushback(): push before peek")	/* impossible */
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
) (*ast) {
	
	return &ast{
		yy_tok:	yy_tok,
		go_type: go_type,
		left: left,
		right: right,
		next: next,
		line_no: l.line_no,
	}
}

func (l *yyLexState) bool_node(yy_tok int, left, right *ast) (*ast) {

	if left != nil && right != nil && left.go_type != right.go_type {
		l.error("operator %s: type mismatch: %s != %s",
				yyToknames[yy_tok - __MIN_YYTOK + 3],
				left.go_type,
				right.go_type,
		)
		return nil
	}
	return l.node(yy_tok, reflect.Bool, left, right, nil)
}

func (l *yyLexState) scalar_node(yy_tok int, go_type reflect.Kind) (*ast) {

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
	for c, eof, err = l.get();  !eof && err == nil;  c, eof, err = l.get() {
		if c == '\n' {
			return
		}
	}
	return err

}

func skip_space(l *yyLexState) (c rune, eof bool, err error) {

	for c, eof, err = l.get();  !eof && err == nil;  c, eof, err = l.get() {
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
	for c, eof, err = l.get();  !eof && err == nil;  c, eof, err = l.get() {
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
		l.pushback(c)		//  first character after ui8
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
 *
 *  Note:
 *	Why ascii only?  unicode ought to suffice.
 */
func (l *yyLexState) scan_word(yylval *yySymType, c rune) (tok int, err error) {
	var eof bool

	w := string(c)
	count := 1

	//  Scan a string of ascii letters, numbers/digits and '_' character.

	for c, eof, err = l.get();  !eof && err == nil;  c, eof, err = l.get() {
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

	for c, eof, err = l.get();  !eof && err == nil;  c, eof, err = l.get(){

		//  double quotes always close the string

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
//  Note: consider changing to big switch{} statement.

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
	if (c > 127) {
		goto PARSE_ERROR
	}

	//  scan a word.  a words starts with a letter

	if (unicode.IsLetter(c)) || c == '_' {
		tok, err = l.scan_word(yylval, c)
		if err != nil {
			goto PARSE_ERROR
		}
		return tok
	}

	//  scan  a decimal uint8

	if unicode.IsNumber(c) {
		err = l.scan_uint8(yylval, c)
		if err != nil {
			goto PARSE_ERROR
		}
		return UINT8
	}

	//  scan a string

	if c == '"' {
		lno := l.line_no	// reset line number on error

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

func (l *yyLexState) mkerror(format string, args...interface{}) error {

	return errors.New(fmt.Sprintf("%s near line %d",
		fmt.Sprintf(format, args...),
		l.line_no,
	))
}

func (l *yyLexState) error(format string, args...interface{}) {

	l.Error(fmt.Sprintf(format, args...))
}

func (l *yyLexState) Error(msg string) {

	if l.err == nil {			//  only report first error
		l.err = l.mkerror("%s", msg)
	}
}

//  enter the yacc dragon

func parse(in io.Reader) (_ *ast, depend_order []string, err error) {

	l := &yyLexState {
		line_no:	1,
		in:		bufio.NewReader(in),
		commands:	make(map[string]*command),
		execed:		make(map[string]bool),
	}

	yyParse(l)

	//  add unqualified exec ... () to dependency list

	var find_unreferenced_EXEC func(a *ast)

	find_unreferenced_EXEC = func(a *ast) {

		if a == nil {
			return
		}
		if a.yy_tok == EXEC && a.command.depend_ref_count == 0 {
			n := a.command.name
			l.depends = append(l.depends, fmt.Sprintf("%s %s", n,n))
		}
		find_unreferenced_EXEC(a.left)
		find_unreferenced_EXEC(a.right)
		find_unreferenced_EXEC(a.next)
	}
	find_unreferenced_EXEC(l.ast_head)

	return l.ast_head, tsort(l.depends), l.err
}
