package dynamodb

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"unicode"
)

const maxExpressionLength = 4096
const maxExpressionTokens = 512
const maxExpressionDepth = 32

type tokenKind int

const (
	tEOF tokenKind = iota
	tWord
	tName
	tValue
	tNumber
	tLParen
	tRParen
	tLBracket
	tRBracket
	tComma
	tDot
	tEq
	tNE
	tLT
	tLE
	tGT
	tGE
	tPlus
	tMinus
)

type exprToken struct {
	kind tokenKind
	text string
	pos  int
}

func lexExpression(input string) ([]exprToken, error) {
	if len(input) == 0 || len(input) > maxExpressionLength {
		return nil, errors.New("expression length is invalid")
	}
	var out []exprToken
	for i := 0; i < len(input); {
		if unicode.IsSpace(rune(input[i])) {
			i++
			continue
		}
		start := i
		var kind tokenKind
		switch input[i] {
		case '(':
			kind, i = tLParen, i+1
		case ')':
			kind, i = tRParen, i+1
		case '[':
			kind, i = tLBracket, i+1
		case ']':
			kind, i = tRBracket, i+1
		case ',':
			kind, i = tComma, i+1
		case '.':
			kind, i = tDot, i+1
		case '+':
			kind, i = tPlus, i+1
		case '-':
			kind, i = tMinus, i+1
		case '=':
			kind, i = tEq, i+1
		case '<':
			i++
			if i < len(input) && input[i] == '=' {
				kind, i = tLE, i+1
			} else if i < len(input) && input[i] == '>' {
				kind, i = tNE, i+1
			} else {
				kind = tLT
			}
		case '>':
			i++
			if i < len(input) && input[i] == '=' {
				kind, i = tGE, i+1
			} else {
				kind = tGT
			}
		case '#', ':':
			prefix := input[i]
			i++
			for i < len(input) && (unicode.IsLetter(rune(input[i])) || unicode.IsDigit(rune(input[i])) || input[i] == '_') {
				i++
			}
			if i == start+1 {
				return nil, fmt.Errorf("invalid placeholder at position %d", start)
			}
			if prefix == '#' {
				kind = tName
			} else {
				kind = tValue
			}
		default:
			if unicode.IsDigit(rune(input[i])) {
				kind = tNumber
				for i < len(input) && unicode.IsDigit(rune(input[i])) {
					i++
				}
			} else if unicode.IsLetter(rune(input[i])) || input[i] == '_' {
				kind = tWord
				for i < len(input) && (unicode.IsLetter(rune(input[i])) || unicode.IsDigit(rune(input[i])) || input[i] == '_') {
					i++
				}
			} else {
				return nil, fmt.Errorf("unexpected character at position %d", i)
			}
		}
		out = append(out, exprToken{kind: kind, text: input[start:i], pos: start})
		if len(out) > maxExpressionTokens {
			return nil, errors.New("expression has too many tokens")
		}
	}
	return append(out, exprToken{kind: tEOF, pos: len(input)}), nil
}

type pathPart struct {
	name  string
	index int
	isIdx bool
}
type expression struct {
	op          string
	path        []pathPart
	value       *Value
	left, right *expression
	args        []*expression
}
type expressionParser struct {
	tokens []exprToken
	pos    int
	names  map[string]string
	values map[string]Value
	depth  int
}

