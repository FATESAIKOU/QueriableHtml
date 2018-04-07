/* A simple library to build queriable html structure.
 *
 * @author: FATESAIKOU
 * @date  : 04/17/2018
 */

package queriableHtml

import (
    "fmt"
    "bytes"
    "strings"
    "regexp"
    "golang.org/x/net/html"
)

type DOMObj struct {
    Atom string
    Attrs map[string]string
    Contents []DOMObj
    TokenType html.TokenType
}

func (self *DOMObj) GetEleByAtom(pattern string) []DOMObj {
    res := []DOMObj{}

    for i := range self.Contents {
        cmp_res, _ := regexp.MatchString(pattern, self.Contents[i].Atom)
        if cmp_res {
            res = append(res, self.Contents[i])
        }
    }

    return res
}

func (self *DOMObj) GetEleByAttr(term string, pattern string) []DOMObj {
    res := []DOMObj{}

    for i := range self.Contents {
        val, ok := self.Contents[i].Attrs[term]
        if !ok {
            continue
        }

        cmp_res, _ := regexp.MatchString(pattern, val)
        if cmp_res {
            res = append(res, self.Contents[i])
        }
    }

    return res
}

func (self *DOMObj) GetEleByQuery(query_str string) []DOMObj {
    query_terms := strings.Split(query_str, ",")
    if len(query_terms) < 1 {
        fmt.Println("Error while parsing the query string!")
        return nil
    }

    switch term := query_terms[0]; term {
    case "*", "*...":
        return self.Contents
    case "Attr":
        return self.GetEleByAttr(query_terms[1], query_terms[2])
    case "Atom":
        return self.GetEleByAtom(query_terms[1])
    default:
        return nil
    }
}

//  The main function for query, query_strs is a list of query string,
//  each query string can be:
//      1. *...                 // pass with no matching as much layer as possible while querying
//      2. *                    // pass with no matching one layer while querying
//      3. Attr,AttrName,regexp // get all the element that it's attribute of AttrName is matched with regexp
//      4. Atom,regexp          // get all the element that it's tag name is matched with regexp
//
//  and while the query string is multiple, the next query string will be apply to the result's child of
//  previous query string.
func (self *DOMObj) Query(query_strs []string) []DOMObj {
    if len(query_strs) < 1 {
        return []DOMObj{*self}
    }

    res := []DOMObj{}
    tmp_res := self.GetEleByQuery(query_strs[0])

    for i := range tmp_res {
        res = append(res,
            tmp_res[i].Query(query_strs[1:])...)
    }

    if query_strs[0] == "*..." {
        for i := range tmp_res {
            res = append(res,
                tmp_res[i].Query(query_strs)...)
        }
    }

    return res
}


/* For constructing */
func BuildScope(body_bytes []byte) (map[int]bool, map[int]bool) {
    type Scope struct {
        Name string
        SInd int
    }
    scope_stack := []Scope{}

    in_seq  := make(map[int]bool)
    out_seq := make(map[int]bool)

    tr := html.NewTokenizer(bytes.NewBuffer(body_bytes))
    cnt := 0
    for {
        tt := tr.Next()
        t  := tr.Token()

        if tt == html.StartTagToken {
            scope_stack = append(scope_stack, Scope{
                Name: t.Data,
                SInd: cnt})
        }

        if tt == html.EndTagToken {
            var m int
            for m = len(scope_stack) - 1; m >=0; m -- {
                if scope_stack[m].Name == t.Data {
                    break
                }
            }

            if m >= 0 {
                in_seq[scope_stack[m].SInd] = true
                out_seq[cnt] = true
                scope_stack = scope_stack[:m]
            }
        }

        if tt == html.ErrorToken {
            break
        }

        cnt ++
    }

    return in_seq, out_seq
}

func LoadAttr(attrs []html.Attribute) map[string]string {
    loaded_attrs := map[string]string{}

    for i := range attrs {
        loaded_attrs[attrs[i].Key] = attrs[i].Val
    }

    return loaded_attrs
}

func NewQueriableHtml(body_bytes []byte) DOMObj {
    root := DOMObj{
        Atom: "<ROOT>",
        Attrs: map[string]string{},
        Contents: []DOMObj{},
        TokenType: html.StartTagToken}

    in_seq, out_seq := BuildScope(body_bytes)
    token_reader := html.NewTokenizer(bytes.NewBuffer(body_bytes))

    cnt := -1
    root.Contents = ParseHtml(in_seq, out_seq, &cnt, token_reader)
    return root
}

func ParseHtml(in_seq map[int]bool, out_seq map[int]bool, cnt *int, tr *html.Tokenizer) []DOMObj {
    obj_list := []DOMObj{}

    for {
        (*cnt) ++
        tt := tr.Next()
        t  := tr.Token()

        if out_seq[*cnt] == true || tt == html.ErrorToken {
            break
        }


        tmp_obj := DOMObj{
            Atom: t.Data,
            Attrs: LoadAttr(t.Attr),
            Contents: nil,
            TokenType: tt}

        if in_seq[*cnt] == true {
            tmp_obj.Contents = ParseHtml(in_seq, out_seq, cnt, tr)
        }
        obj_list = append(obj_list, tmp_obj)
    }

    return obj_list
}
