package grpc

import (
	"context"
	"fmt"

	"github.com/rinx/vald-meta-halodb/internal/info"
	"github.com/rinx/vald-meta-halodb/internal/log"
	"github.com/rinx/vald-meta-halodb/internal/net/grpc/status"
	"github.com/rinx/vald-meta-halodb/internal/observability/trace"
	"github.com/rinx/vald-meta-halodb/pkg/meta/halodb/service"
	"github.com/vdaas/vald/apis/grpc/meta"
	"github.com/vdaas/vald/apis/grpc/payload"
)

type server struct {
	haloDB service.HaloDB
}

func New(opts ...Option) meta.MetaServer {
	s := new(server)

	for _, opt := range append(defaultOpts, opts...) {
		opt(s)
	}
	return s
}

func (s *server) kvKey(key string) string {
	return "kv:" + key
}

func (s *server) vkKey(val string) string {
	return "vk:" + val
}

func (s *server) GetMeta(ctx context.Context, key *payload.Meta_Key) (*payload.Meta_Val, error) {
	ctx, span := trace.StartSpan(ctx, "vald/meta-haloDB.GetMeta")
	defer func() {
		if span != nil {
			span.End()
		}
	}()

	val, err := s.haloDB.Get(s.kvKey(key.GetKey()))
	if err != nil {
		log.Warnf("[GetMeta]\tkey %s not found", key.GetKey())
		if span != nil {
			span.SetStatus(trace.StatusCodeNotFound(err.Error()))
		}
		return nil, status.WrapWithNotFound(fmt.Sprintf("GetMeta API haloDB key %s not found", key.GetKey()), err, info.Get())
	}
	return &payload.Meta_Val{
		Val: val,
	}, nil
}

func (s *server) GetMetas(ctx context.Context, keys *payload.Meta_Keys) (mv *payload.Meta_Vals, err error) {
	ctx, span := trace.StartSpan(ctx, "vald/meta-haloDB.GetMetas")
	defer func() {
		if span != nil {
			span.End()
		}
	}()
	mv = new(payload.Meta_Vals)
	for _, k := range keys.GetKeys() {
		v, err := s.haloDB.Get(s.kvKey(k))
		if err != nil {
			log.Warnf("[GetMetas]\tkeys %#v not found", keys.GetKeys())
			if span != nil {
				span.SetStatus(trace.StatusCodeNotFound(err.Error()))
			}
			return mv, status.WrapWithNotFound(fmt.Sprintf("GetMetas API haloDB entry keys %#v not found", keys.GetKeys()), err, info.Get())
		}
		mv.Vals = append(mv.Vals, v)
	}
	return mv, nil
}

func (s *server) GetMetaInverse(ctx context.Context, val *payload.Meta_Val) (*payload.Meta_Key, error) {
	ctx, span := trace.StartSpan(ctx, "vald/meta-haloDB.GetMetaInverse")
	defer func() {
		if span != nil {
			span.End()
		}
	}()
	key, err := s.haloDB.Get(s.vkKey(val.GetVal()))
	if err != nil {
		log.Warnf("[GetMetaInverse]\tval %s not found", val.GetVal())
		if span != nil {
			span.SetStatus(trace.StatusCodeNotFound(err.Error()))
		}
		return nil, status.WrapWithNotFound(fmt.Sprintf("GetMetaInverse API haloDB val %s not found", val.GetVal()), err, info.Get())
	}
	return &payload.Meta_Key{
		Key: key,
	}, nil
}

func (s *server) GetMetasInverse(ctx context.Context, vals *payload.Meta_Vals) (mk *payload.Meta_Keys, err error) {
	ctx, span := trace.StartSpan(ctx, "vald/meta-haloDB.GetMetasInverse")
	defer func() {
		if span != nil {
			span.End()
		}
	}()
	mk = new(payload.Meta_Keys)
	for _, v := range vals.GetVals() {
		k, err := s.haloDB.Get(s.vkKey(v))
		if err != nil {
			log.Warnf("[GetMetasInverse]\tvals %#v not found", vals.GetVals())
			if span != nil {
				span.SetStatus(trace.StatusCodeNotFound(err.Error()))
			}
			return mk, status.WrapWithNotFound(fmt.Sprintf("GetMetasInverse API haloDB vals %#v not found", vals.GetVals()), err, info.Get())
		}
		mk.Keys = append(mk.Keys, k)
	}
	return mk, nil
}

func (s *server) SetMeta(ctx context.Context, kv *payload.Meta_KeyVal) (_ *payload.Empty, err error) {
	ctx, span := trace.StartSpan(ctx, "vald/meta-haloDB.SetMeta")
	defer func() {
		if span != nil {
			span.End()
		}
	}()
	err = s.haloDB.Put(s.kvKey(kv.GetKey()), kv.GetVal())
	if err != nil {
		log.Errorf("[SetMeta]\tunknown error\t%+v", err)
		if span != nil {
			span.SetStatus(trace.StatusCodeInternal(err.Error()))
		}
		return nil, status.WrapWithInternal(fmt.Sprintf("SetMeta API haloDB key %s val %s failed to store", kv.GetKey(), kv.GetVal()), err, info.Get())
	}
	err = s.haloDB.Put(s.vkKey(kv.GetVal()), kv.GetKey())
	if err != nil {
		log.Errorf("[SetMeta]\tunknown error\t%+v", err)
		if span != nil {
			span.SetStatus(trace.StatusCodeInternal(err.Error()))
		}
		return nil, status.WrapWithInternal(fmt.Sprintf("SetMeta API haloDB key %s val %s failed to store", kv.GetKey(), kv.GetVal()), err, info.Get())
	}
	return new(payload.Empty), nil
}