func parseCondition(input string, names map[string]string, values map[string]Value) (*expression, error) {
	tokens, err := lexExpression(input)
	if err != nil {
		return nil, err
	}
	p := &expressionParser{tokens: tokens, names: names, values: values}
	e, err := p.parseOr()
	if err == nil && p.peek().kind != tEOF {
		err = fmt.Errorf("unexpected token %q at position %d", p.peek().text, p.peek().pos)
	}
	return e, err
}
func (p *expressionParser) peek() exprToken { return p.tokens[p.pos] }
func (p *expressionParser) take() exprToken { t := p.peek(); p.pos++; return t }
func (p *expressionParser) word(s string) bool {
	return p.peek().kind == tWord && strings.EqualFold(p.peek().text, s)
}
func (p *expressionParser) parseOr() (*expression, error) {
	e, err := p.parseAnd()
	for err == nil && p.word("OR") {
		p.take()
		var r *expression
		r, err = p.parseAnd()
		e = &expression{op: "OR", left: e, right: r}
	}
	return e, err
}
func (p *expressionParser) parseAnd() (*expression, error) {
	e, err := p.parseNot()
	for err == nil && p.word("AND") {
		p.take()
		var r *expression
		r, err = p.parseNot()
		e = &expression{op: "AND", left: e, right: r}
	}
	return e, err
}
func (p *expressionParser) parseNot() (*expression, error) {
	if p.word("NOT") {
		p.take()
		e, err := p.parseNot()
		return &expression{op: "NOT", left: e}, err
	}
	return p.parsePredicate()
}
func (p *expressionParser) parsePredicate() (*expression, error) {
	if p.peek().kind == tLParen {
		p.take()
		p.depth++
		if p.depth > maxExpressionDepth {
			return nil, errors.New("expression nesting too deep")
		}
		e, err := p.parseOr()
		p.depth--
		if err != nil {
			return nil, err
		}
		if p.peek().kind != tRParen {
			return nil, fmt.Errorf("missing ) at position %d", p.peek().pos)
		}
		p.take()
		return e, nil
	}
	left, err := p.parseOperand()
	if err != nil {
		return nil, err
	}
	if p.word("BETWEEN") {
		p.take()
		lo, e := p.parseOperand()
		if e != nil {
			return nil, e
		}
		if !p.word("AND") {
			return nil, errors.New("BETWEEN requires AND")
		}
		p.take()
		hi, e := p.parseOperand()
		return &expression{op: "BETWEEN", left: left, args: []*expression{lo, hi}}, e
	}
	if p.word("IN") {
		p.take()
		if p.peek().kind != tLParen {
			return nil, errors.New("IN requires values")
		}
		p.take()
		var args []*expression
		for {
			v, e := p.parseOperand()
			if e != nil {
				return nil, e
			}
			args = append(args, v)
			if p.peek().kind != tComma {
				break
			}
			p.take()
		}
		if p.peek().kind != tRParen {
			return nil, errors.New("IN missing )")
		}
		p.take()
		return &expression{op: "IN", left: left, args: args}, nil
	}
	op := map[tokenKind]string{tEq: "=", tNE: "<>", tLT: "<", tLE: "<=", tGT: ">", tGE: ">="}[p.peek().kind]
	if op == "" {
		return left, nil
	}
	p.take()
	right, err := p.parseOperand()
	return &expression{op: op, left: left, right: right}, err
}
func (p *expressionParser) parseOperand() (*expression, error) {
	if p.peek().kind == tValue {
		t := p.take()
		v, ok := p.values[t.text]
		if !ok {
			return nil, fmt.Errorf("missing value placeholder %s", t.text)
		}
		x := v
		return &expression{op: "VALUE", value: &x}, nil
	}
	if p.peek().kind != tWord && p.peek().kind != tName {
		return nil, fmt.Errorf("expected operand at position %d", p.peek().pos)
	}
	if p.peek().kind == tWord && p.pos+1 < len(p.tokens) && p.tokens[p.pos+1].kind == tLParen {
		name := strings.ToLower(p.take().text)
		p.take()
		var args []*expression
		if p.peek().kind != tRParen {
			for {
				a, e := p.parseOperand()
				if e != nil {
					return nil, e
				}
				args = append(args, a)
				if p.peek().kind != tComma {
					break
				}
				p.take()
			}
		}
		if p.peek().kind != tRParen {
			return nil, errors.New("function missing )")
		}
		p.take()
		switch name {
		case "attribute_exists", "attribute_not_exists", "attribute_type", "begins_with", "contains", "size":
		default:
			return nil, fmt.Errorf("unsupported function %s", name)
		}
		return &expression{op: name, args: args}, nil
	}
	path, err := p.parsePath()
	return &expression{op: "PATH", path: path}, err
}
func (p *expressionParser) parsePath() ([]pathPart, error) {
	var out []pathPart
	for {
		t := p.peek()
		if t.kind != tWord && t.kind != tName {
			return nil, fmt.Errorf("invalid path at position %d", t.pos)
		}
		p.take()
		name := t.text
		if t.kind == tName {
			var ok bool
			name, ok = p.names[t.text]
			if !ok || name == "" {
				return nil, fmt.Errorf("missing name placeholder %s", t.text)
			}
		}
		out = append(out, pathPart{name: name})
		for p.peek().kind == tLBracket {
			p.take()
			n := p.take()
			if n.kind != tNumber {
				return nil, errors.New("list index must be non-negative integer")
			}
			idx, e := strconv.Atoi(n.text)
			if e != nil {
				return nil, e
			}
			if p.peek().kind != tRBracket {
				return nil, errors.New("list index missing ]")
			}
			p.take()
			out = append(out, pathPart{index: idx, isIdx: true})
		}
		if p.peek().kind != tDot {
			break
		}
		p.take()
	}
	return out, nil
}

