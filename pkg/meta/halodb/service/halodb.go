package service

// #cgo CFLAGS: -I${SRCDIR}/../../../../native
// #cgo LDFLAGS: -L${SRCDIR}/../../../../native -lhalodb
//
// #include <stdlib.h>
// #include <libhalodb.h>
import "C"
import (
	"fmt"
	"sync"
	"unsafe"

	"github.com/rinx/vald-meta-halodb/internal/errors"
	"github.com/rinx/vald-meta-halodb/internal/log"
)

type haloDB struct {
	isolate *C.graal_isolate_t
	mu      sync.Mutex
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

	param := &C.graal_create_isolate_params_t{
		reserved_address_space_size: 1024 * 1024 * 500,
	}

	if C.graal_create_isolate(param, &isolate, &thread) != 0 {
		return nil, fmt.Errorf("failed to initialize")
	}

	return &haloDB{
		isolate: isolate,
	}, nil
}

func (h *haloDB) attachThread() (*C.graal_isolatethread_t, error) {
	thread := C.graal_get_current_thread(h.isolate)
	if thread != nil {
		return thread, nil
	}

	var newThread *C.graal_isolatethread_t
	if C.graal_attach_thread(h.isolate, &newThread) != 0 {
		return nil, fmt.Errorf("failed to attach thread")
	}

	return newThread, nil
}

func (h *haloDB) detachThread(thread *C.graal_isolatethread_t) error {
	if C.graal_detach_thread(thread) != 0 {
		return errors.New("failed to detach thread")
	}

	return nil
}

func (h *haloDB) pauseCompaction(thread *C.graal_isolatethread_t) error {
	if C.halodb_pause_compaction(thread) != 0 {
		return errors.New("failed to pause compaction")
	}

	return nil
}

func (h *haloDB) resumeCompaction(thread *C.graal_isolatethread_t) error {
	if C.halodb_resume_compaction(thread) != 0 {
		return errors.New("failed to resume compaction")
	}

	return nil
}

func (h *haloDB) Open(path string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	thread, err := h.attachThread()
	if err != nil {
		return err
	}
	defer func() {
		err = h.detachThread(thread)
		if err != nil {
			log.Error(err)
		}
	}()

	csPath := C.CString(path)
	defer C.free(unsafe.Pointer(csPath))

	if C.halodb_open(thread, csPath) != 0 {
		return errors.New("failed to open halodb")
	}

	return nil
}

func (h *haloDB) Put(key, value string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	thread, err := h.attachThread()
	if err != nil {
		return err
	}
	defer func() {
		err = h.detachThread(thread)
		if err != nil {
			log.Error(err)
		}
	}()

	csKey, csValue := C.CString(key), C.CString(value)
	defer func() {
		C.free(unsafe.Pointer(csKey))
		C.free(unsafe.Pointer(csValue))
	}()

	if C.halodb_put(thread, csKey, csValue) != 0 {
		return errors.Errorf("failed to store %s", key)
	}

	return nil
}

func (h *haloDB) Get(key string) (string, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	thread, err := h.attachThread()
	if err != nil {
		return "", err
	}
	defer func() {
		err = h.detachThread(thread)
		if err != nil {
			log.Error(err)
		}
	}()

	csKey := C.CString(key)
	defer C.free(unsafe.Pointer(csKey))

	res := C.GoString(C.halodb_get(thread, csKey))
	if res == "" {
		return "", errors.Errorf("failed to get %s", key)
	}

	return res, nil
}

func (h *haloDB) Delete(key string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	thread, err := h.attachThread()
	if err != nil {
		return err
	}
	defer func() {
		err = h.detachThread(thread)
		if err != nil {
			log.Error(err)
		}
	}()

	csKey := C.CString(key)
	defer C.free(unsafe.Pointer(csKey))

	if C.halodb_delete(thread, csKey) != 0 {
		return errors.Errorf("failed to delete %s", key)
	}

	return nil
}

func (h *haloDB) Size() (int64, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	thread, err := h.attachThread()
	if err != nil {
		return -1, err
	}
	defer func() {
		err = h.detachThread(thread)
		if err != nil {
			log.Error(err)
		}
	}()

	res := C.halodb_size(thread)

	return *(*int64)(unsafe.Pointer(&res)), nil
}

func (h *haloDB) Close() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	thread, err := h.attachThread()
	if err != nil {
		return err
	}

	if C.halodb_close(thread) != 0 {
		return errors.New("failed to close")
	}

	if C.graal_detach_all_threads_and_tear_down_isolate(thread) != 0 {
		return errors.New("failed to detach all threads and teardown isolate")
	}

	return nil
}