func (s *server) SetMetas(ctx context.Context, kvs *payload.Meta_KeyVals) (_ *payload.Empty, err error) {
	ctx, span := trace.StartSpan(ctx, "vald/meta-haloDB.SetMetas")
	defer func() {
		if span != nil {
			span.End()
		}
	}()
	for _, kv := range kvs.GetKvs() {
		_, err = s.SetMeta(ctx, kv)
		if err != nil {
			log.Errorf("[SetMetas]\tunknown error\t%+v", err)
			if span != nil {
				span.SetStatus(trace.StatusCodeInternal(err.Error()))
			}
			return nil, status.WrapWithInternal("SetMetas API haloDB failed to store", err, info.Get())
		}
	}
	return new(payload.Empty), nil
}

func (s *server) DeleteMeta(ctx context.Context, key *payload.Meta_Key) (*payload.Meta_Val, error) {
	ctx, span := trace.StartSpan(ctx, "vald/meta-haloDB.DeleteMeta")
	defer func() {
		if span != nil {
			span.End()
		}
	}()
	val, err := s.GetMeta(ctx, key)
	if err != nil {
		log.Errorf("[DeleteMeta]\tunknown error\t%+v", err)
		if span != nil {
			span.SetStatus(trace.StatusCodeUnknown(err.Error()))
		}
		return nil, status.WrapWithUnknown(fmt.Sprintf("DeleteMeta API haloDB unknown error occurred key %s", key.GetKey()), err, info.Get())
	}
	err = s.haloDB.Delete(s.kvKey(key.GetKey()))
	if err != nil {
		log.Errorf("[DeleteMeta]\tunknown error\t%+v", err)
		if span != nil {
			span.SetStatus(trace.StatusCodeUnknown(err.Error()))
		}
		return nil, status.WrapWithUnknown(fmt.Sprintf("DeleteMeta API haloDB unknown error occurred key %s", key.GetKey()), err, info.Get())
	}
	return val, nil
}

func (s *server) DeleteMetas(ctx context.Context, keys *payload.Meta_Keys) (mv *payload.Meta_Vals, err error) {
	ctx, span := trace.StartSpan(ctx, "vald/meta-haloDB.DeleteMetas")
	defer func() {
		if span != nil {
			span.End()
		}
	}()
	mv, err = s.GetMetas(ctx, keys)
	if err != nil {
		log.Errorf("[DeleteMetas]\tunknown error\t%+v", err)
		if span != nil {
			span.SetStatus(trace.StatusCodeUnknown(err.Error()))
		}
		return mv, status.WrapWithUnknown(fmt.Sprintf("DeleteMetas API haloDB entry keys %#v unknown error occurred", keys.GetKeys()), err, info.Get())
	}
	for _, k := range keys.GetKeys() {
		err = s.haloDB.Delete(s.kvKey(k))
		if err != nil {
			log.Errorf("[DeleteMetas]\tunknown error\t%+v", err)
			if span != nil {
				span.SetStatus(trace.StatusCodeUnknown(err.Error()))
			}
			return mv, status.WrapWithUnknown(fmt.Sprintf("DeleteMetas API haloDB entry keys %#v unknown error occurred", keys.GetKeys()), err, info.Get())
		}
	}
	return mv, nil
}

func (s *server) DeleteMetaInverse(ctx context.Context, val *payload.Meta_Val) (*payload.Meta_Key, error) {
	ctx, span := trace.StartSpan(ctx, "vald/meta-haloDB.DeleteMetaInverse")
	defer func() {
		if span != nil {
			span.End()
		}
	}()
	key, err := s.GetMetaInverse(ctx, val)
	if err != nil {
		log.Errorf("[DeleteMetaInverse]\tunknown error\t%+v", err)
		if span != nil {
			span.SetStatus(trace.StatusCodeUnknown(err.Error()))
		}
		return nil, status.WrapWithUnknown(fmt.Sprintf("DeleteMetaInverse API val %s unknown error occurred", val.GetVal()), err, info.Get())
	}
	err = s.haloDB.Delete(s.vkKey(val.GetVal()))
	if err != nil {
		log.Errorf("[DeleteMetaInverse]\tunknown error\t%+v", err)
		if span != nil {
			span.SetStatus(trace.StatusCodeUnknown(err.Error()))
		}
		return nil, status.WrapWithUnknown(fmt.Sprintf("DeleteMetaInverse API val %s unknown error occurred", val.GetVal()), err, info.Get())
	}
	return key, nil
}

func (s *server) DeleteMetasInverse(ctx context.Context, vals *payload.Meta_Vals) (mk *payload.Meta_Keys, err error) {
	ctx, span := trace.StartSpan(ctx, "vald/meta-haloDB.DeleteMetasInverse")
	defer func() {
		if span != nil {
			span.End()
		}
	}()
	mk, err = s.GetMetasInverse(ctx, vals)
	if err != nil {
		log.Errorf("[DeleteMetasInverse]\tunknown error\t%+v", err)
		if span != nil {
			span.SetStatus(trace.StatusCodeUnknown(err.Error()))
		}
		return mk, status.WrapWithUnknown(fmt.Sprintf("DeleteMetasInverse API vals %#v unknown error occurred", vals.GetVals()), err, info.Get())
	}
	for _, v := range vals.GetVals() {
		err = s.haloDB.Delete(s.vkKey(v))
		if err != nil {
			log.Errorf("[DeleteMetasInverse]\tunknown error\t%+v", err)
			if span != nil {
				span.SetStatus(trace.StatusCodeUnknown(err.Error()))
			}
			return mk, status.WrapWithUnknown(fmt.Sprintf("DeleteMetasInverse API vals %#v unknown error occurred", vals.GetVals()), err, info.Get())
		}
	}
	return mk, nil
}