func evalCondition(e *expression, item map[string]Value) (bool, error) {
	v, exists, err := evalExpr(e, item)
	if err != nil {
		return false, err
	}
	if !exists || v.Kind != "BOOL" {
		return false, errors.New("condition does not evaluate to boolean")
	}
	return v.Bool, nil
}
func evalExpr(e *expression, item map[string]Value) (Value, bool, error) {
	switch e.op {
	case "VALUE":
		return *e.value, true, nil
	case "PATH":
		return getPath(item, e.path)
	case "AND", "OR":
		a, err := evalCondition(e.left, item)
		if err != nil {
			return Value{}, false, err
		}
		if e.op == "AND" && !a {
			return Value{Kind: "BOOL", Bool: false}, true, nil
		}
		if e.op == "OR" && a {
			return Value{Kind: "BOOL", Bool: true}, true, nil
		}
		b, err := evalCondition(e.right, item)
		return Value{Kind: "BOOL", Bool: b}, true, err
	case "NOT":
		a, err := evalCondition(e.left, item)
		return Value{Kind: "BOOL", Bool: !a}, true, err
	case "attribute_exists", "attribute_not_exists":
		_, ok, err := evalExpr(e.args[0], item)
		if e.op == "attribute_not_exists" {
			ok = !ok
		}
		return Value{Kind: "BOOL", Bool: ok}, true, err
	case "size":
		v, ok, err := evalExpr(e.args[0], item)
		if err != nil || !ok {
			return Value{}, false, err
		}
		n := 0
		switch v.Kind {
		case "S":
			n = len([]byte(v.S))
		case "B":
			n = len(v.B)
		case "L":
			n = len(v.L)
		case "M":
			n = len(v.M)
		case "SS", "NS", "BS":
			n = len(v.Set)
		default:
			return Value{}, false, errors.New("size unsupported for type")
		}
		return Value{Kind: "N", N: strconv.Itoa(n)}, true, nil
	case "attribute_type":
		v, ok, err := evalExpr(e.args[0], item)
		if err != nil || !ok {
			return Value{Kind: "BOOL", Bool: false}, true, err
		}
		typ, ok2, err := evalExpr(e.args[1], item)
		return Value{Kind: "BOOL", Bool: ok2 && typ.Kind == "S" && v.Kind == typ.S}, true, err
	case "begins_with", "contains":
		a, ok, err := evalExpr(e.args[0], item)
		if err != nil || !ok {
			return Value{Kind: "BOOL"}, true, err
		}
		b, ok, err := evalExpr(e.args[1], item)
		if err != nil || !ok {
			return Value{Kind: "BOOL"}, true, err
		}
		yes := false
		if e.op == "begins_with" {
			if a.Kind == "S" && b.Kind == "S" {
				yes = strings.HasPrefix(a.S, b.S)
			} else if a.Kind == "B" && b.Kind == "B" {
				yes = bytes.HasPrefix(a.B, b.B)
			} else {
				return Value{}, false, errors.New("begins_with type mismatch")
			}
		} else {
			switch a.Kind {
			case "S":
				yes = b.Kind == "S" && strings.Contains(a.S, b.S)
			case "L":
				for _, x := range a.L {
					eq, _ := equalValue(x, b)
					yes = yes || eq
				}
			case "SS", "NS", "BS":
				for _, x := range a.Set {
					eq, _ := equalValue(x, b)
					yes = yes || eq
				}
			default:
				return Value{}, false, errors.New("contains type mismatch")
			}
		}
		return Value{Kind: "BOOL", Bool: yes}, true, nil
	case "=", "<>", "<", "<=", ">", ">=":
		a, aok, err := evalExpr(e.left, item)
		if err != nil {
			return Value{}, false, err
		}
		b, bok, err := evalExpr(e.right, item)
		if err != nil {
			return Value{}, false, err
		}
		cmp, comparable := compareValue(a, b)
		yes := false
		if e.op == "=" || e.op == "<>" {
			yes = aok && bok && comparable && cmp == 0
			if e.op == "<>" {
				yes = !yes
			}
		} else if aok && bok && comparable {
			yes = map[string]bool{"<": cmp < 0, "<=": cmp <= 0, ">": cmp > 0, ">=": cmp >= 0}[e.op]
		}
		return Value{Kind: "BOOL", Bool: yes}, true, nil
	case "BETWEEN":
		a, ok, err := evalExpr(e.left, item)
		if err != nil || !ok {
			return Value{Kind: "BOOL"}, true, err
		}
		lo, _, err := evalExpr(e.args[0], item)
		if err != nil {
			return Value{}, false, err
		}
		hi, _, err := evalExpr(e.args[1], item)
		c1, o1 := compareValue(a, lo)
		c2, o2 := compareValue(a, hi)
		return Value{Kind: "BOOL", Bool: o1 && o2 && c1 >= 0 && c2 <= 0}, true, err
	case "IN":
		a, ok, err := evalExpr(e.left, item)
		if err != nil || !ok {
			return Value{Kind: "BOOL"}, true, err
		}
		yes := false
		for _, x := range e.args {
			b, _, er := evalExpr(x, item)
			if er != nil {
				return Value{}, false, er
			}
			eq, _ := equalValue(a, b)
			yes = yes || eq
		}
		return Value{Kind: "BOOL", Bool: yes}, true, nil
	}
	return Value{}, false, fmt.Errorf("unsupported expression %s", e.op)
}
func getPath(item map[string]Value, path []pathPart) (Value, bool, error) {
	if len(path) == 0 || path[0].isIdx {
		return Value{}, false, errors.New("invalid path")
	}
	v, ok := item[path[0].name]
	for _, part := range path[1:] {
		if !ok {
			return Value{}, false, nil
		}
		if part.isIdx {
			if v.Kind != "L" || part.index >= len(v.L) {
				return Value{}, false, nil
			}
			v = v.L[part.index]
		} else {
			if v.Kind != "M" {
				return Value{}, false, nil
			}
			v, ok = v.M[part.name]
		}
	}
	return v, ok, nil
}
func compareValue(a, b Value) (int, bool) {
	if a.Kind != b.Kind {
		return 0, false
	}
	switch a.Kind {
	case "N":
		x, _ := new(big.Rat).SetString(a.N)
		y, _ := new(big.Rat).SetString(b.N)
		return x.Cmp(y), true
	case "S":
		return strings.Compare(a.S, b.S), true
	case "B":
		return bytes.Compare(a.B, b.B), true
	default:
		ac, _ := a.Canonical()
		bc, _ := b.Canonical()
		return bytes.Compare(ac, bc), true
	}
}
func equalValue(a, b Value) (bool, error) { c, ok := compareValue(a, b); return ok && c == 0, nil }
