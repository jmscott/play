/*
 *  Synopsis:
 *	Yacc grammar for 'hoq' language.
 */
%{
package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"unicode"
)

%}

%union {
	string
	uint8

	//  unix command execed by hoq
	command		*command

	//  abstract syntax tree
	ast		*ast
}

//  lowest numbered yytoken.  must be first in list.
%token	__MIN_YYTOK

//  need to lower case
%token	COMMAND  XCOMMAND  EXIT_STATUS
%token	PATH
%token	CALL  WHEN
%token	NAME
%token	STRING
%token	PARSE_ERROR
%token	EQ
%token	NEQ
%token	RE_MATCH  RE_NMATCH
%token	DOLLAR  UINT8
%token	AND  OR  NOT
%token	ARGV  ARGV0  ARGV1
%token	TO_STRING_UINT8

%type	<string>	STRING
%type	<string>	NAME
%type	<ast>		statement  statement_list
%type	<ast>		qualification  boolean
%type	<ast>		expression  expression_list
%type	<command>	XCOMMAND
%type	<ast>		DOLLAR
%type	<ast>		AND  OR
%type	<ast>		RE_MATCH  RE_NMATCH
%type	<uint8>		UINT8
%type	<ast>		argv

%left AND  OR
%left EQ  NEQ  RE_MATCH  RE_NMATCH
%right NOT

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

expression:
	  '$'  UINT8
	  {
	  	$$ = &ast{
			yy_tok:	DOLLAR,
			uint8: $2,
		}
	  }
	|
	  STRING
	  {
	  	$$ = &ast{
			yy_tok:	STRING,
			string: $1,
		}
	  }
	|
	  UINT8
	  {
	  	$$ = &ast{
			yy_tok:	UINT8,
			uint8: $1,
		}
	  }
	|
	  '('  boolean  ')'
	  {
	  	$$ = $2
	  }
	;

expression_list:
	  expression
	|
	  expression_list  ','  expression
	  {
	  	e := $1

		//  linearly find the last expression in the list

		for ;  e.next != nil;  e = e.next {}

		e.next = $3
	  }
	;

argv:
	  /* empty */
	  {
	  	$$ = nil
	  }
	|
	  expression_list
	  {
	  	$$ = &ast{
			yy_tok:	ARGV,
			left:	$1,
		}
	  }
	;
	
boolean:
	  expression  EQ  expression
	  {
	  	$$ = &ast{
			yy_tok: EQ,
			left: $1,
			right: $3,
		}
	  }
	|
	  expression  NEQ  expression
	  {
	  	$$ = &ast{
			yy_tok: NEQ,
			left: $1,
			right: $3,
		}
	  }
	|
	  expression  RE_MATCH  expression
	  {
	  	$$ = &ast{
			yy_tok: RE_MATCH,
			left: $1,
			right: $3,
		}
	  }
	|
	  expression  RE_NMATCH  expression
	  {
	  	$$ = &ast{
			yy_tok: RE_NMATCH,
			left: $1,
			right: $3,
		}
	  }
	|
	  boolean  OR  boolean
	  {
	  	$$ = &ast{
			yy_tok: OR,
			left: $1,
			right: $3,
		}
	  }
	|
	  boolean  AND  boolean
	  {
	  	$$ = &ast{
			yy_tok: AND,
			left: $1,
			right: $3,
		}
	  }
	|
	  NOT  boolean
	  {
	  	$$ = &ast{
			yy_tok:	NOT,
			left: $2,
		}
	  }
	|
	  '('  boolean  ')'
	  {
	  	$$ = $2
	  }
	;

qualification:
	  /*  empty  */
	  {
	  	$$ = nil
	  }
	|
	  WHEN   boolean
	  {
	  	$$ = &ast{
			yy_tok:	WHEN,
			right:	$2,
		}
	  }
	;
	
statement:
	  COMMAND  NAME  '{'  
	  	PATH  '='  STRING  ';'
	  '}'
	  {
		l := yylex.(*yyLexState)

		if $6 == "" {
			l.error("command %s: path is zero length", $2)
			return 0
		}

		l.commands[$2] = &command{
					name: $2,
					path: $6,
				}
		$$ = &ast{
			yy_tok:		COMMAND,
			command:	l.commands[$2],
		}
	  }
	|
	  CALL  XCOMMAND  '('  argv  ')'  qualification  ';'
	  {
	  	$$ = &ast{
			yy_tok:		CALL,
			command:	$2,
			left:		$4,
			right:		$6,
		}
	  }
	;
%%

var keyword = map[string]int{
	"and":			AND,
	"call":			CALL,
	"command":		COMMAND,
	"exit_status":		EXIT_STATUS,
	"not":			NOT,
	"or":			OR,
	"path":			PATH,
	"when":			WHEN,
}

type yyLexState struct {
	//  source code stream
	in				io.RuneReader	//  config source stream

	//  line number in source stream
	line_no				uint8	   //  lexical line number

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
	if (c > 127) {
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

func parse(in io.Reader) (ast *ast, err error) {

	l := &yyLexState {
		line_no:	1,
		in:		bufio.NewReader(in),
		commands:	make(map[string]*command),
	}

	yyParse(l)

	return l.ast_head, l.err
}
