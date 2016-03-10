package tpl

import (
	"reflect"
	"sync"
	"github.com/LindenY/gomdise/trans"
	"errors"
	"fmt"
)

/*
 *
 */
type ActionTemplate interface {
	engrave(actions *[]*trans.Action, args ...interface{})
}

/*
 *
 */
type TemplateCreater func(t reflect.Type) ActionTemplate

/*
 *
 */
type TemplateCache struct {
	sync.RWMutex
	tpls    map[reflect.Type]ActionTemplate
	creater TemplateCreater
}

/*
 *
 */
func (tcache *TemplateCache) GetTemplate(t reflect.Type) ActionTemplate {
	tcache.RLock()
	tpl := tcache.tpls[t]
	tcache.RUnlock()
	if tpl != nil {
		return tpl
	}

	tcache.Lock()
	if tcache.tpls == nil {
		tcache.tpls = make(map[reflect.Type]ActionTemplate)
	}
	var wg sync.WaitGroup
	wg.Add(1)
	tcache.tpls[t] = &proxyTemplate{wg, tpl}
	tcache.Unlock()

	tpl = tcache.creater(t)
	wg.Done()
	tcache.Lock()
	tcache.tpls[t] = tpl
	tcache.Unlock()
	return tpl
}

/*
 *
 */
func newTplCache(creater TemplateCreater) *TemplateCache {
	tcache := new(TemplateCache)
	tcache.tpls = make(map[reflect.Type]ActionTemplate)
	tcache.creater = creater
	return tcache
}

/*
 *
 */
type proxyTemplate struct {
	wg   sync.WaitGroup
	dest ActionTemplate
}

/*
 *
 */
func (pt *proxyTemplate) engrave(actions *[]*trans.Action, args ...interface{}) {
	pt.wg.Wait()
	pt.dest.engrave(actions, args...)
}

/*
 *
 */
type unsupportedTypeTemplate struct {
	typ reflect.Type
	op string
}

func (ust *unsupportedTypeTemplate) engrave(actions *[]*trans.Action, args ...interface{}) {
	panic(errors.New(fmt.Sprintf("Operation[%s] does not support for type: %v \n", ust.op, ust.typ)))
}

func newUnsupportedTypeTemplate(t reflect.Type, op string) *unsupportedTypeTemplate {
	return &unsupportedTypeTemplate{t, op}
}