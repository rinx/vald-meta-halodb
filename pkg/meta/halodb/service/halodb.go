package service

// #cgo CFLAGS: -I${SRCDIR}/../../../../native
// #cgo LDFLAGS: -L${SRCDIR}/../../../../native -lhalodb
//
// #include <stdlib.h>
// #include <libhalodb.h>
import "C"
import (
	"fmt"
	"unsafe"
	"sync"

	"github.com/rinx/vald-meta-halodb/internal/errors"
	"github.com/rinx/vald-meta-halodb/internal/log"
)

type haloDB struct {
	isolate *C.graal_isolate_t
	thread  *C.graal_isolatethread_t
	mu sync.Mutex
}

type HaloDB interface {
	Open(path string) error
	Put(key, value string) error
	Get(key string) (string, error)
	Delete(key string) error
	Size() (int64, error)
	Close() error
}

func New() (HaloDB, error) {
	var isolate *C.graal_isolate_t
	var thread *C.graal_isolatethread_t

	if C.graal_create_isolate(nil, &isolate, &thread) != 0 {
		return nil, fmt.Errorf("failed to initialize")
	}

	return &haloDB{
		isolate: isolate,
		thread:  thread,
	}, nil
}

func (h *haloDB) attachThread() error {
	if C.graal_attach_thread(h.isolate, &h.thread) != 0 {
		return fmt.Errorf("failed to attach thread")
	}

	return nil
}

func (h *haloDB) detachThread() error {
	if C.graal_detach_thread(h.thread) != 0 {
		return errors.New("failed to detach thread")
	}

	return nil
}

func (h *haloDB) Open(path string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	err := h.attachThread()
	if err != nil {
		return err
	}
	defer func() {
		err = h.detachThread()
		if err != nil {
			log.Error("failed to detach")
		}
	}()

	cspath := C.CString(path)
	defer C.free(unsafe.Pointer(cspath))

	if C.halodb_open(h.thread, cspath) != 0 {
		return errors.New("failed to open halodb")
	}

	return nil
}

func (h *haloDB) Put(key, value string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	err := h.attachThread()
	if err != nil {
		return err
	}
	defer func() {
		err = h.detachThread()
		if err != nil {
			log.Error("failed to detach")
		}
	}()

	csKey, csValue := C.CString(key), C.CString(value)
	defer func() {
		C.free(unsafe.Pointer(csKey))
		C.free(unsafe.Pointer(csValue))
	}()

	if C.halodb_put(h.thread, csKey, csValue) != 0 {
		return errors.Errorf("failed to store %s", key)
	}

	return nil
}

func (h *haloDB) Get(key string) (string, error) {
	err := h.attachThread()
	if err != nil {
		return "", err
	}
	defer func() {
		err = h.detachThread()
		if err != nil {
			log.Error("failed to detach")
		}
	}()

	csKey := C.CString(key)
	defer C.free(unsafe.Pointer(csKey))

	res := C.GoString(C.halodb_get(h.thread, csKey))
	if res == "" {
		return "", errors.Errorf("failed to get %s", key)
	}

	return res, nil
}

func (h *haloDB) Delete(key string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	err := h.attachThread()
	if err != nil {
		return err
	}
	defer func() {
		err = h.detachThread()
		if err != nil {
			log.Error("failed to detach")
		}
	}()

	csKey := C.CString(key)
	defer C.free(unsafe.Pointer(csKey))

	if C.halodb_delete(h.thread, csKey) != 0 {
		return errors.Errorf("failed to delete %s", key)
	}

	return nil
}

func (h *haloDB) Size() (int64, error) {
	err := h.attachThread()
	if err != nil {
		return -1, err
	}
	defer func() {
		err = h.detachThread()
		if err != nil {
			log.Error("failed to detach")
		}
	}()

	res := C.halodb_size(h.thread)

	return *(*int64)(unsafe.Pointer(&res)), nil
}

func (h *haloDB) Close() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	err := h.attachThread()
	if err != nil {
		return err
	}
	defer func() {
		err = h.detachThread()
		if err != nil {
			log.Error("failed to detach")
		}
	}()

	if C.halodb_close(h.thread) != 0 {
		return errors.New("failed to close")
	}

	return nil
}
