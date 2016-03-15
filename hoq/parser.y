/*
 *  Synopsis:
 *	Yacc grammar for 'hoq' language.
 */
%{
package main

import (
	"errors"
	"io"
	"unicode"

	"fmt"
)

type command struct {
	name	string
	path	string
}

//  abstract syntax tree that represents the parsed program

type ast struct {

	yy_tok	int		//  lexical token defined by yacc

	string
	*command

	//  children
	left	*ast
	right	*ast

	//  siblings
	next	*ast
}

func init() {
	if yyToknames[3] != "__MIN_YYTOK" {
		panic("yyToknames[3] != __MIN_YYTOK: yacc may have changed")
	}
}

%}

%union {
	string
	command		*command
	ast		*ast
}

//  lowest numbered yytoken.  must be first in list.
%token	__MIN_YYTOK

%token	COMMAND
%token	PATH
%token	NAME
%token	STRING
%token	PARSE_ERROR
%token	EQ
%token	NEQ

%type	<string>	STRING
%type	<string>	NAME
%type	<ast>		statement  statement_list

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
		for ;  s.next != nil;  s = s.next {}	//  find last stmt

		s.next = $2
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

		$$ = &ast{
			yy_tok:		COMMAND,
			command:	&command {
						name:	$2,
						path:	$6,	
					},
		}
	  }
	;
%%

var keyword = map[string]int{
	"command":		COMMAND,
	"path":			PATH,
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
	if !eof {
		l.pushback(c)		/* first character after word */
	}

	if keyword[w] > 0 {		/* got a keyword */
		return keyword[w], nil	/* return yacc generated token */
	}

	yylval.string = w
	return NAME, nil
}

/*
 *  Very simple utf8 string scanning, with no proper escapes for characters.
 */
func (l *yyLexState) scan_string(yylval *yySymType) (eof bool, err error) {
	var c rune
	s := ""

	for c, eof, err = l.get();  !eof && err == nil;  c, eof, err = l.get(){
		if c == '"' {
			yylval.string = s
			return false, nil
		}
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
	/*
	 *  switch(c) statement?
	 */
	if (unicode.IsLetter(c)) || c == '_' {
		tok, err = l.scan_word(yylval, c)
		if err != nil {
			goto PARSE_ERROR
		}
		return tok
	}
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
	if c == '=' {
		tok, err = lookahead(l, '=', EQ, int('='))
		if err != nil {
			goto PARSE_ERROR
		}
		return tok
	}
	if c == '!' {
		tok, err = lookahead(l, '=', NEQ, int('!'))
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

func parse(in io.RuneReader) (
				ast *ast,
				err error,
) {
	l := &yyLexState {
		line_no:	1,
		in:		in,
		eof:		false,
		err:		nil,
	}
	yyParse(l)
	err = l.err
	if err != nil {
		return
	}
	if l.ast_head == nil {
		panic("null ast_head")
	}
	return l.ast_head, err
}
