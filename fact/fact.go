package fact

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"

	"github.com/cayleygraph/cayley"
	"github.com/cayleygraph/cayley/graph"
	_ "github.com/cayleygraph/cayley/graph/bolt"
	"github.com/cayleygraph/cayley/graph/path"
	"github.com/cayleygraph/cayley/quad"
)

// Fact is fact!
type Fact struct {
	G *cayley.Handle
	T *Triple
	Q *Triple
	C *CurrentFact
}

// Triple to represent triple
type Triple struct {
	Subject   string
	Predicate string
	Object    string
}

const engine = "bolt"

// NewFact to Create Fact database if not exist
func NewFact(name string) *Fact {
	graph.IgnoreDuplicates = true
	var g *cayley.Handle
	var err error
	if name != "" {
		if _, err := os.Stat(name); os.IsNotExist(err) {
			graph.InitQuadStore(engine, name, nil)
		}
		g, err = cayley.NewGraph(engine, name, nil)
	} else {
		g, err = cayley.NewMemoryGraph()
	}
	if err != nil {
		log.Fatalln(err)
	}
	return &Fact{G: g, C: NewCurrentFact()}
}

// Stringify to make query result to string
func (f *Fact) Stringify(result []interface{}) string {
	out := []string{}
	for _, v := range result {
		out = append(out, v.(string))
	}
	return strings.Join(out, ", ")
}

// Close to release resource
func (f *Fact) Close() {
	f.G.Close()
}

// Let to set subject
func (f *Fact) Let(subj string) *Fact {
	f.T = &Triple{Subject: subj}
	return f
}

// Has sets quad
func (f *Fact) Has(pred string, obj string) {
	f.T.Predicate = pred
	f.T.Object = obj
	f.save()
}

// Save to store triple to database
func (f *Fact) save() error {
	q := quad.Make(quad.IRI(f.T.Subject), quad.IRI(f.T.Predicate), quad.IRI(f.T.Object), nil)
	e := f.G.AddQuad(q)
	return e
}

// What is to query inward
func (f *Fact) What(subj interface{}, preds ...string) []interface{} {
	pred := ""
	if len(preds) == 0 {
		// Finding something of Current fact
		return f.C.Find(reflect.ValueOf(subj).String())
	}

	pred = preds[0]
	switch reflect.TypeOf(subj).Kind() {
	case reflect.String:
		f.Q = &Triple{Subject: subj.(string), Predicate: pred}
		p := cayley.StartPath(f.G, quad.IRI(f.Q.Subject)).Out(quad.IRI(f.Q.Predicate))
		return f.all(p)
	case reflect.Slice:
		out := []interface{}{}
		s := subj.([]interface{})
		for _, v := range s {
			f.Q = &Triple{Subject: v.(string), Predicate: pred}
			p := cayley.StartPath(f.G, quad.IRI(f.Q.Subject)).Out(quad.IRI(f.Q.Predicate))
			out = append(out, f.all(p)...)
		}
		return out
	default:
		fmt.Println(reflect.ValueOf(subj).String())
	}
	return []interface{}{}
}

// WhoHas is to query outward
func (f *Fact) WhoHas(pred, obj string) []interface{} {
	f.Q = &Triple{Predicate: pred, Object: obj}
	p := cayley.StartPath(f.G, quad.IRI(f.Q.Object)).In(quad.IRI(f.Q.Predicate))
	return f.all(p)
}

func (f *Fact) all(p *path.Path) []interface{} {
	// สร้าง optimized iterator
	it, _ := p.BuildIterator().Optimize()

	//เอา optimized iterator ไปชี้ที่ quad ใน graph
	it, _ = f.G.OptimizeIterator(it)

	// clear iterator
	defer it.Close()

	out := []interface{}{}

	// ลูปดึงค่าออกมา
	for it.Next() {
		token := it.Result()                // ดึง token ออกมา (token เป็น reference)
		value := f.G.NameOf(token)          // ดึง value ที่ผูกกับ token นั้นอยู่
		nativeValue := quad.NativeOf(value) // แปลงเป็น go

		out = append(out, nativeValue) // แสดงค่าออกมา
	}
	if err := it.Err(); err != nil {
		log.Fatalln(err)
	}

	out = f.cleanup(out)

	return out
}

func (f *Fact) normalize(s interface{}) interface{} {
	switch s.(type) {
	case string:
		t := s.(string)
		if len(t) > 2 && t[0] == '<' && t[len(t)-1] == '>' {
			t = t[1 : len(t)-1]
		}
		return t
	case quad.IRI:
		t := s.(quad.IRI).String()
		if len(t) > 2 && t[0] == '<' && t[len(t)-1] == '>' {
			t = t[1 : len(t)-1]
		}
		return t
	default:
		return s
	}
}

func (f *Fact) cleanup(s []interface{}) []interface{} {
	out := []interface{}{}
	for _, v := range s {
		out = append(out, f.normalize(v))
	}
	return out
}

// String to return represent value
func (t *Triple) String() string {
	return fmt.Sprintf("%s --> %s --> %s", t.Subject, t.Predicate, t.Object)
}
